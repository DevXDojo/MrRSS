export interface XPathPreviewNode {
  html?: string;
  base_url?: string;
  path?: string;
  group?: string;
  tag?: string;
  classes?: string[];
  text?: string;
  link?: string;
  image?: string;
  date?: string;
  children?: XPathPreviewNode[];
}

export type XPathPickerField = 'item' | 'title' | 'uri' | 'timestamp' | 'content' | 'thumbnail';
export type XPathSelection = Record<XPathPickerField, string>;

export function flattenPreview(root: XPathPreviewNode): XPathPreviewNode[] {
  return [root, ...(root.children ?? []).flatMap(flattenPreview)];
}

export function previewText(node: XPathPreviewNode): string {
  return [node.text ?? '', ...(node.children ?? []).map(previewText)]
    .join(' ')
    .replace(/\s+/g, ' ')
    .trim();
}

export function relativePickerXPath(
  root: XPathPreviewNode,
  node: XPathPreviewNode,
  field: XPathPickerField
): string | null {
  if (
    !root.path ||
    !node.path ||
    (node.path !== root.path && !node.path.startsWith(`${root.path}/`))
  )
    return null;
  let path = `.${node.path.slice(root.path.length)}`;
  if (field === 'uri') {
    if (!node.link) return null;
    path += '/@href';
  } else if (field === 'thumbnail') {
    if (!node.image) return null;
    path += '/@src';
  } else if (field === 'timestamp' && node.date) path += '/@datetime';
  return path;
}

// Generated paths only: never evaluate arbitrary XPath or render source HTML.
export function previewField(
  root: XPathPreviewNode,
  relative: string,
  nodes: Map<string, XPathPreviewNode>
): string {
  const [path, attr] = relative.split('/@');
  let node: XPathPreviewNode | undefined;
  if (path.startsWith('.//')) {
    const parsed = /^\.\/\/([a-z][a-z0-9-]*)(.*)$/.exec(path);
    if (!parsed) return '';
    const classes = [
      ...parsed[2].matchAll(
        /\[contains\(concat\(' ', normalize-space\(@class\), ' '\), ' ([a-zA-Z_][a-zA-Z0-9_-]*) '\)\]/g
      ),
    ].map((match) => match[1]);
    if (classPredicates(classes) !== parsed[2]) return '';
    const matches = flattenPreview(root).filter(
      (child) =>
        child.path !== root.path &&
        child.tag === parsed[1] &&
        classes.every((name) => child.classes?.includes(name))
    );
    // An ambiguous field must be visible as missing, not silently take the first.
    if (matches.length !== 1) return '';
    node = matches[0];
  } else node = nodes.get(`${root.path}${path.slice(1)}`);
  if (!node) return '';
  if (attr === 'href') return node.link ?? '';
  if (attr === 'src') return node.image ?? '';
  if (attr === 'datetime') return node.date ?? '';
  return previewText(node);
}

function classPredicates(classes: string[]): string {
  return classes
    .map((name) => `[contains(concat(' ', normalize-space(@class), ' '), ' ${name} ')]`)
    .join('');
}

export function inferPickerField(
  first: XPathPreviewNode,
  firstItem: XPathPreviewNode,
  second: XPathPreviewNode,
  secondItem: XPathPreviewNode,
  field: XPathPickerField,
  nodes: Map<string, XPathPreviewNode>
): string | null {
  if (firstItem.path === secondItem.path) return null;
  const firstPath = relativePickerXPath(firstItem, first, field);
  const secondPath = relativePickerXPath(secondItem, second, field);
  if (!firstPath || !secondPath) return null;
  if (firstPath === secondPath) return firstPath;
  if (
    !first.tag ||
    first.tag !== second.tag ||
    first.path === firstItem.path ||
    second.path === secondItem.path
  )
    return null;
  const firstAttribute = firstPath.split('/@')[1];
  const secondAttribute = secondPath.split('/@')[1];
  if (firstAttribute !== secondAttribute) return null;
  const classes = (first.classes ?? []).filter((name) => second.classes?.includes(name));
  const rule = `.//${first.tag}${classPredicates(classes)}${firstAttribute ? '/@' + firstAttribute : ''}`;
  // Both examples must resolve uniquely to exactly the selected value.
  if (
    !previewField(firstItem, rule, nodes) ||
    !previewField(secondItem, rule, nodes) ||
    previewField(firstItem, rule, nodes) !== previewField(firstItem, firstPath, nodes) ||
    previewField(secondItem, rule, nodes) !== previewField(secondItem, secondPath, nodes)
  )
    return null;
  return rule;
}

