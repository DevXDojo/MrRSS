import { describe, expect, it } from 'vitest';
import { exportRules, importRules, type BackupRule } from './ruleBackup';

const rules: BackupRule[] = [
  {
    id: 1,
    name: 'Keep tutorials',
    enabled: true,
    position: 0,
    conditions: [
      {
        id: 1,
        field: 'article_title',
        value: 'tutorial',
        values: [],
        negate: false,
        logic: null,
        operator: 'contains',
      },
    ],
    actions: ['favorite'],
  },
];
const parse = (text: string) => importRules(text, ['article_title'], ['favorite'], rules);
describe('rule backups', () => {
  it('round-trips conditions, order and enabled state with collision-free IDs', () => {
    const imported = parse(exportRules(rules));
    expect(imported[0]).toMatchObject({
      name: rules[0].name,
      enabled: true,
      conditions: rules[0].conditions,
      actions: ['favorite'],
      position: 1,
    });
    expect(imported[0].id).not.toBe(1);
    expect(rules).toHaveLength(1);
  });
  it('accepts legacy arrays and UTF-8 BOM files', () =>
    expect(parse('\uFEFF' + JSON.stringify(rules))).toHaveLength(1));
  it.each([
    'not json',
    JSON.stringify({ format: 'mrrss-rules', version: 99, rules }),
    JSON.stringify([{ ...rules[0], actions: ['run_script'] }]),
    JSON.stringify([
      { ...rules[0], conditions: [{ ...rules[0].conditions[0], field: 'unknown' }] },
    ]),
    JSON.stringify([{ ...rules[0], enabled: 'true' }]),
  ])('rejects invalid backups before changing existing rules', (text) => {
    expect(() => parse(text)).toThrow('invalidBackup');
    expect(rules).toHaveLength(1);
  });
});
