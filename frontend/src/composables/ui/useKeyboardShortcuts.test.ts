import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { flushPromises, mount } from '@vue/test-utils';
import { useAppStore } from '@/stores/app';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';
import { shortcuts, shortcutsEnabled } from './shortcutBindings';
import type { Article } from '@/types/models';

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
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

function setup(isRead: boolean, onToggleTicker?: () => void) {
  const store = useAppStore();
  store.articles = [{ id: 5, is_read: isRead, is_read_later: false } as Article];
  store.currentArticleId = 5;
  const counts = vi.spyOn(store, 'fetchFilterCounts').mockResolvedValue();
  const wrapper = mount({
    setup() {
      useKeyboardShortcuts({
        onAddFeed: vi.fn(),
        onOpenSettings: vi.fn(),
        onMarkAllRead: vi.fn(),
        onToggleTicker,
      });
      return {};
    },
    template: '<div />',
  });
  return { store, counts, wrapper };
}

it('toggles the desktop ticker once per key press and leaves the key unbound on the server', () => {
  const press = (repeat = false) =>
    document.body.dispatchEvent(
      new KeyboardEvent('keydown', {
        key: 'T',
        shiftKey: true,
        repeat,
        bubbles: true,
        cancelable: true,
      })
    );
  const toggle = vi.fn();
  const desktop = setup(true, toggle);
  expect(press()).toBe(false);
  press(true);
  expect(toggle).toHaveBeenCalledOnce();
  desktop.wrapper.unmount();

  const server = setup(true);
  expect(press()).toBe(true);
  server.wrapper.unmount();
});

function pressReadLater(): void {
  document.body.dispatchEvent(
    new KeyboardEvent('keydown', { key: shortcuts.value.toggleReadLaterStatus, bubbles: true })
  );
}

describe('read-later keyboard shortcut', () => {
  it('toggles the feed list and respects disabled shortcuts and text inputs', () => {
    const { wrapper } = setup(true);
    const toggle = vi.fn();
    window.addEventListener('toggle-feed-list', toggle);
    const press = (target: HTMLElement) =>
      target.dispatchEvent(
        new KeyboardEvent('keydown', { key: shortcuts.value.toggleFeedList, bubbles: true })
      );
    press(document.body);
    expect(toggle).toHaveBeenCalledOnce();
    const input = document.createElement('input');
    document.body.append(input);
    press(input);
    shortcutsEnabled.value = false;
    press(document.body);
    expect(toggle).toHaveBeenCalledOnce();
    input.remove();
    window.removeEventListener('toggle-feed-list', toggle);
    wrapper.unmount();
  });
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
