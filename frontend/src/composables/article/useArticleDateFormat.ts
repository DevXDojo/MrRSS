import { useI18n } from 'vue-i18n';
import { useSettings } from '@/composables/core/useSettings';
import { formatDate, formatExactDateTime } from '@/utils/date';

export function useArticleDateFormat() {
  const { locale, t } = useI18n();
  const { settings } = useSettings();

  return {
    formatArticleDate: (date: string): string => formatDate(date, locale.value, t, settings.value),
    formatArticleDateTime: (date: string): string =>
      formatExactDateTime(date, locale.value, settings.value),
  };
}
