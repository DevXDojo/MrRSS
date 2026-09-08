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

function setup(
  overrides: Record<string, string> = {},
  feedMode = 'global',
  empty = false,
  savedState: Record<string, string> = {}
) {
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
  cy.intercept({ method: 'GET', pathname: '/api/articles' }, empty ? [] : [article]).as('articles');
  cy.intercept('GET', '/api/articles/images*', empty ? [] : [article]).as('images');
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
  cy.visit('/', {
    onBeforeLoad(win) {
      Object.entries(savedState).forEach(([key, value]) => win.localStorage.setItem(key, value));
    },
  });
  cy.wait(['@feeds', '@articles']);
}

function openArticle() {
  cy.get('[data-article-id="1"]').click();
  cy.wait('@content');
  cy.get('.prose-content p').should('have.length', 2);
}

describe('Reading interactions', () => {
  for (const [language, unreadTitle, galleryTitle, unreadToggle, emptyText, completedText] of [
    [
      'en-US',
      'Unread',
      'Multimedia Gallery',
      'Show only unread articles',
      'No articles found.',
      "You're all caught up",
    ],
    ['zh-CN', '未读', '多媒体模式', '仅显示未读文章', '未找到文章', '已读完全部文章'],
  ]) {
    it(`centers unread completion and distinguishes ordinary empty galleries in ${language}`, () => {
      setup({ language, translation_enabled: 'false' }, 'global', true);
      cy.get('[data-testid="article-list-empty"]').should('contain', emptyText);
      cy.get(`.smart-activity-bar button[title^="${unreadTitle}"]`).click();
      cy.get('[data-testid="article-list-empty"]')
        .should('contain', completedText)
        .then(($empty) => {
          const element = $empty[0];
          const viewport = element.parentElement!.getBoundingClientRect();
          const first = element.firstElementChild!.getBoundingClientRect();
          const last = element.lastElementChild!.getBoundingClientRect();
          expect(
            Math.abs((first.top + last.bottom) / 2 - (viewport.top + viewport.bottom) / 2)
          ).to.be.lessThan(3);
        });
      cy.get(`.smart-activity-bar button[title="${galleryTitle}"]`).click();
      cy.wait('@images');
      cy.get('[data-testid="gallery-empty"]')
        .should('contain', emptyText)
        .and('not.contain', completedText);
      cy.get(`button[title="${unreadToggle}"]`).click();
      cy.wait('@images').its('request.url').should('contain', 'only_unread=true');
      cy.get('[data-testid="gallery-empty"]')
        .should('contain', completedText)
        .then(($empty) => {
          const element = $empty[0];
          const viewport = element.parentElement!.getBoundingClientRect();
          const first = element.firstElementChild!.getBoundingClientRect();
          const last = element.lastElementChild!.getBoundingClientRect();
          expect(
            Math.abs((first.top + last.bottom) / 2 - (viewport.top + viewport.bottom) / 2)
          ).to.be.lessThan(3);
        });
    });
  }
  it('prevents selecting model labels while keeping model changes and message selection available', () => {
    setup({ ai_chat_enabled: 'true', translation_enabled: 'false' });
    cy.intercept('GET', '/api/ai/profiles', [
      { id: 1, name: 'Model One', is_default: true },
      { id: 2, name: 'Model Two', is_default: false },
    ]);
    cy.intercept('GET', '/api/ai/chat/sessions*', []);
    cy.intercept('POST', '/api/ai-chat', (req) => {
      expect(req.body.profile_id).to.equal(2);
      req.reply({ response: 'Selectable answer', session_id: 1 });
    }).as('modelChat');
    openArticle();
    cy.get('button[title="AI Chat"]').click();
    cy.get('.chat-profile-selector .select-trigger .select-text')
      .should('have.css', 'user-select', 'none');
    cy.get('.chat-profile-selector .select-trigger')
      .should('have.css', 'user-select', 'none')
      .click();
    cy.contains('.chat-profile-selector .select-option', 'Model Two')
      .should('have.css', 'user-select', 'none')
      .click();
    cy.get('.chat-profile-selector .select-trigger').should('contain', 'Model Two');
    cy.get('input[placeholder="Type a message..."]').type('Question{enter}');
    cy.wait('@modelChat');
    cy.contains('.chat-panel .select-text', 'Selectable answer')
      .should('be.visible')
      .and('have.css', 'user-select', 'text');
  });

  it('keeps title editing above message actions without selecting the session on save', () => {
    setup({ ai_chat_enabled: 'true', translation_enabled: 'false' });
    cy.intercept('GET', '/api/ai/profiles', []);
    cy.intercept('GET', '/api/ai/chat/sessions*', [
      { id: 1, article_id: 1, title: 'Saved session', message_count: 1 },
    ]);
    cy.intercept('GET', '/api/ai/chat/messages*', [
      { id: 1, role: 'user', content: 'Previous question', created_at: '' },
    ]).as('chatMessages');
    cy.intercept('PUT', '/api/ai/chat/session*', (req) => {
      expect(req.body.title).to.equal('Renamed session');
      req.reply({ success: true });
    }).as('renameChat');
    openArticle();
    cy.get('button[title="AI Chat"]').click();
    cy.wait('@chatMessages');
    cy.get('[data-testid="chat-session-switcher"]').click();
    cy.get('[data-session-id="1"] button').first().click();
    cy.get('[data-session-id="1"] input').clear().type('Renamed session');
    cy.get('.chat-panel button[title="Copy message"]').should('not.be.visible');
    cy.get('[data-session-id="1"] button').first().click();
    cy.wait('@renameChat');
    cy.get('[data-session-id="1"]').should('be.visible').and('contain', 'Renamed session');
    cy.get('@chatMessages.all').should('have.length', 1);
    cy.get('[data-testid="chat-session-switcher"]').click();
    cy.contains('.chat-panel', 'Previous question').should('be.visible');
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

  it('remembers the measured chat size after dragging, closing, reopening, and reloading', () => {
    cy.viewport(1280, 900);
    setup({ ai_chat_enabled: 'true' }, 'global', false, { FeedListExpanded: 'false' });
    cy.intercept('GET', '/api/ai/profiles', []);
    cy.intercept('GET', '/api/ai/chat/sessions*', []);
    openArticle();
    cy.get('.js-article-chat-button').click();
    cy.get('.chat-panel')
      .should(($panel) => {
        const rect = $panel[0].getBoundingClientRect();
        expect(rect.width).to.equal(500);
        expect(rect.height).to.equal(600);
      })
      .then(($panel) => {
        const rect = $panel[0].getBoundingClientRect();
        cy.get('.chat-panel .cursor-nw-resize').trigger('mousedown', {
          clientX: rect.left + 2,
          clientY: rect.top + 2,
          button: 0,
          force: true,
        });
        cy.document().trigger('mousemove', { clientX: rect.left - 118, clientY: rect.top - 98 });
      });
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(620);
      expect(rect.height).to.equal(700);
    });
    cy.window().then((win) => expect(win.localStorage.getItem('mrrssChatPanelSize')).to.be.null);
    cy.document().trigger('mouseup');
    cy.window().then((win) => {
      expect(JSON.parse(win.localStorage.getItem('mrrssChatPanelSize')!)).to.deep.equal({
        width: 620,
        height: 700,
      });
    });
    cy.get('.chat-panel button[title="Close"]').click();
    cy.get('.chat-panel').should('not.exist');
    cy.get('.js-article-chat-button').click();
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(620);
      expect(rect.height).to.equal(700);
    });
    cy.reload();
    openArticle();
    cy.get('.js-article-chat-button').click();
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(620);
      expect(rect.height).to.equal(700);
    });
  });

  it('temporarily fits a narrow viewport without replacing the preferred chat size', () => {
    cy.viewport(1280, 900);
    const preferred = { width: 680, height: 720 };
    setup({ ai_chat_enabled: 'true' }, 'global', false, {
      FeedListExpanded: 'false',
      mrrssChatPanelSize: JSON.stringify(preferred),
    });
    cy.intercept('GET', '/api/ai/profiles', []);
    cy.intercept('GET', '/api/ai/chat/sessions*', []);
    openArticle();
    cy.get('.js-article-chat-button').click();
    cy.viewport(390, 500);
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(358);
      expect(rect.height).to.equal(444);
      expect(rect.left).to.be.at.least(0);
      expect(rect.top).to.be.at.least(0);
      expect(rect.right).to.be.at.most(390);
      expect(rect.bottom).to.be.at.most(500);
    });
    cy.get('.chat-panel button[title="Close"]').click();
    cy.get('.chat-panel').should('not.exist');
    cy.get('.js-article-chat-button').click();
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(358);
      expect(rect.height).to.equal(444);
    });
    cy.window().then((win) => {
      expect(JSON.parse(win.localStorage.getItem('mrrssChatPanelSize')!)).to.deep.equal(preferred);
    });
    cy.viewport(1280, 900);
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(preferred.width);
      expect(rect.height).to.equal(preferred.height);
    });
    cy.window().then((win) => {
      expect(JSON.parse(win.localStorage.getItem('mrrssChatPanelSize')!)).to.deep.equal(preferred);
    });
  });

  it('uses a 500 by 600 chat default and a viewport-limited 420 by 200 drag minimum', () => {
    cy.viewport(1280, 900);
    setup({ ai_chat_enabled: 'true' });
    cy.intercept('GET', '/api/ai/profiles', []);
    cy.intercept('GET', '/api/ai/chat/sessions*', []);
    openArticle();
    cy.get('.js-article-chat-button').click();
    cy.get('.chat-panel')
      .should(($panel) => {
        const rect = $panel[0].getBoundingClientRect();
        expect(rect.width).to.equal(500);
        expect(rect.height).to.equal(600);
      })
      .then(($panel) => {
        const rect = $panel[0].getBoundingClientRect();
        cy.get('.chat-panel .cursor-nw-resize').trigger('mousedown', {
          clientX: rect.left + 2,
          clientY: rect.top + 2,
          button: 0,
          force: true,
        });
        cy.document().trigger('mousemove', { clientX: rect.right, clientY: rect.bottom });
        cy.document().trigger('mouseup');
      });
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(420);
      expect(rect.height).to.equal(200);
    });
    cy.viewport(320, 230);
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(288);
      expect(rect.height).to.equal(174);
      expect(rect.left).to.be.at.least(0);
      expect(rect.top).to.be.at.least(0);
      expect(rect.right).to.be.at.most(320);
      expect(rect.bottom).to.be.at.most(230);
    });
    cy.viewport(1280, 900);
    cy.get('.chat-panel').should(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      expect(rect.width).to.equal(420);
      expect(rect.height).to.equal(200);
    });
  });

  it('truncates long session titles without pushing header actions outside the chat panel', () => {
    cy.viewport(1280, 900);
    setup({ ai_chat_enabled: 'true' });
    const title = 'Long chat session title '.repeat(30);
    cy.intercept('GET', '/api/ai/profiles', [
      { id: 1, name: 'Long AI profile name '.repeat(10), is_default: true },
    ]);
    cy.intercept('GET', '/api/ai/chat/sessions*', [{ id: 1, article_id: article.id, title }]);
    cy.intercept('GET', '/api/ai/chat/messages*', []).as('chatMessages');
    openArticle();
    cy.get('.js-article-chat-button').click();
    cy.wait('@chatMessages');
    cy.get('.chat-panel').then(($panel) => {
      const rect = $panel[0].getBoundingClientRect();
      cy.get('.chat-panel .cursor-nw-resize').trigger('mousedown', {
        clientX: rect.left + 2,
        clientY: rect.top + 2,
        button: 0,
        force: true,
      });
      cy.document().trigger('mousemove', { clientX: rect.right, clientY: rect.bottom });
      cy.document().trigger('mouseup');
    });
    for (const [viewportWidth, viewportHeight, expectedWidth] of [
      [1280, 900, 420],
      [320, 330, 288],
    ]) {
      cy.viewport(viewportWidth, viewportHeight);
      cy.get('.chat-panel').should(($panel) => {
        const panel = $panel[0];
        const rect = panel.getBoundingClientRect();
        expect(rect.width).to.equal(expectedWidth);
        const header = panel.firstElementChild as HTMLElement;
        expect(header.scrollWidth).to.be.at.most(header.clientWidth);
        for (const button of header.querySelectorAll('button')) {
          const bounds = button.getBoundingClientRect();
          expect(bounds.width).to.be.greaterThan(0);
          expect(bounds.left).to.be.at.least(rect.left);
          expect(bounds.right).to.be.at.most(rect.right);
          expect(bounds.top).to.be.at.least(rect.top);
          expect(bounds.bottom).to.be.at.most(rect.bottom);
        }
      });
      cy.get('[data-testid="chat-session-switcher"] span')
        .should('have.text', title)
        .should(($title) => {
          expect($title[0].scrollWidth).to.be.greaterThan($title[0].clientWidth);
          expect(getComputedStyle($title[0]).textOverflow).to.equal('ellipsis');
        });
      cy.get('.chat-panel button[title="Close"]').should('be.visible');
      cy.get('[data-testid="chat-new-session"]').should('be.visible');
    }
    cy.get('.chat-panel button[title="Close"]').click();
    cy.get('.chat-panel').should('not.exist');
  });
});
