import { computed, onBeforeUnmount, ref, shallowRef, watch, type Ref } from 'vue';
import type { Article } from '@/types/models';

// Keep a presentation-only snapshot while navigation clears the live list.
// Callers must keep it inert and use live articles for actions and navigation.
export function useArticleListTransition(articles: Ref<Article[]>, loading: Ref<boolean>) {
  const previous = shallowRef<Article[]>(articles.value);
  watch([articles, loading], ([items, pending]) => {
    if (items.length > 0 || !pending) previous.value = items;
  });
  const showingPrevious = computed(
    () => loading.value && articles.value.length === 0 && previous.value.length > 0
  );
  const displayedArticles = computed(() =>
    showingPrevious.value ? previous.value : articles.value
  );
  const showLoadingIndicator = ref(false);
  let indicatorTimer: ReturnType<typeof setTimeout> | undefined;
  watch(showingPrevious, (pending) => {
    clearTimeout(indicatorTimer);
    indicatorTimer = undefined;
    if (pending)
      indicatorTimer = setTimeout(() => {
        showLoadingIndicator.value = true;
      }, 150);
    else showLoadingIndicator.value = false;
  });
  onBeforeUnmount(() => clearTimeout(indicatorTimer));
  return { showingPrevious, displayedArticles, showLoadingIndicator };
}
