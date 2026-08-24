import { ref, computed, onMounted, onUnmounted, type Ref, type CSSProperties } from 'vue';
import type { SelectOption } from '@/types/select';

// Global state to track currently open dropdown
let openDropdownId: string | null = null;
let registerOpenDropdown: ((id: string | null) => void) | null = null;

interface UseSelectOptions {
  options: SelectOption[];
  isOpen: Ref<boolean>;
  dropdownId?: string;
  position?: Ref<'bottom' | 'top' | 'auto'>;
  width?: Ref<string>;
  maxWidth?: Ref<string>;
  maxHeight?: Ref<string>;
  // User-provided desired max height (can be CSS value like 'max-h-60')
  desiredMaxHeight?: Ref<string>;
}

interface DropdownPositionStyle extends CSSProperties {}

export function useSelect(options: UseSelectOptions) {
  const selectedIndex = ref(-1);
  const triggerRef = ref<HTMLElement>();
  const dropdownRef = ref<HTMLElement>();
  const shouldTeleport = ref(false); // Whether to teleport to body
  const dropdownPositionStyle = ref<CSSProperties>({
    position: 'absolute',
    left: '0px',
    top: '0px',
    width: '0px',
    zIndex: '50',
  });

  // Reset selected index
  function resetIndex() {
    selectedIndex.value = -1;
  }

  // Register this dropdown as the open one
  function registerAsOpen() {
    if (options.dropdownId) {
      openDropdownId = options.dropdownId;
      // Notify other dropdowns to close
      if (registerOpenDropdown) {
        registerOpenDropdown(options.dropdownId);
      }
    }
  }

  // Unregister this dropdown
  function unregisterAsOpen() {
    if (options.dropdownId && openDropdownId === options.dropdownId) {
      openDropdownId = null;
      if (registerOpenDropdown) {
        registerOpenDropdown(null);
      }
    }
  }

  // Handle other dropdown opening
  function onOtherDropdownOpen(id: string | null) {
    if (options.dropdownId && id !== options.dropdownId) {
      options.isOpen.value = false;
      resetIndex();
    }
  }

  // Click outside to close
  function handleClickOutside(event: MouseEvent) {
    if (
      options.isOpen.value &&
      triggerRef.value &&
      dropdownRef.value &&
      !triggerRef.value.contains(event.target as Node) &&
      !dropdownRef.value.contains(event.target as Node)
    ) {
      options.isOpen.value = false;
      resetIndex();
      unregisterAsOpen();
    }
  }

  // Close dropdown when context menu is triggered
  function handleContextMenu(event: Event) {
    if (options.isOpen.value) {
      // Check if context menu is triggered outside the dropdown
      if (
        triggerRef.value &&
        dropdownRef.value &&
        !triggerRef.value.contains(event.target as Node) &&
        !dropdownRef.value.contains(event.target as Node)
      ) {
        options.isOpen.value = false;
        resetIndex();
        unregisterAsOpen();
      }
    }
  }

  // Close dropdown on mousedown (before click, works better with modals)
  function handleMouseDown(event: MouseEvent) {
    if (
      options.isOpen.value &&
      triggerRef.value &&
      dropdownRef.value &&
      !triggerRef.value.contains(event.target as Node) &&
      !dropdownRef.value.contains(event.target as Node)
    ) {
      options.isOpen.value = false;
      resetIndex();
      unregisterAsOpen();
    }
  }

  // Mouse leaves dropdown area - reset hover state
  function handleMouseLeave() {
    if (options.isOpen.value) {
      selectedIndex.value = -1;
    }
  }

  // Parse CSS max-height value to pixels
  function parseMaxHeight(value: string | undefined): number {
    if (!value) return 240; // default

    // Handle Tailwind classes like 'max-h-60'
    const tailwindMatch = value.match(/max-h-(\d+)/);
    if (tailwindMatch) {
      const num = parseInt(tailwindMatch[1], 10);
      return num * 4; // Tailwind spacing: 1 = 0.25rem = 4px (assuming base 16px)
    }

    // Handle pixel values like '240px'
    const pixelMatch = value.match(/(\d+)px/);
    if (pixelMatch) {
      return parseInt(pixelMatch[1], 10);
    }

    // Handle rem values like '15rem'
    const remMatch = value.match(/([\d.]+)rem/);
    if (remMatch) {
      return parseFloat(remMatch[1]) * 16;
    }

    return 240; // default
  }

  // Find the scrollable container and modal boundaries
  function findContainerBounds() {
    if (!triggerRef.value) return null;

    let element: HTMLElement | null = triggerRef.value;
    let scrollableContainer: HTMLElement | null = null;
    let modalContent: HTMLElement | null = null;
    let useTeleport = false;

    // Walk up the DOM tree to find relevant containers
    while (element && element !== document.body) {
      // Check if this is a scrollable container
      const style = window.getComputedStyle(element);
      const isScrollable =
        (style.overflowY === 'auto' || style.overflowY === 'scroll' || style.overflow === 'auto') &&
        element.scrollHeight > element.clientHeight;

      // Check if this is inside a modal (has data-modal-open attribute)
      const isInModal = element.closest('[data-modal-open]') !== null;

      if (isScrollable && !scrollableContainer) {
        scrollableContainer = element;
      }

      // For settings modal - has header but we should not teleport
      const settingsModal = element.closest('[data-settings-modal]');
      if (settingsModal) {
        // The scrollable content in settings modal
        const contentArea = settingsModal.querySelector('.overflow-y-scroll');
        if (contentArea && contentArea.contains(triggerRef.value)) {
          scrollableContainer = contentArea as HTMLElement;
          modalContent = settingsModal as HTMLElement;
          // Don't teleport - stay in modal's stacking context
          useTeleport = false;
          break;
        }
      }

      // For BaseModal with header/footer - don't teleport to stay in stacking context
      if (isInModal && !modalContent) {
        modalContent = element.closest('[data-modal-open]') as HTMLElement;
        const modalBody = modalContent.querySelector('.overflow-y-scroll');
        if (modalBody && modalBody.contains(triggerRef.value)) {
          scrollableContainer = modalBody as HTMLElement;
          // Don't teleport - this keeps dropdown within modal's stacking context
          useTeleport = false;
          break;
        }
      }

      element = element.parentElement;
    }

    return { scrollableContainer, modalContent, useTeleport };
  }

  // Calculate dropdown position when opened
  function updateDropdownPosition() {
    if (!options.isOpen.value || !triggerRef.value) return;

    const triggerRect = triggerRef.value.getBoundingClientRect();
    const viewportHeight = window.innerHeight;
    const viewportWidth = window.innerWidth;

    // Get desired max height from user or use default
    const desiredMaxHeight = parseMaxHeight(options.desiredMaxHeight?.value);
    const margin = 4; // margin between trigger and dropdown

    // Find container bounds
    const bounds = findContainerBounds();

    // Update teleport flag
    shouldTeleport.value = bounds?.useTeleport || false;

    let position: 'top' | 'bottom' = options.position?.value || 'bottom';

    // Calculate available space considering containers
    let spaceBelow: number;
    let spaceAbove: number;

    if (bounds?.scrollableContainer) {
      const containerRect = bounds.scrollableContainer.getBoundingClientRect();
      // Space within the scrollable container
      spaceBelow = containerRect.bottom - triggerRect.bottom - margin;
      spaceAbove = triggerRect.top - containerRect.top - margin;
    } else {
      // Fall back to viewport
      spaceBelow = viewportHeight - triggerRect.bottom - margin;
      spaceAbove = triggerRect.top - margin;
    }

    // Auto-detect position based on available space
    if (options.position?.value === 'auto') {
      position = spaceAbove > spaceBelow ? 'top' : 'bottom';
    }

    // Calculate max height based on available space and desired max height
    let maxHeight = Math.min(desiredMaxHeight, position === 'bottom' ? spaceBelow : spaceAbove);

    // Ensure minimum height for usability
    const minHeight = 100;
    if (maxHeight < minHeight) {
      maxHeight = minHeight;
    }

    const width = triggerRect.width;

    // Use absolute positioning (within modal's stacking context)
    dropdownPositionStyle.value = {
      position: 'absolute',
      left: '0',
      width: '100%',
      maxHeight: `${maxHeight}px`,
      zIndex: '50',
    };

    // Set top/bottom positioning using margin
    if (position === 'bottom') {
      dropdownPositionStyle.value.top = '100%';
      dropdownPositionStyle.value.marginTop = `${margin}px`;
    } else {
      // Position above trigger - use negative top margin
      dropdownPositionStyle.value.bottom = '100%';
      dropdownPositionStyle.value.marginBottom = `${margin}px`;
      dropdownPositionStyle.value.top = 'auto';
    }
  }

  // Width class
  const widthClass = computed(() => {
    if (options.width?.value) {
      switch (options.width.value) {
        case 'sm':
          return 'w-20 sm:w-24';
        case 'md':
          return 'w-32 sm:w-48';
        case 'lg':
          return 'w-48 sm:w-64';
        default:
          return options.width.value;
      }
    }
    return 'w-full';
  });

  // Max width style
  const maxWidthStyle = computed(() => {
    if (options.maxWidth?.value) {
      return { maxWidth: options.maxWidth.value };
    }
    return {};
  });

  // Max height style for dropdown
  const maxHeightStyle = computed(() => {
    if (options.maxHeight?.value) {
      return { maxHeight: options.maxHeight.value };
    }
    return {};
  });

  // Handle scroll to update position
  function handleScroll() {
    if (options.isOpen.value) {
      updateDropdownPosition();
    }
  }

  // Handle window resize
  function handleResize() {
    if (options.isOpen.value) {
      updateDropdownPosition();
    }
  }

  // Store scrollable container reference for proper event handling
  let scrollableContainerForEvents: HTMLElement | null = null;

  // Set up scroll listener on the proper container when dropdown opens
  function setupScrollListener() {
    // Clean up previous listener
    if (scrollableContainerForEvents) {
      scrollableContainerForEvents.removeEventListener('scroll', handleScroll);
      scrollableContainerForEvents = null;
    }

    // Find and listen to the scrollable container
    const bounds = findContainerBounds();
    if (bounds?.scrollableContainer) {
      scrollableContainerForEvents = bounds.scrollableContainer;
      scrollableContainerForEvents.addEventListener('scroll', handleScroll);
    }
  }

  // Clean up scroll listener
  function cleanupScrollListener() {
    if (scrollableContainerForEvents) {
      scrollableContainerForEvents.removeEventListener('scroll', handleScroll);
      scrollableContainerForEvents = null;
    }
  }

  onMounted(() => {
    document.addEventListener('click', handleClickOutside);
    document.addEventListener('contextmenu', handleContextMenu);
    // Use mousedown in capture phase to close dropdown before click propagation stops
    document.addEventListener('mousedown', handleMouseDown, true);
    window.addEventListener('resize', handleResize);
    // Register for other dropdown notifications
    if (options.dropdownId && !registerOpenDropdown) {
      registerOpenDropdown = onOtherDropdownOpen;
    }
  });

  onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside);
    document.removeEventListener('contextmenu', handleContextMenu);
    document.removeEventListener('mousedown', handleMouseDown, true);
    window.removeEventListener('resize', handleResize);
    cleanupScrollListener();
    unregisterAsOpen();
    if (registerOpenDropdown === onOtherDropdownOpen) {
      registerOpenDropdown = null;
    }
  });

  return {
    selectedIndex,
    triggerRef,
    dropdownRef,
    dropdownPositionStyle,
    widthClass,
    maxWidthStyle,
    maxHeightStyle,
    shouldTeleport,
    resetIndex,
    registerAsOpen,
    unregisterAsOpen,
    handleMouseLeave,
    updateDropdownPosition,
    setupScrollListener,
    cleanupScrollListener,
  };
}
