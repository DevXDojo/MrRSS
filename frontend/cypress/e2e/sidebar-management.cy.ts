/// <reference types="cypress" />

function setup() {
  const settings: Record<string, string> = {
    language: 'en-US',
    theme: 'light',
    layout_mode: 'normal',
    default_view_mode: 'rendered',
    translation_enabled: 'false',
    summary_enabled: 'false',
    full_text_fetch_enabled: 'false',
    update_check_enabled: 'false',
    shortcuts_enabled: 'true',
    sidebar_sort_mode: 'manual',
    sidebar_category_order: '[]',
    sidebar_pinned_items: '[]',
    rules: '[]',
  };
  const feeds = [
    {
      id: 1,
      title: 'One feed',
      category: 'Alpha/One',
      position: 0,
      latest_article_time: '2026-09-01T00:00:00Z',
    },
    {
      id: 2,
      title: 'Two feed',
      category: 'Alpha/Two',
      position: 0,
      latest_article_time: '2026-09-02T00:00:00Z',
    },
    {
      id: 3,
      title: 'Beta feed',
      category: 'Beta',
      position: 0,
      latest_article_time: '2026-09-03T00:00:00Z',
    },
    {
      id: 4,
      title: 'Zulu feed',
      category: 'Zulu',
      position: 0,
      latest_article_time: '2026-09-04T00:00:00Z',
    },
    {
      id: 5,
      title: 'Other feed',
      category: 'Alpha/One',
      position: 1,
      latest_article_time: '2026-09-01T00:00:00Z',
    },
  ].map((feed) => ({ ...feed, url: `https://example.com/feed/${feed.id}` }));
  cy.intercept('/api/**', { statusCode: 200, body: {} });
  cy.intercept('GET', '/api/settings', (req) => req.reply(settings));
  cy.intercept('POST', '/api/settings', (req) => {
    Object.assign(settings, req.body);
    req.reply({ success: true });
  }).as('saveSettings');
  cy.intercept('GET', '/api/feeds', (req) => req.reply(feeds)).as('feeds');
  cy.intercept('GET', '/api/tags', []);
  cy.intercept('GET', '/api/saved-filters', []);
  cy.intercept({ method: 'GET', pathname: '/api/articles' }, (req) =>
    req.reply([
      {
        id: 1,
        feed_id: 1,
        feed_title: 'One feed',
        title: 'Read favorite',
        url: 'https://example.com/article',
        published_at: '2026-09-01T00:00:00Z',
        is_read: true,
        is_favorite: true,
        is_hidden: false,
        is_read_later: false,
      },
    ])
  ).as('articles');
  cy.intercept('GET', '/api/articles/unread-counts', {
    total: 13,
    feed_counts: { 1: 1, 2: 2, 3: 7, 4: 3 },
  });
  cy.intercept('GET', '/api/articles/filter-counts', {
    favorites: { 1: 2, 5: 1 },
    favorites_unread: {},
    unread: { 1: 1, 2: 2, 3: 7, 4: 3 },
  });
  cy.intercept('GET', '/api/articles/content*', { content: '<p>Saved article</p>', cached: true });
  cy.intercept('GET', '/api/progress', { is_running: false });
  cy.intercept('POST', '/api/feeds/reorder', { statusCode: 200, body: {} }).as('reorderFeed');
  cy.intercept('POST', '/api/feeds/category', { statusCode: 200, body: { affected: 1 } }).as(
    'categoryAction'
  );
  cy.visit('/', {
    onBeforeLoad(win) {
      win.localStorage.setItem('FeedListPinned', 'true');
      win.localStorage.setItem('FeedListExpanded', 'true');
      win.localStorage.setItem('showOnlyUnread', 'true');
    },
  });
  cy.wait(['@feeds', '@articles']);
  cy.get('.categories-list').should('be.visible');
}
const header = (path: string) =>
  `.category-container[data-category-path="${path}"] > .category-header`;
const roots = '.categories-list > .category-container';

function drag(source: string, target: string, before = true) {
  cy.window().then((win) => {
    const from = win.document.querySelector(source)!;
    const to = win.document.querySelector(target)!;
    expect(from, source).not.to.be.null;
    expect(to, target).not.to.be.null;
    const transfer = new win.DataTransfer();
    from.dispatchEvent(
      new win.DragEvent('dragstart', { bubbles: true, cancelable: true, dataTransfer: transfer })
    );
    const rect = to.getBoundingClientRect();
    const options = {
      bubbles: true,
      cancelable: true,
      dataTransfer: transfer,
      clientX: rect.left + 8,
      clientY: before ? rect.top + 2 : rect.bottom - 2,
    };
    to.dispatchEvent(new win.DragEvent('dragover', options));
    to.dispatchEvent(new win.DragEvent('drop', options));
    from.dispatchEvent(new win.DragEvent('dragend', { bubbles: true, dataTransfer: transfer }));
  });
}

