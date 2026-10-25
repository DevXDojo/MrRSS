import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount, type VueWrapper } from '@vue/test-utils';
import { defineComponent, h, ref } from 'vue';
import { useViewerFullscreen } from './useViewerFullscreen';

const native = vi.hoisted(() => ({
  desktop: false,
  enter: vi.fn(),
  leave: vi.fn(),
  isFullscreen: vi.fn(),
  on: vi.fn(),
  remove: vi.fn(),
}));
vi.mock('@wailsio/runtime', () => ({
  System: { IsDesktop: () => native.desktop },
  Window: { Fullscreen: native.enter, UnFullscreen: native.leave, IsFullscreen: native.isFullscreen },
  Events: { On: native.on },
}));

let wrapper: VueWrapper | undefined;
let fullscreenElement: Element | null;
beforeEach(() => {
  vi.clearAllMocks();
  native.desktop = false;
  native.enter.mockResolvedValue(undefined);
  native.leave.mockResolvedValue(undefined);
  native.isFullscreen.mockResolvedValue(false);
  native.on.mockReturnValue(native.remove);
  fullscreenElement = null;
  Object.defineProperty(document, 'fullscreenElement', { configurable: true, get: () => fullscreenElement });
  Object.defineProperty(document, 'exitFullscreen', { configurable: true, value: vi.fn(async () => { fullscreenElement = null; }) });
});
afterEach(() => { wrapper?.unmount(); wrapper = undefined; });

function setup() {
  const element = document.createElement('div');
  element.requestFullscreen = vi.fn(async () => { fullscreenElement = element; });
  const onError = vi.fn();
  let api!: ReturnType<typeof useViewerFullscreen>;
  wrapper = mount(defineComponent({ setup() { api = useViewerFullscreen(ref(element), onError); return () => h('div'); } }));
  return { api, element, onError };
}

describe('multimedia viewer fullscreen', () => {
  it('enters browser fullscreen and exits before closing the viewer', async () => {
    const { api, element } = setup();
    expect(api.supported.value).toBe(true);
    await api.toggle();
    expect(element.requestFullscreen).toHaveBeenCalledOnce();
    expect(api.isFullscreen.value).toBe(true);
    expect(api.exitFullscreen()).toBe(true);
    await vi.waitFor(() => expect(api.isFullscreen.value).toBe(false));
    expect(api.exitFullscreen()).toBe(false);
  });

  it('restores native fullscreen owned by the viewer on unmount', async () => {
    native.desktop = true;
    const { api } = setup();
    await api.toggle();
    expect(native.enter).toHaveBeenCalledOnce();
    wrapper!.unmount(); wrapper = undefined;
    expect(native.leave).toHaveBeenCalledOnce();
    expect(native.remove).toHaveBeenCalledOnce();
  });

  it('preserves native fullscreen that was already active', async () => {
    native.desktop = true;
    native.isFullscreen.mockResolvedValue(true);
    const { api } = setup();
    await api.toggle();
    expect(api.isFullscreen.value).toBe(true);
    await api.toggle();
    expect(native.enter).not.toHaveBeenCalled();
    expect(native.leave).not.toHaveBeenCalled();
  });

  it('reports rejected browser requests and remains retryable', async () => {
    const { api, element, onError } = setup();
    vi.mocked(element.requestFullscreen).mockRejectedValueOnce(new Error('Denied'));
    await api.toggle();
    expect(onError).toHaveBeenCalledOnce();
    expect(api.isFullscreen.value).toBe(false);
    expect(api.pending.value).toBe(false);
    await api.toggle();
    expect(api.isFullscreen.value).toBe(true);
  });

  it('releases browser fullscreen when a pending entry finishes after unmount', async () => {
    const { api, element } = setup();
    let resolve!: () => void;
    vi.mocked(element.requestFullscreen).mockImplementation(() => new Promise<void>((done) => { resolve = () => { fullscreenElement = element; done(); }; }));
    const entry = api.toggle();
    wrapper!.unmount(); wrapper = undefined;
    resolve();
    await entry;
    expect(document.exitFullscreen).toHaveBeenCalledOnce();
    expect(fullscreenElement).toBeNull();
  });
});
