import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { nextTick } from 'vue';
import BaseSelect from './BaseSelect.vue';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));

describe('dropdowns in scrolling modals', () => {
  afterEach(() => vi.restoreAllMocks());

  it('opens above a bottom-edge trigger without adding overflow to the form', async () => {
    const backdrop = document.createElement('div');
    backdrop.dataset.modalOpen = 'true';
    const body = document.createElement('div');
    body.className = 'overflow-y-scroll';
    body.style.overflowY = 'scroll';
    backdrop.append(body);
    document.body.append(backdrop);
    vi.spyOn(backdrop, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 0, 800, 600));
    vi.spyOn(body, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 100, 800, 300));
    body.scrollTop = 125;
    const focus = vi.spyOn(HTMLInputElement.prototype, 'focus');
    const wrapper = mount(BaseSelect, {
      attachTo: body,
      props: { modelValue: 'one', searchable: true, options: [{ value: 'one', label: 'One' }] },
    });
    try {
      vi.spyOn(wrapper.get('button').element, 'getBoundingClientRect').mockReturnValue(
        new DOMRect(200, 350, 160, 40)
      );
      await wrapper.get('button').trigger('click');
      await nextTick();
      const dropdown = backdrop.querySelector<HTMLElement>('.select-dropdown');
      expect(dropdown?.parentElement).toBe(backdrop);
      expect(body.querySelector('.select-dropdown')).toBeNull();
      expect(dropdown?.style.bottom).toBe('254px');
      expect(dropdown?.style.maxHeight).toBe('240px');
      expect(body.scrollTop).toBe(125);
      expect(focus).toHaveBeenCalledWith({ preventScroll: true });
      await wrapper.get('button').trigger('click');
      expect(backdrop.querySelector('.select-dropdown')).toBeNull();
    } finally {
      wrapper.unmount();
      backdrop.remove();
    }
  });
});
