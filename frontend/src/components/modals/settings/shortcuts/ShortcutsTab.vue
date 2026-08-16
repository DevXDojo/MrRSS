<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, onMounted, onUnmounted, computed, watch, type Ref, type Component } from 'vue';
import {
  PhKeyboard,
  PhArrowDown,
  PhArrowUp,
  PhArrowRight,
  PhX,
  PhBookOpen,
  PhStar,
  PhClockCountdown,
  PhArrowSquareOut,
  PhArticle,
  PhArrowClockwise,
  PhCheckCircle,
  PhGear,
  PhPlus,
  PhMagnifyingGlass,
  PhListDashes,
  PhCircle,
  PhArrowCounterClockwise,
  PhFunnel,
} from '@phosphor-icons/vue';
import ShortcutItem from './ShortcutItem.vue';
import type { SettingsData } from '@/types/settings';
import { useSettingsAutoSave } from '@/composables/core/useSettingsAutoSave';
import { ButtonControl, SettingWithToggle, TipBox } from '@/components/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Create a computed ref that returns the settings object for auto-save
const settingsRef = computed(() => props.settings);

// Use composable for auto-save with reactivity
useSettingsAutoSave(settingsRef);

// Update settings and dispatch events
function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });

  // Dispatch event to notify keyboard shortcuts system when toggling
  if (key === 'shortcuts_enabled' && typeof window !== 'undefined') {
    window.dispatchEvent(
      new CustomEvent('shortcuts-enabled-changed', {
        detail: { enabled: value },
      })
    );
  }
}

interface Shortcuts {
  nextArticle: string;
  previousArticle: string;
  openArticle: string;
  closeArticle: string;
  toggleReadStatus: string;
  toggleFavoriteStatus: string;
  toggleReadLaterStatus: string;
  openInBrowser: string;
  toggleContentView: string;
  refreshFeeds: string;
  markAllRead: string;
  openSettings: string;
  addFeed: string;
  focusSearch: string;
  toggleFilter: string;
  toggleUnreadFilter: string;
  toggleFavoritesFilter: string;
  toggleReadLaterFilter: string;
  goToAllArticles: string;
  goToUnread: string;
  goToFavorites: string;
  goToReadLater: string;
}

interface ShortcutItemData {
  key: keyof Shortcuts;
  label: string;
  icon: Component;
}

// Default shortcuts configuration
const defaultShortcuts: Shortcuts = {
  nextArticle: 'j',
  previousArticle: 'k',
  openArticle: 'Enter',
  closeArticle: 'Escape',
  toggleReadStatus: 'r',
  toggleFavoriteStatus: 's',
  toggleReadLaterStatus: 'l',
  openInBrowser: 'o',
  toggleContentView: 'v',
  refreshFeeds: 'Shift+r',
  markAllRead: 'Shift+a',
  openSettings: ',',
  addFeed: 'a',
  focusSearch: '/',
  toggleFilter: 'f',
  toggleUnreadFilter: 'Alt+r',
  toggleFavoritesFilter: 'Alt+s',
  toggleReadLaterFilter: 'Alt+l',
  goToAllArticles: '1',
  goToUnread: '2',
  goToFavorites: '3',
  goToReadLater: '4',
};

// Current shortcuts (loaded from settings or use defaults)
const shortcuts: Ref<Shortcuts> = ref({ ...defaultShortcuts });

// Track which shortcut is being edited
const editingShortcut: Ref<keyof Shortcuts | null> = ref(null);
const recordedKey = ref('');

