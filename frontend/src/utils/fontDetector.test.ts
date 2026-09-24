import { afterEach, describe, expect, it, vi } from 'vitest';
import { isFontAvailable, resolveFontFamily } from './fontDetector';

afterEach(() => vi.restoreAllMocks());

describe('local font detection', () => {
  it('detects Chinese faces whose glyph widths equal every fallback', () => {
    let font = '';
    const context = {
      get font() {
        return font;
      },
      set font(value: string) {
        font = value;
      },
      measureText: () => ({ width: 400 }),
      clearRect: vi.fn(),
      fillText: vi.fn(),
      getImageData: () => ({
        data: new Uint8ClampedArray([font.includes('"中文字体"') ? 255 : 0]),
      }),
    };
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(
      context as unknown as CanvasRenderingContext2D
    );
    expect(isFontAvailable('中文字体')).toBe(true);
    expect(isFontAvailable('Missing')).toBe(false);
  });
  it('detects Chinese-only glyphs and fonts matching the default sans family', () => {
    let font = '';
    const context = {
      get font() {
        return font;
      },
      set font(value: string) {
        font = value;
      },
      measureText(sample: string) {
        const baseline = font.endsWith('monospace')
          ? 300
          : font.endsWith('serif') && !font.endsWith('sans-serif')
            ? 200
            : 100;
        const chinese = sample.includes('中文') && font.includes('"中文字体"');
        const system = font.includes('"Default Sans"');
        return { width: chinese ? 400 : system ? 100 : baseline };
      },
    };
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(
      context as unknown as CanvasRenderingContext2D
    );
    expect(isFontAvailable('中文字体')).toBe(true);
    expect(isFontAvailable('Default Sans')).toBe(true);
    expect(isFontAvailable('Missing font')).toBe(false);
  });

  it('quotes custom family names and retains a system fallback', () => {
    expect(resolveFontFamily('自定义字体')).toContain('"自定义字体",');
    expect(resolveFontFamily('A"B')).toContain('"A\\"B",');
    expect(resolveFontFamily('自定义字体')).toContain('system-ui');
  });
});
