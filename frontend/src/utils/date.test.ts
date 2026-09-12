import { describe, expect, it } from 'vitest';
import { formatDate, formatExactDateTime } from './date';

describe('article date preferences', () => {
  const date = '2026-12-31T15:04:00';

  it('formats calendar dates without converting the local day to UTC', () => {
    for (const [format, expected] of [
      ['yyyy-mm-dd', '2026-12-31'],
      ['mm/dd/yyyy', '12/31/2026'],
      ['dd/mm/yyyy', '31/12/2026'],
      ['dd.mm.yyyy', '31.12.2026'],
    ]) {
      expect(formatExactDateTime(date, 'en-US', { date_format: format, time_format: '24h' })).toBe(
        `${expected} 15:04`
      );
    }
  });

  it('supports 12-hour time and disabling relative dates', () => {
    expect(
      formatDate(date, 'en-US', undefined, {
        date_format: 'yyyy-mm-dd',
        time_format: '12h',
        relative_time: false,
      })
    ).toBe('2026-12-31 03:04 PM');
  });

  it('uses midnight rather than 24:00 in the 24-hour clock', () => {
    expect(
      formatExactDateTime('2026-12-31T00:00:00', 'en-US', {
        date_format: 'yyyy-mm-dd',
        time_format: '24h',
      })
    ).toBe('2026-12-31 00:00');
  });

  it('falls back to locale defaults for unknown stored formats', () => {
    const invalid = { date_format: 'unknown', time_format: 'unknown' };
    expect(formatExactDateTime(date, 'en-US', invalid)).toContain('2026');
    expect(formatDate('invalid', 'en-US', undefined, invalid)).toBe('');
    expect(formatExactDateTime('invalid', 'en-US', invalid)).toBe('');
  });
});
