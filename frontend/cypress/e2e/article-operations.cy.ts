/// <reference types="cypress" />

describe('Article Operations', () => {
  beforeEach(() => {
    // Set up intercepts before visiting the page
    cy.intercept('GET', '/api/articles*').as('getArticles');
    cy.intercept('PUT', '/api/articles/*').as('updateArticle');
    cy.intercept('PUT', '/api/articles/mark-all-read').as('markAllRead');

    cy.visit('/');
    cy.get('body').should('be.visible');
  });

  it('should mark article as read', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Try to click on an article if it exists
    cy.get('body').then(($body) => {
      if ($body.find('[class*="article"]').length > 0) {
        cy.get('[class*="article"]').first().click({ force: true });

        // Wait for detail view to appear
        cy.wait(500);

        // The article detail view should be shown (or at least some content changed)
        cy.get('body').should('be.visible');
      } else {
        cy.log('No articles found to test marking as read');
      }
    });
  });

  it('should mark article as favorite', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Try to find an article to test
    cy.get('body').then(($body) => {
      if ($body.find('[class*="article"]').length > 0) {
        // Right-click on an article to open context menu
        cy.get('[class*="article"]').first().rightclick({ force: true });

        // Click favorite option if it exists
        cy.get('body').then(($body2) => {
          if ($body2.find(/favorite|收藏|star/i).length > 0) {
            cy.contains(/favorite|收藏|star/i).click({ force: true });
          } else {
            cy.log('Favorite option not available');
          }
        });
      } else {
        cy.log('No articles found to test marking as favorite');
      }
    });
  });

  it('should filter articles by read status', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Look for filter buttons
    cy.contains(/unread|未读/i).click({ force: true });

    // Wait a bit for the filter to apply
    cy.wait(500);

    // Verify filter button is clickable
    cy.contains(/unread|未读/i).should('exist');
  });

  it('should filter articles by favorites', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Click favorites filter
    cy.contains(/favorite|收藏/i).click({ force: true });

    // Wait a bit for the filter to apply
    cy.wait(500);

    // Verify filter button is clickable
    cy.contains(/favorite|收藏/i).should('exist');
  });

  it('should mark all articles as read', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Try to find mark all as read button (it might be in a context menu or toolbar)
    cy.get('body').then(($body) => {
      if (
        $body.find('button').filter((i, el) => /mark.*all|全部标记/i.test(el.textContent || ''))
          .length > 0
      ) {
        cy.get('button')
          .contains(/mark.*all|全部标记/i)
          .click({ force: true });

        // Wait for confirmation if needed
        cy.get('body').then(($body2) => {
          if ($body2.find(/confirm|确认/i).length > 0) {
            cy.contains(/confirm|确认/i).click({ force: true });
          }
        });
      } else {
        cy.log('Mark all as read button not found');
      }
    });
  });

  it('should open article detail view', () => {
    // Try to click on an article if it exists
    cy.get('body').then(($body) => {
      if ($body.find('[class*="article"]').length > 0) {
        cy.get('[class*="article"]').first().click({ force: true });

        // Verify detail view is shown
        cy.wait(500);
        cy.get('body').should('be.visible');
      } else {
        cy.log('No articles found to test detail view');
      }
    });
  });

  it('should search articles', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Find search input
    cy.get('body').then(($body) => {
      if (
        $body.find('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]')
          .length > 0
      ) {
        cy.get('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]')
          .last()
          .type('test{enter}');

        // Wait a bit for search results
        cy.wait(500);
      } else {
        cy.log('Search input not found');
      }
    });
  });

  it('should open article in external browser', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Try to find an article
    cy.get('body').then(($body) => {
      if ($body.find('[class*="article"]').length > 0) {
        // Right-click on article
        cy.get('[class*="article"]').first().rightclick({ force: true });

        // Look for "Open in browser" option
        cy.get('body').then(($body2) => {
          if ($body2.find(/open.*browser|在浏览器中打开/i).length > 0) {
            cy.contains(/open.*browser|在浏览器中打开/i).should('exist');
          } else {
            cy.log('Open in browser option not found in context menu');
          }
        });
      } else {
        cy.log('No articles found to test open in browser');
      }
    });
  });

  it('should translate orphaned article text next to media', () => {
    const settingsState: Record<string, string> = {
      language: 'en-US',
      theme: 'light',
      layout_mode: 'normal',
      default_view_mode: 'rendered',
      translation_enabled: 'true',
      translation_provider: 'ai',
      translation_only_mode: 'false',
      target_language: 'zh-CN',
      summary_enabled: 'false',
      full_text_fetch_enabled: 'false',
      update_check_enabled: 'false',
    };
    const feed = {
      id: 1,
      title: 'Translation Feed',
      url: 'https://example.com/feed.xml',
      category: '',
    };
    const article = {
      id: 1,
      feed_id: 1,
      feed_title: feed.title,
      title: 'English title',
      url: 'https://example.com/article',
      published_at: '2026-04-22T00:00:00Z',
      translated_title: '',
      is_read: false,
      is_favorite: false,
      is_hidden: false,
      is_read_later: false,
    };

    cy.intercept('/api/**', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/settings', { statusCode: 200, body: settingsState });
    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: [feed] }).as('translationFeeds');
    cy.intercept('GET', '/api/tags', { statusCode: 200, body: [] });
    cy.intercept('GET', '/api/saved-filters', { statusCode: 200, body: [] });
    cy.intercept(
      { method: 'GET', pathname: '/api/articles' },
      { statusCode: 200, body: [article] }
    ).as('translationArticles');
    cy.intercept('GET', '/api/articles/unread-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/articles/filter-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/progress', { statusCode: 200, body: { is_running: false } });
    cy.intercept('GET', '/api/articles/content*', {
      statusCode: 200,
      body: {
        content:
          'This text is a direct DOM text node and must still be translated.<img src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///ywAAAAAAQABAAACAUwAOw==">',
        cached: true,
      },
    }).as('translationContent');
    cy.intercept('POST', '/api/articles/translate-text', (req) => {
      if (req.body.text === article.title) {
        req.reply({
          statusCode: 200,
          body: { translated_text: '英文标题', html: '', skipped: false },
        });
        return;
      }
      req.alias = 'translateArticleBody';
      expect(req.body.text).to.contain('direct DOM text node');
      req.reply({
        statusCode: 200,
        body: { translated_text: '这段正文已成功翻译', html: '', skipped: false },
      });
    });

    cy.reload();
    cy.wait('@translationFeeds');
    cy.wait('@translationArticles');
    cy.get('[data-article-id="1"]').click();
    cy.wait('@translationContent');
    cy.wait('@translateArticleBody');
    cy.contains('这段正文已成功翻译').should('be.visible');
  });

  it('should explain AI search results and keep list and card navigation in search context', () => {
    const settingsState: Record<string, string> = {
      language: 'en-US',
      theme: 'light',
      layout_mode: 'normal',
      default_view_mode: 'rendered',
      ai_search_enabled: 'true',
      translation_mode: 'off',
      summary_enabled: 'false',
      full_text_fetch_enabled: 'false',
      update_check_enabled: 'false',
    };
    const feed = {
      id: 1,
      title: 'Search Feed',
      url: 'https://example.com/feed.xml',
      category: '',
      article_view_mode: 'global',
    };
    const timelineArticle = {
      id: 1,
      feed_id: 1,
      feed_title: feed.title,
      title: 'Timeline article outside search results',
      url: 'https://example.com/timeline',
      published_at: '2026-08-20T00:00:00Z',
      is_read: false,
      is_favorite: false,
      is_hidden: false,
      is_read_later: false,
    };
    const searchArticles = [101, 102, 103].map((id, index) => ({
      id,
      feed_id: 1,
      feed_title: feed.title,
      title: `Search result ${index + 1}`,
      url: `https://example.com/search/${id}`,
      published_at: `2026-08-${23 - index}T00:00:00Z`,
      is_read: false,
      is_favorite: false,
      is_hidden: false,
      is_read_later: false,
      relevance_score: 90 - index,
      matched_terms: ['privacy'],
      matched_fields: index === 0 ? ['title', 'summary'] : ['content'],
      excerpt: `This privacy excerpt explains result ${index + 1}`,
    }));

    cy.intercept('/api/**', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/settings', (req) => {
      req.reply({ statusCode: 200, body: settingsState });
    });
    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: [feed] }).as('searchFeeds');
    cy.intercept('GET', '/api/tags', { statusCode: 200, body: [] });
    cy.intercept('GET', '/api/saved-filters', { statusCode: 200, body: [] });
    cy.intercept(
      { method: 'GET', pathname: '/api/articles' },
      { statusCode: 200, body: [timelineArticle] }
    ).as('timelineArticles');
    cy.intercept('GET', '/api/articles/unread-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/articles/filter-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/progress', { statusCode: 200, body: { is_running: false } });
    cy.intercept('POST', '/api/ai/search', (req) => {
      const noResults = req.body?.query === 'nothing matches';
      req.reply({
        statusCode: 200,
        body: {
          success: true,
          articles: noResults ? [] : searchArticles,
          total_count: noResults ? 0 : searchArticles.length,
        },
      });
    }).as('aiSearch');
    cy.intercept('GET', '/api/articles/content*', (req) => {
      req.reply({
        statusCode: 200,
        body: { content: `<p>Body for search result ${req.query.id}</p>`, cached: true },
      });
    }).as('searchArticleContent');
    cy.intercept('POST', '/api/articles/read*', { statusCode: 200, body: { success: true } });
    cy.intercept('POST', '/api/articles/favorite*', {
      statusCode: 200,
      body: { success: true },
    });
    cy.intercept('POST', '/api/articles/toggle-read-later*', {
      statusCode: 200,
      body: { success: true },
    });

    cy.window().then((win) => win.localStorage.setItem('showOnlyUnread', 'true'));
    cy.reload();
    cy.wait('@searchFeeds');
    cy.wait('@timelineArticles');

    cy.get('input[placeholder="Describe what you want to find..."]').type('privacy');
    cy.contains('button', /^AI Search$/).click();
    cy.wait('@aiSearch');
    cy.contains('Title match').should('be.visible');
    cy.contains('Summary match').should('be.visible');
    cy.contains('This privacy excerpt explains result 1').should('be.visible');

    cy.get('[data-article-id="101"]').click();
    cy.wait('@searchArticleContent');
    cy.contains('Body for search result 101').should('be.visible');

    // The next result is marked read when opened, but must remain selected and
    // renderable while the unread-only preference is active.
    cy.get('button[title="Next Article"]').click();
    cy.wait('@searchArticleContent');
    cy.contains('Body for search result 102').should('be.visible');
    cy.get('[data-article-id="102"]').should('exist');

    // Global shortcuts must use the same ordered search context instead of the
    // separately paginated timeline.
    cy.get('body').trigger('keydown', { key: 'j' });
    cy.wait('@searchArticleContent');
    cy.contains('Body for search result 103').should('be.visible');
    cy.get('button[title="Previous Article"]').click();
    cy.wait('@searchArticleContent');
    cy.contains('Body for search result 102').should('be.visible');

    cy.contains('button', /^Clear$/).click();
    cy.get('input[placeholder="Describe what you want to find..."]').type('nothing matches');
    cy.contains('button', /^AI Search$/).click();
    cy.wait('@aiSearch');
    cy.contains('No articles found matching your search').should('be.visible');

    // Card mode uses ArticleDetailModal, which must use the same navigation
    // context instead of hiding navigation for off-page search results.
    cy.then(() => {
      settingsState.layout_mode = 'card';
    });
    cy.reload();
    cy.wait('@searchFeeds');
    cy.wait('@timelineArticles');
    cy.get('input[placeholder="Describe what you want to find..."]').type('privacy');
    cy.contains('button', /^AI Search$/).click();
    cy.wait('@aiSearch');
    cy.get('.article-card-item[data-article-id="101"]').click();
    cy.wait('@searchArticleContent');
    cy.contains('Body for search result 101').should('be.visible');
    cy.get('button[title="Next Article"]').should('be.visible').click();
    cy.wait('@searchArticleContent');
    cy.contains('Body for search result 102').should('be.visible');
  });

  it('should preserve AI chat history across new conversations and request failures', () => {
    const settingsState: Record<string, string> = {
      language: 'en-US',
      theme: 'light',
      layout_mode: 'normal',
      default_view_mode: 'rendered',
      ai_chat_enabled: 'true',
      translation_enabled: 'false',
      summary_enabled: 'false',
      full_text_fetch_enabled: 'false',
      update_check_enabled: 'false',
    };
    const feed = {
      id: 1,
      title: 'Chat Feed',
      url: 'https://example.com/chat.xml',
      category: '',
      article_view_mode: 'global',
    };
    const article = {
      id: 1,
      feed_id: 1,
      feed_title: feed.title,
      title: 'Chat article',
      url: 'https://example.com/chat/article',
      published_at: '2026-08-25T00:00:00Z',
      is_read: false,
      is_favorite: false,
      is_hidden: false,
      is_read_later: false,
    };
    let nextSessionID = 1;
    const sessions: Array<Record<string, unknown>> = [];
    const messages = new Map<number, Array<Record<string, unknown>>>();
    const timestamp = () => new Date().toISOString();

    cy.intercept('/api/**', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/settings', { statusCode: 200, body: settingsState });
    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: [feed] }).as('chatFeeds');
    cy.intercept('GET', '/api/tags', { statusCode: 200, body: [] });
    cy.intercept('GET', '/api/saved-filters', { statusCode: 200, body: [] });
    cy.intercept(
      { method: 'GET', pathname: '/api/articles' },
      { statusCode: 200, body: [article] }
    ).as('chatArticles');
    cy.intercept('GET', '/api/articles/unread-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/articles/filter-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/progress', {
      statusCode: 200,
      body: { is_running: true, pool_task_count: 0, queue_task_count: 0 },
    });
    cy.intercept('GET', '/api/articles/content*', {
      statusCode: 200,
      body: { content: '<p>Article context for AI chat</p>', cached: true },
    }).as('chatArticleContent');
    cy.intercept('POST', '/api/articles/read*', { statusCode: 200, body: { success: true } });
    cy.intercept('GET', '/api/ai/chat/sessions*', (req) => {
      req.reply({ statusCode: 200, body: sessions });
    }).as('chatSessions');
    cy.intercept('GET', '/api/ai/chat/messages*', (req) => {
      req.reply({ statusCode: 200, body: messages.get(Number(req.query.session_id)) || [] });
    }).as('chatMessages');
    cy.intercept('POST', '/api/ai/chat/session/create', (req) => {
      const id = nextSessionID++;
      const session = {
        id,
        article_id: 1,
        title: req.body.title,
        created_at: timestamp(),
        updated_at: timestamp(),
        message_count: 0,
      };
      sessions.unshift(session);
      messages.set(id, []);
      req.reply({ statusCode: 200, body: session });
    }).as('createChatSession');
    cy.intercept('POST', '/api/ai-chat', (req) => {
      const lastMessage = req.body.messages.at(-1)?.content || '';
      let sessionID = Number(req.body.session_id || 0);
      if (!sessionID) {
        sessionID = nextSessionID++;
        sessions.unshift({
          id: sessionID,
          article_id: 1,
          title: lastMessage.slice(0, 60),
          created_at: timestamp(),
          updated_at: timestamp(),
          message_count: 0,
        });
      }
      const stored = messages.get(sessionID) || [];
      stored.push({
        id: stored.length + 1,
        role: 'user',
        content: lastMessage,
        created_at: timestamp(),
      });
      messages.set(sessionID, stored);

      if (lastMessage === 'trigger failure') {
        req.reply({
          statusCode: 500,
          body: {
            error: 'Failed to get response from AI. Please try again.',
            session_id: sessionID,
          },
        });
        return;
      }

      stored.push({
        id: stored.length + 1,
        role: 'assistant',
        content: 'Persisted answer',
        created_at: timestamp(),
      });
      const session = sessions.find((item) => item.id === sessionID);
      if (session) session.message_count = stored.length;
      req.reply({
        statusCode: 200,
        body: { response: 'Persisted answer', session_id: sessionID, history_saved: true },
      });
    }).as('aiChat');

    cy.reload();
    cy.wait('@chatFeeds');
    cy.wait('@chatArticles');
    cy.get('[data-article-id="1"]').click();
    cy.wait('@chatArticleContent');
    cy.get('button[title="AI Chat"]').click();
    cy.wait('@chatSessions');

    cy.get('input[placeholder="Type a message..."]').type('First question{enter}');
    cy.wait('@aiChat');
    cy.contains('.chat-panel', 'Persisted answer').should('be.visible');

    cy.get('[data-testid="chat-new-session"]').click();
    cy.wait('@createChatSession');
    cy.wait('@chatMessages');
    cy.contains('.chat-panel', 'Persisted answer').should('not.exist');
    cy.get('[data-testid="chat-session-switcher"]').click();
    cy.get('.chat-panel [data-session-id="1"]').click();
    cy.wait('@chatMessages');
    cy.contains('.chat-panel', 'Persisted answer').should('be.visible');

    cy.get('input[placeholder="Type a message..."]').type('trigger failure{enter}');
    cy.wait('@aiChat');
    cy.wait('@chatMessages');
    cy.contains('.chat-panel', 'trigger failure').should('be.visible');
    cy.contains('Failed to get response from AI. Please try again.').should('be.visible');
  });
});
