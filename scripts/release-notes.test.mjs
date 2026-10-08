import { test } from 'node:test';
import assert from 'node:assert/strict';
import { extractVersion, releaseNotes } from './release-notes.mjs';

test('extracts only the exact release and preserves subsection headings', () => {
  const markdown = '## [1.3.400]\r\nwrong\r\n## [1.3.40] - 2026-10-08\r\n\r\n### Fixed\r\n- correct\r\n## [1.3.39]\r\nold';
  assert.equal(extractVersion(markdown, '1.3.40'), '### Fixed\n- correct');
  assert.equal(extractVersion(markdown, '1.3.41'), '');
});

test('includes Chinese when available and supports older English-only releases', () => {
  assert.equal(releaseNotes('## [1.3.40]\nFix', '## [1.3.40]\n修复', '1.3.40'), '## English\n\nFix\n\n## 中文\n\n修复\n');
  assert.equal(releaseNotes('## [1.3.39]\nOld', '', '1.3.39'), 'Old\n');
  assert.equal(releaseNotes('', '', '1.3.40'), 'Release version 1.3.40\n');
});

test('excludes general release guidance after the oldest version in both languages', () => {
  const en = '## [1.1.0]\n\n### Added\n\n- First release\n\n---\n\n## Release Notes\nGeneral guidance';
  const zh = '## [1.1.0]\n\n### 新增\n\n- 首次发布\n\n---\n\n## 发布说明\n通用说明';
  const notes = releaseNotes(en, zh, '1.1.0');
  assert.ok(notes.includes('First release') && notes.includes('首次发布'));
  assert.ok(!notes.includes('General guidance') && !notes.includes('通用说明'));
});
