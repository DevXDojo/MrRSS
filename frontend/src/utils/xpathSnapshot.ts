// Only our nonce-bearing inspector executes. The iframe has an opaque origin,
// no forms/popups/navigation privileges, and receives already sanitized markup.
export function createXPathSnapshot(html: string, baseURL: string, token: string): string {
  const escape = (value: string) =>
    value.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;');
  const csp = `default-src 'none'; script-src 'nonce-${token}'; style-src 'unsafe-inline' http: https:; img-src http: https: data:; font-src http: https: data:; connect-src 'none'; form-action 'none'; base-uri http: https:`;
  const head = `<meta http-equiv="Content-Security-Policy" content="${csp}"><meta name="referrer" content="no-referrer"><base href="${escape(baseURL)}">`;
  const script = `<script nonce="${token}">
  (() => {
    const token = ${JSON.stringify(token)};
    const overlay = document.createElement('div');
    overlay.style.cssText = 'position:fixed;pointer-events:none;z-index:2147483647;border:2px solid #2563eb;background:rgba(37,99,235,.12);box-sizing:border-box;display:none';
    document.body.append(overlay);
    const groupLayer = document.createElement('div');
    groupLayer.style.cssText = 'position:fixed;inset:0;pointer-events:none;z-index:2147483646';
    document.body.append(groupLayer);
    const indexed = new Map(Array.from(document.querySelectorAll('[data-mrrss-path]')).map(el => [el.dataset.mrrssPath, el]));
    let matches = [];
    let target = null;
    let selected = null;
    function draw(el) {
      if (!el) { overlay.style.display = 'none'; return; }
      const r = el.getBoundingClientRect();
      Object.assign(overlay.style, {display:'block',left:r.left+'px',top:r.top+'px',width:r.width+'px',height:r.height+'px'});
    }
    function drawMatches() {
      groupLayer.replaceChildren();
      for (const el of matches) {
        const r = el.getBoundingClientRect();
        if (!r.width || !r.height || r.bottom < 0 || r.top > innerHeight || r.right < 0 || r.left > innerWidth) continue;
        const mark = document.createElement('div');
        mark.style.cssText = 'position:fixed;pointer-events:none;border:1px solid #16a34a;background:rgba(22,163,74,.05);box-sizing:border-box';
        Object.assign(mark.style, {left:r.left+'px',top:r.top+'px',width:r.width+'px',height:r.height+'px'});
        groupLayer.append(mark);
        if (groupLayer.childElementCount >= 200) break;
      }
    }
    document.addEventListener('mousemove', e => { target = e.target.closest('[data-mrrss-path]'); draw(target); }, true);
    document.addEventListener('mouseleave', () => { target = null; draw(selected); });
    document.addEventListener('click', e => {
      e.preventDefault(); e.stopImmediatePropagation();
      const el = e.target.closest('[data-mrrss-path]');
      if (el) { selected = el; draw(el); parent.postMessage({type:'mrrss-xpath-pick',token,path:el.dataset.mrrssPath}, '*'); }
    }, true);
    document.addEventListener('submit', e => e.preventDefault(), true);
    document.addEventListener('keydown', e => {
      if (e.key === 'Enter') { e.preventDefault(); parent.postMessage({type:'mrrss-xpath-confirm',token}, '*'); }
    });
    let redrawPending = false;
    function redraw() {
      if (redrawPending) return;
      redrawPending = true;
      requestAnimationFrame(() => { redrawPending = false; draw(target || selected); drawMatches(); });
    }
    window.addEventListener('scroll', redraw, true);
    window.addEventListener('resize', redraw);
    window.addEventListener('load', redraw);
    new ResizeObserver(redraw).observe(document.body);
    window.addEventListener('message', e => {
      if (e.source !== parent || e.data?.token !== token || e.data?.type !== 'mrrss-xpath-highlight') return;
      selected = indexed.get(e.data.path);
      matches = Array.isArray(e.data.matches) ? e.data.matches.map(path => indexed.get(path)).filter(Boolean) : [];
      target = null;
      draw(selected);
      drawMatches();
    });
    parent.postMessage({type:'mrrss-xpath-ready',token}, '*');
  })();
  </script>`;
  return html
    .replace(/<head[^>]*>/i, (match) => match + head)
    .replace(/<\/body>/i, script + '</body>');
}
