// Type definitions for models

export interface Tag {
  id: number;
  name: string;
  color: string; // Hex color code
  position?: number;
}

export interface Article {
  id: number;
  feed_id: number;
  feed_title?: string;
  feed_name?: string; // Alias for feed_title (used in filters/rules)
  title: string;
  original_title?: string;
  translated_title?: string;
  url: string;
  image_url?: string; // Article thumbnail image
  audio_url?: string; // Podcast audio file URL
  video_url?: string; // YouTube video embed URL
  published_at: string;
  is_read: boolean;
  is_favorite: boolean;
  is_hidden: boolean;
  is_read_later: boolean;
  author?: string; // Article author
  summary?: string; // Cached AI-generated summary
  freshrss_item_id?: string; // FreshRSS/Google Reader item ID
}

export interface Feed {
  id: number;
  url: string;
  title: string;
  category: string;
  last_fetched_at: string;
  position?: number; // Position within category for custom ordering
  is_discovered?: boolean;
  website_url?: string;
  image_url?: string;
  last_error?: string;
  script_path?: string;
  hide_from_timeline?: boolean;
  proxy_url?: string;
  proxy_enabled?: boolean;
  refresh_interval?: number;
  is_image_mode?: boolean;
  // XPath support
  type?: string;
  xpath_item?: string;
  xpath_item_title?: string;
  xpath_item_content?: string;
  xpath_item_uri?: string;
  xpath_item_author?: string;
  xpath_item_timestamp?: string;
  xpath_item_time_format?: string;
  xpath_item_thumbnail?: string;
  xpath_item_categories?: string;
  xpath_item_uid?: string;
  article_view_mode?: string; // Article view mode override ('global', 'webpage', 'rendered', 'external')
  auto_expand_content?: string; // Auto expand content mode ('global', 'enabled', 'disabled')
  // Email/Newsletter support
  email_address?: string;
  email_imap_server?: string;
  email_imap_port?: number;
  email_username?: string;
  email_password?: string;
  email_folder?: string;
  // FreshRSS integration
  is_freshrss_source?: boolean; // Whether this feed is from FreshRSS sync
  freshrss_stream_id?: string; // FreshRSS stream ID (e.g., "feed/http://...")
  // Statistics
  latest_article_time?: string; // Latest article publish time
  articles_per_month?: number; // Average articles per month (calculated from last 90 days)
  last_update_status?: string; // Last update status ("success" or "failed")
  // Tags (populated by API handlers)
  tags?: Tag[]; // Tags assigned to this feed
}

export interface UnreadCounts {
  total: number;
  feedCounts: Record<number, number>;
}

export interface RefreshProgress {
  isRunning: boolean;
  errors?: Record<number, string>; // Map of feed ID to error message
  pool_task_count?: number; // Tasks currently in pool
  article_click_count?: number; // Article click triggered tasks
  queue_task_count?: number; // Tasks in queue
  pool_tasks?: PoolTaskInfo[]; // Detailed pool task information
  queue_tasks?: QueueTaskInfo[]; // Detailed queue task information (max 3)
}

export interface PoolTaskInfo {
  feed_id: number;
  feed_title: string;
  reason: number; // TaskReason enum value
  created_at: string;
}

export interface QueueTaskInfo {
  feed_id: number;
  feed_title: string;
  position: number;
}

export interface UpdateInfo {
  has_update: boolean;
  latest_version: string;
  current_version: string;
  download_url: string;
  release_notes: string;
  is_portable: boolean;
}

export interface Settings {
  update_interval: string;
  auto_cleanup_enabled: string;
  max_cache_size_mb: string;
  max_article_age_days: string;
  translation_enabled: string;
  target_language: string;
  translation_provider: string;
  deepl_api_key: string;
  language: string;
  theme: string;
  default_view_mode: string;
  show_hidden_articles: string;
  startup_on_boot: string;
}

export interface DiscoveredFeed {
  url: string;
  title: string;
  description?: string;
  articles?: Article[];
}

export interface Rule {
  id: number;
  name: string;
  enabled: boolean;
  condition: RuleCondition;
  actions: RuleAction[];
}

export interface RuleCondition {
  type: 'always' | 'filter';
  filter?: FilterCondition[];
}

export interface FilterCondition {
  field:
    | 'feed_name'
    | 'feed_category'
    | 'feed_tags'
    | 'article_title'
    | 'is_read'
    | 'is_favorite'
    | 'is_hidden'
    | 'is_read_later';
  operator: 'contains' | 'equals' | 'not_equals';
  value: string;
  logic?: 'and' | 'or' | 'not';
}

export type RuleAction =
  | { type: 'favorite' }
  | { type: 'unfavorite' }
  | { type: 'hide' }
  | { type: 'unhide' }
  | { type: 'mark_read' }
  | { type: 'mark_unread' }
  | { type: 'read_later' }
  | { type: 'remove_read_later' };

export interface KeyboardShortcut {
  action: string;
  key: string;
  defaultKey: string;
}
