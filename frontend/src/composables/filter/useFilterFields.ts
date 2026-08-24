/**
 * Composable for filter field utilities
 */
import { computed, type ComputedRef } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import type { FieldOption, OperatorOption, LogicOption, FilterCondition } from '@/types/filter';

export function useFilterFields() {
  const store = useAppStore();
  const { t, locale } = useI18n();

  /**
   * Available field options for filtering
   */
  const fieldOptions: FieldOption[] = [
    { value: 'feed_name', labelKey: 'modal.feed.feedName', multiSelect: true },
    { value: 'feed_category', labelKey: 'modal.feed.feedCategory', multiSelect: true },
    { value: 'feed_tags', labelKey: 'modal.feed.feedTags', multiSelect: true },
    { value: 'article_title', labelKey: 'article.parts.articleTitle', multiSelect: false },
    { value: 'article_content', labelKey: 'modal.filter.articleContent', multiSelect: false },
    { value: 'author', labelKey: 'modal.filter.author', multiSelect: false },
    { value: 'url', labelKey: 'modal.filter.url', multiSelect: false },
    { value: 'feed_type', labelKey: 'modal.filter.feedType', multiSelect: true },
    { value: 'published_after', labelKey: 'modal.filter.publishedAfter', multiSelect: false },
    { value: 'published_before', labelKey: 'modal.filter.publishedBefore', multiSelect: false },
    {
      value: 'published_after_hours',
      labelKey: 'modal.filter.publishedAfterHours',
      multiSelect: false,
      numberField: true,
    },
    {
      value: 'published_after_days',
      labelKey: 'modal.filter.publishedAfterDays',
      multiSelect: false,
      numberField: true,
    },
    {
      value: 'feed_articles_per_month',
      labelKey: 'modal.filter.feedArticlesPerMonth',
      multiSelect: false,
      numberField: true,
    },
    {
      value: 'feed_last_update_status',
      labelKey: 'modal.filter.feedLastUpdateStatus',
      multiSelect: false,
    },
    {
      value: 'is_read',
      labelKey: 'modal.filter.readStatus',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'is_favorite',
      labelKey: 'modal.filter.favoriteStatus',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'is_read_later',
      labelKey: 'modal.filter.readLaterStatus',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'is_hidden',
      labelKey: 'modal.filter.hiddenStatus',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'has_summary',
      labelKey: 'modal.filter.hasSummary',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'has_translation',
      labelKey: 'modal.filter.hasTranslation',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'has_image',
      labelKey: 'modal.filter.hasImage',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'has_audio',
      labelKey: 'modal.filter.hasAudio',
      multiSelect: false,
      booleanField: true,
    },
    {
      value: 'has_video',
      labelKey: 'modal.filter.hasVideo',
      multiSelect: false,
      booleanField: true,
    },
  ];

  /**
   * Operator options for text fields
   */
  const textOperatorOptions: OperatorOption[] = [
    { value: 'contains', labelKey: 'modal.filter.contains' },
    { value: 'exact', labelKey: 'modal.filter.exactMatch' },
  ];

  /**
   * Boolean value options
   */
  const booleanOptions: OperatorOption[] = [
    { value: 'true', labelKey: 'common.action.yes' },
    { value: 'false', labelKey: 'common.action.no' },
  ];

  /**
   * Logic connector options
   */
  const logicOptions: LogicOption[] = [
    { value: 'and', labelKey: 'modal.filter.and' },
    { value: 'or', labelKey: 'modal.filter.or' },
  ];

  /**
   * Get available feed names
   */
  const feedNames: ComputedRef<string[]> = computed(() => {
    return store.feeds.map((f) => f.title);
  });

  /**
   * Get available feed categories
   */
  const feedCategories: ComputedRef<string[]> = computed(() => {
    const categories = new Set<string>();
    store.feeds.forEach((f) => {
      if (f.category) {
        categories.add(f.category);
      }
    });
    return Array.from(categories);
  });

  /**
   * Get available feed types (as type codes, not translated text)
   * Type codes: "regular", "freshrss", "rsshub", "script", "xpath", "email"
   */
  const feedTypes: ComputedRef<string[]> = computed(() => {
    const typeSet = new Set<string>();
    store.feeds.forEach((f) => {
      // Determine feed type based on feed properties
      let typeCode: string;
      if (f.is_freshrss_source) {
        typeCode = 'freshrss';
      } else if (f.url && f.url.startsWith('rsshub://')) {
        typeCode = 'rsshub';
      } else if (f.script_path) {
        typeCode = 'script';
      } else if (f.type === 'email') {
        typeCode = 'email';
      } else if (f.type === 'HTML+XPath' || f.type === 'XML+XPath') {
        typeCode = 'xpath';
      } else {
        // Default: regular RSS/Atom feed
        typeCode = 'regular';
      }
      // Store type code directly, not translated text
      typeSet.add(typeCode);
    });
    return Array.from(typeSet);
  });

  /**
   * Get available feed tags
   */
  const feedTags: ComputedRef<string[]> = computed(() => {
    const tagSet = new Set<string>();
    store.feeds.forEach((f) => {
      f.tags?.forEach((t) => tagSet.add(t.name));
    });
    return Array.from(tagSet);
  });

  /**
   * Check if field is a date field
   */
  function isDateField(field: string): boolean {
    return field === 'published_after' || field === 'published_before';
  }

  /**
   * Check if field supports multiple values
   */
  function isMultiSelectField(field: string): boolean {
    return (
      field === 'feed_name' ||
      field === 'feed_category' ||
      field === 'feed_type' ||
      field === 'feed_tags'
    );
  }

  /**
   * Check if field is a boolean field
   */
  function isBooleanField(field: string): boolean {
    return (
      field === 'is_read' ||
      field === 'is_favorite' ||
      field === 'is_read_later' ||
      field === 'is_hidden' ||
      field === 'has_summary' ||
      field === 'has_translation' ||
      field === 'has_image' ||
      field === 'has_audio' ||
      field === 'has_video'
    );
  }

  /**
   * Check if field needs an operator selector
   */
  function needsOperator(field: string): boolean {
    // article_title, article_content, author, and url need the contains/exact/regex operator
    return (
      field === 'article_title' ||
      field === 'article_content' ||
      field === 'author' ||
      field === 'url'
    );
  }

  /**
   * Handle field change - reset appropriate values
   */
  function onFieldChange(condition: FilterCondition): void {
    if (isDateField(condition.field)) {
      condition.operator = null;
      condition.value = '';
      condition.values = [];
    } else if (isMultiSelectField(condition.field)) {
      condition.operator = 'contains'; // Always contains for multi-select
      condition.value = '';
      condition.values = [];
    } else if (isBooleanField(condition.field)) {
      condition.operator = null;
      condition.value = 'true'; // Default to true (read/favorited)
      condition.values = [];
    } else {
      condition.operator = 'contains';
      condition.value = '';
      condition.values = [];
    }
  }

  /**
   * Get display text for multi-select dropdown
   */
  function getMultiSelectDisplayText(condition: FilterCondition, labelKey: string): string {
    if (!condition.values || condition.values.length === 0) {
      return t(labelKey);
    }

    if (condition.values.length === 1) {
      return condition.values[0];
    }

    // Show first item and total count
    // For Chinese: "xxx等N个" means "xxx and N items in total"
    // For English: "xxx and N more" means N additional items
    const firstItem = condition.values[0];
    const totalCount = condition.values.length;
    const remaining = totalCount - 1;

    // Use different count based on locale
    if (locale.value === 'zh') {
      return `${firstItem} ${t('common.text.andNMore', { count: totalCount })}`;
    }
    return `${firstItem} ${t('common.text.andNMore', { count: remaining })}`;
  }

  return {
    fieldOptions,
    textOperatorOptions,
    booleanOptions,
    logicOptions,
    feedNames,
    feedCategories,
    feedTypes,
    feedTags,
    isDateField,
    isMultiSelectField,
    isBooleanField,
    needsOperator,
    onFieldChange,
    getMultiSelectDisplayText,
  };
}
