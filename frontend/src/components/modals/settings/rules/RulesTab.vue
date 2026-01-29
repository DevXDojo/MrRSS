<script setup lang="ts">
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { ref, onMounted, watch, type Ref } from 'vue';
import { PhLightning, PhPlus } from '@phosphor-icons/vue';
import RuleEditorModal from '../../rules/RuleEditorModal.vue';
import RuleItem from './RuleItem.vue';
import type { Condition } from '@/composables/rules/useRuleOptions';
import type { SettingsData } from '@/types/settings';
import { ButtonControl, SettingGroup, SettingItem } from '@/components/settings';

const store = useAppStore();
const { t } = useI18n();

interface Rule {
  id: number;
  name: string;
  enabled: boolean;
  conditions: Condition[];
  actions: string[];
  position?: number; // Optional for backward compatibility
}

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Rules list
const rules: Ref<Rule[]> = ref([]);

// Drag and drop state
const draggingRuleId: Ref<number | null> = ref(null);
const dragOverRuleId: Ref<number | null> = ref(null);
const dropBeforeTarget: Ref<boolean> = ref(true);

// Modal states
const showRuleEditor = ref(false);
const editingRule: Ref<Rule | null> = ref(null);
const applyingRuleId: Ref<number | null> = ref(null);

// Load rules from settings
onMounted(() => {
  loadRules();
});

function loadRules() {
  if (props.settings.rules) {
    try {
      const parsed =
        typeof props.settings.rules === 'string'
          ? JSON.parse(props.settings.rules)
          : props.settings.rules;

      // Add position field to rules that don't have it (backward compatibility)
      const loadedRules = Array.isArray(parsed) ? parsed : [];
      rules.value = loadedRules.map((rule: Rule, index: number) => ({
        ...rule,
        position: rule.position ?? index,
      }));

      // Sort rules by position
      rules.value.sort((a, b) => (a.position || 0) - (b.position || 0));
    } catch (e) {
      console.error('Error parsing rules:', e);
      rules.value = [];
    }
  }
}

// Watch for settings changes
watch(
  () => props.settings.rules,
  () => {
    loadRules();
  },
  { immediate: true }
);

// Save rules to settings
async function saveRules() {
  try {
    // No transformation needed - feed_type values are already codes
    const updatedSettings = { ...props.settings, rules: JSON.stringify(rules.value) };
    emit('update:settings', updatedSettings);
    await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ rules: updatedSettings.rules }),
    });
  } catch (e) {
    console.error('Error saving rules:', e);
  }
}

// Add new rule
function addRule() {
  editingRule.value = null;
  showRuleEditor.value = true;
}

// Edit existing rule
function editRule(rule: Rule): void {
  editingRule.value = { ...rule };
  showRuleEditor.value = true;
}

// Delete rule
async function deleteRule(ruleId: number): Promise<void> {
  const confirmed = await window.showConfirm({
    title: t('modal.rule.deleteConfirmTitle'),
    message: t('modal.rule.deleteConfirmMessage'),
    confirmText: t('common.delete'),
    cancelText: t('common.cancel'),
    isDanger: true,
  });

  if (!confirmed) return;

  rules.value = rules.value.filter((r) => r.id !== ruleId);
  await saveRules();
  window.showToast(t('modal.rule.deletedSuccess'), 'success');
}

// Toggle rule enabled state
async function toggleRuleEnabled(rule: Rule): Promise<void> {
  rule.enabled = !rule.enabled;
  await saveRules();
}

// Save rule from editor
async function handleSaveRule(rule: Rule): Promise<void> {
  // Check if this is a new rule (editingRule is null or has no id)
  const isNew = !editingRule.value || !editingRule.value.id;

  if (editingRule.value && editingRule.value.id) {
    // Update existing rule
    const index = rules.value.findIndex((r) => r.id === rule.id);
    if (index !== -1) {
      rules.value[index] = rule;
    }
  } else {
    // Add new rule
    rule.id = Date.now();
    rule.enabled = true;
    rule.position = rules.value.length; // Add to the end
    rules.value.push(rule);
  }

  await saveRules();
  showRuleEditor.value = false;
  window.showToast(t('modal.rule.savedSuccess'), 'success');

  // Apply rule to existing articles when adding a new rule
  if (isNew && rule.enabled) {
    await applyRule(rule);
  }
}

// Apply rule now
async function applyRule(rule: Rule): Promise<void> {
  applyingRuleId.value = rule.id;

  try {
    // No transformation needed - feed_type values are already codes
    const res = await fetch('/api/rules/apply', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(rule),
    });

    if (res.ok) {
      const data = await res.json();
      window.showToast(t('modal.rule.ruleAppliedSuccess', { count: data.affected }), 'success');
      store.fetchArticles();
      store.fetchUnreadCounts();
    } else {
      window.showToast(t('common.errors.savingSettings'), 'error');
    }
  } catch (e) {
    console.error('Error applying rule:', e);
    window.showToast(t('common.errors.savingSettings'), 'error');
  } finally {
    applyingRuleId.value = null;
  }
}

// Drag and drop handlers
function onDragStart(ruleId: number, event: DragEvent) {
  draggingRuleId.value = ruleId;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', String(ruleId));

    // Set a custom drag image if possible (better UX)
    const target = event.target as HTMLElement;
    if (target instanceof HTMLElement) {
      try {
        event.dataTransfer.setDragImage(target, 0, 0);
      } catch (e) {
        // Fallback for browsers that don't support custom drag image
        console.debug('Could not set custom drag image:', e);
      }
    }
  }
}

