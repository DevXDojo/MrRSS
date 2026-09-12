import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { flushPromises, mount } from '@vue/test-utils';
import { useAppStore } from '@/stores/app';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';
import { shortcuts, shortcutsEnabled } from './shortcutBindings';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ locale: { value: 'en' }, t: (key: string) => key }),
}));

const originalToast = window.showToast;
beforeEach(() => {
  setActivePinia(createPinia());
  shortcutsEnabled.value = true;
  window.showToast = vi.fn();
});
afterEach(() => {
  vi.unstubAllGlobals();
  window.showToast = originalToast;
});

function setup(isRead: boolean) {
  const store = useAppStore();
  store.articles = [{ id: 5, is_read: isRead, is_read_later: false } as Article];
  store.currentArticleId = 5;
  const counts = vi.spyOn(store, 'fetchFilterCounts').mockResolvedValue();
  const wrapper = mount({
    setup() {
      useKeyboardShortcuts({ onAddFeed: vi.fn(), onOpenSettings: vi.fn(), onMarkAllRead: vi.fn() });
      return {};
    },
    template: '<div />',
  });
  return { store, counts, wrapper };
}

function pressReadLater(): void {
  document.body.dispatchEvent(
    new KeyboardEvent('keydown', { key: shortcuts.value.toggleReadLaterStatus, bubbles: true })
  );
}

describe('read-later keyboard shortcut', () => {
  it.each([true, false])(
    'preserves reading state (%s) and updates counts on success',
    async (isRead) => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}')));
      const { store, counts, wrapper } = setup(isRead);
      pressReadLater();
      await flushPromises();
      expect(store.articles[0].is_read).toBe(isRead);
      expect(store.articles[0].is_read_later).toBe(true);
      expect(counts).toHaveBeenCalledOnce();
      pressReadLater();
      await flushPromises();
      expect(store.articles[0].is_read).toBe(isRead);
      expect(store.articles[0].is_read_later).toBe(false);
      wrapper.unmount();
    }
  );
  it('ignores repeats while pending and restores the flag on HTTP failure', async () => {
    let finish!: (value: Response) => void;
    const fetchMock = vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          finish = resolve;
        })
    );
    vi.stubGlobal('fetch', fetchMock);
    const { store, counts, wrapper } = setup(true);
    pressReadLater();
    pressReadLater();
    expect(fetchMock).toHaveBeenCalledOnce();
    finish(new Response('{}', { status: 500 }));
    await flushPromises();
    expect(store.articles[0].is_read).toBe(true);
    expect(store.articles[0].is_read_later).toBe(false);
    expect(counts).not.toHaveBeenCalled();
    expect(window.showToast).toHaveBeenCalledWith('common.errors.savingSettings', 'error');
    wrapper.unmount();
  });
});
