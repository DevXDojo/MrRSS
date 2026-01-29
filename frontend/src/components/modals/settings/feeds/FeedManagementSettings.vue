<script setup lang="ts">
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted, onUnmounted, type Ref } from 'vue';
import {
  PhRss,
  PhPlus,
  PhTrash,
  PhFolder,
  PhPencil,
  PhSortAscending,
  PhCode,
  PhEyeSlash,
  PhCheckCircle,
  PhXCircle,
  PhImage,
  PhMagnifyingGlass,
  PhX,
  PhTag,
} from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import { formatRelativeTime } from '@/utils/date';
import { SettingGroup, ButtonControl } from '@/components/settings';
import BatchActionsDropdown from './BatchActionsDropdown.vue';
import BatchTagSelectorModal from './BatchTagSelectorModal.vue';
import { useFeedManagement } from '@/composables/feed/useFeedManagement';

const store = useAppStore();
const { t, locale } = useI18n();
const { addTagsToFeeds } = useFeedManagement();

const emit = defineEmits<{
  'add-feed': [];
  'edit-feed': [feed: Feed];
  'delete-feed': [id: number];
  'batch-delete': [ids: number[]];
  'batch-move': [ids: number[]];
  'batch-add-tags': [ids: number[]];
  'batch-set-image-mode': [ids: number[]];
  'batch-unset-image-mode': [ids: number[]];
  'select-feed': [feedId: number];
  'manage-tags': [];
}>();

const selectedFeeds: Ref<number[]> = ref([]);
const searchQuery = ref('');

// Batch tag selector state
const showBatchTagSelector = ref(false);
const pendingFeedIdsForTags = ref<number[]>([]);

function handleShowBatchTagSelector(event: Event) {
  const customEvent = event as CustomEvent<{ feedIds: number[] }>;
  pendingFeedIdsForTags.value = customEvent.detail.feedIds;
  showBatchTagSelector.value = true;
}

async function handleBatchTagsConfirm(tagIds: number[]) {
  await addTagsToFeeds(pendingFeedIdsForTags.value, tagIds);
  pendingFeedIdsForTags.value = [];
  selectedFeeds.value = [];
  showBatchTagSelector.value = false;
}

function handleBatchTagsClose() {
  pendingFeedIdsForTags.value = [];
  showBatchTagSelector.value = false;
}

// Listen for batch tag selector event
onMounted(() => {
  window.addEventListener('show-batch-tag-selector', handleShowBatchTagSelector);
});

onUnmounted(() => {
  window.removeEventListener('show-batch-tag-selector', handleShowBatchTagSelector);
});

// Sorting state
type SortField =
  | 'name'
  | 'date'
  | 'category'
  | 'latest_article'
  | 'articles_per_month'
  | 'update_status';
type SortDirection = 'asc' | 'desc';
const sortField = ref<SortField>('name');
const sortDirection = ref<SortDirection>('asc');

// Filtered and sorted feeds
const filteredFeeds = computed(() => {
  if (!store.feeds) return [];
  let feeds = [...store.feeds];

  // Apply search filter
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    feeds = feeds.filter(
      (feed) =>
        feed.title.toLowerCase().includes(query) ||
        feed.url.toLowerCase().includes(query) ||
        (feed.category && feed.category.toLowerCase().includes(query))
    );
  }

  return feeds;
});

