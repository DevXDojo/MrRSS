import { describe, it, expect } from 'vitest';
import { createXPathSnapshot } from './xpathSnapshot';
describe('isolated XPath snapshot', () => {
  it('places the CSP before page resources and enables only its nonce-bearing inspector', () => {
    const html = createXPathSnapshot(
      '<html><head><link rel="stylesheet" href="/site.css"></head><body><p>Page</p></body></html>',
      'https://example.com/"unsafe',
      'abc123'
    );
    expect(html.indexOf('Content-Security-Policy')).toBeLessThan(html.indexOf('<link'));
    expect(html).toContain("script-src 'nonce-abc123'");
    expect(html).toContain("connect-src 'none'");
    expect(html).toContain("form-action 'none'");
    expect(html).toContain('href="https://example.com/&quot;unsafe"');
    expect(html).toContain('<script nonce="abc123">');
    expect(html).toContain('e.source !== parent');
    expect(html).toContain('e.preventDefault()');
  });
});
