<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import BaseSelect from '@/components/common/BaseSelect.vue';
import type { SelectOption } from '@/types/select';
import type { ProxyMode, RefreshMode } from '@/composables/feed/useFeedForm';

interface Props {
  imageGalleryEnabled: boolean;
  isImageMode: boolean;
  hideFromTimeline: boolean;
  articleViewMode: 'global' | 'webpage' | 'rendered' | 'external';
  autoExpandContent: 'global' | 'enabled' | 'disabled';
  proxyMode: ProxyMode;
  proxyType: string;
  proxyHost: string;
  proxyPort: string;
  proxyUsername: string;
  proxyPassword: string;
  refreshMode: RefreshMode;
  refreshInterval: number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:isImageMode': [value: boolean];
  'update:hideFromTimeline': [value: boolean];
  'update:articleViewMode': [value: 'global' | 'webpage' | 'rendered' | 'external'];
  'update:autoExpandContent': [value: 'global' | 'enabled' | 'disabled'];
  'update:proxyMode': [value: ProxyMode];
  'update:proxyType': [value: string];
  'update:proxyHost': [value: string];
  'update:proxyPort': [value: string];
  'update:proxyUsername': [value: string];
  'update:proxyPassword': [value: string];
  'update:refreshMode': [value: RefreshMode];
  'update:refreshInterval': [value: number];
}>();

const { t } = useI18n();

// Article View Mode options
const articleViewModeOptions = computed<SelectOption[]>(() => [
  { value: 'global', label: t('setting.feed.useGlobalSettings') },
  { value: 'webpage', label: t('setting.reading.viewAsWebpage') },
  { value: 'rendered', label: t('setting.reading.viewAsRendered') },
  { value: 'external', label: t('article.action.viewModeExternal') },
]);

// Auto Expand Content options
const autoExpandContentOptions = computed<SelectOption[]>(() => [
  { value: 'global', label: t('setting.feed.useGlobalSettings') },
  { value: 'enabled', label: t('common.form.enabled') },
  { value: 'disabled', label: t('common.form.disabled') },
]);

// Proxy Mode options
const proxyModeOptions = computed<SelectOption[]>(() => [
  { value: 'global', label: t('setting.network.useGlobalProxy') },
  { value: 'custom', label: t('setting.network.useCustomProxy') },
  { value: 'none', label: t('setting.network.noProxy') },
]);

// Proxy Type options
const proxyTypeOptions = computed<SelectOption[]>(() => [
  { value: 'http', label: t('setting.network.httpProxy') },
  { value: 'https', label: t('setting.network.httpsProxy') },
  { value: 'socks5', label: t('setting.network.socks5Proxy') },
]);

// Refresh Mode options
const refreshModeOptions = computed<SelectOption[]>(() => [
  { value: 'global', label: t('setting.feed.useGlobalRefresh') },
  { value: 'intelligent', label: t('setting.feed.useIntelligentInterval') },
  { value: 'custom', label: t('setting.feed.useCustomInterval') },
  { value: 'never', label: t('setting.feed.neverRefresh') },
]);

// Helper functions for emit with proper typing
function handleArticleViewModeChange(value: string | number) {
  emit('update:articleViewMode', value as 'global' | 'webpage' | 'rendered' | 'external');
}

function handleAutoExpandContentChange(value: string | number) {
  emit('update:autoExpandContent', value as 'global' | 'enabled' | 'disabled');
}

function handleProxyModeChange(value: string | number) {
  emit('update:proxyMode', value as ProxyMode);
}

function handleProxyTypeChange(value: string | number) {
  emit('update:proxyType', String(value));
}

function handleRefreshModeChange(value: string | number) {
  emit('update:refreshMode', value as RefreshMode);
}
</script>

