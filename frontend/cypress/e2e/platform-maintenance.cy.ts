/// <reference types="cypress" />
export {};
function setup() {
  cy.intercept('/api/**', { statusCode: 200, body: {} });
  cy.intercept('GET', '/api/settings', {
    language: 'en-US',
    theme: 'light',
    custom_css_file: 'theme.css',
    default_view_mode: 'rendered',
    translation_enabled: 'false',
    summary_enabled: 'false',
    full_text_fetch_enabled: 'false',
    auto_show_all_content: 'false',
    update_check_enabled: 'false',
  });
  cy.intercept('GET', '/api/custom-css', {
    headers: { 'Content-Type': 'text/css' },
    body: ':root { --accent-color: rgb(101, 41, 151); }',
  }).as('css');
  cy.intercept('GET', '/api/feeds', [
    { id: 1, title: 'Local Feed', url: 'https://example.org/feed', category: '' },
  ]).as('feeds');
  cy.intercept('GET', '/api/tags', []);
  cy.intercept('GET', '/api/saved-filters', []);
  cy.intercept({ method: 'GET', pathname: '/api/articles' }, [
    {
      id: 1,
      feed_id: 1,
      feed_title: 'Local Feed',
      title: 'Local article',
      url: 'https://example.org/1',
      published_at: '2026-09-05T00:00:00Z',
      is_read: false,
      is_favorite: false,
      is_read_later: false,
    },
  ]).as('articles');
  cy.intercept('GET', '/api/articles/content*', { content: '<p>Local reader content</p>' });
  cy.intercept('GET', '/api/progress', { is_running: false });
  cy.intercept('https://fonts.googleapis.com/**', { forceNetworkError: true });
  cy.intercept('https://unpkg.com/**', { forceNetworkError: true });
  cy.visit('/', {
    onBeforeLoad(win) {
      win.localStorage.setItem('FeedListPinned', 'true');
      win.localStorage.setItem('FeedListExpanded', 'true');
    },
  });
  cy.wait(['@feeds', '@articles', '@css']);
}
describe('Platform and application maintenance', () => {
  it('starts without external UI resources and keeps the theme across reader navigation', () => {
    setup();
    cy.get('head script[src^="https://"], head link[rel="stylesheet"][href^="https://"]').should(
      'not.exist'
    );
    cy.get('style[data-custom-css="application"]')
      .should('have.length', 1)
      .and('contain', '101, 41, 151');
    cy.get('[data-article-id="1"]').click();
    cy.get('.prose-content').should('contain', 'Local reader content');
    cy.get('[data-feed-id="1"]').click();
    cy.wait('@articles');
    cy.get('style[data-custom-css="application"]')
      .should('have.length', 1)
      .and('contain', '101, 41, 151');
    cy.document().then((doc) =>
      expect(
        doc
          .defaultView!.getComputedStyle(doc.documentElement)
          .getPropertyValue('--accent-color')
          .trim()
      ).to.equal('rgb(101, 41, 151)')
    );
  });
  it('replaces a theme while no reader is open', () => {
    setup();
    cy.get('.prose-content').should('not.exist');
    cy.intercept('GET', '/api/custom-css', {
      headers: { 'Content-Type': 'text/css' },
      body: ':root { --accent-color: rgb(1, 2, 3); }',
    }).as('replacement');
    cy.window().then((win) => win.dispatchEvent(new Event('custom-css-changed')));
    cy.wait('@replacement');
    cy.get('style[data-custom-css="application"]')
      .should('have.length', 1)
      .and('contain', '1, 2, 3');
  });
});
