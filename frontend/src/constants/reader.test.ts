import { describe, expect, it } from 'vitest';
import {
  READER_FONT_SIZE,
  READER_LINE_HEIGHT,
  READER_MAX_WIDTH,
  clampReaderSetting,
} from './reader';

describe('reader preferences', () => {
  it.each([
    [10, READER_FONT_SIZE, 14],
    [30, READER_FONT_SIZE, 22],
    [100, READER_LINE_HEIGHT, 130],
    [240, READER_LINE_HEIGHT, 200],
    [400, READER_MAX_WIDTH, 520],
    [1000, READER_MAX_WIDTH, 840],
  ])('clamps %s to its supported range', (value, range, expected) => {
    expect(clampReaderSetting(value, range.min, range.max, range.default)).toBe(expected);
  });

  it('uses the specified fallback for invalid persisted values', () => {
    expect(clampReaderSetting('invalid', 14, 22, 17)).toBe(17);
  });
});
