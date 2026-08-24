import { computed, type ComputedRef } from 'vue';
import { useAppStore } from '@/stores/app';

export interface Condition {
  id: number;
  logic?: 'and' | 'or' | null;
  negate: boolean;
  field: string;
  operator?: string | null;
  value: string;
  values: string[];
}

interface FieldOption {
  value: string;
  labelKey: string;
  multiSelect: boolean;
  booleanField?: boolean;
}

interface ActionOption {
  value: string;
  labelKey: string;
}

export function useRuleOptions() {
  const store = useAppStore();

  // Field options for conditions
  const fieldOptions: FieldOption[] = [
    { value: 'feed_name', labelKey: 'modal.feed.feedName', multiSelect: true },
    { value: 'feed_category', labelKey: 'modal.feed.feedCategory', multiSelect: true },
    { value: 'feed_tags', labelKey: 'modal.feed.feedTags', multiSelect: true },
    { value: 'article_title', labelKey: 'article.parts.articleTitle', multiSelect: false },
    { value: 'article_content', labelKey: 'modal.filter.articleContent', multiSelect: false },
    { value: 'author', labelKey: 'modal.filter.author', multiSelect: false },
    { value: 'url', labelKey: 'modal.filter.url', multiSelect: false },
    { value: 'feed_type', labelKey: 'modal.filter.feedType', multiSelect: true },
    {
      value: 'is_image_mode_feed',
      labelKey: 'modal.filter.isImageModeFeed',
      multiSelect: false,
      booleanField: true,
    },
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
      value: 'is_hidden',
      labelKey: 'modal.filter.hiddenStatus',
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

  // Operator options for article title
  const textOperatorOptions: Array<{ value: string; labelKey: string }> = [
    { value: 'contains', labelKey: 'modal.filter.contains' },
    { value: 'exact', labelKey: 'modal.filter.exactMatch' },
    { value: 'regex', labelKey: 'modal.filter.regex' },
  ];

  // Boolean value options
  const booleanOptions: Array<{ value: string; labelKey: string }> = [
    { value: 'true', labelKey: 'common.action.yes' },
    { value: 'false', labelKey: 'common.action.no' },
  ];

  // Action options
  const actionOptions: ActionOption[] = [
    { value: 'favorite', labelKey: 'setting.rule.actionFavorite' },
    { value: 'unfavorite', labelKey: 'setting.rule.actionUnfavorite' },
    { value: 'hide', labelKey: 'setting.rule.actionHide' },
    { value: 'unhide', labelKey: 'setting.rule.actionUnhide' },
    { value: 'mark_read', labelKey: 'setting.rule.actionMarkRead' },
    { value: 'mark_unread', labelKey: 'setting.rule.actionMarkUnread' },
    { value: 'read_later', labelKey: 'setting.rule.actionReadLater' },
    { value: 'remove_read_later', labelKey: 'setting.rule.actionRemoveReadLater' },
  ];

  // Feed names for multi-select
  const feedNames: ComputedRef<string[]> = computed(() => {
    return store.feeds.map((f) => f.title);
  });

  // Feed categories for multi-select
  const feedCategories: ComputedRef<string[]> = computed(() => {
    const categories = new Set<string>();
    store.feeds.forEach((f) => {
      if (f.category) {
        categories.add(f.category);
      }
    });
    return Array.from(categories);
  });

  // Feed types for multi-select (as type codes, not translated text)
  // Type codes: "regular", "freshrss", "rsshub", "script", "xpath", "email"
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

  // Feed tags for multi-select
  const feedTags: ComputedRef<string[]> = computed(() => {
    const tagSet = new Set<string>();
    store.feeds.forEach((f) => {
      f.tags?.forEach((t) => tagSet.add(t.name));
    });
    return Array.from(tagSet);
  });

  return {
    fieldOptions,
    textOperatorOptions,
    booleanOptions,
    actionOptions,
    feedNames,
    feedCategories,
    feedTypes,
    feedTags,
  };
}

// Helper functions for field types
export function isDateField(field: string): boolean {
  return field === 'published_after' || field === 'published_before';
}

export function isMultiSelectField(field: string): boolean {
  return (
    field === 'feed_name' ||
    field === 'feed_category' ||
    field === 'feed_type' ||
    field === 'feed_tags'
  );
}

export function isBooleanField(field: string): boolean {
  return (
    field === 'is_read' ||
    field === 'is_favorite' ||
    field === 'is_hidden' ||
    field === 'is_read_later' ||
    field === 'is_image_mode_feed' ||
    field === 'has_summary' ||
    field === 'has_translation' ||
    field === 'has_image' ||
    field === 'has_audio' ||
    field === 'has_video'
  );
}

export function needsOperator(field: string): boolean {
  return (
    field === 'article_title' ||
    field === 'article_content' ||
    field === 'author' ||
    field === 'url'
  );
}

export function isNumberField(field: string): boolean {
  return (
    field === 'published_after_hours' ||
    field === 'published_after_days' ||
    field === 'feed_articles_per_month'
  );
}