const sortedFeeds = computed(() => {
  const feeds = [...filteredFeeds.value];

  feeds.sort((a, b) => {
    let comparison = 0;

    if (sortField.value === 'name') {
      comparison = a.title.localeCompare(b.title, undefined, { sensitivity: 'base' });
    } else if (sortField.value === 'date') {
      // Use feed ID as proxy for add time (higher ID = newer)
      comparison = a.id - b.id;
    } else if (sortField.value === 'category') {
      const catA = a.category || '';
      const catB = b.category || '';
      comparison = catA.localeCompare(catB, undefined, { sensitivity: 'base' });
    } else if (sortField.value === 'latest_article') {
      // Sort by latest article time
      const timeA = a.latest_article_time ? new Date(a.latest_article_time).getTime() : 0;
      const timeB = b.latest_article_time ? new Date(b.latest_article_time).getTime() : 0;
      comparison = timeA - timeB;
    } else if (sortField.value === 'articles_per_month') {
      // Sort by articles per month
      const countA = a.articles_per_month || 0;
      const countB = b.articles_per_month || 0;
      comparison = countA - countB;
    } else if (sortField.value === 'update_status') {
      // Sort by update status (failed first, then success)
      const statusA = a.last_update_status || 'success';
      const statusB = b.last_update_status || 'success';
      comparison = statusA.localeCompare(statusB);
    }

    return sortDirection.value === 'asc' ? comparison : -comparison;
  });

  return feeds;
});

// Feed count statistics
const totalFeeds = computed(() => store.feeds?.length || 0);
const selectedCount = computed(() => selectedFeeds.value.length);

const isAllSelected = computed(() => {
  if (!store.feeds || store.feeds.length === 0) return false;
  // Get non-FreshRSS feeds (RSSHub feeds can be selected)
  const nonManagedFeeds = store.feeds.filter((f) => !f.is_freshrss_source);
  if (nonManagedFeeds.length === 0) return false;
  // Check if all non-managed feeds are selected
  return nonManagedFeeds.every((f) => selectedFeeds.value.includes(f.id));
});

function toggleSort(field: SortField) {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc';
  } else {
    sortField.value = field;
    sortDirection.value = 'asc';
  }
}

function toggleSelectAll(e: Event) {
  const target = e.target as HTMLInputElement;
  if (!store.feeds) return;
  if (target.checked) {
    // Select only non-FreshRSS feeds (RSSHub feeds can be selected)
    selectedFeeds.value = store.feeds.filter((f) => !f.is_freshrss_source).map((f) => f.id);
  } else {
    selectedFeeds.value = [];
  }
}

function handleAddFeed() {
  emit('add-feed');
}

function handleEditFeed(feed: Feed) {
  emit('edit-feed', feed);
}

function handleDeleteFeed(id: number) {
  emit('delete-feed', id);
}

