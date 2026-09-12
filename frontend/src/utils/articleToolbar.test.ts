import { describe, expect, it } from 'vitest';
import { parseToolbarLayout, toolbarActions } from './articleToolbar';

describe('article toolbar preferences', () => {
  it('preserves order and hidden buttons while adding missing actions', () => {
    const layout = parseToolbarLayout('[{"id":"browser","visible":false},{"id":"read"}]');
    expect(layout.slice(0, 2)).toEqual([
      { id: 'browser', visible: false },
      { id: 'read', visible: true },
    ]);
    expect(layout).toHaveLength(toolbarActions.length);
  });
  it('ignores unknown IDs, duplicates, and malformed settings', () => {
    const layout = parseToolbarLayout(
      '[{"id":"unknown"},{"id":"read"},{"id":"read","visible":false},null]'
    );
    expect(layout.filter((item) => item.id === 'read')).toEqual([{ id: 'read', visible: true }]);
    for (const raw of ['invalid', '{}', 'null', '[]']) {
      expect(parseToolbarLayout(raw).map((item) => item.id)).toEqual(
        toolbarActions.map((item) => item.id)
      );
    }
  });
});
