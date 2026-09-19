import { describe, expect, it } from 'vitest';
import {
  flattenPreview,
  previewField,
  relativePickerXPath,
  type XPathPreviewNode,
} from './xpathPicker';

describe('visual XPath selection', () => {
  const item: XPathPreviewNode = { path: '/html[1]/body[1]/ul[1]/li[1]', tag: 'li' };
  const link: XPathPreviewNode = {
    path: `${item.path}/a[1]`,
    tag: 'a',
    link: 'https://example.com/one',
    children: [{ text: '<script>plain text</script>' }],
  };
  item.children = [link];

  it('generates relative field paths and refuses elements outside the container', () => {
    expect(relativePickerXPath(item, link, 'uri')).toBe('./a[1]/@href');
    expect(relativePickerXPath(item, link, 'title')).toBe('./a[1]');
    expect(relativePickerXPath(item, link, 'thumbnail')).toBeNull();
    expect(
      relativePickerXPath(item, { path: '/html[1]/body[1]/ul[1]/li[2]/a[1]' }, 'title')
    ).toBeNull();
    expect(relativePickerXPath(item, item, 'content')).toBe('.');
  });

  it('previews title and attribute values without interpreting HTML', () => {
    const nodes = new Map(
      flattenPreview(item)
        .filter((node) => node.path)
        .map((node) => [node.path!, node])
    );
    expect(previewField(item, './a[1]', nodes)).toBe('<script>plain text</script>');
    expect(previewField(item, './a[1]/@href', nodes)).toBe('https://example.com/one');
    expect(previewField(item, './missing[1]', nodes)).toBe('');
  });
});
