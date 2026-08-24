import { ref } from 'vue';
import type { Condition } from './useRuleOptions';
import { isDateField, isMultiSelectField, isBooleanField } from './useRuleOptions';

export function useRuleConditions() {
  function addCondition(conditions: Condition[]): void {
    conditions.push({
      id: Date.now(),
      logic: conditions.length > 0 ? 'and' : null,
      negate: false,
      field: 'article_title',
      operator: 'contains',
      value: '',
      values: [],
    });
  }

  function removeCondition(conditions: Condition[], index: number): void {
    conditions.splice(index, 1);
    if (conditions.length > 0 && index === 0) {
      conditions[0].logic = null;
    }
  }

  function onFieldChange(condition: Condition): void {
    if (isDateField(condition.field)) {
      condition.operator = null;
      condition.value = '';
      condition.values = [];
    } else if (isMultiSelectField(condition.field)) {
      condition.operator = 'contains';
      condition.value = '';
      condition.values = [];
    } else if (isBooleanField(condition.field)) {
      condition.operator = null;
      condition.value = 'true';
      condition.values = [];
    } else {
      condition.operator = 'contains';
      condition.value = '';
      condition.values = [];
    }
  }

  function toggleNegate(condition: Condition): void {
    condition.negate = !condition.negate;
  }

  return {
    addCondition,
    removeCondition,
    onFieldChange,
    toggleNegate,
  };
}
