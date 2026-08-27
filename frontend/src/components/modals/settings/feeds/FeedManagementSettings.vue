<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhArrowClockwise,
  PhCheckCircle,
  PhCode,
  PhEyeSlash,
  PhFolder,
  PhImage,
  PhMagnifyingGlass,
  PhPencil,
  PhPlus,
  PhRss,
  PhSortAscending,
  PhTag,
  PhTrash,
  PhWarningCircle,
  PhX,
} from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import type { SelectOption } from '@/types/select';
import { useAppStore } from '@/stores/app';
import { useFeedManagement } from '@/composables/feed/useFeedManagement';
import { formatRelativeTime } from '@/utils/date';
import BaseSelect from '@/components/common/BaseSelect.vue';
import { ButtonControl, SettingGroup } from '@/components/settings';
import BatchActionsDropdown from './BatchActionsDropdown.vue';
import BatchTagSelectorModal from './BatchTagSelectorModal.vue';

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
  'manage-tags': [];
}>();

type SortField =
  'original' | 'name' | 'category' | 'latest_article' | 'articles_per_month' | 'update_status';
type SortDirection = 'asc' | 'desc';
type FeedIconStage = 'primary' | 'favicon' | 'fallback';

const selectedFeeds: Ref<number[]> = ref([]);
const searchQuery = ref('');
const sortField = ref<SortField>('original');
const sortDirection = ref<SortDirection>('asc');
const feedIconStages = ref<Record<number, FeedIconStage>>({});
const showBatchTagSelector = ref(false);
const pendingFeedIdsForTags = ref<number[]>([]);

const sortOptions = computed<SelectOption[]>(() => [
  { value: 'original', label: t('modal.feed.originalOrder') },
  { value: 'name', label: t('sidebar.sort.byName') },
  { value: 'category', label: t('sidebar.sort.byCategory') },
  { value: 'latest_article', label: t('sidebar.sort.byLatestArticle') },
  { value: 'articles_per_month', label: t('sidebar.sort.byArticlesPerMonth') },
  { value: 'update_status', label: t('sidebar.sort.byUpdateStatus') },
]);

const filteredFeeds = computed(() => {
  const query = searchQuery.value.trim().toLocaleLowerCase();
  if (!query) return [...store.feeds];

  return store.feeds.filter((feed) =>
    [feed.title, feed.url, feed.category, feed.website_url, feed.email_address, feed.script_path]
      .filter(Boolean)
      .some((value) => String(value).toLocaleLowerCase().includes(query))
  );
});

const sortedFeeds = computed(() => {
  const feeds = filteredFeeds.value.map((feed, index) => ({ feed, index }));
  if (sortField.value === 'original') return feeds.map(({ feed }) => feed);

  feeds.sort((left, right) => {
    const a = left.feed;
    const b = right.feed;
    let comparison = 0;

    switch (sortField.value) {
      case 'name':
        comparison = a.title.localeCompare(b.title, undefined, { sensitivity: 'base' });
        break;
      case 'category':
        comparison = (a.category || '').localeCompare(b.category || '', undefined, {
          sensitivity: 'base',
        });
        break;
      case 'latest_article':
        comparison = getTimestamp(a.latest_article_time) - getTimestamp(b.latest_article_time);
        break;
      case 'articles_per_month':
        comparison = (a.articles_per_month || 0) - (b.articles_per_month || 0);
        break;
      case 'update_status':
        comparison = (a.last_update_status || '').localeCompare(b.last_update_status || '');
        break;
    }

    if (comparison === 0) comparison = left.index - right.index;
    return sortDirection.value === 'asc' ? comparison : -comparison;
  });

  return feeds.map(({ feed }) => feed);
});

const selectableSortedFeeds = computed(() =>
  sortedFeeds.value.filter((feed) => !feed.is_freshrss_source)
);
const totalFeeds = computed(() => store.feeds.length);
const selectedCount = computed(() => selectedFeeds.value.length);
const isInitialLoading = computed(() => store.feedsLoading && store.feeds.length === 0);
const hasBlockingError = computed(() => !!store.feedsLoadError && store.feeds.length === 0);
const isAllSelected = computed(
  () =>
    selectableSortedFeeds.value.length > 0 &&
    selectableSortedFeeds.value.every((feed) => selectedFeeds.value.includes(feed.id))
);

watch(
  () => selectableSortedFeeds.value.map((feed) => feed.id),
  (visibleIds) => {
    const visibleIdSet = new Set(visibleIds);
    selectedFeeds.value = selectedFeeds.value.filter((id) => visibleIdSet.has(id));
  }
);

