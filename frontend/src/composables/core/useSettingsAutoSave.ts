/**
 * Composable for auto-saving settings with debouncing
 */
import { ref, watch, onMounted, onUnmounted, type Ref, computed, isRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import type { SettingsData } from '@/types/settings';
import { settingsDefaults } from '@/config/defaults';
import { buildAutoSavePayload } from './useSettings.generated';

export function useSettingsAutoSave(settings: Ref<SettingsData> | (() => SettingsData)) {
  const { locale } = useI18n();
  const store = useAppStore();

  let saveTimeout: ReturnType<typeof setTimeout> | null = null;
  let isInitialLoad = true;

  // Convert to ref if it's a getter function
  const settingsRef = isRef(settings) ? settings : computed(settings);

  // Track previous translation settings
  const prevTranslationSettings: Ref<{
    enabled: boolean;
    targetLang: string;
    provider: string;
  }> = ref({
    enabled: settingsDefaults.translation_enabled,
    targetLang: settingsDefaults.target_language,
    provider: settingsDefaults.translation_provider,
  });

  // Track previous article display settings to prevent unnecessary refreshes
  const prevArticleDisplaySettings: Ref<{
    showHiddenArticles: boolean;
  }> = ref({
    showHiddenArticles: settingsDefaults.show_hidden_articles,
  });

  // Track previous layout mode setting to prevent unnecessary width resets
  const prevLayoutMode: Ref<string> = ref(settingsDefaults.layout_mode);

  // Track previous summary settings
  const prevSummarySettings: Ref<{
    enabled: boolean;
    provider: string;
    triggerMode: string;
    length: string;
  }> = ref({
    enabled: settingsDefaults.summary_enabled,
    provider: settingsDefaults.summary_provider,
    triggerMode: settingsDefaults.summary_trigger_mode,
    length: settingsDefaults.summary_length,
  });

  /**
   * Initialize translation tracking
   */
  onMounted(() => {
    setTimeout(() => {
      prevTranslationSettings.value = {
        enabled: settingsRef.value.translation_enabled,
        targetLang: settingsRef.value.target_language,
        provider: settingsRef.value.translation_provider,
      };
      prevArticleDisplaySettings.value = {
        showHiddenArticles: settingsRef.value.show_hidden_articles,
      };
      prevLayoutMode.value = settingsRef.value.layout_mode;
      prevSummarySettings.value = {
        enabled: settingsRef.value.summary_enabled,
        provider: settingsRef.value.summary_provider,
        triggerMode: settingsRef.value.summary_trigger_mode,
        length: settingsRef.value.summary_length,
      };
      isInitialLoad = false;
    }, 100);
  });

  /**
   * Save settings to backend
   */
  async function autoSave() {
    try {
      // Skip translation clearing on initial load
      if (isInitialLoad) {
        return;
      }

      // Check if translation settings changed
      const translationChanged =
        prevTranslationSettings.value.enabled !== settingsRef.value.translation_enabled ||
        prevTranslationSettings.value.provider !== settingsRef.value.translation_provider ||
        (settingsRef.value.translation_enabled &&
          prevTranslationSettings.value.targetLang !== settingsRef.value.target_language);

      // Always apply basic settings immediately (theme, language, etc.)
      // even if validation fails - these don't require API keys
      locale.value = settingsRef.value.language;
      store.setTheme(settingsRef.value.theme as 'light' | 'dark' | 'auto');
      store.startAutoRefresh(settingsRef.value.update_interval);

      // Notify components about default view mode change
      window.dispatchEvent(
        new CustomEvent('default-view-mode-changed', {
          detail: {
            mode: settingsRef.value.default_view_mode,
          },
        })
      );

      // Note: Validation is used for UI feedback only (showing red borders on invalid fields).
      // We do NOT block saving settings to the backend based on validation.
      // This allows users to save their preferences immediately, even if some fields are invalid.
      // The backend saves all provided values to the database without validation.
      // Features that require valid settings (e.g., translation with API keys) will check
      // for valid values at runtime and fail gracefully if settings are incomplete/invalid.

      // Save to backend using generated payload (alphabetically sorted)
      await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(buildAutoSavePayload(settingsRef)),
      });

      // Clear and re-translate if translation settings changed
      if (translationChanged) {
        await fetch('/api/articles/clear-translations', { method: 'POST' });
        // Update tracking
        prevTranslationSettings.value = {
          enabled: settingsRef.value.translation_enabled,
          targetLang: settingsRef.value.target_language,
          provider: settingsRef.value.translation_provider,
        };
        // Notify ArticleList about translation settings change
        window.dispatchEvent(
          new CustomEvent('translation-settings-changed', {
            detail: {
              enabled: settingsRef.value.translation_enabled,
              targetLang: settingsRef.value.target_language,
            },
          })
        );
        // Refresh articles to show without translations, then re-translate if enabled
        store.fetchArticles();
      }

      // Refresh articles if show_hidden_articles changed
      if (
        settingsRef.value.show_hidden_articles !==
        prevArticleDisplaySettings.value.showHiddenArticles
      ) {
        store.fetchArticles();
        // Update tracking
        prevArticleDisplaySettings.value.showHiddenArticles =
          settingsRef.value.show_hidden_articles;
      }

      // Notify about show_article_preview_images change
      window.dispatchEvent(
        new CustomEvent('show-preview-images-changed', {
          detail: {
            value: settingsRef.value.show_article_preview_images,
          },
        })
      );

      // Notify about image_gallery_enabled change
      window.dispatchEvent(
        new CustomEvent('image-gallery-setting-changed', {
          detail: {
            enabled: settingsRef.value.image_gallery_enabled,
          },
        })
      );

      // Notify about auto_show_all_content change
      window.dispatchEvent(
        new CustomEvent('auto-show-all-content-changed', {
          detail: {
            value: settingsRef.value.auto_show_all_content,
          },
        })
      );

      // Notify about layout_mode change only if it actually changed
      if (settingsRef.value.layout_mode !== prevLayoutMode.value) {
        window.dispatchEvent(
          new CustomEvent('layout-mode-changed', {
            detail: {
              mode: settingsRef.value.layout_mode,
            },
          })
        );
        // Update tracking
        prevLayoutMode.value = settingsRef.value.layout_mode;
      }

      // Check if summary settings changed
      const summaryChanged =
        prevSummarySettings.value.enabled !== settingsRef.value.summary_enabled ||
        prevSummarySettings.value.provider !== settingsRef.value.summary_provider ||
        prevSummarySettings.value.triggerMode !== settingsRef.value.summary_trigger_mode ||
        prevSummarySettings.value.length !== settingsRef.value.summary_length;

      if (summaryChanged) {
        window.dispatchEvent(
          new CustomEvent('summary-settings-changed', {
            detail: {
              enabled: settingsRef.value.summary_enabled,
              provider: settingsRef.value.summary_provider,
              triggerMode: settingsRef.value.summary_trigger_mode,
              length: settingsRef.value.summary_length,
            },
          })
        );
        // Update tracking
        prevSummarySettings.value = {
          enabled: settingsRef.value.summary_enabled,
          provider: settingsRef.value.summary_provider,
          triggerMode: settingsRef.value.summary_trigger_mode,
          length: settingsRef.value.summary_length,
        };
      }

      // Notify all components that settings have been updated
      // This ensures components like ArticleList can re-fetch settings
      window.dispatchEvent(new CustomEvent('settings-updated'));
    } catch (e) {
      console.error('Error auto-saving settings:', e);
    }
  }

  /**
   * Debounced auto-save function
   */
  function debouncedAutoSave() {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
    }
    saveTimeout = setTimeout(autoSave, 500); // Wait 500ms after last change
  }

  // Watch the entire settings object for changes
  watch(() => settingsRef.value, debouncedAutoSave, { deep: true });

  // Clean up timeout on unmount to prevent memory leaks
  onUnmounted(() => {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
      saveTimeout = null;
    }
  });

  return {
    autoSave,
    debouncedAutoSave,
  };
}
