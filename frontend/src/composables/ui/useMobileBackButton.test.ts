import { afterEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, nextTick, ref } from 'vue';
import { mount } from '@vue/test-utils';
import { useMobileBackButton } from './useMobileBackButton';
import { notificationZIndex, useModalClose } from './useModalClose';

afterEach(() => vi.restoreAllMocks());

describe('mobile reader navigation', () => {
  it('closes the highest modal before the reader on Back', async () => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => ({ matches: true }))
    );
    const article = ref(false);
    const modal = ref(false);
    const push = vi.spyOn(history, 'pushState').mockImplementation(() => {});
    const TestModal = defineComponent({
      setup() {
        useModalClose(() => {
          modal.value = false;
        }, 110);
      },
      template: '<div />',
    });
    const wrapper = mount(
      defineComponent({
        components: { TestModal },
        setup() {
          useMobileBackButton(
            () => article.value,
            () => {
              article.value = false;
            }
          );
          return { modal };
        },
        template: '<TestModal v-if="modal" />',
      })
    );
    article.value = true;
    await nextTick();
    const articleState = push.mock.calls[0][0];
    modal.value = true;
    await nextTick();
    await nextTick();
    expect(push).toHaveBeenCalledTimes(2);
    expect(notificationZIndex.value).toBe(120);
    window.dispatchEvent(new PopStateEvent('popstate', { state: articleState }));
    await nextTick();
    await nextTick();
    expect(modal.value).toBe(false);
    expect(article.value).toBe(true);
    window.dispatchEvent(new PopStateEvent('popstate', { state: null }));
    await nextTick();
    expect(article.value).toBe(false);
    wrapper.unmount();
    vi.unstubAllGlobals();
  });

  it('does not add history for desktop selections', async () => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => ({ matches: false }))
    );
    const article = ref(false);
    const push = vi.spyOn(history, 'pushState');
    const wrapper = mount(
      defineComponent({
        setup() {
          useMobileBackButton(
            () => article.value,
            () => {
              article.value = false;
            }
          );
        },
        template: '<div />',
      })
    );
    article.value = true;
    await nextTick();
    expect(push).not.toHaveBeenCalled();
    wrapper.unmount();
    vi.unstubAllGlobals();
  });
});
