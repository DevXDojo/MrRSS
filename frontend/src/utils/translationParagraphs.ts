// Normalize RSS fragments into paragraphs without separating inline emphasis or links.
export function wrapOrphanedTextNodes(container: Element): void {
  const inlineTags = new Set([
    'A',
    'SPAN',
    'STRONG',
    'EM',
    'B',
    'I',
    'U',
    'S',
    'SMALL',
    'SUB',
    'SUP',
    'BR',
    'CODE',
    'KBD',
    'MARK',
    'ABBR',
    'TIME',
    'DEL',
    'INS',
    'CITE',
    'Q',
  ]);
  const blocks = [container, ...Array.from(container.querySelectorAll('div,section,article'))];
  for (const block of blocks) {
    if (block.closest('pre,code,kbd,.katex,.translation-text')) continue;
    let run: ChildNode[] = [];
    const flush = () => {
      if (run.some((node) => node.textContent?.trim())) {
        const paragraph = document.createElement('p');
        block.insertBefore(paragraph, run[0]);
        run.forEach((node) => paragraph.appendChild(node));
      }
      run = [];
    };
    for (const node of Array.from(block.childNodes)) {
      if (
        node.nodeType === Node.TEXT_NODE ||
        (node instanceof Element && inlineTags.has(node.tagName))
      ) {
        run.push(node);
      } else {
        flush();
      }
    }
    flush();
  }
}
