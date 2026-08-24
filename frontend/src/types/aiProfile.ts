/**
 * AI Profile types for multi-configuration support
 */

export interface AIProfile {
  id: number;
  name: string;
  api_key?: string; // Masked in list view, only shown when editing
  endpoint: string;
  model: string;
  custom_headers: string; // JSON string of key-value pairs
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface AIProfileTestResult {
  profile_id: number;
  profile_name: string;
  config_valid: boolean;
  connection_success: boolean;
  model_available: boolean;
  response_time_ms: number;
  error_message?: string;
}

export interface AIProfileFormData {
  name: string;
  api_key: string;
  endpoint: string;
  model: string;
  custom_headers: string;
  is_default: boolean;
}

// Default values for a new profile
export const defaultAIProfileFormData: AIProfileFormData = {
  name: '',
  api_key: '',
  endpoint: 'https://api.openai.com/v1/chat/completions',
  model: 'gpt-4o-mini',
  custom_headers: '',
  is_default: false,
};
