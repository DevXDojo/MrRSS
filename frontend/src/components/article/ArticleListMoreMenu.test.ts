import { mount } from '@vue/test-utils';
import { describe, expect, it, vi } from 'vitest';
import ArticleListMoreMenu from './ArticleListMoreMenu.vue';
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));
describe('article list more menu', () => {
  it('keeps grouping and filters in the menu and closes on selection or Escape', async () => {
    const wrapper = mount(ArticleListMoreMenu, {
      props: { groupBy: 'none', filterCount: 2 },
      attachTo: document.body,
    });
    expect(wrapper.find('[role="group"]').exists()).toBe(false);
    await wrapper.get('button').trigger('click');
    const choices = wrapper.findAll('[role="group"] button');
    expect(choices).toHaveLength(3);
    await choices[1].trigger('click');
    expect(wrapper.emitted('group')).toEqual([['date']]);
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false');
    await wrapper.get('button').trigger('click');
    await wrapper.findAll('button')[1].trigger('click');
    expect(wrapper.emitted('filter')).toHaveLength(1);
    await wrapper.get('button').trigger('click');
    await wrapper.trigger('keydown', { key: 'Escape' });
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false');
    wrapper.unmount();
  });
});
