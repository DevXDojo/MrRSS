<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import {
  PhMusicNotes,
  PhSpeakerHigh,
  PhPlay,
  PhPause,
  PhSpinner,
  PhRewind,
  PhFastForward,
} from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';

interface Props {
  audioUrl: string;
  articleTitle: string;
}

const props = defineProps<Props>();

const { t } = useI18n();

const audioRef = ref<HTMLAudioElement | null>(null);
const isPlaying = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const buffered = ref(0); // Buffered progress
const isLoading = ref(false); // Loading state
const hasLoadedMetadata = ref(false); // Metadata loaded state

// Local audio controls (not global settings)
const playbackSpeed = ref(1.0);
const volume = ref(1.0);

// Load metadata on mount to display duration immediately
onMounted(() => {
  if (audioRef.value) {
    // Load metadata to get duration without starting playback
    audioRef.value.load();
  }
});

// Speed options
const speedOptions = [0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0];
const currentSpeedIndex = ref(2); // Default to 1.0 (index 2)

// Show loading state
let loadingTimeout: number | null = null;
function showLoading() {
  if (loadingTimeout !== null) {
    clearTimeout(loadingTimeout);
  }
  // Show loading after a short delay to avoid flickering
  loadingTimeout = window.setTimeout(() => {
    isLoading.value = true;
  }, 200);
}

function hideLoading() {
  if (loadingTimeout !== null) {
    clearTimeout(loadingTimeout);
    loadingTimeout = null;
  }
  isLoading.value = false;
}

// Format time in MM:SS format
function formatTime(seconds: number): string {
  if (!isFinite(seconds)) return '0:00';
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

// Toggle play/pause
async function togglePlay() {
  if (!audioRef.value) return;

  if (isPlaying.value) {
    audioRef.value.pause();
  } else {
    // Show loading state immediately when trying to play
    showLoading();
    try {
      await audioRef.value.play();
    } catch (err) {
      console.error('[AudioPlayer] Failed to play audio:', err);
      hideLoading();
      window.showToast(t('article.audioPlayer.audioPlaybackError'), 'error');
    }
  }
}

// Handle audio events
function onPlay() {
  isPlaying.value = true;
  // Don't hide loading immediately, wait for actual audio playback
}

function onPause() {
  isPlaying.value = false;
  hideLoading();
}

function onTimeUpdate() {
  if (!audioRef.value) return;
  currentTime.value = audioRef.value.currentTime;
  updateBufferedProgress();
  // Hide loading when we're actually playing and making progress
  if (isLoading.value && isPlaying.value && currentTime.value > 0) {
    hideLoading();
  }
}

function onLoadedMetadata() {
  if (!audioRef.value) return;
  duration.value = audioRef.value.duration;
  hasLoadedMetadata.value = true;
  updateBufferedProgress();
}

function onEnded() {
  isPlaying.value = false;
  currentTime.value = 0;
  hideLoading();
}

function onWaiting() {
  // Browser is buffering data, show loading if we're playing
  if (isPlaying.value) {
    showLoading();
  }
}

function onCanPlay() {
  // Audio has enough data to start playing
  hasLoadedMetadata.value = true;
  // Always hide loading when audio can play
  // This ensures loading state is cleared once data is ready
  hideLoading();
}

// Handle when audio actually starts playing (fired when playback resumes)
function onPlaying() {
  hideLoading();
}

// Handle seeking start
function onSeeking() {
  // When user seeks, check if we need to show loading
  // Only show loading if we're currently playing
  if (audioRef.value && isPlaying.value) {
    const currentTime = audioRef.value.currentTime;
    const buffered = audioRef.value.buffered;
    let isBuffered = false;

    if (buffered.length > 0) {
      for (let i = 0; i < buffered.length; i++) {
        const start = buffered.start(i);
        const end = buffered.end(i);
        if (currentTime >= start && currentTime <= end) {
          isBuffered = true;
          break;
        }
      }
    }

    if (!isBuffered) {
      showLoading();
    }
  }
}

// Handle seek complete
function onSeeked() {
  // Seek is complete, hide loading if we're not playing
  // The canplay event will hide it when ready to play
  updateBufferedProgress();
  if (!isPlaying.value) {
    hideLoading();
  }
}

// Update buffered progress
function updateBufferedProgress() {
  if (!audioRef.value || !duration.value) {
    buffered.value = 0;
    return;
  }

  try {
    const audioBuffered = audioRef.value.buffered;
    if (audioBuffered && audioBuffered.length > 0) {
      const bufferedEnd = audioBuffered.end(audioBuffered.length - 1);
      buffered.value = (bufferedEnd / duration.value) * 100;
    } else {
      buffered.value = 0;
    }
  } catch {
    // If accessing buffered fails, just set to 0
    buffered.value = 0;
  }
}

// Watch for time updates to update buffer
watch(currentTime, () => {
  updateBufferedProgress();
});

// Handle dragging on progress bar
const isDragging = ref(false);
let progressBarRect: DOMRect | null = null;

function onProgressMouseDown(event: MouseEvent) {
  if (!audioRef.value) return;
  const progressBar = event.currentTarget as HTMLElement;
  isDragging.value = true;
  progressBarRect = progressBar.getBoundingClientRect();

  // Seek to initial position
  const newTime = calculateSeekPosition(event);
  seekToTime(newTime);

  const handleMouseMove = (e: MouseEvent) => {
    if (isDragging.value && progressBarRect) {
      const seekTime = calculateSeekPosition(e);
      seekToTime(seekTime);
    }
  };

  const handleMouseUp = () => {
    isDragging.value = false;
    progressBarRect = null;
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);
  };

  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', handleMouseUp);
}