function handleBatchDelete() {
  if (selectedFeeds.value.length === 0) return;
  emit('batch-delete', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchMove() {
  if (selectedFeeds.value.length === 0) return;
  emit('batch-move', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchAddTags() {
  if (selectedFeeds.value.length === 0) return;
  emit('batch-add-tags', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchSetImageMode() {
  if (selectedFeeds.value.length === 0) return;
  emit('batch-set-image-mode', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchUnsetImageMode() {
  if (selectedFeeds.value.length === 0) return;
  emit('batch-unset-image-mode', selectedFeeds.value);
  selectedFeeds.value = [];
}

function getFavicon(url: string): string {
  try {
    return `https://www.google.com/s2/favicons?domain=${new URL(url).hostname}`;
  } catch {
    return '';
  }
}

function isScriptFeed(feed: Feed): boolean {
  return !!feed.script_path;
}

function isXPathFeed(feed: Feed): boolean {
  return feed.type === 'HTML+XPath' || feed.type === 'XML+XPath';
}

function isEmailFeed(feed: Feed): boolean {
  return feed.type === 'email';
}

function isFreshRSSFeed(feed: Feed): boolean {
  return !!feed.is_freshrss_source;
}

function isRSSHubFeed(feed: Feed): boolean {
  return feed.url.startsWith('rsshub://');
}

async function handleFeedClick(feed: Feed, event: Event) {
  // Don't select feed if clicking on checkbox, edit button, or delete button
  const target = event.target as HTMLElement;
  if (
    target.tagName === 'INPUT' ||
    target.tagName === 'BUTTON' ||
    target.closest('button') ||
    target.closest('input[type="checkbox"]')
  ) {
    return;
  }
  // Reset to 'all' filter first to ensure proper navigation
  await store.setFilter('all');
  // Select the feed and emit event to close settings modal
  store.setFeed(feed.id);
  emit('select-feed', feed.id);
}

function handleManageTags() {
  emit('manage-tags');
}
</script>

<template>
  <SettingGroup :icon="PhRss" :title="t('modal.feed.manageFeeds')">
    <div class="flex flex-wrap justify-between gap-1.5 sm:gap-2 mb-2">
      <div class="flex flex-wrap gap-1.5 sm:gap-2">
        <ButtonControl
          :label="t('setting.feed.addFeed')"
          :icon="PhPlus"
          type="secondary"
          class="py-1.5 px-2.5 sm:px-3"
          @click="handleAddFeed"
        />
        <ButtonControl
          :label="t('common.action.deleteSelected')"
          :icon="PhTrash"
          :disabled="selectedFeeds.length === 0"
          type="danger"
          class="py-1.5 px-2.5 sm:px-3"
          @click="handleBatchDelete"
        />
        <BatchActionsDropdown
          :disabled="selectedFeeds.length === 0"
          @move="handleBatchMove"
          @add-tags="handleBatchAddTags"
          @set-image-mode="handleBatchSetImageMode"
          @unset-image-mode="handleBatchUnsetImageMode"
        />
      </div>
      <ButtonControl
        :label="t('modal.tag.manageTags')"
        :icon="PhTag"
        type="secondary"
        class="py-1.5 px-2.5 sm:px-3"
        @click="handleManageTags"
      />
    </div>

    <div class="border border-border rounded-lg bg-bg-secondary">
      <!-- Table Header -->
      <div
        class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-1.5 sm:p-2 border-b border-border bg-bg-tertiary"
      >
        <div class="flex items-center gap-2 flex-wrap">
          <label class="flex items-center gap-2 cursor-pointer select-none">
            <input
              type="checkbox"
              :checked="isAllSelected"
              class="w-3.5 h-3.5 sm:w-4 sm:h-4 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer"
              @change="toggleSelectAll"
            />
            <span class="hidden sm:inline text-xs sm:text-sm">{{
              t('common.search.selectAll')
            }}</span>
            <span class="text-xs text-text-tertiary"
              >({{
                t('common.search.totalAndSelected', { total: totalFeeds, selected: selectedCount })
              }})</span
            >
          </label>
        </div>
        <div class="flex items-center gap-1 flex-wrap justify-between sm:justify-end">
          <div class="flex items-center gap-1 flex-wrap">
            <PhSortAscending :size="16" class="text-text-secondary" />
            <button
              :class="[
                'px-1.5 py-0.5 text-xs rounded transition-colors whitespace-nowrap',
                sortField === 'name'
                  ? 'bg-accent text-white'
                  : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
              ]"
              @click="toggleSort('name')"
            >
              {{ t('sidebar.sort.byName') }}
              <span v-if="sortField === 'name'">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
            </button>
            <button
              :class="[
                'px-1.5 py-0.5 text-xs rounded transition-colors whitespace-nowrap',
                sortField === 'category'
                  ? 'bg-accent text-white'
                  : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
              ]"
              @click="toggleSort('category')"
            >
              {{ t('sidebar.sort.byCategory') }}
              <span v-if="sortField === 'category'">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
            </button>
            <button
              :class="[
                'px-1.5 py-0.5 text-xs rounded transition-colors whitespace-nowrap',
                sortField === 'latest_article'
                  ? 'bg-accent text-white'
                  : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
              ]"
              :title="t('sidebar.sort.byLatestArticle')"
              @click="toggleSort('latest_article')"
            >
              {{ t('sidebar.sort.latest') }}
              <span v-if="sortField === 'latest_article'">{{
                sortDirection === 'asc' ? '↑' : '↓'
              }}</span>
            </button>
            <button
              :class="[
                'px-1.5 py-0.5 text-xs rounded transition-colors whitespace-nowrap',
                sortField === 'articles_per_month'
                  ? 'bg-accent text-white'
                  : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
              ]"
              :title="t('sidebar.sort.byArticlesPerMonth')"
              @click="toggleSort('articles_per_month')"
            >
              {{ t('sidebar.sort.frequency') }}
              <span v-if="sortField === 'articles_per_month'">{{
                sortDirection === 'asc' ? '↑' : '↓'
              }}</span>
            </button>
            <button
              :class="[
                'px-1.5 py-0.5 text-xs rounded transition-colors whitespace-nowrap',
                sortField === 'update_status'
                  ? 'bg-accent text-white'
                  : 'bg-bg-secondary text-text-primary hover:bg-bg-primary',
              ]"
              :title="t('sidebar.sort.byUpdateStatus')"
              @click="toggleSort('update_status')"
            >
              {{ t('common.form.status') }}
              <span v-if="sortField === 'update_status'">{{
                sortDirection === 'asc' ? '↑' : '↓'
              }}</span>
            </button>
          </div>
          <!-- Search Box -->
          <div class="ml-2 relative w-28 sm:w-40 shrink-0">
            <PhMagnifyingGlass
              :size="14"
              class="absolute left-2 top-1/2 -translate-y-1/2 text-text-secondary"
            />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('common.search.searchFeeds')"
              class="w-full pl-7 pr-7 py-1 text-xs sm:text-sm bg-bg-secondary border border-border rounded focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
            />
            <PhX
              v-if="searchQuery"
              :size="14"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-text-secondary cursor-pointer hover:text-text-primary"
              @click="searchQuery = ''"
            />
          </div>
        </div>
      </div>

      <!-- Scrollable Content -->
      <div class="overflow-y-auto max-h-64 sm:max-h-96 lg:max-h-[32rem] scroll-smooth">
        <!-- Column Header (Desktop) -->
        <div
          class="hidden lg:grid grid-cols-[16px,16px,2fr,100px,110px,40px,44px,52px] gap-2 px-2 py-1.5 bg-bg-tertiary border-b border-border text-xs text-text-secondary font-medium"
        >
          <div></div>
          <div></div>
          <div>{{ t('common.form.title') }}</div>
          <div>{{ t('common.form.category') }}</div>
          <div class="text-center">{{ t('sidebar.sort.latest') }}</div>
          <div class="text-center">{{ t('sidebar.sort.frequency') }}</div>
          <div class="text-center">{{ t('common.form.status') }}</div>
          <div></div>
        </div>

        <!-- Column Header (Medium screens) -->
        <div
          class="hidden sm:grid lg:hidden grid-cols-[16px,16px,1fr,90px,100px,40px,44px,52px] gap-2 px-2 py-1.5 bg-bg-tertiary border-b border-border text-xs text-text-secondary font-medium"
        >
          <div></div>
          <div></div>
          <div>{{ t('common.form.title') }}</div>
          <div>{{ t('common.form.category') }}</div>
          <div class="text-center">{{ t('sidebar.sort.latest') }}</div>
          <div class="text-center">{{ t('sidebar.sort.frequency') }}</div>
          <div class="text-center">{{ t('common.form.status') }}</div>
          <div></div>
        </div>

        <!-- Feed Rows -->
        <div
          v-for="feed in sortedFeeds"
          :key="feed.id"
          :class="[
            'grid grid-cols-[auto,auto,1fr,auto] sm:grid-cols-[16px,16px,1fr,90px,100px,40px,44px,52px] lg:grid-cols-[16px,16px,2fr,100px,110px,40px,44px,52px] gap-1.5 sm:gap-2 p-1.5 sm:p-2 border-b border-border last:border-0 items-center cursor-pointer',
            feed.is_freshrss_source ? 'bg-info/10' : 'bg-bg-primary hover:bg-bg-secondary',
          ]"
          @click="handleFeedClick(feed, $event)"
        >
          <!-- Checkbox -->
          <input
            v-model="selectedFeeds"
            type="checkbox"
            :value="feed.id"
            :disabled="feed.is_freshrss_source"
            class="w-3.5 h-3.5 sm:w-4 sm:h-4 shrink-0 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer"
            :class="{
              'cursor-not-allowed opacity-50': feed.is_freshrss_source,
            }"
          />

          <!-- Favicon -->
          <div class="w-4 h-4 flex items-center justify-center shrink-0">
            <img
              :src="getFavicon(feed.url)"
              class="w-full h-full object-contain"
              @error="
                ($event: Event) => {
                  const target = $event.target as HTMLImageElement;
                  if (target) target.style.display = 'none';
                }
              "
            />
          </div>

          <!-- Title Column -->
          <div class="min-w-0 flex-1">
            <div
              class="font-medium text-xs sm:text-sm flex items-center gap-1 sm:gap-2 overflow-hidden"
            >
              <span class="truncate" :title="feed.title">{{ feed.title }}</span>
              <!-- Feed Type Indicators -->
              <img
                v-if="feed.is_freshrss_source"
                src="/assets/plugin_icons/freshrss.svg"
                class="w-4 h-4 sm:w-4 sm:h-4 shrink-0 inline"
                :title="t('setting.freshrss.syncedFeed')"
                alt="FreshRSS"
              />
              <img
                v-if="isRSSHubFeed(feed)"
                src="/assets/plugin_icons/rsshub.svg"
                class="w-4 h-4 sm:w-4 sm:h-4 shrink-0 inline"
                :title="t('setting.rsshub.feed')"
                alt="RSSHub"
              />
              <PhImage
                v-if="feed.is_image_mode"
                :size="14"
                class="text-accent shrink-0 inline"
                :title="t('setting.feed.imageMode')"
              />
              <PhEyeSlash
                v-if="feed.hide_from_timeline"
                :size="14"
                class="text-text-secondary shrink-0 inline"
                :title="t('setting.reading.hideFromTimeline')"
              />
              <!-- Tags (limited to prevent overflow) -->
              <div
                v-if="feed.tags && feed.tags.length > 0"
                class="flex gap-0.5 shrink-0 overflow-hidden"
              >
                <span
                  v-for="tag in feed.tags.slice(0, 3)"
                  :key="tag.id"
                  class="text-[9px] px-1.5 py-0.5 rounded text-white whitespace-nowrap leading-tight shrink-0"
                  :style="{ backgroundColor: tag.color }"
                  :title="tag.name"
                >
                  {{ tag.name }}
                </span>
                <span
                  v-if="feed.tags.length > 3"
                  class="text-[9px] px-1.5 py-0.5 rounded bg-bg-tertiary text-text-secondary whitespace-nowrap leading-tight shrink-0"
                  :title="
                    feed.tags
                      .slice(3)
                      .map((t) => t.name)
                      .join(', ')
                  "
                >
                  +{{ feed.tags.length - 3 }}
                </span>
              </div>
            </div>
            <!-- Mobile-only URL display -->
            <div class="text-xs text-text-secondary truncate sm:hidden">
              <span
                v-if="isFreshRSSFeed(feed)"
                class="text-info"
                :title="t('setting.freshrss.syncedFeed')"
              >
                {{ feed.url }}
              </span>
              <span
                v-else-if="isRSSHubFeed(feed)"
                class="text-info"
                :title="t('setting.rsshub.feed')"
              >
                {{ feed.url }}
              </span>
              <span
                v-else-if="isScriptFeed(feed)"
                class="flex items-center gap-1"
                :title="t('setting.customization.script')"
              >
                <PhCode :size="12" class="inline text-accent" />
                {{ feed.script_path }}
              </span>
              <span v-else-if="isXPathFeed(feed)" class="text-accent" :title="feed.type">
                [{{ feed.type }}] {{ feed.url }}
              </span>
              <span
                v-else-if="isEmailFeed(feed)"
                class="text-accent"
                :title="t('modal.feed.email')"
              >
                [{{ t('modal.feed.email') }}]
                <span v-if="feed.email_address">{{ feed.email_address }}</span>
              </span>
              <span v-else>{{ feed.url }}</span>
            </div>
          </div>

          <!-- Category Column (Desktop) -->
          <div class="hidden sm:block min-w-0">
            <div class="text-sm text-text-secondary truncate flex items-center gap-1">
              <PhFolder v-if="feed.category" :size="14" class="inline shrink-0" />
              <span class="truncate">{{ feed.category || '-' }}</span>
            </div>
          </div>

          <!-- Latest Article Time (Desktop) -->
          <div class="hidden sm:block min-w-0 text-sm text-text-secondary truncate text-center">
            <span v-if="feed.latest_article_time" :title="t('sidebar.sort.latest')">
              {{ formatRelativeTime(feed.latest_article_time, locale, t) }}
            </span>
            <span v-else class="text-text-tertiary">-</span>
          </div>

          <!-- Articles Per Month (Desktop) -->
          <div class="hidden sm:block min-w-0 text-sm text-text-secondary truncate text-center">
            <span :title="t('sidebar.sort.frequency')">
              {{
                feed.articles_per_month !== null && feed.articles_per_month !== undefined
                  ? feed.articles_per_month
                  : 0
              }}
            </span>
          </div>

          <!-- Update Status (Desktop) -->
          <div class="hidden sm:flex min-w-0 items-center justify-center">
            <PhCheckCircle
              v-if="feed.last_update_status === 'success'"
              :size="18"
              class="text-green-500"
              :title="t('setting.update.updateSuccess')"
            />
            <PhXCircle
              v-else-if="feed.last_update_status === 'failed'"
              :size="18"
              class="text-red-500"
              :title="feed.last_error || t('setting.update.updateFailed')"
            />
            <span v-else class="text-text-tertiary text-sm">?</span>
          </div>

          <!-- Actions -->
          <div class="flex gap-0.5 sm:gap-1 shrink-0">
            <button
              class="text-accent hover:bg-bg-tertiary p-1 rounded text-sm"
              :title="feed.is_freshrss_source ? t('setting.freshrss.feedLocked') : t('common.edit')"
              :disabled="feed.is_freshrss_source"
              :class="{
                'cursor-not-allowed opacity-50': feed.is_freshrss_source,
              }"
              @click="!feed.is_freshrss_source && handleEditFeed(feed)"
            >
              <PhPencil :size="16" class="sm:w-4 sm:h-4" />
            </button>
            <button
              class="text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 p-1 rounded text-sm"
              :title="
                feed.is_freshrss_source ? t('setting.freshrss.feedLocked') : t('common.delete')
              "
              :disabled="feed.is_freshrss_source"
              :class="{
                'cursor-not-allowed opacity-50': feed.is_freshrss_source,
              }"
              @click="!feed.is_freshrss_source && handleDeleteFeed(feed.id)"
            >
              <PhTrash :size="16" class="sm:w-4 sm:h-4" />
            </button>
          </div>
        </div>

        <!-- Empty State -->
        <div
          v-if="sortedFeeds.length === 0"
          class="flex flex-col items-center justify-center py-8 text-text-secondary"
        >
          <PhRss :size="32" class="mb-2" />
          <p class="text-sm">
            {{ searchQuery ? t('common.search.noSearchResults') : t('modal.feed.noFeeds') }}
          </p>
        </div>
      </div>
    </div>
  </SettingGroup>

  <!-- Batch Tag Selector Modal (Teleported to body) -->
  <Teleport to="body">
    <BatchTagSelectorModal
      :show="showBatchTagSelector"
      @close="handleBatchTagsClose"
      @confirm="handleBatchTagsConfirm"
    />
  </Teleport>
</template>

<style scoped>
@reference "../../../../style.css";
</style>
