/// <reference types="cypress" />

describe('Article body translation', () => {
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

    cy.visit('/');
    cy.wait('@translationFeeds');
    cy.wait('@translationArticles');
    cy.get('[data-article-id="1"]').click();
    cy.wait('@translationContent');
    cy.wait('@translateArticleBody');
    cy.contains('这段正文已成功翻译').should('be.visible');
  });
});
