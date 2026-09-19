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
    let target = null;
    let selected = null;
    function draw(el) {
      if (!el) { overlay.style.display = 'none'; return; }
      const r = el.getBoundingClientRect();
      Object.assign(overlay.style, {display:'block',left:r.left+'px',top:r.top+'px',width:r.width+'px',height:r.height+'px'});
    }
    document.addEventListener('mousemove', e => { target = e.target.closest('[data-mrrss-path]'); draw(target); }, true);
    document.addEventListener('mouseleave', () => draw(selected), true);
    document.addEventListener('click', e => {
      e.preventDefault(); e.stopImmediatePropagation();
      const el = e.target.closest('[data-mrrss-path]');
      if (el) { selected = el; draw(el); parent.postMessage({type:'mrrss-xpath-pick',token,path:el.dataset.mrrssPath}, '*'); }
    }, true);
    document.addEventListener('submit', e => e.preventDefault(), true);
    window.addEventListener('scroll', () => draw(target || selected), true);
    window.addEventListener('resize', () => draw(selected));
    window.addEventListener('message', e => {
      if (e.source !== parent || e.data?.token !== token || e.data?.type !== 'mrrss-xpath-highlight') return;
      selected = Array.from(document.querySelectorAll('[data-mrrss-path]')).find(el => el.dataset.mrrssPath === e.data.path);
      draw(selected);
    });
    parent.postMessage({type:'mrrss-xpath-ready',token}, '*');
  })();
  </script>`;
  return html
    .replace(/<head[^>]*>/i, (match) => match + head)
    .replace(/<\/body>/i, script + '</body>');
}