// Shortcut groups for display
const shortcutGroups = computed<Array<{ label: string; items: ShortcutItemData[] }>>(() => [
  {
    label: t('shortcut.category.navigation'),
    items: [
      { key: 'nextArticle', label: t('article.navigation.nextArticle'), icon: PhArrowDown },
      { key: 'previousArticle', label: t('article.navigation.previousArticle'), icon: PhArrowUp },
      { key: 'openArticle', label: t('article.action.openArticle'), icon: PhArrowRight },
      { key: 'closeArticle', label: t('article.action.closeArticle'), icon: PhX },
      {
        key: 'goToAllArticles',
        label: t('article.navigation.goToAllArticles'),
        icon: PhListDashes,
      },
      { key: 'goToUnread', label: t('article.navigation.goToUnread'), icon: PhCircle },
      { key: 'goToFavorites', label: t('article.navigation.goToFavorites'), icon: PhStar },
      {
        key: 'goToReadLater',
        label: t('article.navigation.goToReadLater'),
        icon: PhClockCountdown,
      },
    ],
  },
  {
    label: t('shortcut.category.articles'),
    items: [
      { key: 'toggleReadStatus', label: t('shortcut.toggle.readStatus'), icon: PhBookOpen },
      {
        key: 'toggleFavoriteStatus',
        label: t('article.action.toggleFavoriteStatus'),
        icon: PhStar,
      },
      {
        key: 'toggleReadLaterStatus',
        label: t('shortcut.toggle.readLaterStatus'),
        icon: PhClockCountdown,
      },
      {
        key: 'openInBrowser',
        label: t('article.action.openInBrowserShortcut'),
        icon: PhArrowSquareOut,
      },
      { key: 'toggleContentView', label: t('shortcut.toggle.contentView'), icon: PhArticle },
    ],
  },
  {
    label: t('shortcut.category.other'),
    items: [
      {
        key: 'refreshFeeds',
        label: t('article.action.refreshFeedsShortcut'),
        icon: PhArrowClockwise,
      },
      { key: 'markAllRead', label: t('article.action.markAllReadShortcut'), icon: PhCheckCircle },
      { key: 'openSettings', label: t('setting.shortcut.openSettingsShortcut'), icon: PhGear },
      { key: 'addFeed', label: t('setting.shortcut.addFeedShortcut'), icon: PhPlus },
      { key: 'focusSearch', label: t('setting.shortcut.focusFeedSearch'), icon: PhMagnifyingGlass },
      { key: 'toggleFilter', label: t('shortcut.toggle.filter'), icon: PhFunnel },
      { key: 'toggleUnreadFilter', label: t('shortcut.toggle.unreadFilter'), icon: PhCircle },
      { key: 'toggleFavoritesFilter', label: t('shortcut.toggle.favoritesFilter'), icon: PhStar },
      {
        key: 'toggleReadLaterFilter',
        label: t('shortcut.toggle.readLaterFilter'),
        icon: PhClockCountdown,
      },
    ],
  },
]);

// Load shortcuts from settings
onMounted(() => {
  if (props.settings.shortcuts) {
    try {
      const parsed =
        typeof props.settings.shortcuts === 'string'
          ? JSON.parse(props.settings.shortcuts)
          : props.settings.shortcuts;
      shortcuts.value = { ...defaultShortcuts, ...parsed };
    } catch (e) {
      console.error('Error parsing shortcuts:', e);
      shortcuts.value = { ...defaultShortcuts };
    }
  }

  // Add global keyboard listener for recording
  window.addEventListener('keydown', handleKeyRecord, true);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyRecord, true);
});

// Start editing a shortcut
function startEditing(shortcutKey: keyof Shortcuts) {
  editingShortcut.value = shortcutKey;
  recordedKey.value = '';
}

// Stop editing
function stopEditing() {
  editingShortcut.value = null;
  recordedKey.value = '';
}

