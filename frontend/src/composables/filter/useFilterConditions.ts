/**
 * Composable for managing filter conditions
 */
import { ref, type Ref } from 'vue';
import type { FilterCondition } from '@/types/filter';
import { useFilterFields } from './useFilterFields';

export function useFilterConditions(initialConditions: FilterCondition[] = []) {
  const { isMultiSelectField } = useFilterFields();

  const conditions: Ref<FilterCondition[]> = ref([]);

  /**
   * Initialize conditions from props
   */
  function initializeConditions(filters: FilterCondition[]): void {
    if (filters && filters.length > 0) {
      conditions.value = JSON.parse(JSON.stringify(filters));
    }
  }

  /**
   * Add a new condition
   */
  function addCondition(): void {
    conditions.value.push({
      id: Date.now(),
      logic: conditions.value.length > 0 ? 'and' : null,
      negate: false,
      field: 'article_title',
      operator: 'contains',
      value: '',
      values: [],
    });
  }

  /**
   * Remove a condition by index
   */
  function removeCondition(index: number): void {
    conditions.value.splice(index, 1);
    // Reset first condition's logic to null
    if (conditions.value.length > 0 && index === 0) {
      conditions.value[0].logic = null;
    }
  }

  /**
   * Toggle negate flag on a condition
   */
  function toggleNegate(index: number): void {
    conditions.value[index].negate = !conditions.value[index].negate;
  }

  /**
   * Clear all conditions
   */
  function clearConditions(): void {
    conditions.value = [];
  }

  /**
   * Validate and get valid conditions
   */
  function getValidConditions(): FilterCondition[] {
    return conditions.value.filter((c) => {
      if (isMultiSelectField(c.field)) {
        return c.values && c.values.length > 0;
      }
      return c.value;
    });
  }

  // Initialize if provided
  if (initialConditions.length > 0) {
    initializeConditions(initialConditions);
  }

  return {
    conditions,
    initializeConditions,
    addCondition,
    removeCondition,
    toggleNegate,
    clearConditions,
    getValidConditions,
  };
}