describe('Sidebar and subscription management', () => {
  it('persists root and nested category drag order without moving their feeds', () => {
    setup();
    cy.get('button[title="Edit"]').click();
    drag(`${header('Beta')} [draggable]`, header('Alpha'));
    cy.wait('@saveSettings').its('request.body.sidebar_category_order').should('contain', 'Beta');
    cy.get(roots).first().should('have.attr', 'data-category-path', 'Beta');
    drag(`${header('Alpha/Two')} [draggable]`, header('Alpha/One'));
    cy.wait('@saveSettings');
    cy.get('.category-container[data-category-path="Alpha"] > .feeds-list > .category-container')
      .first()
      .should('have.attr', 'data-category-path', 'Alpha/Two');
    cy.reload();
    cy.get(roots).first().should('have.attr', 'data-category-path', 'Beta');
    cy.get('.category-container[data-category-path="Alpha/Two"] [data-feed-id="2"]').should(
      'exist'
    );
  });

  it('uses the correct nested folder when dropping a feed over an icon', () => {
    setup();
    cy.get('button[title="Edit"]').click();
    drag('[data-feed-id="5"] .drag-handle', '[data-feed-id="1"] svg');
    cy.wait('@reorderFeed')
      .its('request.body')
      .should('deep.equal', { feed_id: 5, category: 'Alpha/One', position: 0 });
  });

  it('sorts by count and recency while preserving pinned categories', () => {
    setup();
    cy.get('select[aria-label="Sort categories and feeds"]').select('count_desc');
    cy.wait('@saveSettings');
    cy.get(roots).first().should('have.attr', 'data-category-path', 'Beta');
    cy.get(header('Zulu')).rightclick();
    cy.contains('Pin to top of this level').click();
    cy.wait('@saveSettings');
    cy.get(roots).first().should('have.attr', 'data-category-path', 'Zulu');
    cy.get('select[aria-label="Sort categories and feeds"]').select('name_asc');
    cy.wait('@saveSettings');
    cy.get(roots).first().should('have.attr', 'data-category-path', 'Zulu');
    cy.reload();
    cy.get(roots).first().should('have.attr', 'data-category-path', 'Zulu');
    cy.get(header('Zulu')).rightclick();
    cy.contains(/^Unpin$/).click();
    cy.wait('@saveSettings');
    cy.get('select[aria-label="Sort categories and feeds"]').select('latest');
    cy.wait('@saveSettings');
    cy.get(roots).first().should('have.attr', 'data-category-path', 'Zulu');
  });

  it('shows total favorite counts and keeps read favorites visible', () => {
    setup();
    cy.get('button[title^="Favorites"]').click();
    cy.get(roots).should('have.length', 1);
    cy.get(header('Alpha')).find('.unread-badge').should('have.text', '3');
    cy.get(header('Alpha/One')).find('.unread-badge').should('have.text', '3');
    cy.get('[data-feed-id="1"]').find('.unread-badge').should('have.text', '2');
    cy.get('[data-article-id="1"]').should('contain', 'Read favorite');
    cy.get('button[title="Show only unread"]').should('not.exist');
    cy.get('@articles.all').then((requests) => {
      const favorites = requests.filter((request: any) =>
        request.request.url.includes('filter=favorites')
      );
      expect(favorites).to.have.length.greaterThan(0);
      expect(favorites.at(-1)!.request.url).not.to.contain('only_unread=true');
    });
  });

  it('confirms category scope before dissolving or unsubscribing', () => {
    setup();
    cy.get(header('Beta')).rightclick();
    cy.contains('Dissolve category').click();
    cy.contains('Keep all 1 subscriptions').should('be.visible');
    cy.contains('button', /^Cancel$/).click();
    cy.get('@categoryAction.all').should('have.length', 0);
    cy.get(header('Beta')).rightclick();
    cy.contains('Unsubscribe category').click();
    cy.contains('including favorites').should('be.visible');
    cy.contains('button', /^Confirm$/).click();
    cy.wait('@categoryAction')
      .its('request.body')
      .should('deep.equal', { category: 'Beta', action: 'unsubscribe' });
  });

  it('imports rule backups without applying them to existing articles and rejects invalid backups', () => {
    setup();
    cy.get('button[title^="Settings"]').click();
    cy.contains('button', /^Rules$/).click();
    cy.intercept('POST', '/api/rules/apply', () => {
      throw new Error('Import must not apply rules to old articles');
    });
    const backup = {
      format: 'mrrss-rules',
      version: 1,
      rules: [
        {
          id: 1,
          name: 'Imported favorites',
          enabled: true,
          conditions: [
            {
              id: 1,
              field: 'article_title',
              value: 'tutorial',
              values: [],
              negate: false,
              operator: 'contains',
            },
          ],
          actions: ['favorite'],
        },
      ],
    };
    cy.get('input[type="file"][accept=".json,application/json"]').selectFile(
      { contents: Cypress.Buffer.from(JSON.stringify(backup)), fileName: 'rules.json' },
      { force: true }
    );
    cy.contains('Append 1 rules').should('be.visible');
    cy.contains('button', /^Confirm$/).click();
    cy.contains('Imported favorites').should('be.visible');
    cy.get('input[type="file"][accept=".json,application/json"]').selectFile(
      { contents: Cypress.Buffer.from('{}'), fileName: 'invalid.json' },
      { force: true }
    );
    cy.contains('Invalid or unsupported MrRSS rules backup').should('be.visible');
    cy.contains('Imported favorites').should('be.visible');
    cy.contains('button', 'Export rules').click();
    cy.readFile('cypress/downloads/mrrss-rules.json')
      .its('rules.0.name')
      .should('equal', 'Imported favorites');
  });
});
