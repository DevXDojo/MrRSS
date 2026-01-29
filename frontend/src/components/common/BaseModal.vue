<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useModalClose } from '@/composables/ui/useModalClose';

interface Props {
  // Title
  title?: string;
  // Size presets
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl' | 'full';
  // Custom width (overrides size)
  width?: string;
  // Custom max width
  maxWidth?: string;
  // Height control
  height?: 'auto' | 'full' | string;
  // Maximum height (default: 90vh)
  maxHeight?: string;
  // z-index (default: 50)
  zIndex?: number;
  // Show close button in header (default: true)
  closable?: boolean;
  // Close on click outside (default: true)
  closeOnClickOutside?: boolean;
  // Show footer (default: false)
  showFooter?: boolean;
  // Custom classes
  containerClass?: string;
  headerClass?: string;
  bodyClass?: string;
  footerClass?: string;
  // Animation
  animation?: boolean;
  // Loading state
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  size: 'md',
  width: undefined,
  maxWidth: undefined,
  height: 'auto',
  maxHeight: '90vh',
  zIndex: 50,
  closable: true,
  closeOnClickOutside: true,
  showFooter: false,
  containerClass: '',
  headerClass: '',
  bodyClass: '',
  footerClass: '',
  animation: true,
  loading: false,
});

const emit = defineEmits<{
  close: [];
}>();

// Modal close handling
useModalClose(() => handleClose(), props.zIndex);

// Track if modal is open
const isOpen = ref(true);

// Computed size classes
const sizeClasses = computed(() => {
  if (props.width) return '';

  const sizeMap = {
    xs: 'max-w-xs',
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
    '2xl': 'max-w-2xl',
    '3xl': 'max-w-3xl',
    '4xl': 'max-w-4xl',
    full: 'max-w-full',
  };

  return sizeMap[props.size];
});

// Computed container styles
const containerStyles = computed(() => {
  const styles: Record<string, string> = {};

  // Handle custom width
  if (props.width) {
    styles.width = props.width;
  }

  // Handle custom maxWidth
  if (props.maxWidth) {
    styles.maxWidth = props.maxWidth;
  }

  // Handle height - for 'full' use 100% on mobile, auto with max-height on desktop
  if (props.height === 'full') {
    styles.height = '100vh';
    // Add max-height for desktop
    if (props.maxHeight) {
      styles.maxHeight = props.maxHeight;
    } else {
      styles.maxHeight = '90vh';
    }
  } else if (props.height && props.height !== 'auto') {
    styles.height = props.height;
  }

  // Handle custom maxHeight (only apply if explicitly provided)
  if (props.maxHeight) {
    styles.maxHeight = props.maxHeight;
  }

  return styles;
});

// Computed container classes
const containerSizeClasses = computed(() => {
  const classes = ['w-full', 'mx-2', 'sm:mx-4'];

  if (!props.width && !props.maxWidth) {
    classes.push(sizeClasses.value);
  }

  // Handle height class
  if (props.height === 'full') {
    classes.push('h-full');
  }

  return classes;
});

// Handle close
function handleClose() {
  if (props.loading) return;
  isOpen.value = false;
  emit('close');
}

// Handle click outside
function handleBackdropClick(event: MouseEvent) {
  if (props.closeOnClickOutside && !props.loading) {
    // Check if click was on the backdrop (not on modal content)
    if ((event.target as HTMLElement).dataset.modalBackdrop === 'true') {
      handleClose();
    }
  }
}

// Handle ESC key from useModalClose
// This is automatically handled by useModalClose composable

// Prevent body scroll when modal is open
onMounted(() => {
  document.body.style.overflow = 'hidden';
});

onUnmounted(() => {
  document.body.style.overflow = '';
});
</script>

<template>
  <div
    :class="[
      'fixed inset-0 flex items-center justify-center bg-black/50 backdrop-blur-sm p-2 sm:p-4',
      animation ? 'animate-fade-in' : '',
    ]"
    :style="{ zIndex }"
    :data-modal-backdrop="closeOnClickOutside ? 'true' : undefined"
    :data-modal-open="isOpen ? 'true' : undefined"
    @click.self="handleBackdropClick"
  >
    <div
      :class="[
        'bg-bg-primary rounded-none sm:rounded-2xl shadow-2xl border border-border overflow-hidden flex flex-col',
        containerSizeClasses,
        containerClass,
      ]"
      :style="containerStyles"
      @click.stop
    >
      <!-- Header -->
      <div
        v-if="title || $slots.header || closable"
        :class="[
          'p-3 sm:p-5 border-b border-border flex justify-between items-center shrink-0',
          headerClass,
        ]"
      >
        <div v-if="title || $slots.header">
          <slot name="header">
            <h3 class="text-base sm:text-lg font-semibold m-0">{{ title }}</h3>
          </slot>
        </div>
        <button
          v-if="closable && !loading"
          class="text-text-secondary hover:text-text-primary transition-colors text-2xl cursor-pointer"
          :disabled="loading"
          @click="handleClose"
        >
          &times;
        </button>
      </div>

      <!-- Body -->
      <div
        :class="['flex-1 overflow-y-auto', title || $slots.header ? '' : 'p-4 sm:p-6', bodyClass]"
      >
        <slot></slot>
      </div>

      <!-- Footer -->
      <div
        v-if="showFooter || $slots.footer"
        :class="[
          'p-3 sm:p-5 border-t border-border bg-bg-secondary shrink-0 rounded-none sm:rounded-b-2xl',
          footerClass,
        ]"
      >
        <slot name="footer"></slot>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

/* Desktop max-height for rounded modals */
@media (min-width: 640px) {
  .rounded-2xl {
    max-height: 90vh;
  }
}

/* Ensure proper scrolling within modal body */
.overflow-y-auto {
  max-height: calc(90vh - 120px); /* Adjust based on header/footer */
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