// Calculate seek time from mouse event
function calculateSeekPosition(event: MouseEvent): number {
  if (!progressBarRect || !duration.value) return 0;
  const clickX = event.clientX - progressBarRect.left;
  const percentage = Math.max(0, Math.min(1, clickX / progressBarRect.width));
  return percentage * duration.value;
}

// Seek to specific time and handle loading state
function seekToTime(newTime: number) {
  if (!audioRef.value) return;

  // Show loading state if seeking to unbuffered region
  const buffered = audioRef.value.buffered;
  let isBuffered = false;
  if (buffered.length > 0) {
    for (let i = 0; i < buffered.length; i++) {
      const start = buffered.start(i);
      const end = buffered.end(i);
      if (newTime >= start && newTime <= end) {
        isBuffered = true;
        break;
      }
    }
  }

  // If seeking to unbuffered region, show loading state
  if (!isBuffered && isPlaying.value) {
    showLoading();
  }

  audioRef.value.currentTime = newTime;
}

// Computed progress percentage
const progressPercentage = computed(() => {
  if (!duration.value) return 0;
  return (currentTime.value / duration.value) * 100;
});

// Change playback speed
function cycleSpeed() {
  currentSpeedIndex.value = (currentSpeedIndex.value + 1) % speedOptions.length;
  playbackSpeed.value = speedOptions[currentSpeedIndex.value];
  if (audioRef.value) {
    audioRef.value.playbackRate = playbackSpeed.value;
  }
}

// Change volume
function onVolumeChange(event: Event) {
  const target = event.target as HTMLInputElement;
  volume.value = parseFloat(target.value);
  if (audioRef.value) {
    audioRef.value.volume = volume.value;
  }
}

// Skip backward 10 seconds
function skipBackward() {
  if (!audioRef.value) return;
  audioRef.value.currentTime = Math.max(0, audioRef.value.currentTime - 10);
}

// Skip forward 10 seconds
function skipForward() {
  if (!audioRef.value) return;
  audioRef.value.currentTime = Math.min(duration.value, audioRef.value.currentTime + 10);
}

// Extract filename from audio URL
const downloadFilename = computed(() => {
  try {
    const url = new URL(props.audioUrl);
    const pathname = url.pathname;
    const filename = pathname.substring(pathname.lastIndexOf('/') + 1);
    // If filename has no extension or is empty, use article title with .mp3
    if (!filename || !filename.includes('.')) {
      return `${props.articleTitle}.mp3`;
    }
    return filename;
  } catch {
    // Fallback if URL parsing fails
    return `${props.articleTitle}.mp3`;
  }
});
</script>

