import { ref, computed, type Ref } from 'vue';
import { useAppStore } from '@/stores/app';
import type { Article } from '@/types/models';
import type { FilterCondition } from '@/types/filter';

export function useArticleFilter() {
  const store = useAppStore();
  // Use computed to get references to the store's filter state
  const activeFilters = computed({
    get: () => store.activeFilters,
    set: (value) => store.setActiveFilters(value),
  });
  const filteredArticlesFromServer = computed({
    get: () => store.filteredArticlesFromServer,
    set: (value) => store.setFilteredArticlesFromServer(value),
  });
  const isFilterLoading = computed({
    get: () => store.isFilterLoading,
    set: (value) => store.setIsFilterLoading(value),
  });
  const filterPage = ref(1);
  const filterHasMore = ref(true);
  const filterTotal = ref(0);

  // Reset filter state
  function resetFilterState(): void {
    store.setFilteredArticlesFromServer([]);
    filterPage.value = 1;
    filterHasMore.value = true;
    filterTotal.value = 0;
  }

  // Fetch filtered articles from server with pagination
  async function fetchFilteredArticles(filters: FilterCondition[], append = false): Promise<void> {
    if (filters.length === 0) {
      resetFilterState();
      return;
    }

    store.setIsFilterLoading(true);
    try {
      const page = append ? filterPage.value : 1;

      const res = await fetch('/api/articles/filter', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          conditions: filters,
          page: page,
          limit: 50,
        }),
      });

      if (res.ok) {
        const data = await res.json();
        const articles = data.articles || [];

        if (append) {
          store.setFilteredArticlesFromServer([...store.filteredArticlesFromServer, ...articles]);
        } else {
          store.setFilteredArticlesFromServer(articles);
          filterPage.value = 1;
        }

        // Ensure filtered articles are also in the store for article detail view
        articles.forEach((article) => {
          const existingIndex = store.articles.findIndex((a) => a.id === article.id);
          if (existingIndex === -1) {
            // Article not in store, add it
            store.articles.push(article);
          } else {
            // Article already in store, update it
            store.articles[existingIndex] = article;
          }
        });

        filterHasMore.value = data.has_more;
        filterTotal.value = data.total;
      } else {
        console.error('Error fetching filtered articles');
        if (!append) {
          store.setFilteredArticlesFromServer([]);
        }
      }
    } catch (e) {
      console.error('Error fetching filtered articles:', e);
      if (!append) {
        store.setFilteredArticlesFromServer([]);
      }
    } finally {
      store.setIsFilterLoading(false);
    }
  }

  // Load more filtered articles
  async function loadMoreFilteredArticles(): Promise<void> {
    if (isFilterLoading.value || !filterHasMore.value) return;

    filterPage.value++;
    await fetchFilteredArticles(activeFilters.value, true);
  }

  // Clear all filters
  function clearAllFilters(): void {
    store.setActiveFilters([]);
    store.setFilteredArticlesFromServer([]);
    filterPage.value = 1;
    filterHasMore.value = true;
    filterTotal.value = 0;
  }

  return {
    activeFilters,
    filteredArticlesFromServer,
    isFilterLoading,
    filterPage,
    filterHasMore,
    filterTotal,
    resetFilterState,
    fetchFilteredArticles,
    loadMoreFilteredArticles,
    clearAllFilters,
  };
}
