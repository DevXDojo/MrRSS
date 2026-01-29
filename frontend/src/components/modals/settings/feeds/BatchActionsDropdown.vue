<script setup lang="ts">
import { ref, type Ref } from 'vue';
import { PhDotsThree, PhFolder, PhTag, PhImage } from '@phosphor-icons/vue';

interface Props {
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
});

const emit = defineEmits<{
  move: [];
  'add-tags': [];
  'set-image-mode': [];
  'unset-image-mode': [];
}>();

const showMenu = ref(false);
const menuRef: Ref<HTMLDivElement | null> = ref(null);

function toggleMenu() {
  if (props.disabled) return;
  showMenu.value = !showMenu.value;
}

function handleAction(action: string) {
  showMenu.value = false;
  switch (action) {
    case 'move':
      emit('move');
      break;
    case 'add-tags':
      emit('add-tags');
      break;
    case 'set-image-mode':
      emit('set-image-mode');
      break;
    // case 'unset-image-mode':
    //   emit('unset-image-mode');
    //   break;
    default:
      // Handle unrecognized action silently
      break;
  }
}

function handleClickOutside(event: MouseEvent) {
  if (menuRef.value && event.target instanceof Node && !menuRef.value.contains(event.target)) {
    showMenu.value = false;
  }
}

// Add click outside listener when menu is shown
import { watch, onUnmounted } from 'vue';

watch(showMenu, (newValue) => {
  if (newValue) {
    setTimeout(() => {
      document.addEventListener('click', handleClickOutside);
    }, 0);
  } else {
    document.removeEventListener('click', handleClickOutside);
  }
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<template>
  <div ref="menuRef" class="relative">
    <button class="batch-more-btn" :disabled="disabled" @click="toggleMenu">
      <PhDotsThree :size="16" class="sm:w-5 sm:h-5" />
    </button>

    <!-- Dropdown Menu -->
    <div
      v-if="showMenu"
      class="absolute left-0 top-full mt-1 bg-bg-primary border border-border rounded-lg shadow-xl py-1 min-w-[200px] z-50 animate-fade-in"
    >
      <!-- Move to Category -->
      <div
        class="px-4 py-2 flex items-center gap-3 cursor-pointer hover:bg-bg-tertiary text-sm transition-colors text-text-primary"
        @click="handleAction('move')"
      >
        <PhFolder :size="18" class="text-text-secondary" />
        <span>{{ $t('common.action.moveSelected') }}</span>
      </div>

      <!-- Add Tags -->
      <div
        class="px-4 py-2 flex items-center gap-3 cursor-pointer hover:bg-bg-tertiary text-sm transition-colors text-text-primary"
        @click="handleAction('add-tags')"
      >
        <PhTag :size="18" class="text-text-secondary" />
        <span>{{ $t('common.action.addTags') }}</span>
      </div>

      <div class="h-px bg-border my-1"></div>

      <!-- Set Image Mode -->
      <div
        class="px-4 py-2 flex items-center gap-3 cursor-pointer hover:bg-bg-tertiary text-sm transition-colors text-text-primary"
        @click="handleAction('set-image-mode')"
      >
        <PhImage :size="18" class="text-text-secondary" />
        <span>{{ $t('common.action.setImageMode') }}</span>
      </div>

      <!-- Unset Image Mode (Commented out but preserved for future use) -->
      <!--
      <div
        class="px-4 py-2 flex items-center gap-3 cursor-pointer hover:bg-bg-tertiary text-sm transition-colors text-text-primary"
        @click="handleAction('unset-image-mode')"
      >
        <PhProhibit :size="18" class="text-text-secondary" />
        <span>{{ $t('common.action.unsetImageMode') }}</span>
      </div>
      -->
    </div>
  </div>
</template>

<style scoped>
.batch-more-btn {
  @apply py-2.5 px-3 sm:py-3 sm:px-4 rounded-md flex items-center justify-center gap-1.5 sm:gap-2 font-medium transition-colors shrink-0 bg-bg-tertiary border border-border text-text-primary hover:bg-bg-secondary cursor-pointer;
}

.batch-more-btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.animate-fade-in {
  animation: fadeIn 0.1s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
