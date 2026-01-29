/**
 * Types for article filtering
 */

export interface FilterCondition {
  id: number;
  logic?: 'and' | 'or' | null;
  negate: boolean;
  field: string;
  operator?: string | null;
  value: string;
  values: string[];
}

export interface FieldOption {
  value: string;
  labelKey: string;
  multiSelect: boolean;
  booleanField?: boolean;
}

export interface OperatorOption {
  value: string;
  labelKey: string;
}

export interface LogicOption {
  value: 'and' | 'or';
  labelKey: string;
}

/**
 * Saved filter - a user-saved collection of filter conditions
 */
export interface SavedFilter {
  id: number;
  name: string;
  conditions: string; // JSON string of FilterCondition[]
  position: number;
  created_at: string;
  updated_at: string;
}
