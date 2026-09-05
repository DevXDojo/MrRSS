import { describe, expect, it } from 'vitest';
import { compareSidebarRows, type SidebarSortMode } from './useSidebarSort';

const rows = [
  { name: 'Zulu', count: 1, latest: 10, position: 0, pinned: false },
  { name: 'Alpha', count: 7, latest: 30, position: 2, pinned: false },
  { name: 'Beta', count: 3, latest: 20, position: 1, pinned: false },
];
describe('sidebar sorting', () => {
  it.each<[SidebarSortMode, string[]]>([
    ['manual', ['Zulu', 'Beta', 'Alpha']],
    ['name_asc', ['Alpha', 'Beta', 'Zulu']],
    ['name_desc', ['Zulu', 'Beta', 'Alpha']],
    ['count_asc', ['Zulu', 'Beta', 'Alpha']],
    ['count_desc', ['Alpha', 'Beta', 'Zulu']],
    ['latest', ['Alpha', 'Beta', 'Zulu']],
  ])('sorts by %s with stable names', (mode, expected) => {
    expect([...rows].sort((a, b) => compareSidebarRows(a, b, mode)).map((row) => row.name)).toEqual(
      expected
    );
  });
  it('keeps pinned items first even in count order', () => {
    const values = rows.map((row) => ({ ...row, pinned: row.name === 'Zulu' }));
    expect(values.sort((a, b) => compareSidebarRows(a, b, 'count_desc'))[0].name).toBe('Zulu');
  });
});
