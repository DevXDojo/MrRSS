<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { PhYoutubeLogo, PhPlayCircle } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';
import { isYouTubeUrl } from '@/utils/youtube';
import { isBilibiliUrl } from '@/utils/bilibili';

interface Props {
  videoUrl: string;
  articleTitle: string;
}

const props = defineProps<Props>();

const { t } = useI18n();

const iframeRef = ref<HTMLIFrameElement | null>(null);
const isLoading = ref(true);

// Check if this is a YouTube video
const isYouTube = computed(() => isYouTubeUrl(props.videoUrl));

// Check if this is a Bilibili video
const isBilibili = computed(() => isBilibiliUrl(props.videoUrl));

// Get video platform name
const videoPlatform = computed(() => {
  if (isYouTube.value) return 'YouTube';
  if (isBilibili.value) return 'Bilibili';
  return 'Video';
});

// Get video platform icon color
const iconColor = computed(() => {
  if (isYouTube.value) return 'text-red-600';
  if (isBilibili.value) return 'text-pink-600';
  return 'text-accent';
});

function onLoad() {
  isLoading.value = false;
}

function onError() {
  isLoading.value = false;
  window.showToast(t('article.videoPlayer.videoLoadError'), 'error');
}

// Open video in new tab
function openInNewTab() {
  if (isYouTube.value) {
    // Convert embed URL back to watch URL
    const watchURL = props.videoUrl.replace('/embed/', '/watch?v=');
    window.open(watchURL, '_blank');
  } else if (isBilibili.value) {
    // For Bilibili, the video_url is already the iframe URL
    // Extract BVID and open the watch page
    const bvidMatch = props.videoUrl.match(/bvid=([A-Za-z0-9]+)/);
    if (bvidMatch && bvidMatch[1]) {
      window.open(`https://www.bilibili.com/video/${bvidMatch[1]}`, '_blank');
    } else {
      // Fallback: open the video_url as-is
      window.open(props.videoUrl, '_blank');
    }
  } else {
    window.open(props.videoUrl, '_blank');
  }
}

onMounted(() => {
  if (iframeRef.value) {
    iframeRef.value.addEventListener('load', onLoad);
    iframeRef.value.addEventListener('error', onError);
  }
});

onUnmounted(() => {
  if (iframeRef.value) {
    iframeRef.value.removeEventListener('load', onLoad);
    iframeRef.value.removeEventListener('error', onError);
  }
});
</script>

<template>
  <div class="bg-bg-secondary border border-border rounded-lg overflow-hidden mb-4 sm:mb-6">
    <!-- Header -->
    <div class="flex items-center justify-between p-3 border-b border-border">
      <div class="flex items-center gap-2">
        <PhYoutubeLogo v-if="isYouTube" :size="20" :class="iconColor + ' flex-shrink-0'" />
        <PhPlayCircle v-else-if="isBilibili" :size="20" :class="iconColor + ' flex-shrink-0'" />
        <span class="text-sm font-medium text-text-primary">{{
          t('article.videoPlayer.videoPlayer', { platform: videoPlatform })
        }}</span>
      </div>
      <button
        class="text-xs text-accent hover:underline"
        :title="t('article.videoPlayer.openInPlatform', { platform: videoPlatform })"
        @click="openInNewTab"
      >
        {{ t('article.videoPlayer.openInPlatform', { platform: videoPlatform }) }}
      </button>
    </div>

    <!-- Video Player -->
    <div class="relative w-full" style="padding-bottom: 56.25%">
      <!-- 16:9 Aspect Ratio -->
      <iframe
        ref="iframeRef"
        :src="videoUrl"
        :title="articleTitle"
        class="absolute top-0 left-0 w-full h-full border-none"
        allow="
          accelerometer;
          autoplay;
          clipboard-write;
          encrypted-media;
          gyroscope;
          picture-in-picture;
          web-share;
        "
        allowfullscreen
        :sandbox="
          isBilibili ? 'allow-forms allow-scripts allow-same-origin allow-presentation' : undefined
        "
      />

      <!-- Loading indicator -->
      <div
        v-if="isLoading"
        class="absolute inset-0 flex items-center justify-center bg-bg-tertiary"
      >
        <div
          class="animate-spin rounded-full h-12 w-12 border-4 border-accent border-t-transparent"
        ></div>
      </div>
    </div>
  </div>
</template>