<template>
  <!-- Advanced Settings Section (Collapsible) -->
  <div class="mb-3 sm:mb-4 space-y-3 sm:space-y-4">
    <!-- Image Mode Toggle (only shown if image gallery is enabled) -->
    <div
      v-if="props.imageGalleryEnabled"
      class="p-3 rounded-lg bg-bg-secondary border border-border"
    >
      <label class="flex items-center justify-between cursor-pointer">
        <div>
          <span class="font-semibold text-xs sm:text-sm text-text-primary">{{
            t('setting.feed.imageMode')
          }}</span>
          <p class="text-[10px] sm:text-xs text-text-secondary mt-0.5">
            {{ t('setting.feed.imageModeDesc') }}
          </p>
        </div>
        <input
          :checked="props.isImageMode"
          type="checkbox"
          class="toggle"
          @change="emit('update:isImageMode', ($event.target as HTMLInputElement).checked)"
        />
      </label>
    </div>

    <!-- Hide from Timeline Toggle -->
    <div class="p-3 rounded-lg bg-bg-secondary border border-border">
      <label class="flex items-center justify-between cursor-pointer">
        <div>
          <span class="font-semibold text-xs sm:text-sm text-text-primary">{{
            t('setting.reading.hideFromTimeline')
          }}</span>
          <p class="text-[10px] sm:text-xs text-text-secondary mt-0.5">
            {{ t('setting.reading.hideFromTimelineDesc') }}
          </p>
        </div>
        <input
          :checked="props.hideFromTimeline"
          type="checkbox"
          class="toggle"
          @change="emit('update:hideFromTimeline', ($event.target as HTMLInputElement).checked)"
        />
      </label>
    </div>

    <!-- Article View Mode -->
    <div class="p-3 rounded-lg bg-bg-secondary border border-border">
      <label class="block mb-1.5 font-semibold text-xs sm:text-sm text-text-primary">
        {{ t('setting.feed.articleViewMode') }}
      </label>
      <p class="text-[10px] sm:text-xs text-text-secondary mb-2">
        {{ t('setting.feed.articleViewModeDesc') }}
      </p>
      <BaseSelect
        :model-value="props.articleViewMode"
        :options="articleViewModeOptions"
        @update:model-value="handleArticleViewModeChange"
      />
    </div>

    <!-- Auto Expand Content -->
    <div class="p-3 rounded-lg bg-bg-secondary border border-border">
      <label class="block mb-1.5 font-semibold text-xs sm:text-sm text-text-primary">
        {{ t('setting.feed.autoExpandContent') }}
      </label>
      <p class="text-[10px] sm:text-xs text-text-secondary mb-2">
        {{ t('setting.feed.autoExpandContentDesc') }}
      </p>
      <BaseSelect
        :model-value="props.autoExpandContent"
        :options="autoExpandContentOptions"
        @update:model-value="handleAutoExpandContentChange"
      />
    </div>

    <!-- Proxy Settings -->
    <div class="p-3 rounded-lg bg-bg-secondary border border-border space-y-3">
      <div>
        <label class="block mb-1.5 font-semibold text-xs sm:text-sm text-text-primary">
          {{ t('modal.feed.proxy') }}
        </label>
        <p class="text-[10px] sm:text-xs text-text-secondary mb-2">
          {{ t('modal.feed.proxyDesc') }}
        </p>
        <BaseSelect
          :model-value="props.proxyMode"
          :options="proxyModeOptions"
          @update:model-value="handleProxyModeChange"
        />
      </div>

      <!-- Custom Proxy Configuration -->
      <div v-if="props.proxyMode === 'custom'" class="space-y-2.5 pl-3 border-l-2 border-accent/30">
        <!-- Proxy Type -->
        <div>
          <label class="block mb-1 text-[10px] sm:text-xs font-medium text-text-secondary">
            {{ t('modal.feed.proxyType') }}
          </label>
          <BaseSelect
            :model-value="props.proxyType"
            :options="proxyTypeOptions"
            size="sm"
            @update:model-value="handleProxyTypeChange"
          />
        </div>

        <!-- Proxy Host and Port -->
        <div class="grid grid-cols-3 gap-2">
          <div class="col-span-2">
            <label class="block mb-1 text-[10px] sm:text-xs font-medium text-text-secondary">
              {{ t('modal.feed.proxyHost') }} <span class="text-red-500">*</span>
            </label>
            <input
              :value="props.proxyHost"
              type="text"
              :placeholder="t('setting.network.proxyHostPlaceholder')"
              :class="[
                'input-field text-xs sm:text-sm',
                props.proxyMode === 'custom' && !props.proxyHost.trim() ? 'border-red-500' : '',
              ]"
              @input="emit('update:proxyHost', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="block mb-1 text-[10px] sm:text-xs font-medium text-text-secondary">
              {{ t('modal.feed.proxyPort') }} <span class="text-red-500">*</span>
            </label>
            <input
              :value="props.proxyPort"
              type="text"
              placeholder="8080"
              :class="[
                'input-field text-center text-xs sm:text-sm',
                props.proxyMode === 'custom' && !props.proxyPort.trim() ? 'border-red-500' : '',
              ]"
              @input="emit('update:proxyPort', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>

        <!-- Proxy Authentication (Optional) -->
        <div class="grid grid-cols-2 gap-2">
          <div>
            <label class="block mb-1 text-[10px] sm:text-xs font-medium text-text-secondary">
              {{ t('modal.feed.proxyUsername') }}
            </label>
            <input
              :value="props.proxyUsername"
              type="text"
              :placeholder="t('setting.network.proxyUsernamePlaceholder')"
              class="input-field text-xs sm:text-sm"
              @input="emit('update:proxyUsername', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="block mb-1 text-[10px] sm:text-xs font-medium text-text-secondary">
              {{ t('modal.feed.proxyPassword') }}
            </label>
            <input
              :value="props.proxyPassword"
              type="password"
              :placeholder="t('setting.network.proxyPasswordPlaceholder')"
              class="input-field text-xs sm:text-sm"
              @input="emit('update:proxyPassword', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Refresh Settings -->
    <div class="p-3 rounded-lg bg-bg-secondary border border-border space-y-3">
      <div>
        <label class="block mb-1.5 font-semibold text-xs sm:text-sm text-text-primary">
          {{ t('modal.feed.refreshMode') }}
        </label>
        <p class="text-[10px] sm:text-xs text-text-secondary mb-2">
          {{ t('modal.feed.refreshModeDesc') }}
        </p>
        <BaseSelect
          :model-value="props.refreshMode"
          :options="refreshModeOptions"
          :position="'auto'"
          @update:model-value="handleRefreshModeChange"
        />
      </div>

      <!-- Custom Refresh Interval -->
      <div v-if="props.refreshMode === 'custom'" class="pl-3 border-l-2 border-accent/30">
        <label class="block mb-1 text-[10px] sm:text-xs font-medium text-text-secondary">
          {{ t('modal.feed.refreshInterval') }}
        </label>
        <div class="flex items-center gap-2">
          <input
            :value="props.refreshInterval"
            type="number"
            min="5"
            max="1440"
            :placeholder="t('modal.feed.refreshIntervalPlaceholder')"
            class="input-field flex-1 text-xs sm:text-sm"
            @input="
              emit(
                'update:refreshInterval',
                parseInt(($event.target as HTMLInputElement).value) || 0
              )
            "
          />
          <span class="text-xs text-text-secondary shrink-0">{{
            t('common.time.minutesShort')
          }}</span>
        </div>
        <p class="text-[10px] text-text-secondary mt-1">
          {{ t('modal.feed.refreshIntervalDesc') }}
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply w-full p-2 sm:p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-xs sm:text-sm focus:border-accent focus:outline-none transition-colors;
}

.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}

.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}

.toggle:checked::after {
  transform: translateX(20px);
}
</style>
