/**
 * Settings types for SettingsModal and related components
 *
 * Note: SettingsData interface is auto-generated in settings.generated.ts
 * This file contains non-settings-related types and re-exports SettingsData for convenience
 */

// Re-export SettingsData from the generated file for convenience
export type { SettingsData } from './settings.generated';

export interface NetworkInfo {
  speed_level: 'slow' | 'medium' | 'fast';
  bandwidth_mbps: number;
  latency_ms: number;
  max_concurrency: number;
  detection_time: string;
  detection_success: boolean;
  error_message?: string;
}

export interface AITestInfo {
  config_valid: boolean;
  connection_success: boolean;
  model_available: boolean;
  response_time_ms: number;
  test_time: string;
  error_message?: string;
}

export interface UpdateInfo {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  download_url: string;
  asset_name: string;
  is_portable: boolean;
  error?: string;
}

export interface DownloadResponse {
  success: boolean;
  file_path: string;
}

export interface InstallResponse {
  success: boolean;
}

export type TabName =
  | 'general'
  | 'reading'
  | 'feeds'
  | 'content'
  | 'ai'
  | 'rules'
  | 'network'
  | 'plugins'
  | 'shortcuts'
  | 'statistics'
  | 'about';
