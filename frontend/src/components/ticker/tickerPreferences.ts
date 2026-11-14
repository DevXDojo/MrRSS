import type { CSSProperties } from 'vue';
import type { SettingsData } from '@/types/settings';

export const tickerPalettes = [
  { id: 'midnight', start: '#1e293b', end: '#0f172a', text: '#f8fafc' },
  { id: 'paper', start: '#f8fafc', end: '#e2e8f0', text: '#0f172a' },
  { id: 'forest', start: '#183d35', end: '#102923', text: '#ecfdf5' },
  { id: 'warm', start: '#f5ede0', end: '#e8d8bf', text: '#483724' },
] as const;

export const tickerNumber = (value: number, fallback: number, min: number, max: number) =>
  Number.isFinite(value) ? Math.min(max, Math.max(min, value)) : fallback;

export function tickerFeedIDs(settings: SettingsData): number[] {
  // An empty value predates multi-select. Zero explicitly selects all feeds.
  const selection = settings.ticker_feed_ids || String(settings.ticker_feed_id);
  return [...new Set(selection.split(',').map(Number))].filter(
    (id) => Number.isSafeInteger(id) && id > 0
  );
}
// Legacy color values are kept in storage, while the interface and rendering
// use only the supported presets.
export function tickerPalette(settings: SettingsData) {
  return (
    tickerPalettes.find(
      (palette) =>
        palette.start === settings.ticker_background_color &&
        palette.end === settings.ticker_background_color_end &&
        palette.text === settings.ticker_text_color
    ) ?? tickerPalettes[0]
  );
}

export function tickerStyle(settings: SettingsData): CSSProperties {
  const { start, end, text } = tickerPalette(settings);
  const rgb = [1, 3, 5].map((index) => parseInt(end.slice(index, index + 2), 16));
  const light = rgb[0] * 0.2126 + rgb[1] * 0.7152 + rgb[2] * 0.0722 > 150;
  return {
    '--ticker-start': start,
    '--ticker-end': end,
    '--ticker-ink': text,
    '--ticker-chrome': light ? '#1e293b' : '#f1f5f9',
    '--ticker-size': `${tickerNumber(settings.ticker_font_size, 18, 12, 40)}px`,
  };
}

export function tickerPayload(settings: SettingsData): Record<string, string> {
  return Object.fromEntries(
    Object.entries(settings)
      .filter(
        ([key]) =>
          key.startsWith('ticker_') && !key.startsWith('ticker_window_') && key !== 'ticker_width'
      )
      .map(([key, value]) => [key, String(value)])
  );
}