// Handle key recording
function handleKeyRecord(e: KeyboardEvent) {
  if (!editingShortcut.value) return;

  e.preventDefault();
  e.stopPropagation();

  // Handle Escape to clear the shortcut
  if (e.key === 'Escape' && !e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey) {
    // Clear the shortcut
    shortcuts.value[editingShortcut.value] = '';
    saveShortcuts();
    window.showToast(t('setting.shortcut.shortcutsCleared'), 'info');
    stopEditing();
    return;
  }

  // Build key combination
  let key = '';
  if (e.ctrlKey) key += 'Ctrl+';
  if (e.altKey) key += 'Alt+';
  if (e.shiftKey) key += 'Shift+';
  if (e.metaKey) key += 'Meta+';

  // Get the actual key
  let actualKey = e.key;

  // Skip modifier keys alone
  if (['Control', 'Alt', 'Shift', 'Meta'].includes(actualKey)) {
    return;
  }

  // Normalize key names
  if (actualKey === ' ') actualKey = 'Space';
  else if (actualKey.length === 1) actualKey = actualKey.toLowerCase();

  key += actualKey;

  // Check for conflicts
  const conflictKey = Object.entries(shortcuts.value).find(
    ([k, v]) => v === key && k !== editingShortcut.value
  );

  if (conflictKey) {
    window.showToast(t('setting.shortcut.shortcutsConflict'), 'warning');
    stopEditing();
    return;
  }

  // Update the shortcut
  shortcuts.value[editingShortcut.value] = key;
  saveShortcuts();
  window.showToast(t('setting.shortcut.shortcutsUpdated'), 'success');
  stopEditing();
}

// Save shortcuts to settings
async function saveShortcuts() {
  try {
    // Update props.settings.shortcuts
    const updatedSettings = { ...props.settings, shortcuts: JSON.stringify(shortcuts.value) };
    emit('update:settings', updatedSettings);

    // The parent component will handle auto-save via the watcher
    // But we also dispatch an event to notify the app
    window.dispatchEvent(
      new CustomEvent('shortcuts-changed', {
        detail: { shortcuts: shortcuts.value },
      })
    );
  } catch (e) {
    console.error('Error saving shortcuts:', e);
  }
}

// Reset all shortcuts to defaults
function resetToDefaults() {
  shortcuts.value = { ...defaultShortcuts };
  saveShortcuts();
  window.showToast(t('setting.shortcut.shortcutsUpdated'), 'success');
}

// Watch for settings changes from parent
watch(
  () => props.settings.shortcuts,
  (newVal) => {
    if (newVal) {
      try {
        const parsed = typeof newVal === 'string' ? JSON.parse(newVal) : newVal;
        shortcuts.value = { ...defaultShortcuts, ...parsed };
      } catch (e) {
        console.error('Error parsing shortcuts:', e);
      }
    }
  },
  { immediate: true }
);
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2 sm:gap-3">
        <PhKeyboard :size="20" class="text-text-secondary sm:w-6 sm:h-6" />
        <div>
          <h3 class="font-semibold text-sm sm:text-base">{{ t('setting.shortcut.shortcuts') }}</h3>
          <p class="text-xs text-text-secondary hidden sm:block">
            {{ t('setting.shortcut.shortcutsDesc') }}
          </p>
        </div>
      </div>
      <ButtonControl
        :label="t('common.action.resetToDefault')"
        :icon="PhArrowCounterClockwise"
        type="secondary"
        @click="resetToDefaults"
      />
    </div>

    <!-- Enable/Disable Shortcuts Toggle -->
    <SettingWithToggle
      :icon="PhKeyboard"
      :title="t('setting.shortcut.shortcutsEnabled')"
      :description="t('setting.shortcut.shortcutsEnabledDesc')"
      :model-value="settings.shortcuts_enabled === true"
      @update:model-value="updateSetting('shortcuts_enabled', $event)"
    />

    <!-- Tip moved to top with improved styling -->
    <TipBox type="tip" :title="t('common.action.escToClear')" />

    <div v-for="group in shortcutGroups" :key="group.label" class="setting-group">
      <label
        class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
      >
        {{ group.label }}
      </label>

      <div class="space-y-2">
        <ShortcutItem
          v-for="item in group.items"
          :key="item.key"
          :item="item"
          :shortcut-value="shortcuts[item.key as keyof Shortcuts]"
          :is-editing="editingShortcut === item.key"
          @edit="startEditing(item.key as keyof Shortcuts)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped></style>
