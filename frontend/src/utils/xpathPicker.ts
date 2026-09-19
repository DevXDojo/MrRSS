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
  const node = nodes.get(`${root.path}${path.slice(1)}`);
  if (!node) return '';
  if (attr === 'href') return node.link ?? '';
  if (attr === 'src') return node.image ?? '';
  if (attr === 'datetime') return node.date ?? '';
  return previewText(node);
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
