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
  it('shows the bound article and requires an explicit new chat before sending from another reader article', () => {
    let sends = 0;
    const firstArticle = {
      ...article,
      title: 'LongUnbrokenArticleTitle'.repeat(80),
      feed_title: 'LongUnbrokenFeedSource'.repeat(80),
    };
    const secondArticle = {
      ...article,
      id: 2,
      title: 'Second article',
      feed_title: 'Second source',
      url: 'https://example.com/second',
    };
    const sessions = [{ id: 11, article_id: 1, title: 'First article session', message_count: 1 }];
    setup({ ai_chat_enabled: 'true', translation_enabled: 'false' });
    cy.intercept({ method: 'GET', pathname: '/api/articles' }, [firstArticle, secondArticle]).as(
      'contextArticles'
    );
    cy.intercept('GET', '/api/articles/content*', (req) => {
      req.reply({
        content:
          Number(req.query.id) === 2
            ? '<p>Second article body.</p><p>Second context.</p>'
            : '<p>First article body.</p><p>First context.</p>',
        cached: true,
      });
    }).as('contextContent');
    cy.intercept('GET', '/api/ai/profiles', []);
    cy.intercept('GET', '/api/ai/chat/sessions*', (req) => {
      req.reply(sessions.filter((session) => session.article_id === Number(req.query.article_id)));
    });
    cy.intercept('GET', '/api/ai/chat/messages*', (req) => {
      req.reply(
        Number(req.query.session_id) === 11
          ? [{ id: 1, role: 'user', content: 'Question about first article', created_at: '' }]
          : { statusCode: 500, body: {} }
      );
    }).as('contextMessages');
    cy.intercept('POST', '/api/ai/chat/session/create', (req) => {
      expect(req.body.article_id).to.equal(2);
      const session = { id: 22, article_id: 2, title: req.body.title, message_count: 0 };
      sessions.push(session);
      req.reply({ delay: 300, body: session });
    }).as('newContext');
    cy.intercept('POST', '/api/ai-chat', (req) => {
      sends++;
      expect(req.body.session_id).to.equal(22);
      expect(req.body.article_id).to.equal(2);
      expect(req.body.article_title).to.equal(secondArticle.title);
      expect(req.body.article_url).to.equal(secondArticle.url);
      expect(req.body.article_content).to.contain('Second article body');
      expect(req.body.article_content).not.to.contain('First article body');
      expect(req.body.messages).to.have.length(1);
      req.reply({ response: 'Answer for second article', session_id: 22 });
    }).as('contextChat');
    cy.reload();
    cy.wait('@contextArticles');
    cy.get('[data-article-id="1"]').click();
    cy.wait('@contextContent');
    cy.get('button[title="AI Chat"]').click();
    cy.wait('@contextMessages');
    cy.get('[data-testid="chat-context-article"]')
      .should('contain', firstArticle.title)
      .and('contain', firstArticle.feed_title);
    cy.get('[data-testid="chat-session-switcher"]').click();
    cy.get('[data-testid="chat-context-article"]').should('be.visible');
    cy.get('[data-session-id="11"]').should('be.visible').click();
    cy.wait('@contextMessages');
    cy.get('input[placeholder="Type a message..."]').type('Draft for first article');
    cy.get('[data-article-id="2"]').click();
    cy.wait('@contextContent');
    cy.get('[data-testid="chat-context-article"]').should(
      'have.attr',
      'data-context-article-id',
      '1'
    );
    cy.contains('.chat-panel', 'Question about first article').should('be.visible');
    cy.get('input[placeholder="Type a message..."]')
      .should('be.disabled')
      .trigger('keydown', { key: 'Enter', force: true });
    cy.then(() => expect(sends).to.equal(0));
    cy.get('.chat-panel').invoke('css', 'width', '420px');
    cy.get('.chat-panel').should(($panel) => {
      const panel = $panel[0];
      const panelRect = panel.getBoundingClientRect();
      const card = panel.querySelector('[data-testid="chat-context-article"]')!;
      const input = panel.querySelector('input[placeholder="Type a message..."]')!;
      const inputRow = input.parentElement!.parentElement!;
      expect(panelRect.width).to.equal(420);
      for (const element of [card, inputRow, input]) {
        const rect = element.getBoundingClientRect();
        expect(rect.width).to.be.greaterThan(0);
        expect(rect.left).to.be.at.least(panelRect.left);
        expect(rect.right).to.be.at.most(panelRect.right);
      }
    });
    for (const height of [200, 174]) {
      cy.get('.chat-panel').invoke('css', 'height', `${height}px`);
      cy.get('input[placeholder="Type a message..."]').then(($input) => {
        const panelRect = $input[0].closest('.chat-panel')!.getBoundingClientRect();
        const inputRect = $input[0].getBoundingClientRect();
        expect(inputRect.top).to.be.at.least(panelRect.top);
        expect(inputRect.bottom).to.be.at.most(panelRect.bottom);
      });
    }
    cy.get('.chat-panel').invoke('css', 'height', '600px');
    cy.get('[data-article-id="1"]').click();
    cy.wait('@contextContent');
    cy.get('input[placeholder="Type a message..."]')
      .should('be.enabled')
      .and('have.value', 'Draft for first article');
    cy.get('[data-article-id="2"]').click();
    cy.wait('@contextContent');
    cy.get('[data-testid="chat-new-context"]').click();
    cy.get('input[placeholder="Type a message..."]').should('be.disabled');
    cy.wait('@newContext');
    cy.wait('@contextMessages').its('response.statusCode').should('equal', 500);
    cy.get('[data-testid="chat-context-article"]')
      .should('have.attr', 'data-context-article-id', '2')
      .and('contain', secondArticle.title)
      .and('contain', secondArticle.feed_title);
    cy.get('input[placeholder="Type a message..."]')
      .should('be.enabled')
      .and('have.value', '')
      .type('Question about second article{enter}');
    cy.wait('@contextChat');
    cy.contains('.chat-panel', 'Answer for second article').should('be.visible');
  });

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
