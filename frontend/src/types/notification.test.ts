import { afterEach, describe, expect, it, vi } from 'vitest';
import { newChannel, newRule } from './notification';

afterEach(() => vi.unstubAllGlobals());

describe('notification identifiers', () => {
  it('creates destinations and rules on LAN HTTP without randomUUID', () => {
    const getRandomValues = globalThis.crypto.getRandomValues.bind(globalThis.crypto);
    vi.stubGlobal('crypto', { getRandomValues });
    const channel = newChannel('telegram', 'Personal');
    const rule = newRule('digest', 'Daily', [channel.id]);
    expect(channel.id).toMatch(/^[a-f0-9]{32}$/);
    expect(rule.id).toMatch(/^[a-f0-9]{32}$/);
    expect(rule.id).not.toBe(channel.id);
    expect(rule.channel_ids).toEqual([channel.id]);
  });
});
