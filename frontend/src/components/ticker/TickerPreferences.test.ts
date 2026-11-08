import { expect, it } from 'vitest';
import { generateInitialSettings } from '@/composables/core/useSettings.generated';
import { tickerPayload, tickerStyle } from './tickerPreferences';

it('does not expose internal window coordinates or unrelated settings in preference writes', () => {
  const payload = tickerPayload({
    ...generateInitialSettings(),
    ticker_window_x: 500,
    ticker_window_y: 600,
  });
  expect(payload.ticker_enabled).toBe('true');
  expect(payload.ticker_position).toBe('floating');
  expect(payload).not.toHaveProperty('ticker_window_x');
  expect(payload).not.toHaveProperty('ticker_window_y');
  expect(payload).not.toHaveProperty('ticker_width');
  expect(payload).not.toHaveProperty('ai_api_key');
});
it('bounds invalid appearance values and keeps controls readable on light backgrounds', () => {
  const style = tickerStyle({
    ...generateInitialSettings(),
    ticker_font_size: NaN,
    ticker_background_color: '#f8fafc',
    ticker_background_color_end: '#e2e8f0',
    ticker_text_color: '#0f172a',
  });
  expect(style['--ticker-start']).toBe('#f8fafc');
  expect(style['--ticker-size']).toBe('18px');
  expect(style['--ticker-chrome']).toBe('#1e293b');
});
it('uses a supported preset for legacy custom colors', () => {
  const style = tickerStyle({
    ...generateInitialSettings(),
    ticker_background_color: 'url(https://example.org)',
    ticker_background_color_end: '#ffffff',
  });
  expect(style['--ticker-start']).toBe('#1e293b');
  expect(style['--ticker-end']).toBe('#0f172a');
});
