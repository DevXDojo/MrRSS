import { ref, type Ref } from 'vue';
import type { SavedFilter } from '@/types/filter';
import type { FilterCondition } from '@/types/filter';

export function useSavedFilters() {
  const savedFilters: Ref<SavedFilter[]> = ref([]);
  const isLoading = ref(false);
  const error: Ref<string | null> = ref(null);

  // Fetch all saved filters
  async function fetchSavedFilters(): Promise<void> {
    isLoading.value = true;
    error.value = null;
    try {
      const res = await fetch('/api/saved-filters');
      if (res.ok) {
        const data = await res.json();
        // Ensure data is an array before assigning
        savedFilters.value = Array.isArray(data) ? data : [];
      } else {
        throw new Error('Failed to fetch saved filters');
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Unknown error';
      console.error('Error fetching saved filters:', e);
      // Ensure savedFilters is an array even on error
      savedFilters.value = [];
    } finally {
      isLoading.value = false;
    }
  }

  // Create new saved filter
  async function createSavedFilter(
    name: string,
    conditions: FilterCondition[]
  ): Promise<SavedFilter | null> {
    const conditionsJson = JSON.stringify(conditions);

    const requestBody = {
      name,
      conditions: conditionsJson,
    };

    const res = await fetch('/api/saved-filters', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestBody),
    });

    if (res.ok) {
      const newFilter = await res.json();
      // Ensure savedFilters is an array before pushing
      if (Array.isArray(savedFilters.value)) {
        savedFilters.value.push(newFilter);
      } else {
        savedFilters.value = [newFilter];
      }
      return newFilter;
    } else if (res.status === 409) {
      // Conflict - filter with same name already exists
      const errorData = await res.json();
      throw new Error(errorData.error || 'A filter with this name already exists');
    } else {
      const errorText = await res.text();
      throw new Error(`Failed to create saved filter: ${errorText}`);
    }
  }

  // Update existing saved filter
  async function updateSavedFilter(
    id: number,
    name: string,
    conditions: FilterCondition[]
  ): Promise<boolean> {
    try {
      const res = await fetch(`/api/saved-filters/filter?id=${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name,
          conditions: JSON.stringify(conditions),
        }),
      });

      if (res.ok) {
        await fetchSavedFilters(); // Refresh list
        return true;
      } else {
        throw new Error('Failed to update saved filter');
      }
    } catch (e) {
      console.error('Error updating saved filter:', e);
      return false;
    }
  }

  // Delete saved filter
  async function deleteSavedFilter(id: number): Promise<boolean> {
    try {
      const res = await fetch(`/api/saved-filters/filter?id=${id}`, {
        method: 'DELETE',
      });

      if (res.ok) {
        // Ensure savedFilters is an array before filtering
        if (Array.isArray(savedFilters.value)) {
          savedFilters.value = savedFilters.value.filter((f) => f.id !== id);
        } else {
          savedFilters.value = [];
        }
        return true;
      } else {
        throw new Error('Failed to delete saved filter');
      }
    } catch (e) {
      console.error('Error deleting saved filter:', e);
      return false;
    }
  }

  // Reorder saved filters
  async function reorderSavedFilters(filters: SavedFilter[]): Promise<boolean> {
    try {
      const res = await fetch('/api/saved-filters/reorder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(filters),
      });

      if (res.ok) {
        // Ensure filters is an array before assigning
        savedFilters.value = Array.isArray(filters) ? filters : [];
        return true;
      } else {
        throw new Error('Failed to reorder saved filters');
      }
    } catch (e) {
      console.error('Error reordering saved filters:', e);
      return false;
    }
  }

  // Parse conditions from JSON string
  function parseConditions(conditionsJson: string): FilterCondition[] {
    try {
      return JSON.parse(conditionsJson);
    } catch {
      return [];
    }
  }

  return {
    savedFilters,
    isLoading,
    error,
    fetchSavedFilters,
    createSavedFilter,
    updateSavedFilter,
    deleteSavedFilter,
    reorderSavedFilters,
    parseConditions,
  };
}
