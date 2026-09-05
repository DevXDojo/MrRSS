import { describe, expect, it } from 'vitest';
import { wrapOrphanedTextNodes } from './translationParagraphs';

describe('translation paragraphs', () => {
  it('keeps links and emphasis in the same sentence next to media', () => {
    const container = document.createElement('div');
    container.innerHTML = 'Read <strong>this</strong> <a href="https://example.com">article</a> now.<img src="test.png"><span>Another sentence.</span>';
    const link = container.querySelector('a');
    wrapOrphanedTextNodes(container);
    expect([...container.children].map((el) => el.tagName)).toEqual(['P', 'IMG', 'P']);
    expect(container.firstElementChild?.textContent).toBe('Read this article now.');
    expect(container.querySelector('p a')).toBe(link);
    const normalized = container.innerHTML;
    wrapOrphanedTextNodes(container);
    expect(container.innerHTML).toBe(normalized);
  });

  it('leaves code and rendered formula internals untouched', () => {
    const container = document.createElement('div');
    container.innerHTML = '<pre><code><div>const x = 1;</div></code></pre><div class="katex"><div>formula</div></div><section>Text<img src="test.png"></section>';
    wrapOrphanedTextNodes(container);
    expect(container.querySelector('pre p,.katex p')).toBeNull();
    expect(container.querySelector('section p')?.textContent).toBe('Text');
  });
});
