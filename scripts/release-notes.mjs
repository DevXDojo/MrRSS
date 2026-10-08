import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { pathToFileURL } from 'node:url';

export function extractVersion(markdown, version) {
  const lines = markdown.split(/\r?\n/);
  const start = lines.findIndex((line) => line.startsWith(`## [${version}]`));
  if (start < 0) return '';
  const next = lines.findIndex((line, index) => index > start && line.startsWith('## ['));
  return lines.slice(start + 1, next < 0 ? undefined : next).join('\n').trim();
}

export function releaseNotes(english, chinese, version) {
  const en = extractVersion(english, version) || `Release version ${version}`;
  const zh = extractVersion(chinese, version);
  return zh ? `## English\n\n${en}\n\n## 中文\n\n${zh}\n` : `${en}\n`;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const version = process.argv[2];
  if (!version || !/^\d+\.\d+\.\d+(?:-[\w.-]+)?$/.test(version)) {
    throw new Error('Supply a release version, e.g. 1.3.40');
  }
  const read = (path) => existsSync(path) ? readFileSync(path, 'utf8') : '';
  writeFileSync('RELEASE_NOTES.md', releaseNotes(read('CHANGELOG.md'), read('CHANGELOG.zh-CN.md'), version));
}
