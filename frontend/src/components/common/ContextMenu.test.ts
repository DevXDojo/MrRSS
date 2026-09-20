import { mount } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import ContextMenu from './ContextMenu.vue';

afterEach(() => {
  document.body.innerHTML = '';
});

describe('ContextMenu', () => {
  it('closes when an overlay stops bubbling outside clicks', async () => {
    const wrapper = mount(ContextMenu, {
      attachTo: document.body,
      props: {
        items: [{ label: 'Action', action: 'action' }],
        x: 10,
        y: 10,
      },
    });
    await new Promise((resolve) => window.setTimeout(resolve, 0));

    const overlay = document.createElement('div');
    overlay.addEventListener('click', (event) => event.stopPropagation());
    document.body.appendChild(overlay);
    overlay.dispatchEvent(new MouseEvent('click', { bubbles: true }));

    expect(wrapper.emitted('close')).toHaveLength(1);
    wrapper.unmount();
  });
});
