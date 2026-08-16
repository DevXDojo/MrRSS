export const READER_FONT_SIZE = {
  min: 14,
  max: 22,
  step: 1,
  default: 17,
} as const;

export const READER_LINE_HEIGHT = {
  min: 130,
  max: 200,
  step: 5,
  default: 165,
} as const;

export const READER_MAX_WIDTH = {
  min: 520,
  max: 840,
  step: 20,
  default: 680,
} as const;

export function clampReaderSetting(value: unknown, min: number, max: number, fallback: number) {
  const numericValue = Number(value);
  if (!Number.isFinite(numericValue)) return fallback;
  return Math.min(max, Math.max(min, numericValue));
}