<template>
  <div class="bg-bg-secondary border border-border rounded-lg p-4 mb-4 sm:mb-6">
    <div class="flex items-center gap-3 mb-3">
      <PhMusicNotes :size="20" class="text-accent flex-shrink-0" />
      <span class="text-sm font-medium text-text-primary">{{
        t('article.audioPlayer.podcastAudio')
      }}</span>
    </div>

    <!-- Audio element (hidden) -->
    <audio
      ref="audioRef"
      :src="audioUrl"
      preload="metadata"
      @play="onPlay"
      @pause="onPause"
      @playing="onPlaying"
      @seeking="onSeeking"
      @seeked="onSeeked"
      @timeupdate="onTimeUpdate"
      @loadedmetadata="onLoadedMetadata"
      @ended="onEnded"
      @waiting="onWaiting"
      @canplay="onCanPlay"
      @progress="updateBufferedProgress"
    />

    <!-- Custom audio controls -->
    <div class="space-y-3">
      <!-- Progress bar row -->
      <div class="flex items-center gap-3">
        <!-- Skip backward button -->
        <button
          class="flex items-center justify-center w-8 h-8 rounded-full bg-bg-tertiary hover:bg-bg-hover transition-colors flex-shrink-0"
          :title="t('article.audioPlayer.skipBackward')"
          @click="skipBackward"
        >
          <PhRewind :size="16" class="text-text-primary" />
        </button>

        <!-- Play/Pause button -->
        <button
          class="flex items-center justify-center w-10 h-10 rounded-full bg-accent hover:bg-accent/90 transition-colors flex-shrink-0 relative"
          :title="isPlaying ? t('article.audioPlayer.pause') : t('article.audioPlayer.play')"
          @click="togglePlay"
        >
          <PhSpinner v-if="isLoading" :size="20" class="text-white animate-spin" />
          <PhPlay v-else-if="!isPlaying" :size="20" class="text-white ml-0.5" />
          <PhPause v-else :size="20" class="text-white" />
        </button>

        <!-- Skip forward button -->
        <button
          class="flex items-center justify-center w-8 h-8 rounded-full bg-bg-tertiary hover:bg-bg-hover transition-colors flex-shrink-0"
          :title="t('article.audioPlayer.skipForward')"
          @click="skipForward"
        >
          <PhFastForward :size="16" class="text-text-primary" />
        </button>

        <!-- Progress bar -->
        <div class="flex-1 flex items-center gap-2">
          <span class="text-xs text-text-secondary min-w-[40px] text-right">{{
            formatTime(currentTime)
          }}</span>
          <div
            class="flex-1 h-2 bg-bg-tertiary rounded-full cursor-pointer relative group"
            @mousedown="onProgressMouseDown"
          >
            <!-- Buffered progress -->
            <div
              class="absolute top-0 left-0 h-full bg-bg-hover rounded-full transition-all duration-300"
              :style="{ width: `${Math.min(buffered, 100)}%` }"
            />
            <!-- Played progress -->
            <div
              class="absolute top-0 left-0 h-full bg-accent rounded-full transition-all duration-75"
              :style="{ width: `${progressPercentage}%` }"
            />
            <!-- Draggable thumb (visible on hover and during drag) -->
            <div
              class="absolute top-1/2 -translate-y-1/2 w-3 h-3 bg-accent rounded-full shadow-lg opacity-0 group-hover:opacity-100 transition-opacity duration-200"
              :class="{ 'opacity-100': isDragging }"
              :style="{ left: `calc(${progressPercentage}% - 6px)` }"
            />
            <!-- Loading text indicator -->
            <span
              v-if="isLoading"
              class="absolute left-1/2 -translate-x-1/2 top-1/2 -translate-y-1/2 text-[10px] text-text-secondary font-medium px-2 py-0.5 bg-bg-tertiary/95 rounded-full backdrop-blur-sm whitespace-nowrap z-10"
            >
              {{ t('common.pagination.loading') }}...
            </span>
          </div>
          <span class="text-xs text-text-secondary min-w-[40px]">{{ formatTime(duration) }}</span>
        </div>
      </div>

      <!-- Download and controls row -->
      <div class="flex items-center justify-between pt-3 border-t border-border">
        <!-- Download link -->
        <a
          :href="audioUrl"
          :download="downloadFilename"
          class="text-xs text-accent hover:underline flex items-center gap-1"
          target="_blank"
        >
          {{ t('common.contextMenu.downloadAudio') }}
        </a>

        <!-- Controls -->
        <div class="flex items-center gap-3">
          <!-- Volume control -->
          <div class="flex items-center gap-1.5">
            <PhSpeakerHigh :size="14" class="text-text-secondary flex-shrink-0" />
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              :value="volume"
              class="w-20 h-1.5 bg-bg-tertiary rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-accent [&::-webkit-slider-thumb]:cursor-pointer [&::-webkit-slider-thumb]:transition-all [&::-webkit-slider-thumb]:hover:scale-125"
              :title="t('article.audioPlayer.volume')"
              @input="onVolumeChange"
            />
            <span class="text-xs text-text-secondary w-[35px] text-right"
              >{{ Math.round(volume * 100) }}%</span
            >
          </div>

          <!-- Playback speed control -->
          <button
            class="flex items-center justify-center px-3 py-1 rounded-md bg-bg-tertiary hover:bg-bg-tertiary/80 transition-colors text-xs font-medium text-text-primary w-[52px]"
            :title="t('article.audioPlayer.playbackSpeed')"
            @click="cycleSpeed"
          >
            <span class="tabular-nums">{{ playbackSpeed }}x</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
