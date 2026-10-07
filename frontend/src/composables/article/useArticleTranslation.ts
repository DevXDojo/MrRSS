import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';
import {
  requestTranslation,
  notifyTranslationError,
  resetTranslationCooldown,
} from '@/utils/translationRequest';

interface TranslationSettings {
  enabled: boolean;
  targetLang: string;
  translationOnlyMode: boolean;
  triggerMode: 'auto' | 'manual';
}

export function useArticleTranslation() {
  const { t } = useI18n();
  const translationSettings = ref<TranslationSettings>({
    enabled: false,
    targetLang: 'en',
    translationOnlyMode: false,
    triggerMode: 'auto',
  });
  const translatingArticles: Ref<Set<number>> = ref(new Set());
  let observer: IntersectionObserver | null = null;
  let observerRoot: HTMLElement | null = null;
  let currentArticles = new Map<number, Article>();
  const observed = new Set<Element>();
  const visible = new Set<number>();
  const attempts = new Map<string, { retryAt: number; failures: number; title?: string }>();
  let retryTimer: ReturnType<typeof setTimeout> | null = null;
  let observeTimer: ReturnType<typeof setTimeout> | null = null;
  let disposed = false;
  let settingsRevision = 0;
  const keyFor = (article: Article) =>
    JSON.stringify([article.id, article.title, translationSettings.value.targetLang]);

  function translateVisible(): void {
    if (
      disposed ||
      !translationSettings.value.enabled ||
      translationSettings.value.triggerMode === 'manual'
    )
      return;
    for (const id of visible) {
      const article = currentArticles.get(id);
      if (article && (!article.translated_title || article.translated_title === article.title))
        void translateArticle(article, true);
    }
  }

  function scheduleRetry(): void {
    if (retryTimer) clearTimeout(retryTimer);
    retryTimer = null;
    if (
      disposed ||
      !translationSettings.value.enabled ||
      translationSettings.value.triggerMode === 'manual'
    )
      return;
    const deadlines = [...visible].flatMap((id) => {
      const article = currentArticles.get(id);
      const attempt = article && attempts.get(keyFor(article));
      return attempt && attempt.title === undefined && !translatingArticles.value.has(id)
        ? [attempt.retryAt]
        : [];
    });
    if (!deadlines.length) return;
    retryTimer = setTimeout(
      () => {
        retryTimer = null;
        translateVisible();
      },
      Math.max(1, Math.min(...deadlines) - Date.now())
    );
  }

  // Load translation settings
  async function loadTranslationSettings(): Promise<void> {
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      translationSettings.value = {
        enabled: data.translation_enabled === 'true',
        targetLang: data.target_language || 'en',
        translationOnlyMode: data.translation_only_mode === 'true',
        triggerMode: data.translation_trigger_mode === 'manual' ? 'manual' : 'auto',
      };
    } catch (e) {
      console.error('Error loading translation settings:', e);
    }
  }

  // Setup intersection observer for auto-translation
  function setupIntersectionObserver(listRef: HTMLElement | null, articles: Article[]): void {
    currentArticles = new Map(articles.map((article) => [article.id, article]));
    if (observer && observerRoot !== listRef) {
      observer.disconnect();
      observer = null;
      observed.clear();
      visible.clear();
    }
    observerRoot = listRef;
    if (!observer)
      observer = new IntersectionObserver(
        (entries) => {
          // Check if translation is still enabled before processing
          if (
            !translationSettings.value.enabled ||
            translationSettings.value.triggerMode === 'manual'
          ) {
            return;
          }

          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              const articleId = parseInt((entry.target as HTMLElement).dataset.articleId || '0');
              visible.add(articleId);
              const article = currentArticles.get(articleId);

              // Check if translation is needed:
              // - No translation exists, OR
              // - Translation equals original title (indicates failed/skipped translation)
              const needsTranslation =
                article &&
                (!article.translated_title || article.translated_title === article.title);

              // Only translate if article exists, needs translation, and is not already being translated
              if (needsTranslation && !translatingArticles.value.has(articleId)) {
                void translateArticle(article, true);
              }
            } else {
              visible.delete(parseInt((entry.target as HTMLElement).dataset.articleId || '0'));
            }
          });
        },
        {
          root: listRef,
          rootMargin: '100px',
          threshold: 0.1,
        }
      );

    // Automatically observe all current article elements
    if (listRef && translationSettings.value.enabled) {
      // Use setTimeout to ensure DOM is updated
      if (observeTimer) clearTimeout(observeTimer);
      observeTimer = setTimeout(() => {
        observeTimer = null;
        const cards = listRef.querySelectorAll('[data-article-id]');
        cards.forEach((card) => observeArticle(card));
        translateVisible();
      }, 0);
    }
  }

  // Translate an article
  async function translateArticle(article: Article, automatic = false): Promise<void> {
    // Don't translate if translation is disabled
    if (!translationSettings.value.enabled) return;
    if (translatingArticles.value.has(article.id)) return;
    if (disposed) return;
    const key = keyFor(article);
    const revision = settingsRevision;
    const isCurrent = () =>
      !disposed &&
      translationSettings.value.enabled &&
      revision === settingsRevision &&
      key === keyFor(article);
    const previous = attempts.get(key);
    if (automatic && previous) {
      if (previous.title !== undefined) {
        article.translated_title = previous.title;
        return;
      }
      if (Date.now() < previous.retryAt) return;
    }

    translatingArticles.value.add(article.id);

    try {
      const requestBody = {
        article_id: article.id,
        title: article.title,
        target_language: translationSettings.value.targetLang,
      };

      const res = await requestTranslation('/api/articles/translate', requestBody, isCurrent);
      if (!isCurrent()) return;

      if (res.ok) {
        const data = await res.json();
        if (!isCurrent()) return;

        // Update the article in the store
        // Backend returns translated_title even when skipped (returns original title)
        article.translated_title = data.translated_title;
        const current = currentArticles.get(article.id);
        if (current && keyFor(current) === key) current.translated_title = data.translated_title;
        attempts.set(key, {
          retryAt: Infinity,
          failures: 0,
          title: data.translated_title || article.title,
        });

        // Show notification if AI limit was reached
        if (data.limit_reached) {
          window.showToast(t('article.translation.aiLimitReached'), 'warning');
        }
      } else {
        const failures = (previous?.failures || 0) + 1;
        const seconds = Number(res.headers.get('Retry-After'));
        attempts.set(key, {
          failures,
          retryAt:
            Date.now() +
            Math.max(
              seconds > 0 ? seconds * 1000 : 0,
              Math.min(300000, 30000 * 2 ** Math.min(failures - 1, 4))
            ),
        });
        notifyTranslationError(
          t(
            res.status === 429
              ? 'common.errors.translationRateLimited'
              : 'common.errors.translatingTitle'
          )
        );
      }
    } catch {
      if (isCurrent()) {
        const failures = (previous?.failures || 0) + 1;
        attempts.set(key, {
          failures,
          retryAt: Date.now() + Math.min(300000, 30000 * 2 ** Math.min(failures - 1, 4)),
        });
        notifyTranslationError(t('common.errors.translating'));
      }
    } finally {
      translatingArticles.value.delete(article.id);
      while (attempts.size > 512) attempts.delete(attempts.keys().next().value!);
      scheduleRetry();
    }
  }

  // Observe an article element
  function observeArticle(el: Element | null): void {
    if (el && observer && translationSettings.value.enabled && !observed.has(el)) {
      observed.add(el);
      observer.observe(el);
    }
  }

  // Rows are unmounted once they leave the rendered window; releasing them here
  // keeps the observer from retaining detached elements.
  function unobserveArticle(el: Element | null): void {
    if (el && observer) {
      observer.unobserve(el);
      observed.delete(el);
      visible.delete(parseInt((el as HTMLElement).dataset.articleId || '0'));
    }
  }

  // Update translation settings from event
  function handleTranslationSettingsChange(
    enabled: boolean,
    targetLang: string,
    triggerMode?: string
  ): void {
    settingsRevision++;
    resetTranslationCooldown();
    attempts.clear();
    translationSettings.value = {
      enabled,
      targetLang,
      translationOnlyMode: translationSettings.value.translationOnlyMode,
      triggerMode: triggerMode === 'manual' ? 'manual' : 'auto',
    };

    // Disconnect observer if translation is disabled
    if (!enabled && observer) {
      observer.disconnect();
      observer = null;
      observed.clear();
      visible.clear();
    }
    // Re-observe if translation is enabled
    else if (enabled && observer) {
      if (observeTimer) clearTimeout(observeTimer);
      observeTimer = setTimeout(() => {
        observeTimer = null;
        const cards = observerRoot?.querySelectorAll('[data-article-id]') || [];
        cards.forEach((card) => observer?.observe(card));
        translateVisible();
      }, 100);
    }
    scheduleRetry();
  }

  // Cleanup
  function cleanup(): void {
    disposed = true;
    if (retryTimer) clearTimeout(retryTimer);
    if (observeTimer) clearTimeout(observeTimer);
    observed.clear();
    visible.clear();
    if (observer) {
      observer.disconnect();
      observer = null;
    }
  }

  return {
    translationSettings,
    translatingArticles,
    loadTranslationSettings,
    setupIntersectionObserver,
    translateArticle,
    observeArticle,
    unobserveArticle,
    handleTranslationSettingsChange,
    cleanup,
  };
}
