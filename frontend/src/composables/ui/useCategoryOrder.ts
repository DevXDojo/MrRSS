import { computed, provide, ref, onUnmounted, type InjectionKey, type Ref } from 'vue';
import { useSettings } from '@/composables/core/useSettings';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';

export function parseCategoryOrder(value: string): string[] {
  try {
    const result: unknown = JSON.parse(value);
    return Array.isArray(result) ? [...new Set(result.filter((item): item is string => typeof item === 'string'))] : [];
  } catch { return []; }
}

export const categoryParent = (path: string) => path.slice(0, Math.max(0, path.lastIndexOf('/')));

interface CategoryDrag {
  source: Ref<string | null>;
  preview: Ref<{ path: string; before: boolean } | null>;
  start: (path: string, event: DragEvent) => void;
  over: (path: string, event: DragEvent) => boolean;
  drop: (path: string, event: DragEvent) => boolean;
  end: () => void;
}
export const categoryDragKey: InjectionKey<CategoryDrag> = Symbol('categoryDrag');

export function useCategoryOrder() {
  const { settings } = useSettings();
  const store = useAppStore();
  const { t } = useI18n();
  const order = computed(() => parseCategoryOrder(settings.value.sidebar_category_order));
  const source = ref<string | null>(null);
  const preview = ref<{ path: string; before: boolean } | null>(null);
  const saving = ref(false);

  function entries<T>(children: Record<string, T>, parent = ''): [string, T][] {
    const path = (name: string) => parent ? `${parent}/${name}` : name;
    const position = (name: string) => {
      const index = order.value.indexOf(path(name));
      return index < 0 ? Number.MAX_SAFE_INTEGER : index;
    };
    return Object.entries(children).sort(([a], [b]) => position(a) - position(b) || a.localeCompare(b));
  }

  function end() { source.value = null; preview.value = null; }
  function start(path: string, event: DragEvent) {
    if (saving.value) { event.preventDefault(); return; }
    event.stopPropagation();
    source.value = path;
    event.dataTransfer?.setData('application/x-mrrss-category', path);
    if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
  }
  function over(path: string, event: DragEvent) {
    if (source.value === null) return false;
    event.stopPropagation();
    if (source.value === path || categoryParent(source.value) !== categoryParent(path)) {
      preview.value = null;
      if (event.dataTransfer) event.dataTransfer.dropEffect = 'none';
      return true;
    }
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'move';
    const header = event.currentTarget as HTMLElement;
    const rect = header.getBoundingClientRect();
    preview.value = { path, before: event.clientY < rect.top + rect.height / 2 };
    return true;
  }
  async function save(sourcePath: string, targetPath: string, before: boolean) {
    const parent = categoryParent(sourcePath);
    const paths = new Set<string>();
    for (const feed of store.feeds) {
      const parts = (feed.category || '').split('/');
      for (let i = 1; i <= parts.length; i++) {
        const path = parts.slice(0, i).join('/');
        if (path && categoryParent(path) === parent) paths.add(path);
      }
    }
    const siblings = [...paths].sort((a, b) => {
      const rank = (path: string) => order.value.includes(path) ? order.value.indexOf(path) : Number.MAX_SAFE_INTEGER;
      return rank(a) - rank(b) || a.localeCompare(b);
    }).filter((path) => path !== sourcePath);
    const index = siblings.indexOf(targetPath);
    if (index < 0 || !paths.has(sourcePath)) return;
    siblings.splice(index + (before ? 0 : 1), 0, sourcePath);
    const value = JSON.stringify([...order.value.filter((path) => !paths.has(path)), ...siblings]);
    saving.value = true;
    try {
      const response = await fetch('/api/settings', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sidebar_category_order: value }),
      });
      if (!response.ok) throw new Error('Category order save failed');
      settings.value.sidebar_category_order = value;
      window.showToast(t('sidebar.order.saved'), 'success');
    } catch {
      window.showToast(t('common.errors.savingSettings'), 'error');
    } finally { saving.value = false; }
  }
  function drop(path: string, event: DragEvent) {
    if (source.value === null) return false;
    event.preventDefault(); event.stopPropagation();
    if (preview.value?.path === path && categoryParent(source.value) === categoryParent(path)) {
      void save(source.value, path, preview.value.before);
    }
    end();
    return true;
  }
  provide(categoryDragKey, { source, preview, start, over, drop, end });
  onUnmounted(end);
  return { entries };
}
