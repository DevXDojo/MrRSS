/// <reference types="cypress" />

const image = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw==';
const article = {
  id: 1,
  feed_id: 1,
  feed_title: 'Reading Feed',
  title: 'English title',
  url: 'https://example.com/article',
  published_at: '2026-09-01T00:00:00Z',
  translated_title: '',
  is_read: false,
  is_favorite: false,
  is_hidden: false,
  is_read_later: false,
  image_url: image,
};

function setup(overrides: Record<string, string> = {}, feedMode = 'global') {
  const settings: Record<string, string> = {
    language: 'en-US',
    theme: 'light',
    layout_mode: 'normal',
    default_view_mode: 'rendered',
    translation_enabled: 'true',
    translation_provider: 'ai',
    translation_only_mode: 'false',
    translation_trigger_mode: 'manual',
    target_language: 'zh-CN',
    summary_enabled: 'false',
    full_text_fetch_enabled: 'false',
    update_check_enabled: 'false',
    image_gallery_enabled: 'true',
    shortcuts_enabled: 'true',
    ...overrides,
  };
  cy.intercept('/api/**', { statusCode: 200, body: {} });
  cy.intercept('GET', '/api/settings', (req) => req.reply(settings));
  cy.intercept('POST', '/api/settings', (req) => {
    Object.entries(req.body).forEach(([key, value]) => {
      settings[key] = String(value);
    });
    req.reply({ success: true });
  });
  cy.intercept('GET', '/api/feeds', [
    {
      id: 1,
      title: article.feed_title,
      url: 'https://example.com/feed',
      category: '',
      article_view_mode: feedMode,
    },
  ]).as('feeds');
  cy.intercept('GET', '/api/tags', []);
  cy.intercept('GET', '/api/saved-filters', []);
  cy.intercept({ method: 'GET', pathname: '/api/articles' }, [article]).as('articles');
  cy.intercept('GET', '/api/articles/images*', [article]).as('images');
  cy.intercept('GET', '/api/articles/extract-images*', { images: [image] });
  cy.intercept('GET', '/api/articles/unread-counts', {});
  cy.intercept('GET', '/api/articles/filter-counts', {});
  cy.intercept('GET', '/api/progress', { is_running: false });
  cy.intercept('GET', '/api/articles/content*', {
    content:
      '<p>First paragraph with enough English words to translate.</p><p>Second paragraph stays unchanged.</p>',
    cached: true,
  }).as('content');
  cy.intercept('POST', '/api/browser/open', { statusCode: 200, body: {} }).as('openBrowser');
  cy.visit('/');
  cy.wait(['@feeds', '@articles']);
}

function openArticle() {
  cy.get('[data-article-id="1"]').click();
  cy.wait('@content');
  cy.get('.prose-content p').should('have.length', 2);
}

describe('Reading interactions', () => {
  it('translates only the requested title or paragraph in manual mode, and retries failures', () => {
    let calls = 0;
    let paragraphCalls = 0;
    setup();
    cy.intercept('POST', '/api/articles/translate', () => {
      throw new Error('Manual mode must not translate list titles automatically');
    });
    cy.intercept('POST', '/api/articles/translate-text', (req) => {
      calls++;
      if (req.body.text === article.title) {
        req.alias = 'titleTranslation';
        req.reply({ translated_text: '翻译后的标题', skipped: false });
      } else {
        paragraphCalls++;
        expect(req.body.text).to.contain('First paragraph');
        expect(req.body.text).not.to.contain('Second paragraph');
        req.alias = 'paragraphTranslation';
        req.reply(
          paragraphCalls === 1
            ? { statusCode: 500, body: { error: 'temporary failure' } }
            : { translated_text: '第一段已翻译', skipped: false }
        );
      }
    });
    openArticle();
    cy.contains('Translate on demand')
      .should('be.visible')
      .then(() => expect(calls).to.equal(0));
    cy.get('button[title="Translate title"]').click();
    cy.wait('@titleTranslation');
    cy.contains('翻译后的标题').should('be.visible');
    cy.get('.prose-content p').first().click({ ctrlKey: true });
    cy.wait('@paragraphTranslation');
    cy.get('.translation-text').should('not.exist');
    cy.get('.prose-content p').first().click({ ctrlKey: true });
    cy.wait('@paragraphTranslation');
    cy.get('.translation-text').should('have.length', 1).and('contain', '第一段已翻译');
    cy.get('.prose-content p').eq(1).should('contain', 'Second paragraph stays unchanged');
    cy.then(() => expect(calls).to.equal(3));
  });

  it('searches a captured text selection and copies the article link', () => {
    setup();
    openArticle();
    cy.window().then((win) => {
      const paragraph = win.document.querySelector('.prose-content p')!;
      const range = win.document.createRange();
      range.selectNodeContents(paragraph);
      win.getSelection()!.removeAllRanges();
      win.getSelection()!.addRange(range);
    });
    cy.get('.prose-content p').first().trigger('contextmenu');
    cy.contains('Search with Google').should('be.visible');
    cy.contains('Search with Bing').click();
    cy.wait('@openBrowser')
      .its('request.body.url')
      .should(
        'equal',
        'https://www.bing.com/search?q=' +
          encodeURIComponent('First paragraph with enough English words to translate.')
      );
    cy.window().then((win) => {
      cy.stub(win.navigator.clipboard, 'writeText').resolves().as('copyText');
    });
    cy.get('button[aria-label="Copy Link"]').click();
    cy.get('@copyText').should('have.been.calledWith', article.url);
  });

  it('keeps icon hit targets on tooltip buttons and reflects changed shortcuts', () => {
    setup();
    cy.get('button[title^="Settings"]')
      .should('have.attr', 'title', 'Settings (,)')
      .find('svg')
      .then(($icon) => {
        const icon = $icon[0];
        const rect = icon.getBoundingClientRect();
        expect(
          icon.ownerDocument.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2)
            ?.tagName
        ).to.equal('BUTTON');
      });
    cy.window().then((win) =>
      win.dispatchEvent(
        new CustomEvent('shortcuts-changed', { detail: { shortcuts: { openSettings: 'Ctrl+,' } } })
      )
    );
    cy.get('button[title^="Settings"]').should('have.attr', 'title', 'Settings (Ctrl+,)');
  });

  for (const [feedMode, globalMode] of [
    ['external', 'rendered'],
    ['global', 'external'],
  ]) {
    it(`opens gallery articles externally with ${feedMode} feed / ${globalMode} global preference`, () => {
      setup({ default_view_mode: globalMode }, feedMode);
      cy.get('[title="Multimedia Gallery"]').click();
      cy.wait('@images');
      cy.contains('English title').click();
      cy.wait('@openBrowser').its('request.body.url').should('equal', article.url);
      cy.get('iframe').should('not.exist');
      cy.get('[role="dialog"][aria-modal="true"]').should('not.exist');
    });
  }

  it('allows an explicit rendered feed preference to override global external mode', () => {
    setup({ default_view_mode: 'external' }, 'rendered');
    cy.get('[title="Multimedia Gallery"]').click();
    cy.wait('@images');
    cy.contains('English title').click();
    cy.get('[role="dialog"][aria-modal="true"]').should('be.visible');
    cy.get('@openBrowser.all').should('have.length', 0);
  });
});
