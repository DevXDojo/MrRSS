import { describe, expect, it } from 'vitest';
import { sanitizeArticleHTML } from './articleHtml';
import { withLazyImages } from './lazyImages';

describe('article HTML rendering boundary', () => {
  it.each([
    '<p>safe</p><p><noscript></p><img src=x onerror="alert(1)"></noscript></p>',
    '<noscript><p title="</noscript><img src=x onerror=alert(1)>\">',
    '<img src=x ONERROR=alert(1)><details open ontoggle=alert(1)>text</details>',
    '<a href="java&#x73;cript:alert(1)">link</a><object data="x"></object><embed src="x">',
    '<svg><foreignObject><img src=x onerror=alert(1)></foreignObject></svg>',
    '<math><mtext><table><mglyph><style><!--</style><img title="--><img src=x onerror=alert(1)>">',
    '<iframe src="https://evil.example" srcdoc="<script>alert(1)</script>"></iframe>',
  ])('blocks active content through template and rendered DOM reparsing: %s', (html) => {
    const prepared = withLazyImages(html);
    const body = document.createElement('div');
    body.innerHTML = sanitizeArticleHTML(prepared);
    expect(body.querySelector('script, noscript, iframe, object, embed, svg, style')).toBeNull();
    for (const element of body.querySelectorAll('*')) {
      for (const attr of element.attributes) {
        expect(attr.name).not.toMatch(/^on/i);
        expect(attr.value).not.toMatch(/^\s*javascript:/i);
      }
    }
  });

  it('preserves reader formatting, math, lazy images and sandboxed video embeds', () => {
    const html = sanitizeArticleHTML('<pre><code class="language-go">x</code></pre><math><mi>x</mi></math><img src="https://example.org/a.jpg" loading="lazy"><iframe src="https://www.youtube.com/embed/123" srcdoc="bad" sandbox="allow-top-navigation"></iframe>');
    expect(html).toContain('class="language-go"');
    expect(html).toContain('<math><mi>x</mi></math>');
    expect(html).toContain('loading="lazy"');
    expect(html).toContain('referrerpolicy="no-referrer"');
    expect(html).toContain('sandbox="allow-scripts allow-same-origin allow-presentation"');
    expect(html).not.toContain('srcdoc');
    expect(html).not.toContain('allow-top-navigation');
    expect(sanitizeArticleHTML(html)).toBe(html);
  });

  it('rejects lookalike video hosts and unsafe protocols', () => {
    expect(sanitizeArticleHTML('<iframe src="https://youtube.com.evil.example/embed/x"></iframe><iframe src="javascript:alert(1)"></iframe>')).toBe('');
  });
});
