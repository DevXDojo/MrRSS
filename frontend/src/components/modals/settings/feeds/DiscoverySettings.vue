<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhPlay, PhBinoculars } from '@phosphor-icons/vue';
import { ButtonControl, TipBox, SettingGroup, SettingItem } from '@/components/settings';

const { t } = useI18n();

const emit = defineEmits<{
  'discover-all': [];
}>();

function handleDiscoverAll() {
  emit('discover-all');
}
</script>

<template>
  <SettingGroup :icon="PhBinoculars" :title="t('modal.feed.feedDiscovery')">
    <TipBox type="info" :title="t('modal.discovery.discoveryLongRunningWarning')" />

    <!-- Discover All Feeds -->
    <SettingItem
      :icon="PhBinoculars"
      :title="t('modal.discovery.discoverAllFeeds')"
      :description="t('modal.discovery.discoverAllFeedsDesc')"
    >
      <ButtonControl
        :label="t('modal.discovery.startDiscovery')"
        :icon="PhPlay"
        type="secondary"
        @click="handleDiscoverAll"
      />
    </SettingItem>
  </SettingGroup>
</template>

<style scoped></style>
