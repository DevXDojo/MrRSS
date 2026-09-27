import { nextTick, onBeforeUnmount, ref, watch, type Ref } from 'vue';

export function useArticleListScrollReset(
  scope: Ref<string>,
  loading: Ref<boolean>,
  reset: () => void,
  suspend: () => void,
  resume: () => void
) {
  const resetting = ref(false);
  let revision = 0;
  watch(
    scope,
    () => {
      revision++;
      resetting.value = true;
      suspend();
      reset();
    },
    { flush: 'sync' }
  );

  watch(
    [scope, loading],
    async () => {
      const current = ++revision;
      if (!resetting.value) return;
      await nextTick();
      if (current !== revision || loading.value) return;
      // The old presentation snapshot may have preserved the scroll offset.
      // Reset again after the new rows have actually replaced it.
      reset();
      await nextTick();
      if (current !== revision || loading.value) return;
      resetting.value = false;
      resume();
    },
    { flush: 'post' }
  );

  onBeforeUnmount(() => {
    revision++;
  });
  return resetting;
}
