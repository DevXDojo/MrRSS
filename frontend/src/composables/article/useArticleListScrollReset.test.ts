import { describe, expect, it, vi } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { ref } from 'vue';
import { useArticleListScrollReset } from './useArticleListScrollReset';

describe('article list navigation scroll', () => {
  it('suspends reading immediately and resets again after the replacement list arrives', async () => {
    const scope = ref('A');
    const loading = ref(false);
    const reset = vi.fn();
    const suspend = vi.fn();
    const resume = vi.fn();
    let resetting!: ReturnType<typeof useArticleListScrollReset>;
    const wrapper = mount({
      setup() {
        resetting = useArticleListScrollReset(scope, loading, reset, suspend, resume);
      },
      template: '<div />',
    });
    scope.value = 'B';
    loading.value = true;
    expect(resetting.value).toBe(true);
    expect(suspend).toHaveBeenCalledOnce();
    await flushPromises();
    expect(resume).not.toHaveBeenCalled();
    scope.value = 'C';
    await flushPromises();
    loading.value = false;
    await flushPromises();
    expect(reset).toHaveBeenCalledTimes(3);
    expect(resume).toHaveBeenCalledOnce();
    expect(resetting.value).toBe(false);
    // Background refresh/pagination does not reset a reader's current position.
    loading.value = true;
    await flushPromises();
    loading.value = false;
    await flushPromises();
    expect(reset).toHaveBeenCalledTimes(3);
    wrapper.unmount();
  });
});
