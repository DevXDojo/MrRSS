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