function getTimestamp(value?: string): number {
  if (!value) return 0;
  const timestamp = new Date(value).getTime();
  return Number.isNaN(timestamp) ? 0 : timestamp;
}

function getFriendlyErrorMessage(error: string): string {
  if (!error) return '';
  if (/timeout/i.test(error)) return t('modal.feed.errorTimeout');
  if (/connection/i.test(error)) return t('modal.feed.errorConnection');
  if (/dns/i.test(error)) return t('modal.feed.errorDNS');
  if (/certificate|ssl|tls/i.test(error)) return t('modal.feed.errorCertificate');
  if (error.includes('404')) return t('modal.feed.errorNotFound');
  if (/401|403/.test(error)) return t('modal.feed.errorUnauthorized');
  if (/500|502|503/.test(error)) return t('modal.feed.errorServer');
  if (/xml|parse|invalid/i.test(error)) return t('modal.feed.errorInvalidFormat');
  return error;
}

function getFavicon(url: string): string {
  try {
    const parsedUrl = new URL(url);
    if (!['http:', 'https:'].includes(parsedUrl.protocol)) return '';
    return `https://www.google.com/s2/favicons?domain=${parsedUrl.hostname}`;
  } catch {
    return '';
  }
}

function getFeedIconSource(feed: Feed): string {
  const stage = feedIconStages.value[feed.id] || 'primary';
  if (stage === 'fallback') return '';
  if (stage === 'primary' && feed.image_url) return feed.image_url;
  return getFavicon(feed.website_url || feed.url);
}

function handleFeedIconError(feed: Feed) {
  const stage = feedIconStages.value[feed.id] || 'primary';
  const favicon = getFavicon(feed.website_url || feed.url);
  feedIconStages.value = {
    ...feedIconStages.value,
    [feed.id]: stage === 'primary' && !!feed.image_url && !!favicon ? 'favicon' : 'fallback',
  };
}

function getFeedSource(feed: Feed): string {
  if (feed.type === 'email') return feed.email_address || t('modal.feed.email');
  if (feed.script_path) return feed.script_path;
  return feed.url || feed.website_url || '';
}

function getFeedSourceTitle(feed: Feed): string {
  const parts = [getFeedSource(feed)];
  if (feed.category) parts.push(feed.category);
  return parts.filter(Boolean).join(' · ');
}

function isXPathFeed(feed: Feed): boolean {
  return feed.type === 'HTML+XPath' || feed.type === 'XML+XPath';
}

function isRSSHubFeed(feed: Feed): boolean {
  return feed.url.startsWith('rsshub://');
}

function handleSortField(value: string | number) {
  sortField.value = value as SortField;
}

function toggleSortDirection() {
  if (sortField.value === 'original') return;
  sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc';
}

function toggleSelectAll(event: Event) {
  const target = event.target as HTMLInputElement;
  selectedFeeds.value = target.checked ? selectableSortedFeeds.value.map((feed) => feed.id) : [];
}

function handleFeedClick(feed: Feed) {
  if (feed.is_freshrss_source) return;
  emit('edit-feed', feed);
}

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

