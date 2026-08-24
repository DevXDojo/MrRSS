import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';

interface TranslationSettings {
  enabled: boolean;
  targetLang: string;
  translationOnlyMode: boolean;
}

export function useArticleTranslation() {
  const { t } = useI18n();
  const translationSettings = ref<TranslationSettings>({
    enabled: false,
    targetLang: 'en',
    translationOnlyMode: false,
  });
  const translatingArticles: Ref<Set<number>> = ref(new Set());
  let observer: IntersectionObserver | null = null;

  // Load translation settings
  async function loadTranslationSettings(): Promise<void> {
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      translationSettings.value = {
        enabled: data.translation_enabled === 'true',
        targetLang: data.target_language || 'en',
        translationOnlyMode: data.translation_only_mode === 'true',
      };
    } catch (e) {
      console.error('Error loading translation settings:', e);
    }
  }

  // Setup intersection observer for auto-translation
  function setupIntersectionObserver(listRef: HTMLElement | null, articles: Article[]): void {
    if (observer) {
      observer.disconnect();
    }

    observer = new IntersectionObserver(
      (entries) => {
        // Check if translation is still enabled before processing
        if (!translationSettings.value.enabled) {
          return;
        }

        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const articleId = parseInt((entry.target as HTMLElement).dataset.articleId || '0');
            const article = articles.find((a) => a.id === articleId);

            // Check if translation is needed:
            // - No translation exists, OR
            // - Translation equals original title (indicates failed/skipped translation)
            const needsTranslation =
              article && (!article.translated_title || article.translated_title === article.title);

            // Only translate if article exists, needs translation, and is not already being translated
            if (needsTranslation && !translatingArticles.value.has(articleId)) {
              translateArticle(article);
            }
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
      setTimeout(() => {
        const cards = listRef.querySelectorAll('[data-article-id]');
        cards.forEach((card) => observer?.observe(card));
      }, 0);
    }
  }

  // Translate an article
  async function translateArticle(article: Article): Promise<void> {
    // Don't translate if translation is disabled
    if (!translationSettings.value.enabled) return;
    if (translatingArticles.value.has(article.id)) return;

    translatingArticles.value.add(article.id);

    try {
      const requestBody = {
        article_id: article.id,
        title: article.title,
        target_language: translationSettings.value.targetLang,
      };

      const res = await fetch('/api/articles/translate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestBody),
      });

      if (res.ok) {
        const data = await res.json();

        // Update the article in the store
        // Backend returns translated_title even when skipped (returns original title)
        article.translated_title = data.translated_title;

        // Show notification if AI limit was reached
        if (data.limit_reached) {
          window.showToast(t('article.translation.aiLimitReached'), 'warning');
        }
      } else {
        window.showToast(t('common.errors.translatingTitle'), 'error');
      }
    } catch {
      window.showToast(t('common.errors.translating'), 'error');
    } finally {
      translatingArticles.value.delete(article.id);
    }
  }

  // Observe an article element
  function observeArticle(el: Element | null): void {
    if (el && observer && translationSettings.value.enabled) {
      observer.observe(el);
    }
  }

  // Update translation settings from event
  function handleTranslationSettingsChange(enabled: boolean, targetLang: string): void {
    translationSettings.value = {
      enabled,
      targetLang,
      translationOnlyMode: translationSettings.value.translationOnlyMode,
    };

    // Disconnect observer if translation is disabled
    if (!enabled && observer) {
      observer.disconnect();
      observer = null;
    }
    // Re-observe if translation is enabled
    else if (enabled && observer) {
      setTimeout(() => {
        const cards = document.querySelectorAll('[data-article-id]');
        cards.forEach((card) => observer?.observe(card));
      }, 100);
    }
  }

  // Cleanup
  function cleanup(): void {
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
    handleTranslationSettingsChange,
    cleanup,
  };
}
