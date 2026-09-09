<script setup lang="ts">
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { computed, nextTick, ref, onMounted, watch, type Component, type Ref } from 'vue';
import GeneralTab from './settings/general/GeneralTab.vue';
import ReadingDisplayTab from './settings/reading/ReadingDisplayTab.vue';
import FeedsTab from './settings/feeds/FeedsTab.vue';
import ContentTab from './settings/content/ContentTab.vue';
import AITab from './settings/ai/AITab.vue';
import NetworkTab from './settings/network/NetworkTab.vue';
import PluginsTab from './settings/plugins/PluginsTab.vue';
import ShortcutsTab from './settings/shortcuts/ShortcutsTab.vue';
import RulesTab from './settings/rules/RulesTab.vue';
import StatisticsTab from './settings/statistics/StatisticsTab.vue';
import AboutTab from './settings/about/AboutTab.vue';
import DiscoverAllFeedsModal from './discovery/DiscoverAllFeedsModal.vue';
import {
  PhGear,
  PhSlidersHorizontal,
  PhBookOpen,
  PhRss,
  PhTextT,
  PhBrain,
  PhFunnel,
  PhGlobe,
  PhPuzzlePiece,
  PhKeyboard,
  PhChartBar,
  PhInfo,
  PhMagnifyingGlass,
  PhX,
} from '@phosphor-icons/vue';
import type { TabName } from '@/types/settings';
import type { ThemePreference } from '@/stores/app';
import { useSettings } from '@/composables/core/useSettings';
import { useAppUpdates } from '@/composables/core/useAppUpdates';
import { useFeedManagement } from '@/composables/feed/useFeedManagement';
import { useModalClose, LARGE_MODAL_Z_INDEX } from '@/composables/ui/useModalClose';

const store = useAppStore();
const { t, tm } = useI18n();

interface Props {
  initialTab?: TabName;
}

const props = withDefaults(defineProps<Props>(), {
  initialTab: 'general',
});

// Modal close handling; nested modals are placed above this layer automatically.
const { zIndex: modalZIndex } = useModalClose(() => emit('close'), LARGE_MODAL_Z_INDEX);

// Use composables
const { settings, fetchSettings, applySettings } = useSettings();
const {
  updateInfo,
  checkingUpdates,
  downloadingUpdate,
  installingUpdate,
  downloadProgress,
  downloadProgressKnown,
  downloadBytesWritten,
  downloadTotalBytes,
  downloadErrorCode,
  checkForUpdates: handleCheckUpdates,
  downloadAndInstallUpdate: handleDownloadInstallUpdate,
} = useAppUpdates();
const {
  handleImportOPML,
  handleExportOPML,
  handleCleanupDatabase,
  handleAddFeed,
  handleEditFeed,
  handleDeleteFeed,
  handleBatchDelete,
  handleBatchMove,
  handleBatchAddTags,
  handleBatchSetImageMode,
  handleBatchUnsetImageMode,
} = useFeedManagement();

const emit = defineEmits<{
  close: [];
}>();

const activeTab: Ref<TabName> = ref(props.initialTab);
const showDiscoverAllModal = ref(false);
const settingsContentRef = ref<HTMLElement | null>(null);
const settingsSearchRef = ref<HTMLElement | null>(null);
const settingsSearchQuery = ref('');
const settingsSearchOpen = ref(false);
const highlightedSearchIndex = ref(0);

const settingsTabs: Array<{
  id: TabName;
  icon: Component;
  labelKey: string;
  searchNamespaces: string[];
}> = [
  {
    id: 'general',
    icon: PhSlidersHorizontal,
    labelKey: 'setting.tab.general',
    searchNamespaces: ['setting.general', 'setting.update', 'setting.database'],
  },
  {
    id: 'reading',
    icon: PhBookOpen,
    labelKey: 'setting.tab.readingAndDisplay',
    searchNamespaces: ['setting.reading', 'setting.typography', 'setting.customization'],
  },
  {
    id: 'feeds',
    icon: PhRss,
    labelKey: 'sidebar.feedList.feeds',
    searchNamespaces: ['setting.feed', 'modal.feed', 'modal.discovery'],
  },
  {
    id: 'content',
    icon: PhTextT,
    labelKey: 'setting.tab.content',
    searchNamespaces: ['setting.content', 'setting.translation'],
  },
  {
    id: 'ai',
    icon: PhBrain,
    labelKey: 'setting.tab.ai',
    searchNamespaces: ['setting.ai', 'aiErrors'],
  },
  {
    id: 'rules',
    icon: PhFunnel,
    labelKey: 'modal.rule.rules',
    searchNamespaces: ['setting.rule', 'modal.rule'],
  },
  {
    id: 'network',
    icon: PhGlobe,
    labelKey: 'setting.tab.network',
    searchNamespaces: ['setting.network'],
  },
  {
    id: 'plugins',
    icon: PhPuzzlePiece,
    labelKey: 'setting.tab.plugins',
    searchNamespaces: [
      'setting.plugins',
      'setting.freshrss',
      'setting.rsshub',
      'setting.notion',
      'setting.obsidian',
      'setting.zotero',
    ],
  },
  {
    id: 'shortcuts',
    icon: PhKeyboard,
    labelKey: 'setting.shortcut.shortcuts',
    searchNamespaces: ['setting.shortcut'],
  },
  {
    id: 'statistics',
    icon: PhChartBar,
    labelKey: 'setting.statistic.statistics',
    searchNamespaces: ['setting.statistic'],
  },
  {
    id: 'about',
    icon: PhInfo,
    labelKey: 'setting.tab.about',
    searchNamespaces: ['setting.about', 'setting.update'],
  },
];