function onDragEnd() {
  // Add a small delay to allow visual feedback to complete
  setTimeout(() => {
    draggingRuleId.value = null;
    dragOverRuleId.value = null;
    dropBeforeTarget.value = true;
  }, 50);
}

function onDragOver(targetRuleId: number, event: DragEvent) {
  event.preventDefault();
  if (!draggingRuleId.value || draggingRuleId.value === targetRuleId) {
    return;
  }

  dragOverRuleId.value = targetRuleId;

  // Calculate drop position based on mouse Y position
  if (event.target instanceof HTMLElement) {
    const targetElement = event.target.closest('.rule-item');
    if (targetElement) {
      const rect = targetElement.getBoundingClientRect();
      const relativeY = event.clientY - rect.top;
      const threshold = rect.height / 3; // Use 1/3 threshold for better precision
      dropBeforeTarget.value = relativeY < threshold;
    }
  }
}

function onDragLeave(event: DragEvent) {
  // Only clear if we're actually leaving the rule item
  const target = event.target as HTMLElement;
  const relatedTarget = event.relatedTarget as HTMLElement;

  if (relatedTarget && target.contains(relatedTarget)) {
    return;
  }
}

async function onDrop(targetRuleId: number, event: DragEvent) {
  event.preventDefault();
  event.stopPropagation();

  const draggedId = draggingRuleId.value;
  if (!draggedId || draggedId === targetRuleId) {
    onDragEnd();
    return;
  }

  // Find the indices BEFORE any modifications
  const draggedIndex = rules.value.findIndex((r) => r.id === draggedId);
  const targetIndex = rules.value.findIndex((r) => r.id === targetRuleId);

  if (draggedIndex === -1 || targetIndex === -1) {
    onDragEnd();
    return;
  }

  // Remove the dragged rule from its current position
  const [draggedRule] = rules.value.splice(draggedIndex, 1);

  // Calculate the correct insertion index
  // If dragged item was before target, after removing it, target's index shifts down by 1
  let insertIndex = targetIndex;
  if (draggedIndex < targetIndex) {
    insertIndex = targetIndex - 1;
  }

  // Insert at the calculated position
  if (dropBeforeTarget.value) {
    // Insert before target (already adjusted insertIndex points to target)
    rules.value.splice(insertIndex, 0, draggedRule);
  } else {
    // Insert after target
    rules.value.splice(insertIndex + 1, 0, draggedRule);
  }

  // Update all positions to match current array order
  rules.value.forEach((rule, index) => {
    rule.position = index;
  });

  // Save the new order
  await saveRules();

  onDragEnd();
}
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <SettingGroup :icon="PhLightning" :title="t('modal.rule.rules')">
      <!-- Header with description and add button -->
      <SettingItem
        :icon="PhLightning"
        :title="t('modal.rule.rules')"
        :description="t('modal.rule.rulesDesc')"
        class="mb-2 sm:mb-3"
      >
        <ButtonControl
          :label="t('setting.rule.addRule')"
          :icon="PhPlus"
          type="secondary"
          @click="addRule"
        />
      </SettingItem>

      <!-- Empty state -->
      <div v-if="rules.length === 0" class="empty-state">
        <PhLightning :size="48" class="mx-auto mb-3 opacity-30" />
        <p class="text-text-secondary text-sm sm:text-base">{{ t('setting.rule.noRules') }}</p>
        <p class="text-text-secondary text-xs mt-1">{{ t('setting.rule.noRulesHint') }}</p>
      </div>

      <!-- Rules List -->
      <div v-else class="rules-list">
        <transition-group name="rule-reorder">
          <RuleItem
            v-for="rule in rules"
            :key="rule.id"
            :rule="rule"
            :is-applying="applyingRuleId === rule.id"
            :is-dragging="draggingRuleId === rule.id"
            :is-drag-over="dragOverRuleId === rule.id"
            :is-drop-before="dragOverRuleId === rule.id && dropBeforeTarget"
            @toggle-enabled="toggleRuleEnabled(rule)"
            @apply="applyRule(rule)"
            @edit="editRule(rule)"
            @delete="deleteRule(rule.id)"
            @drag-start="onDragStart(rule.id, $event)"
            @drag-end="onDragEnd"
            @drag-over="onDragOver(rule.id, $event)"
            @drag-leave="onDragLeave($event)"
            @drop="onDrop(rule.id, $event)"
          />
        </transition-group>
      </div>
    </SettingGroup>
  </div>

  <!-- Rule Editor Modal (Teleported to body) -->
  <Teleport to="body">
    <RuleEditorModal
      v-if="showRuleEditor"
      :show="showRuleEditor"
      :rule="editingRule"
      @close="showRuleEditor = false"
      @save="handleSaveRule"
    />
  </Teleport>
</template>

<style scoped>
@reference "../../../../style.css";

.empty-state {
  @apply text-center py-8 sm:py-12;
}

/* Rules list container */
.rules-list {
  @apply space-y-2 sm:space-y-3;
  position: relative;
}

/* Reorder transitions for Vue's transition-group */
.rule-reorder-move,
.rule-reorder-enter-active,
.rule-reorder-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.rule-reorder-enter-from {
  opacity: 0;
  transform: translateY(-20px);
}

.rule-reorder-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

/* Ensure leaving items are taken out of layout flow */
.rule-reorder-leave-active {
  position: absolute;
  width: 100%;
  z-index: 0;
}
</style>
