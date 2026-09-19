export interface XPathPreviewNode {
  path?: string;
  group?: string;
  tag?: string;
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
  return [node.text ?? '', ...(node.children ?? []).map(previewText)].join(' ').replace(/\s+/g, ' ').trim();
}

export function relativePickerXPath(root: XPathPreviewNode, node: XPathPreviewNode, field: XPathPickerField): string | null {
  if (!root.path || !node.path || (node.path !== root.path && !node.path.startsWith(`${root.path}/`))) return null;
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
export function previewField(root: XPathPreviewNode, relative: string, nodes: Map<string, XPathPreviewNode>): string {
  const [path, attr] = relative.split('/@');
  const node = nodes.get(`${root.path}${path.slice(1)}`);
  if (!node) return '';
  if (attr === 'href') return node.link ?? '';
  if (attr === 'src') return node.image ?? '';
  if (attr === 'datetime') return node.date ?? '';
  return previewText(node);
}
