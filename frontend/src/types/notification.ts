export type NotificationProvider = 'telegram' | 'discord' | 'feishu';
export interface NotificationChannel {
  id: string;
  name: string;
  provider: NotificationProvider;
  enabled: boolean;
  token: string;
  chat_id: string;
  thread_id: number;
  webhook: string;
  secret: string;
}
export interface NotificationCondition {
  field: 'title' | 'summary' | 'feed' | 'category' | 'author' | 'url' | 'favorite' | 'read_later';
  operator: 'contains' | 'not_contains' | 'equals' | 'not_equals';
  value: string;
}
export interface NotificationRule {
  id: string;
  name: string;
  enabled: boolean;
  mode: 'alert' | 'digest';
  channel_ids: string[];
  match: 'all' | 'any';
  conditions: NotificationCondition[];
  interval_minutes: number;
  time: string;
  weekdays: number[];
  max_items: number;
  include_summary: boolean;
  include_link: boolean;
  unread_only: boolean;
}
export interface NotificationConfig {
  revision: number;
  enabled: boolean;
  timezone: string;
  quiet_enabled: boolean;
  quiet_start: string;
  quiet_end: string;
  channels: NotificationChannel[];
  rules: NotificationRule[];
}
export interface NotificationDelivery {
  id: number;
  channel_id: string;
  rule_id: string;
  label: string;
  status: 'sent' | 'pending' | 'failed' | 'cancelled';
  attempts: number;
  error_code: string;
  next_attempt: number;
  created_at: number;
}
export interface NotificationPreview {
  matched: number;
  scanned: number;
  messages: string[];
}

function notificationID(): string {
  // randomUUID requires a secure context; server mode can also run over LAN HTTP.
  if (typeof globalThis.crypto.randomUUID === 'function') return globalThis.crypto.randomUUID();
  return Array.from(globalThis.crypto.getRandomValues(new Uint8Array(16)), (byte) =>
    byte.toString(16).padStart(2, '0')
  ).join('');
}

export function newChannel(provider: NotificationProvider, name: string): NotificationChannel {
  return {
    id: notificationID(),
    name,
    provider,
    enabled: true,
    token: '',
    chat_id: '',
    thread_id: 0,
    webhook: '',
    secret: '',
  };
}
export function newRule(
  mode: 'alert' | 'digest',
  name: string,
  channels: string[]
): NotificationRule {
  return {
    id: notificationID(),
    name,
    enabled: true,
    mode,
    channel_ids: channels,
    match: 'all',
    conditions: [],
    interval_minutes: 5,
    time: '09:00',
    weekdays: [],
    max_items: 5,
    include_summary: true,
    include_link: true,
    unread_only: false,
  };
}
