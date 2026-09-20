import { describe, expect, it } from 'vitest';
import {
  flattenPreview,
  matchesPickerGroup,
  containingPickerItem,
  pickerLink,
  suggestPickerItems,
  suggestOptionalPickerFields,
  inferPickerField,
  previewField,
  relativePickerXPath,
  type XPathPreviewNode,
  type XPathPickerField,
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
    expect(previewField(item, '', nodes)).toBe('');
  });
});

it('suggests optional fields only when all article rows have one compatible element', () => {
  const rows: XPathPreviewNode[] = [1, 2, 3].map((i) => ({
    path: `/ul[1]/li[${i}]`,
    tag: 'li',
    children: [
      { path: `/ul[1]/li[${i}]/time[1]`, tag: 'time', date: '2026-09-20' },
      { path: `/ul[1]/li[${i}]/img[1]`, tag: 'img', image: 'https://example.org/image.png' },
    ],
  }));
  const nodes = () => new Map(rows.flatMap(flattenPreview).map((node) => [node.path!, node]));
  expect(suggestOptionalPickerFields(rows[0], rows, nodes())).toEqual({
    timestamp: './time[1]/@datetime',
    thumbnail: './img[1]/@src',
  });
  rows[1].children!.push({
    path: '/ul[1]/li[2]/img[2]',
    tag: 'img',
    image: 'https://example.org/two.png',
  });
  expect(suggestOptionalPickerFields(rows[0], rows, nodes()).thumbnail).toBeUndefined();
  rows[2].children = rows[2].children!.filter((node) => node.tag !== 'time');
  expect(suggestOptionalPickerFields(rows[0], rows, nodes()).timestamp).toBeUndefined();
});

it.each<XPathPickerField>(['title', 'uri', 'timestamp', 'content', 'thumbnail'])(
  'calibrates %s across different article wrappers without accepting ambiguous matches',
  (field) => {
    const tag = field === 'uri' ? 'a' : field === 'thumbnail' ? 'img' : 'span';
    const firstItem: XPathPreviewNode = { path: '/ul[1]/li[1]', tag: 'li' };
    const secondItem: XPathPreviewNode = { path: '/ul[1]/li[2]', tag: 'li' };
    const first: XPathPreviewNode = {
      path: firstItem.path + '/div[1]/' + tag + '[1]',
      tag,
      classes: ['field'],
      text: 'First',
      link: 'https://example.com/1',
      image: 'https://example.com/1.png',
      date: '2026-09-19',
    };
    const second: XPathPreviewNode = {
      ...first,
      path: secondItem.path + '/section[1]/' + tag + '[2]',
      text: 'Second',
      link: 'https://example.com/2',
      image: 'https://example.com/2.png',
      date: '2026-09-20',
    };
    firstItem.children = [first];
    secondItem.children = [second];
    const nodes = new Map([firstItem, secondItem, first, second].map((node) => [node.path!, node]));
    const rule = inferPickerField(first, firstItem, second, secondItem, field, nodes);
    expect(rule).toContain(`.//${tag}[contains(`);
    expect(previewField(secondItem, rule!, nodes)).toBe(
      previewField(secondItem, relativePickerXPath(secondItem, second, field)!, nodes)
    );
    expect(inferPickerField(first, firstItem, first, firstItem, field, nodes)).toBeNull();
    secondItem.children.push({ ...second, path: second.path + '/duplicate[1]' });
    expect(inferPickerField(first, firstItem, second, secondItem, field, nodes)).toBeNull();
  }
);

it('uses two title examples to remove a one-off featured class from the article rule', () => {
  const rows: XPathPreviewNode[] = [1, 2, 3].map((index) => ({
    path: `/ul[1]/li[${index}]`,
    group: '/ul[1]/li',
    tag: 'li',
    classes: index === 1 ? ['article', 'featured'] : ['article'],
    children: [
      {
        path: `/ul[1]/li[${index}]/a[1]`,
        tag: 'a',
        text: `Title ${index}`,
        link: `https://example.com/${index}`,
      },
    ],
  }));
  const nodes = new Map(rows.flatMap(flattenPreview).map((node) => [node.path!, node]));
  expect(suggestPickerItems(rows[0].children![0], nodes)).toEqual([]);
  const suggestion = suggestPickerItems(rows[0].children![0], nodes, rows[1].children![0])[0];
  expect(suggestion.classes).toEqual(['article']);
  expect(suggestion.group).not.toContain('featured');
  expect(rows.filter((row) => matchesPickerGroup(row, suggestion))).toHaveLength(3);
});

it('infers repeated articles from a nested linked title and excludes metadata rows', () => {
  const rows: XPathPreviewNode[] = [1, 3, 5].map((index) => {
    const path = `/html[1]/body[1]/table[1]/tbody[1]/tr[${index}]`;
    return {
      path,
      group: '/html[1]/body[1]/table[1]/tbody[1]/tr',
      tag: 'tr',
      classes: ['article'],
      children: [
        {
          path: path + '/td[1]',
          tag: 'td',
          children: [
            {
              path: path + '/td[1]/a[1]',
              tag: 'a',
              link: `https://example.com/${index}`,
              children: [
                { path: path + '/td[1]/a[1]/span[1]', tag: 'span', text: `Title ${index}` },
              ],
            },
          ],
        },
      ],
    };
  });
  const metadata = { path: '/html[1]/body[1]/table[1]/tbody[1]/tr[2]', tag: 'tr' };
  const nodes = new Map(
    [...rows.flatMap(flattenPreview), metadata].map((node) => [node.path!, node])
  );
  const title = nodes.get(rows[1].path + '/td[1]/a[1]/span[1]')!;
  expect(suggestPickerItems(title, nodes)).toEqual([rows[1]]);
  expect(suggestPickerItems(metadata, nodes)).toEqual([]);
  rows[0].children = [];
  rows[2].children = [];
  const incomplete = new Map(rows.flatMap(flattenPreview).map((node) => [node.path!, node]));
  expect(suggestPickerItems(title, incomplete)).toEqual([]);
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