interface SettingsSearchResult {
  id: string;
  tab: TabName;
  tabLabel: string;
  label: string;
}

function collectSearchText(value: unknown): string[] {
  if (typeof value === 'string') return [value];
  if (Array.isArray(value)) return value.flatMap(collectSearchText);
  if (value && typeof value === 'object') {
    return Object.values(value as Record<string, unknown>).flatMap(collectSearchText);
  }
  return [];
}

const settingsSearchResults = computed<SettingsSearchResult[]>(() => {
  const query = settingsSearchQuery.value.trim().toLocaleLowerCase();
  if (!query) return [];

  const results: Array<SettingsSearchResult & { rank: number }> = [];
  const seen = new Set<string>();

  for (const tab of settingsTabs) {
    const tabLabel = t(tab.labelKey);
    const searchableText = [
      tabLabel,
      ...tab.searchNamespaces.flatMap((namespace) => collectSearchText(tm(namespace))),
    ];

    for (const text of searchableText) {
      const label = text.trim().replace(/\s+/g, ' ');
      const normalized = label.toLocaleLowerCase();
      const id = `${tab.id}:${normalized}`;
      if (!label || !normalized.includes(query) || seen.has(id)) continue;

      seen.add(id);
      results.push({
        id,
        tab: tab.id,
        tabLabel,
        label,
        rank: normalized === query ? 0 : normalized.startsWith(query) ? 1 : 2,
      });
    }
  }

  return results
    .sort((left, right) => left.rank - right.rank || left.label.length - right.label.length)
    .slice(0, 12)
    .map(({ rank: _rank, ...result }) => result);
});

function selectSettingsTab(tab: TabName) {
  activeTab.value = tab;
  clearSettingsSearch();
}

async function selectSettingsSearchResult(result: SettingsSearchResult) {
  activeTab.value = result.tab;
  settingsSearchOpen.value = false;

  await nextTick();
  const candidates =
    settingsContentRef.value?.querySelectorAll<HTMLElement>(
      '.setting-item, .sub-setting-item, .setting-item-col, .setting-group'
    ) || [];
  const target = result.label.toLocaleLowerCase();
  const query = settingsSearchQuery.value.trim().toLocaleLowerCase();
  const match = Array.from(candidates)
    .filter((element) => {
      const text = element.textContent?.toLocaleLowerCase() || '';
      return text.includes(target) || text.includes(query);
    })
    .sort((left, right) => (left.textContent?.length || 0) - (right.textContent?.length || 0))[0];
  if (!match) return;

  match.scrollIntoView({ behavior: 'smooth', block: 'center' });
  match.setAttribute('data-settings-search-match', 'true');
  window.setTimeout(() => match.removeAttribute('data-settings-search-match'), 1600);
}

function clearSettingsSearch() {
  settingsSearchQuery.value = '';
  settingsSearchOpen.value = false;
  highlightedSearchIndex.value = 0;
}

function handleSettingsSearchKeydown(event: KeyboardEvent) {
  const resultCount = settingsSearchResults.value.length;
  if (event.key === 'Escape') {
    settingsSearchOpen.value = false;
    return;
  }
  if (!resultCount) return;

  if (event.key === 'ArrowDown') {
    event.preventDefault();
    settingsSearchOpen.value = true;
    highlightedSearchIndex.value = (highlightedSearchIndex.value + 1) % resultCount;
  } else if (event.key === 'ArrowUp') {
    event.preventDefault();
    settingsSearchOpen.value = true;
    highlightedSearchIndex.value =
      (highlightedSearchIndex.value - 1 + resultCount) % resultCount;
  } else if (event.key === 'Enter') {
    event.preventDefault();
    const result = settingsSearchResults.value[highlightedSearchIndex.value];
    if (result) void selectSettingsSearchResult(result);
  }
}

