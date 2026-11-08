import { nextTick, onMounted, onUnmounted, watch } from 'vue';
import { closeTopModal, openModalCount } from './useModalClose';

// Add history only for open reader/modal layers. The list remains a normal
// browser entry, so a further Back can leave the application.
export function useMobileBackButton(hasArticle: () => boolean, closeArticle: () => void) {
  const session = `${Date.now()}-${Math.random()}`;
  let depth = 0;
  let navigating = false;
  let mounted = false;
  const desiredDepth = () => Number(hasArticle()) + openModalCount.value;
  const reconcile = () => {
    if (!mounted || navigating) return;
    const desired = desiredDepth();
    if (desired < depth) {
      navigating = true;
      history.go(desired - depth);
    } else if (window.matchMedia('(max-width: 767px)').matches) {
      while (depth < desired) {
        depth++;
        history.pushState({ ...history.state, mrrssMobile: { session, depth } }, '');
      }
    }
  };
  const onPopState = async (event: PopStateEvent) => {
    const marker = event.state?.mrrssMobile;
    const target = marker?.session === session ? marker.depth : 0;
    if (!depth && !navigating) return;
    navigating = true;
    depth = target;
    // Await each unmount before closing the next layer.
    while (desiredDepth() > target) {
      const before = desiredDepth();
      if (!closeTopModal()) closeArticle();
      await nextTick();
      if (desiredDepth() >= before) break;
    }
    navigating = false;
    reconcile();
  };
  watch(desiredDepth, reconcile, { flush: 'post' });
  onMounted(() => {
    mounted = true;
    window.addEventListener('popstate', onPopState);
    reconcile();
  });
  onUnmounted(() => {
    mounted = false;
    window.removeEventListener('popstate', onPopState);
  });
}
