<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhSparkle, PhArrowUpRight, PhTicket, PhX } from '@phosphor-icons/vue';
import {
  advertisementText,
  useAdvertisementsStore,
  type Advertisement,
} from '@/stores/advertisements';
import { openInBrowser } from '@/utils/browser';

defineProps<{ ad: Advertisement; popup?: boolean }>();
const store = useAdvertisementsStore();
const { t, locale } = useI18n();

async function snooze() {
  if (!(await store.snooze())) {
    window.showToast?.(t('setting.advertisement.saveFailed'), 'error');
  }
}
</script>

<template>
  <aside
    class="partner-ad relative overflow-hidden rounded-2xl border p-4 sm:p-5 text-text-primary shadow-lg"
    :aria-label="t('setting.advertisement.label')"
  >
    <div class="relative flex items-start justify-between gap-3">
      <span
        class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-accent"
      >
        <PhSparkle :size="18" weight="fill" />
        {{ t('setting.advertisement.label') }}
      </span>
      <button
        v-if="popup"
        type="button"
        class="shrink-0 cursor-pointer rounded-lg p-1 text-text-secondary hover:bg-bg-tertiary focus-visible:outline-accent"
        :aria-label="t('common.close')"
        @click="store.popupDismissed = true"
      >
        <PhX :size="18" />
      </button>
    </div>
    <h3 class="relative mt-3 mb-2 break-words text-lg font-bold leading-snug">
      {{ advertisementText(ad.title, locale) }}
    </h3>
    <p
      class="relative m-0 whitespace-pre-line break-words text-sm leading-relaxed text-text-secondary"
      :class="{ 'line-clamp-3': popup }"
    >
      {{ advertisementText(ad.description, locale) }}
    </p>
    <div
      v-if="!popup && ad.blocks?.length"
      class="relative mt-3 space-y-2 text-sm text-text-secondary"
    >
      <template v-for="(block, index) in ad.blocks" :key="index">
        <ul v-if="block.type === 'list'" class="m-0 list-disc space-y-1 pl-5">
          <li v-for="(item, itemIndex) in block.items" :key="itemIndex" class="break-words">
            {{ advertisementText(item, locale) }}
          </li>
        </ul>
        <p
          v-else-if="block.type === 'text' || block.type === 'callout'"
          class="m-0 whitespace-pre-line break-words"
          :class="{
            'rounded-lg border border-border bg-bg-secondary p-3': block.type === 'callout',
          }"
        >
          {{ advertisementText(block.text, locale) }}
        </p>
      </template>
    </div>
    <div
      v-if="ad.coupon_code"
      class="relative mt-4 flex flex-wrap items-center gap-2 rounded-xl border border-dashed border-accent p-3 text-sm"
    >
      <PhTicket :size="20" class="shrink-0 text-accent" />
      <span>{{ t('setting.advertisement.coupon') }}</span>
      <code class="select-all break-all font-bold text-accent">{{ ad.coupon_code }}</code>
      <span class="w-full text-xs text-text-secondary">{{
        t('setting.advertisement.discount')
      }}</span>
    </div>
    <div class="relative mt-4 flex flex-wrap items-center gap-3">
      <button
        type="button"
        class="flex cursor-pointer items-center gap-2 rounded-xl bg-accent px-4 py-2 text-sm font-semibold text-white transition-opacity hover:opacity-90 focus-visible:outline-accent"
        @click="openInBrowser(ad.url)"
      >
        {{ t('setting.advertisement.visit') }} <PhArrowUpRight :size="16" />
      </button>
      <button
        type="button"
        class="cursor-pointer rounded-lg py-2 text-xs text-text-secondary hover:text-text-primary focus-visible:outline-accent disabled:opacity-50"
        :disabled="store.saving"
        @click="snooze"
      >
        {{ t('setting.advertisement.snooze') }}
      </button>
    </div>
  </aside>
</template>

<style scoped>
.partner-ad {
  border-color: color-mix(in srgb, var(--accent-color) 45%, var(--border-color));
  background:
    radial-gradient(
      ellipse at top right,
      color-mix(in srgb, var(--accent-color) 22%, transparent),
      transparent 70%
    ),
    var(--bg-secondary);
  box-shadow: 0 8px 32px color-mix(in srgb, var(--accent-color) 16%, transparent);
}
</style>