function handleBatchDelete() {
  if (!selectedFeeds.value.length) return;
  emit('batch-delete', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchMove() {
  if (!selectedFeeds.value.length) return;
  emit('batch-move', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchAddTags() {
  if (!selectedFeeds.value.length) return;
  emit('batch-add-tags', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchSetImageMode() {
  if (!selectedFeeds.value.length) return;
  emit('batch-set-image-mode', selectedFeeds.value);
  selectedFeeds.value = [];
}

function handleBatchUnsetImageMode() {
  if (!selectedFeeds.value.length) return;
  emit('batch-unset-image-mode', selectedFeeds.value);
  selectedFeeds.value = [];
}

onMounted(() => window.addEventListener('show-batch-tag-selector', handleShowBatchTagSelector));
onUnmounted(() =>
  window.removeEventListener('show-batch-tag-selector', handleShowBatchTagSelector)
);
</script>

<template>
  <SettingGroup :icon="PhRss" :title="t('modal.feed.manageFeeds')">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap gap-2">
        <ButtonControl
          :label="t('setting.feed.addFeed')"
          :icon="PhPlus"
          type="secondary"
          class="px-3 py-1.5"
          @click="emit('add-feed')"
        />
        <ButtonControl
          :label="t('common.action.deleteSelected')"
          :icon="PhTrash"
          :disabled="selectedFeeds.length === 0"
          type="danger"
          class="px-3 py-1.5"
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
        class="px-3 py-1.5"
        @click="emit('manage-tags')"
      />
    </div>

    <div class="overflow-hidden rounded-lg border border-border bg-bg-primary">
      <div
        class="flex flex-col gap-2 border-b border-border bg-bg-secondary px-3 py-2 sm:flex-row sm:items-center sm:justify-between"
      >
        <label
          class="flex cursor-pointer select-none items-center gap-2 text-sm text-text-secondary"
        >
          <input
            type="checkbox"
            :checked="isAllSelected"
            class="h-4 w-4 cursor-pointer rounded border-border text-accent focus:ring-2 focus:ring-accent"
            data-testid="feed-select-all"
            @change="toggleSelectAll"
          />
          <span>{{ t('common.search.selectAll') }}</span>
          <span class="text-xs text-text-tertiary">
            {{
              t('common.search.totalAndSelected', { total: totalFeeds, selected: selectedCount })
            }}
          </span>
        </label>

        <div class="flex min-w-0 items-center gap-2">
          <BaseSelect
            :model-value="sortField"
            :options="sortOptions"
            size="xs"
            width="w-40"
            bg-mode="secondary"
            data-testid="feed-sort-select"
            @update:model-value="handleSortField"
          />
          <button
            type="button"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-text-secondary transition-colors hover:bg-bg-tertiary hover:text-text-primary disabled:cursor-default disabled:opacity-40"
            :class="{ 'rotate-180': sortDirection === 'desc' }"
            :disabled="sortField === 'original'"
            :title="
              sortDirection === 'asc'
                ? t('modal.feed.sortAscending')
                : t('modal.feed.sortDescending')
            "
            data-testid="feed-sort-direction"
            @click="toggleSortDirection"
          >
            <PhSortAscending :size="18" />
          </button>
          <div class="relative min-w-0 flex-1 sm:w-44 sm:flex-none">
            <PhMagnifyingGlass
              :size="15"
              class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-text-tertiary"
            />
            <input
              v-model="searchQuery"
              type="search"
              :placeholder="t('common.search.searchFeeds')"
              class="h-8 w-full rounded-md border border-border bg-bg-primary pl-8 pr-8 text-sm text-text-primary outline-none transition-colors placeholder:text-text-tertiary focus:border-accent focus:ring-1 focus:ring-accent"
              data-testid="feed-search"
            />
            <button
              v-if="searchQuery"
              type="button"
              class="absolute right-1.5 top-1/2 flex -translate-y-1/2 items-center rounded p-0.5 text-text-tertiary hover:bg-bg-tertiary hover:text-text-primary"
              :title="t('common.action.clear')"
              @click="searchQuery = ''"
            >
              <PhX :size="14" />
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="store.feedsLoadError && store.feeds.length > 0"
        class="flex items-center justify-between gap-3 border-b border-border bg-yellow-50 px-3 py-2 text-xs text-yellow-800 dark:bg-yellow-950/30 dark:text-yellow-300"
        data-testid="feed-load-warning"
      >
        <span class="min-w-0 truncate" :title="store.feedsLoadError">
          {{ t('modal.feed.loadFailedKeepingData') }}
        </span>
        <button
          type="button"
          class="shrink-0 rounded px-2 py-1 font-medium hover:bg-yellow-100 dark:hover:bg-yellow-900/40"
          @click="store.fetchFeeds()"
        >
          {{ t('modal.feed.retry') }}
        </button>
      </div>

      <div class="max-h-[32rem] overflow-auto scroll-smooth" data-testid="feed-list">
        <template v-if="isInitialLoading">
          <div class="min-w-[740px]">
            <div
              v-for="index in 6"
              :key="index"
              class="grid min-h-16 grid-cols-[16px_32px_minmax(220px,1fr)_100px_96px_64px_48px_72px] items-center gap-2 border-b border-border/70 px-3 py-2 last:border-0"
              data-testid="feed-list-loading"
            >
              <div class="h-4 w-4 animate-pulse rounded bg-bg-tertiary" />
              <div class="h-8 w-8 animate-pulse rounded-lg bg-bg-tertiary" />
              <div class="min-w-0 space-y-2">
                <div class="h-3.5 w-2/5 animate-pulse rounded bg-bg-tertiary" />
                <div class="h-3 w-3/5 animate-pulse rounded bg-bg-tertiary" />
              </div>
              <div
                v-for="column in 5"
                :key="column"
                class="h-3 animate-pulse rounded bg-bg-tertiary"
              />
            </div>
          </div>
        </template>

        <div
          v-else-if="hasBlockingError"
          class="flex min-h-48 flex-col items-center justify-center gap-3 px-6 py-8 text-center"
          data-testid="feed-load-error"
        >
          <span
            class="flex h-11 w-11 items-center justify-center rounded-full bg-yellow-100 text-yellow-600 dark:bg-yellow-950/40 dark:text-yellow-300"
          >
            <PhWarningCircle :size="24" />
          </span>
          <div>
            <p class="text-sm font-medium text-text-primary">{{ t('modal.feed.loadFailed') }}</p>
            <p class="mt-1 max-w-sm text-xs text-text-tertiary" :title="store.feedsLoadError || ''">
              {{ t('modal.feed.loadFailedDesc') }}
            </p>
          </div>
          <button
            type="button"
            class="flex items-center gap-1.5 rounded-md bg-accent px-3 py-1.5 text-sm font-medium text-white transition-opacity hover:opacity-90"
            @click="store.fetchFeeds()"
          >
            <PhArrowClockwise :size="16" />
            {{ t('modal.feed.retry') }}
          </button>
        </div>

        <template v-else-if="sortedFeeds.length > 0">
          <div class="min-w-[740px]" data-testid="feed-list-grid">
            <div
              class="sticky top-0 z-10 grid grid-cols-[16px_32px_minmax(220px,1fr)_100px_96px_64px_48px_72px] items-center gap-2 border-b border-border bg-bg-secondary px-3 py-2 text-xs font-medium text-text-secondary"
              data-testid="feed-column-headers"
            >
              <span aria-hidden="true" />
              <span aria-hidden="true" />
              <span data-testid="feed-column-name">{{ t('modal.feed.feedName') }}</span>
              <span data-testid="feed-column-category">{{ t('common.form.category') }}</span>
              <span class="text-center" data-testid="feed-column-latest">{{
                t('sidebar.sort.latest')
              }}</span>
              <span class="text-center" data-testid="feed-column-frequency">{{
                t('sidebar.sort.frequency')
              }}</span>
              <span class="text-center" data-testid="feed-column-status">{{
                t('common.form.status')
              }}</span>
              <span class="text-center" data-testid="feed-column-actions">{{
                t('modal.rule.actions')
              }}</span>
            </div>

            <div
              v-for="feed in sortedFeeds"
              :key="feed.id"
              :data-feed-id="feed.id"
              :class="[
                'group grid min-h-16 grid-cols-[16px_32px_minmax(220px,1fr)_100px_96px_64px_48px_72px] items-center gap-2 border-b border-border/70 px-3 py-2 transition-colors last:border-0',
                feed.is_freshrss_source
                  ? 'cursor-default bg-info/5'
                  : 'cursor-pointer bg-bg-primary hover:bg-bg-secondary',
              ]"
              data-testid="feed-row"
              @click="handleFeedClick(feed)"
            >
              <input
                v-model="selectedFeeds"
                type="checkbox"
                :value="feed.id"
                :disabled="feed.is_freshrss_source"
                :aria-label="feed.title"
                class="h-4 w-4 cursor-pointer rounded border-border text-accent focus:ring-2 focus:ring-accent disabled:cursor-not-allowed disabled:opacity-40"
                @click.stop
              />

              <div
                class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-lg bg-bg-tertiary text-text-tertiary"
              >
                <img
                  v-if="getFeedIconSource(feed)"
                  :src="getFeedIconSource(feed)"
                  :alt="feed.title"
                  class="h-full w-full object-cover"
                  loading="lazy"
                  @error="handleFeedIconError(feed)"
                />
                <PhRss v-else :size="18" data-testid="feed-icon-fallback" />
              </div>

              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-1.5">
                  <span
                    class="min-w-0 truncate text-sm font-medium text-text-primary group-hover:text-accent"
                    :title="feed.title"
                    data-testid="feed-title"
                  >
                    {{ feed.title }}
                  </span>
                  <img
                    v-if="feed.is_freshrss_source"
                    src="/assets/plugin_icons/freshrss.svg"
                    class="h-4 w-4 shrink-0"
                    :title="t('setting.freshrss.syncedFeed')"
                    alt="FreshRSS"
                  />
                  <img
                    v-if="isRSSHubFeed(feed)"
                    src="/assets/plugin_icons/rsshub.svg"
                    class="h-4 w-4 shrink-0"
                    :title="t('setting.rsshub.feed')"
                    alt="RSSHub"
                  />
                  <PhCode
                    v-if="feed.script_path || isXPathFeed(feed)"
                    :size="15"
                    class="shrink-0 text-accent"
                    :title="feed.type || t('setting.customization.script')"
                  />
                  <PhImage
                    v-if="feed.is_image_mode"
                    :size="15"
                    class="shrink-0 text-accent"
                    :title="t('setting.feed.imageMode')"
                  />
                  <PhEyeSlash
                    v-if="feed.hide_from_timeline"
                    :size="15"
                    class="shrink-0 text-text-secondary"
                    :title="t('setting.reading.hideFromTimeline')"
                  />
                  <span
                    v-for="tag in (feed.tags || []).slice(0, 2)"
                    :key="tag.id"
                    class="hidden max-w-20 shrink-0 truncate rounded px-1.5 py-0.5 text-[10px] text-white xl:inline"
                    :style="{ backgroundColor: tag.color }"
                    :title="tag.name"
                  >
                    {{ tag.name }}
                  </span>
                </div>
                <div
                  class="mt-1 truncate text-xs text-text-tertiary"
                  :title="getFeedSourceTitle(feed)"
                  data-testid="feed-source"
                >
                  {{ getFeedSource(feed) }}
                </div>
              </div>

              <div
                class="flex min-w-0 items-center gap-1 truncate text-xs text-text-secondary"
                :title="feed.category || '-'"
                data-testid="feed-category"
              >
                <PhFolder v-if="feed.category" :size="13" class="shrink-0" />
                <span class="truncate">{{ feed.category || '-' }}</span>
              </div>

              <div
                class="truncate text-center text-xs text-text-secondary"
                :title="feed.latest_article_time || '-'"
                data-testid="feed-latest"
              >
                {{
                  feed.latest_article_time
                    ? formatRelativeTime(feed.latest_article_time, locale, t)
                    : '-'
                }}
              </div>

              <div
                class="truncate text-center text-xs text-text-secondary"
                :title="String(feed.articles_per_month ?? 0)"
                data-testid="feed-frequency"
              >
                {{ feed.articles_per_month ?? 0 }}
              </div>

              <div
                class="flex items-center justify-center"
                :title="
                  feed.last_update_status === 'failed'
                    ? getFriendlyErrorMessage(feed.last_error || '')
                    : feed.last_update_status || ''
                "
                data-testid="feed-status"
              >
                <PhCheckCircle
                  v-if="feed.last_update_status === 'success'"
                  :size="18"
                  class="text-green-500"
                  :title="t('setting.update.updateSuccess')"
                />
                <PhWarningCircle
                  v-else-if="feed.last_update_status === 'failed'"
                  :size="18"
                  class="text-yellow-500"
                  :title="getFriendlyErrorMessage(feed.last_error || '')"
                />
                <span v-else class="h-1.5 w-1.5 rounded-full bg-text-tertiary/50" />
              </div>

              <div class="flex items-center justify-center gap-0.5" data-testid="feed-actions">
                <button
                  type="button"
                  class="rounded-md p-1.5 text-text-secondary transition-colors hover:bg-bg-tertiary hover:text-accent disabled:cursor-not-allowed disabled:opacity-40"
                  :title="
                    feed.is_freshrss_source ? t('setting.freshrss.feedLocked') : t('common.edit')
                  "
                  :disabled="feed.is_freshrss_source"
                  data-testid="feed-edit"
                  @click.stop="emit('edit-feed', feed)"
                >
                  <PhPencil :size="16" />
                </button>
                <button
                  type="button"
                  class="rounded-md p-1.5 text-text-secondary transition-colors hover:bg-red-50 hover:text-red-500 disabled:cursor-not-allowed disabled:opacity-40 dark:hover:bg-red-950/30 dark:hover:text-red-400"
                  :title="
                    feed.is_freshrss_source ? t('setting.freshrss.feedLocked') : t('common.delete')
                  "
                  :disabled="feed.is_freshrss_source"
                  data-testid="feed-delete"
                  @click.stop="emit('delete-feed', feed.id)"
                >
                  <PhTrash :size="16" />
                </button>
              </div>
            </div>
          </div>
        </template>

        <div
          v-else
          class="flex min-h-40 flex-col items-center justify-center px-6 py-8 text-center text-text-secondary"
          data-testid="feed-empty-state"
        >
          <span class="mb-2 flex h-10 w-10 items-center justify-center rounded-full bg-bg-tertiary">
            <PhRss :size="22" />
          </span>
          <p class="text-sm">
            {{ searchQuery ? t('common.search.noSearchResults') : t('modal.feed.noFeeds') }}
          </p>
        </div>
      </div>
    </div>
  </SettingGroup>

  <Teleport to="body">
    <BatchTagSelectorModal
      :show="showBatchTagSelector"
      @close="handleBatchTagsClose"
      @confirm="handleBatchTagsConfirm"
    />
  </Teleport>
</template>