function handleSettingsSearchFocusOut(event: FocusEvent) {
  if (
    event.relatedTarget instanceof Node &&
    settingsSearchRef.value?.contains(event.relatedTarget)
  ) {
    return;
  }
  settingsSearchOpen.value = false;
}

watch(settingsSearchQuery, (query) => {
  highlightedSearchIndex.value = 0;
  settingsSearchOpen.value = Boolean(query.trim());
});

onMounted(async () => {
  try {
    const data = await fetchSettings();
    applySettings(data, (theme: string) => store.setTheme(theme as ThemePreference));
  } catch (e) {
    console.error('Error loading settings:', e);
  }
});

function handleDiscoverAll() {
  showDiscoverAllModal.value = true;
}
</script>

<template>
  <div
    class="fixed inset-0 flex items-center justify-center bg-black/50 backdrop-blur-sm"
    :style="{ zIndex: modalZIndex }"
    data-modal-open="true"
    data-settings-modal="true"
  >
    <div
      class="bg-bg-primary w-full max-w-5xl h-full sm:h-[800px] sm:max-h-[90vh] flex flex-col rounded-none sm:rounded-2xl shadow-2xl border border-border overflow-hidden animate-fade-in mx-2 sm:mx-4 my-2 sm:my-4"
    >
      <div
        class="grid grid-cols-[minmax(0,1fr)_minmax(12rem,20rem)_minmax(0,1fr)] items-center gap-3 border-b border-border p-3 sm:p-5 shrink-0"
      >
        <h3
          class="justify-self-start text-text-secondary sm:text-lg font-semibold m-0 flex items-center gap-2"
        >
          <PhGear :size="20" :weight="'fill'" class="sm:w-6 sm:h-6" />
          {{ t('setting.tab.settingsTitle') }}
        </h3>
        <div
          ref="settingsSearchRef"
          class="relative w-full justify-self-center"
          @focusout="handleSettingsSearchFocusOut"
        >
          <PhMagnifyingGlass
            :size="16"
            class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary"
          />
          <input
            v-model="settingsSearchQuery"
            type="text"
            class="w-full rounded-lg border border-border bg-bg-secondary py-2 pl-9 pr-9 text-sm text-text-primary outline-none transition-colors focus:border-accent"
            :placeholder="t('setting.search.placeholder')"
            :aria-label="t('setting.search.placeholder')"
            :aria-expanded="settingsSearchOpen"
            aria-autocomplete="list"
            role="combobox"
            @focus="settingsSearchOpen = Boolean(settingsSearchQuery.trim())"
            @keydown="handleSettingsSearchKeydown"
          />
          <button
            v-if="settingsSearchQuery"
            type="button"
            class="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary"
            :title="t('setting.search.clear')"
            @click="clearSettingsSearch"
          >
            <PhX :size="14" />
          </button>
          <div
            v-if="settingsSearchQuery.trim() && settingsSearchOpen"
            class="absolute right-0 top-[calc(100%+0.5rem)] z-50 w-full min-w-72 overflow-hidden rounded-xl border border-border bg-bg-primary py-1.5 shadow-2xl"
            role="listbox"
          >
            <button
              v-for="(result, index) in settingsSearchResults"
              :key="result.id"
              type="button"
              role="option"
              :aria-selected="highlightedSearchIndex === index"
              :class="[
                'flex w-full items-center gap-3 px-3 py-2 text-left transition-colors',
                highlightedSearchIndex === index ? 'bg-bg-tertiary' : 'hover:bg-bg-secondary',
              ]"
              @mouseenter="highlightedSearchIndex = index"
              @mousedown.prevent
              @click="selectSettingsSearchResult(result)"
            >
              <component
                :is="settingsTabs.find((tab) => tab.id === result.tab)?.icon"
                :size="18"
                class="shrink-0 text-text-secondary"
              />
              <span class="min-w-0">
                <span class="block truncate text-sm text-text-primary">{{ result.label }}</span>
                <span class="block text-xs text-text-secondary">{{ result.tabLabel }}</span>
              </span>
            </button>
            <p
              v-if="settingsSearchResults.length === 0"
              class="m-0 px-3 py-5 text-center text-sm text-text-secondary"
            >
              {{ t('setting.search.noResults') }}
            </p>
          </div>
        </div>
        <button
          type="button"
          class="flex h-10 w-10 cursor-pointer items-center justify-center justify-self-end rounded-lg text-text-secondary transition-colors hover:bg-bg-tertiary hover:text-text-primary"
          :aria-label="t('common.close')"
          :title="t('common.close')"
          @click="emit('close')"
        >
          <PhX :size="22" />
        </button>
      </div>

      <div class="flex flex-1 min-h-0 overflow-hidden">
        <!-- Sidebar Navigation -->
        <div class="w-48 sm:w-56 border-r border-border bg-bg-secondary shrink-0 overflow-y-scroll">
          <nav class="p-2 space-y-1">
            <button
              v-for="tab in settingsTabs"
              :key="tab.id"
              :class="['sidebar-tab-btn', activeTab === tab.id ? 'active' : '']"
              @click="selectSettingsTab(tab.id)"
            >
              <component :is="tab.icon" :size="22" />
              <span>{{ t(tab.labelKey) }}</span>
            </button>
          </nav>
        </div>

        <!-- Content Area -->
        <div
          ref="settingsContentRef"
          class="settings-content flex-1 overflow-y-scroll p-3 sm:p-6 min-h-0 scroll-smooth overscroll-contain"
          data-settings-content
        >
          <GeneralTab
            v-if="activeTab === 'general'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <ReadingDisplayTab
            v-if="activeTab === 'reading'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <FeedsTab
            v-if="activeTab === 'feeds'"
            :settings="settings"
            @import-opml="handleImportOPML"
            @export-opml="handleExportOPML"
            @cleanup-database="handleCleanupDatabase"
            @add-feed="handleAddFeed"
            @edit-feed="handleEditFeed"
            @delete-feed="handleDeleteFeed"
            @batch-delete="handleBatchDelete"
            @batch-move="handleBatchMove"
            @batch-add-tags="handleBatchAddTags"
            @batch-set-image-mode="handleBatchSetImageMode"
            @batch-unset-image-mode="handleBatchUnsetImageMode"
            @discover-all="handleDiscoverAll"
            @update:settings="settings = $event"
          />

          <ContentTab
            v-if="activeTab === 'content'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <AITab
            v-if="activeTab === 'ai'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <NetworkTab
            v-if="activeTab === 'network'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <PluginsTab
            v-if="activeTab === 'plugins'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <RulesTab
            v-if="activeTab === 'rules'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <ShortcutsTab
            v-if="activeTab === 'shortcuts'"
            :settings="settings"
            @update:settings="settings = $event"
          />

          <StatisticsTab v-if="activeTab === 'statistics'" />

          <AboutTab
            v-if="activeTab === 'about'"
            :update-info="updateInfo"
            :checking-updates="checkingUpdates"
            :downloading-update="downloadingUpdate"
            :installing-update="installingUpdate"
            :download-progress="downloadProgress"
            :download-progress-known="downloadProgressKnown"
            :download-bytes-written="downloadBytesWritten"
            :download-total-bytes="downloadTotalBytes"
            :download-error-code="downloadErrorCode"
            @check-updates="handleCheckUpdates"
            @download-install-update="handleDownloadInstallUpdate"
          />
        </div>
      </div>
    </div>
  </div>

  <!-- Discover All Feeds Modal (Teleported to body) -->
  <Teleport to="body">
    <DiscoverAllFeedsModal :show="showDiscoverAllModal" @close="showDiscoverAllModal = false" />
  </Teleport>
</template>

<style scoped>
@reference "../../style.css";
.sidebar-tab-btn {
  @apply w-full flex items-center gap-3 px-3 py-2.5 rounded-lg bg-transparent text-text-secondary font-medium cursor-pointer transition-all relative;
}

.sidebar-tab-btn:hover {
  background-color: rgba(128, 128, 128, 0.1);
  color: var(--text-primary);
}

.sidebar-tab-btn.active {
  @apply text-accent;
  background-color: rgba(128, 128, 128, 0.08);
}

.sidebar-tab-btn.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 6px;
  bottom: 6px;
  width: 3px;
  background: var(--accent-color);
  border-radius: 0 2px 2px 0;
}

.settings-content {
  overflow-anchor: none;
}

:deep([data-settings-search-match='true']) {
  outline: 2px solid var(--accent-color);
  outline-offset: 2px;
  transition: outline-color 0.2s ease;
}

.btn-primary {
  @apply bg-accent text-white border-none px-5 py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors;
}

.animate-fade-in {
  animation: modalFadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalFadeIn {
  from {
    transform: translateY(-20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
</style>
