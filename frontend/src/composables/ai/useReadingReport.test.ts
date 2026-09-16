import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { useReadingReport } from './useReadingReport';

vi.mock('@/utils/aiError', () => ({
  readAIError: async () => ({ message: 'Provider unavailable' }),
}));
const sources = [{ id: 1, article_id: 5, title: 'Article', kind: 'cached' }];
const answer = { report: { overview: 'Useful report' }, sources, model: 'fixture' };
const json = (data: unknown, status = 200) => new Response(JSON.stringify(data), { status });
afterEach(() => vi.unstubAllGlobals());

function setup() {
  let state!: ReturnType<typeof useReadingReport>;
  const wrapper = mount(
    defineComponent({
      setup() {
        state = useReadingReport();
        return () => null;
      },
    })
  );
  return { state, wrapper };
}

describe('reading report request lifecycle', () => {
  it('ignores stale source previews', async () => {
    let finish!: (value: Response) => void;
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockImplementationOnce(
          () =>
            new Promise<Response>((resolve) => {
              finish = resolve;
            })
        )
        .mockResolvedValueOnce(json({ sources }))
    );
    const { state, wrapper } = setup();
    const old = state.preview([1]);
    await state.preview([5]);
    finish(json({ sources: [{ article_id: 1, kind: 'missing' }] }));
    await old;
    expect(state.sources.value[0].article_id).toBe(5);
    wrapper.unmount();
  });

  it('keeps the last good report on failure and ignores a late cancelled result', async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce(json({ sources }))
      .mockResolvedValueOnce(json(answer))
      .mockResolvedValueOnce(json({}, 503));
    vi.stubGlobal('fetch', fetcher);
    const { state, wrapper } = setup();
    await state.preview([5]);
    await state.generate([5], 2, 'Focus');
    expect(JSON.parse(fetcher.mock.calls[1][1].body)).toEqual({
      article_ids: [5],
      profile_id: 2,
      focus: 'Focus',
    });
    await state.generate([5], 2, 'Other focus');
    expect(state.result.value?.report.overview).toBe('Useful report');
    expect(state.error.value).toBe('Provider unavailable');
    let finish!: (value: Response) => void;
    fetcher.mockImplementationOnce(
      () =>
        new Promise<Response>((resolve) => {
          finish = resolve;
        })
    );
    const pending = state.generate([5], 2, 'New focus');
    state.cancel();
    expect(fetcher.mock.calls[3][1].signal.aborted).toBe(true);
    finish(json({ ...answer, report: { overview: 'Late result' } }));
    await pending;
    expect(state.result.value?.report.overview).toBe('Useful report');
    expect(state.generating.value).toBe(false);
    wrapper.unmount();
  });

  it('aborts provider work on unmount and does not generate from empty coverage', async () => {
    const fetcher = vi.fn().mockResolvedValueOnce(json({ sources }));
    vi.stubGlobal('fetch', fetcher);
    const { state, wrapper } = setup();
    await state.generate([5], 0, '');
    expect(fetcher).not.toHaveBeenCalled();
    await state.preview([5]);
    fetcher.mockImplementationOnce(() => new Promise<Response>(() => {}));
    void state.generate([5], 0, '');
    wrapper.unmount();
    await flushPromises();
    expect(fetcher.mock.calls[1][1].signal.aborted).toBe(true);
  });
});
