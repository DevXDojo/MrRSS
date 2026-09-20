import { describe, expect, it } from 'vitest';
import { settingsSearchKeys } from './settingsSearch';

describe('settings search catalog', () => {
  it('indexes rendered setting labels instead of hidden status and modal copy', () => {
    const keys = Object.values(settingsSearchKeys).flat();

    expect(keys).toContain('setting.ai.aiChatEnabled');
    expect(keys).toContain('setting.network.enableProxy');
    expect(keys).not.toContain('setting.ai.aiTestFailed');
    expect(keys).not.toContain('modal.feed.errorCertificate');
    expect(keys).not.toContain('setting.database.directoryFailed');
  });
});
