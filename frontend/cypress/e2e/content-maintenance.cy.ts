/// <reference types="cypress" />
const entries = Array.from({ length: 60 }, (_, index) => ({
  id: index + 1,
  feed_id: 1,
  feed_title: 'Content Feed',
  title: `Article ${index + 1}`,
  url: `https://example.org/article/${index + 1}`,
  published_at: new Date(Date.UTC(2026, 8, 5) - index * 60000).toISOString(),
  is_read: false,
  is_favorite: false,
  is_read_later: false,
  translated_title: '',
  image_url: `https://example.org/photo/${index + 1}.svg`,
}));
function setup(imageMode = false, auto = true) {
  cy.intercept('/api/**', { statusCode: 200, body: {} });
  cy.intercept('GET', '/api/settings', {
    language: 'en-US',
    theme: 'light',
    default_view_mode: 'rendered',
    translation_enabled: 'false',
    summary_enabled: 'false',
    full_text_fetch_enabled: 'true',
    auto_show_all_content: String(auto),
    image_gallery_enabled: 'true',
    media_cache_enabled: 'false',
    update_check_enabled: 'false',
  });
  cy.intercept('GET', '/api/feeds', [
    {
      id: 1,
      title: 'Content Feed',
      url: 'https://example.org/feed',
      category: '',
      is_image_mode: imageMode,
    },
  ]).as('feeds');
  cy.intercept('GET', '/api/tags', []);
  cy.intercept('GET', '/api/saved-filters', []);
  cy.intercept({ method: 'GET', pathname: '/api/articles' }, entries.slice(0, 2)).as('articles');
  cy.intercept('GET', '/api/progress', { is_running: false });
  cy.intercept('GET', '/api/articles/content*', { content: '', cached: false }).as('content');
  cy.intercept('GET', '/api/articles/unread-counts', {});
  cy.intercept('GET', '/api/articles/filter-counts', {});
  cy.intercept('GET', '/api/articles/extract-images*', { images: [] });
  cy.intercept('GET', '/api/articles/images*', (req) => {
    const page = Number(new URL(req.url).searchParams.get('page') || 1);
    req.reply(entries.slice((page - 1) * 30, page * 30));
  }).as('images');
  cy.intercept('GET', '/api/media/proxy*', (req) => {
    const id = Number(
      (new URL(req.url).searchParams.get('url') || '').match(/photo\/(\d+)/)?.[1] || 1
    );
    req.reply({
      headers: { 'Content-Type': 'image/svg+xml' },
      body: `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="${id % 3 === 0 ? 900 : 240}"><rect width="100%" height="100%" fill="#709ca8"/></svg>`,
    });
  });
  cy.visit('/', {
    onBeforeLoad(win) {
      win.localStorage.setItem('FeedListPinned', 'true');
      win.localStorage.setItem('FeedListExpanded', 'true');
    },
  });
  cy.wait(['@feeds', '@articles']);
}
describe('Content maintenance', () => {
  it('automatically fetches ordinary empty RSS entries and keeps the full article', () => {
    setup();
    let calls = 0;
    cy.intercept('POST', '/api/articles/fetch-full*', (req) => {
      calls++;
      req.reply({ content: '<p>Recovered complete article.</p>' });
    }).as('full');
    cy.get('[data-article-id="1"]').click();
    cy.wait('@full');
    cy.get('.prose-content').should('contain', 'Recovered complete article.');
    cy.then(() => expect(calls).to.equal(1));
  });
  it('keeps the current article when a previous full-text request finishes later', () => {
    setup();
    cy.intercept('POST', '/api/articles/fetch-full*', (req) => {
      const id = new URL(req.url).searchParams.get('id');
      req.reply({
        delay: id === '1' ? 900 : 0,
        body: { content: `<p>Full content for ${id}</p>` },
      });
    });
    cy.get('[data-article-id="1"]').click();
    cy.wait('@content');
    cy.get('[data-article-id="2"]').click();
    cy.get('.prose-content').should('contain', 'Full content for 2');
    cy.wait(1000);
    cy.get('.prose-content')
      .should('contain', 'Full content for 2')
      .and('not.contain', 'Full content for 1');
  });
  it('provides manual full text for empty RSS when automatic expansion is disabled', () => {
    setup(false, false);
    cy.intercept('POST', '/api/articles/fetch-full*', { content: '<p>Manual full text</p>' }).as(
      'full'
    );
    cy.get('[data-article-id="1"]').click();
    cy.wait('@content');
    cy.contains('button', 'Fetch Full Article').click();
    cy.wait('@full');
    cy.get('.prose-content').should('contain', 'Manual full text');
  });
  it('saves selectors and a replacement Cookie, then hides the credential', () => {
    setup();
    cy.intercept('GET', '/api/feeds/content-options*', {
      content_selector: '',
      remove_selector: '',
      cookie_origin: 'https://example.org',
      has_cookie: true,
    }).as('options');
    cy.intercept('POST', '/api/feeds/content-options*', (req) => {
      expect(req.body.content_selector).to.equal('article .body');
      expect(req.body.remove_selector).to.equal('.ads');
      expect(req.body.cookie).to.equal('session=replacement');
      req.reply({ ...req.body, cookie: undefined, has_cookie: true });
    }).as('saveOptions');
    cy.get('[data-feed-id="1"]').first().trigger('contextmenu');
    cy.contains('Edit Subscription').click();
    cy.contains('Show Advanced Settings').click();
    cy.wait('@options');
    cy.get('[data-testid="content-selector"]').type('article .body');
    cy.get('[data-testid="remove-selector"]').type('.ads');
    cy.get('[data-testid="feed-cookie"]').type('session=replacement');
    cy.contains('button', 'Save extraction settings').click();
    cy.wait('@saveOptions');
    cy.get('[data-testid="feed-cookie"]').should('have.value', '');
    cy.contains('Extraction settings saved').should('be.visible');
  });
  it('loads many lazy images without overlapping cards or losing items on resize', () => {
    setup(true, false);
    cy.get('[title="Multimedia Gallery"]').click();
    cy.wait('@images');
    cy.get('[data-testid="gallery-scroll"]').scrollTo('bottom');
    cy.get('[data-gallery-article]').should('have.length', 60);
    cy.viewport(960, 720);
    cy.get('[data-gallery-article]')
      .should('have.length', 60)
      .then(($cards) => {
        const ids = [...$cards].map((card) => card.getAttribute('data-gallery-article'));
        expect(new Set(ids).size).to.equal(60);
        for (const card of $cards) {
          const next = card.nextElementSibling;
          if (next)
            expect(card.getBoundingClientRect().bottom).to.be.at.most(
              next.getBoundingClientRect().top + 1
            );
          expect(card.getBoundingClientRect().width).to.be.greaterThan(50);
        }
      });
  });
});
