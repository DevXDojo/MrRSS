import { describe, expect, it } from 'vitest';
import {
  flattenPreview,
  matchesPickerGroup,
  containingPickerItem,
  pickerLink,
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

it('matches class-filtered article rows and allows fields in any matching item', () => {
  const one = { path: '/table[1]/tr[1]', tag: 'tr', classes: ['athing'] };
  const two = { path: '/table[1]/tr[3]', tag: 'tr', classes: ['athing', 'selected'] };
  expect(matchesPickerGroup(two, one)).toBe(true);
  expect(matchesPickerGroup({ path: '/table[1]/tr[2]', tag: 'tr' }, one)).toBe(false);
  const link = { path: two.path + '/a[1]', tag: 'a', link: 'https://example.com' };
  const text = { path: link.path + '/span[1]', tag: 'span' };
  const nodes = new Map([one, two, link, text].map((node) => [node.path, node]));
  expect(containingPickerItem(text, [one, two])).toBe(two);
  expect(pickerLink(text, two, nodes)).toBe(link);
  expect(relativePickerXPath(two, link, 'uri')).toBe('./a[1]/@href');
});
