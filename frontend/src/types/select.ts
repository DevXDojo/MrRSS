/**
 * Select component type definitions
 */

export interface SelectOption {
  value: string | number;
  label: string;
  disabled?: boolean;
  icon?: any; // Component
  style?: Record<string, string>; // For font preview and other custom styles
  color?: string; // For TagSelector color display
}

export interface SelectOptionGroup {
  label: string;
  options: SelectOption[];
}

export type SelectOptions = SelectOption[] | SelectOptionGroup[];

/**
 * Type guard to check if an option is a group
 */
export function isOptionGroup(
  option: SelectOption | SelectOptionGroup
): option is SelectOptionGroup {
  return typeof option === 'object' && option !== null && 'label' in option && 'options' in option;
}

/**
 * Flatten grouped options into a single array
 */
export function flattenOptions(options: SelectOptions): SelectOption[] {
  const result: SelectOption[] = [];

  for (const option of options) {
    if (isOptionGroup(option)) {
      result.push(...option.options);
    } else {
      result.push(option);
    }
  }

  return result;
}

/**
 * Find an option by value in grouped or flat options
 */
export function findOption(
  options: SelectOptions,
  value: string | number
): SelectOption | undefined {
  const flat = flattenOptions(options);
  return flat.find((opt) => opt.value === value);
}