export function matchesPickerGroup(node: XPathPreviewNode, item: XPathPreviewNode): boolean {
  const parent = (path?: string) => path?.slice(0, path.lastIndexOf('/'));
  return (
    !!node.path &&
    node.tag === item.tag &&
    parent(node.path) === parent(item.path) &&
    (item.classes ?? []).every((name) => node.classes?.includes(name))
  );
}

export function containingPickerItem(node: XPathPreviewNode, items: XPathPreviewNode[]) {
  return items.find((item) => node.path === item.path || node.path?.startsWith(`${item.path}/`));
}

export function pickerLink(
  node: XPathPreviewNode,
  root: XPathPreviewNode,
  nodes: Map<string, XPathPreviewNode>
) {
  let current: XPathPreviewNode | undefined = node;
  while (current && current.path?.startsWith(root.path!)) {
    if (current.link) return current;
    current = nodes.get(current.path.slice(0, current.path.lastIndexOf('/')));
  }
  const links = flattenPreview(node).filter((child) => child.link);
  return links.length === 1 ? links[0] : undefined;
}

// Suggest repeated ancestors, but let the user confirm the highlighted group.
// A suggestion must extract the selected title and a link in most siblings.
export function suggestPickerItems(
  title: XPathPreviewNode,
  nodes: Map<string, XPathPreviewNode>,
  secondTitle?: XPathPreviewNode | null
): XPathPreviewNode[] {
  const suggestions: { node: XPathPreviewNode; score: number }[] = [];
  let node = nodes.get(title.path?.slice(0, title.path.lastIndexOf('/')) ?? '');
  while (node?.path && !['html', 'body'].includes(node.tag ?? '')) {
    let group = node;
    let secondItem: XPathPreviewNode | undefined;
    if (secondTitle) {
      secondItem = secondTitle;
      while (secondItem?.path && !matchesPickerGroup(secondItem, { ...node, classes: [] })) {
        secondItem = nodes.get(secondItem.path.slice(0, secondItem.path.lastIndexOf('/')));
      }
      if (!secondItem || secondItem.path === node.path) {
        node = nodes.get(node.path.slice(0, node.path.lastIndexOf('/')));
        continue;
      }
      const classes = (node.classes ?? []).filter((name) => secondItem!.classes?.includes(name));
      group = {
        ...node,
        classes,
        group:
          node.path.slice(0, node.path.lastIndexOf('[')) +
          classes
            .map((name) => `[contains(concat(' ', normalize-space(@class), ' '), ' ${name} ')]`)
            .join(''),
      };
    }
    const siblings = [...nodes.values()].filter((other) => matchesPickerGroup(other, group));
    const link = pickerLink(title, node, nodes);
    let titlePath = relativePickerXPath(node, title, 'title');
    let linkPath = link && relativePickerXPath(node, link, 'uri');
    if (secondTitle && secondItem) {
      const secondLink = pickerLink(secondTitle, secondItem, nodes);
      titlePath = inferPickerField(title, node, secondTitle, secondItem, 'title', nodes);
      linkPath =
        link && secondLink && inferPickerField(link, node, secondLink, secondItem, 'uri', nodes);
      if (!titlePath || !linkPath) {
        node = nodes.get(node.path.slice(0, node.path.lastIndexOf('/')));
        continue;
      }
    }
    if (siblings.length >= 2 && titlePath && linkPath) {
      const valid = siblings.filter(
        (sibling) =>
          previewField(sibling, titlePath, nodes).trim() && previewField(sibling, linkPath, nodes)
      ).length;
      if (valid >= 2 && valid / siblings.length >= 0.8) {
        const semantic = ['article', 'li', 'tr'].includes(node.tag ?? '') ? 10 : 0;
        suggestions.push({
          node: group,
          score: semantic + Math.min(valid, 10) + valid / siblings.length,
        });
      }
    }
    node = nodes.get(node.path.slice(0, node.path.lastIndexOf('/')));
  }
  return suggestions.sort((a, b) => b.score - a.score).map(({ node }) => node);
}
