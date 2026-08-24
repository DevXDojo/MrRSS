<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { PhCheckCircle, PhXCircle, PhWarning, PhInfo, PhX } from '@phosphor-icons/vue';

type ToastType = 'info' | 'success' | 'error' | 'warning';

interface Props {
  message: string;
  type?: ToastType;
  duration?: number;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  duration: 3000,
});

const emit = defineEmits<{
  close: [];
}>();

const show = ref(true);

onMounted(() => {
  if (props.duration > 0) {
    setTimeout(() => {
      show.value = false;
      setTimeout(() => emit('close'), 300);
    }, props.duration);
  }
});

function handleClose() {
  show.value = false;
  setTimeout(() => emit('close'), 300);
}
</script>

<template>
  <div v-if="show" :class="['toast', `toast-${type}`, show ? 'toast-show' : 'toast-hide']">
    <div class="flex items-center gap-3">
      <PhCheckCircle v-if="type === 'success'" :size="20" />
      <PhXCircle v-else-if="type === 'error'" :size="20" />
      <PhWarning v-else-if="type === 'warning'" :size="20" />
      <PhInfo v-else :size="20" />
      <span class="flex-1 toast-message selectable-text">{{ message }}</span>
      <button class="text-xl opacity-70 hover:opacity-100 transition-opacity" @click="handleClose">
        <PhX :size="20" />
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";
.toast {
  @apply z-[60] px-5 py-3 rounded-lg shadow-lg border min-w-[300px] max-w-md;
}
.toast-show {
  animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-hide {
  animation: slideOut 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-info {
  @apply bg-blue-50 border-blue-200 text-blue-900;
}
:global(.dark-mode) .toast-info {
  @apply bg-blue-950 border-blue-700 text-blue-100;
}
.toast-success {
  @apply bg-green-50 border-green-200 text-green-900;
}
:global(.dark-mode) .toast-success {
  @apply bg-green-950 border-green-700 text-green-100;
}
.toast-error {
  @apply bg-red-50 border-red-200 text-red-900;
}
:global(.dark-mode) .toast-error {
  @apply bg-red-950 border-red-700 text-red-100;
}
.toast-warning {
  @apply bg-orange-50 border-orange-200 text-orange-900;
}
:global(.dark-mode) .toast-warning {
  @apply bg-orange-950 border-orange-700 text-orange-100;
}
.toast-message {
  user-select: text;
  -webkit-user-select: text;
  -moz-user-select: text;
  -ms-user-select: text;
  cursor: text;
}
@keyframes slideIn {
  from {
    transform: translateY(-20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
@keyframes slideOut {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-20px);
    opacity: 0;
  }
}
</style>
