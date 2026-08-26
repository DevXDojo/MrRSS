/// <reference types="cypress" />

describe('Settings Persistence', () => {
  beforeEach(() => {
    // Set up intercepts before visiting the page
    cy.intercept('GET', '/api/settings').as('getSettings');
    cy.intercept('POST', '/api/settings').as('saveSettings');
    cy.intercept('GET', '/api/feeds').as('getFeeds');

    cy.visit('/');

    // Wait for the app to be fully loaded
    cy.get('body').should('be.visible');
  });

  it('should persist theme changes after closing and reopening settings', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings modal - find the gear icon button
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

    // Wait for settings modal to be visible
    cy.contains(/settings|设置/i).should('be.visible');

    // Ensure we're on the general tab (or navigate to it)
    cy.contains(/general|常规/i).click({ force: true });

    cy.contains('.setting-item', /theme|主题/i)
      .find('button.select-trigger')
      .click();
    cy.contains('.select-option', /dark|暗色|深色/i).click({ force: true });
    cy.wait('@saveSettings', { timeout: 5000 }).its('request.body.theme').should('eq', 'dark');

    // Close the settings modal
    cy.get('body').type('{esc}');

    // Wait a bit for modal to close
    cy.wait(500);

    // Reopen settings to verify the change persisted
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

    // Wait for settings to load again
    cy.wait('@getSettings');

    cy.contains('.setting-item', /theme|主题/i)
      .find('button.select-trigger')
      .should(($button) => {
        expect($button.text()).to.match(/dark|暗色|深色/i);
      });
  });

  it('should persist language changes', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

    // Navigate to general tab if not already there
    cy.contains(/general|常规/i).click({ force: true });

    // Wait for settings to load
    cy.wait('@getSettings');

    let expectedLanguage = '';
    cy.contains('.setting-item', /language|语言/i)
      .find('button.select-trigger')
      .invoke('text')
      .then((currentLanguage) => {
        const chooseChinese = /english|英语/i.test(currentLanguage);
        expectedLanguage = chooseChinese ? 'zh-CN' : 'en-US';
        cy.contains('.setting-item', /language|语言/i)
          .find('button.select-trigger')
          .click();
        cy.contains('.select-option', chooseChinese ? /chinese|中文/i : /english|英语/i).click({
          force: true,
        });
        cy.wait('@saveSettings', { timeout: 5000 })
          .its('request.body.language')
          .should('eq', expectedLanguage);
      });

    // Close settings
    cy.get('body').type('{esc}');
    cy.wait(500);

    // Reopen settings to verify
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    cy.contains('.setting-item', /language|语言/i)
      .find('button.select-trigger')
      .should(($button) => {
        const expectedLabel = expectedLanguage === 'zh-CN' ? /chinese|中文/i : /english|英语/i;
        expect($button.text()).to.match(expectedLabel);
      });
  });

  it('should persist update interval changes', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

    // Navigate to general tab (update settings are here)
    cy.contains(/general|常规/i).click({ force: true });

    // Wait for settings to load
    cy.wait('@getSettings');

    cy.contains('.setting-item', /refresh mode|刷新模式/i)
      .scrollIntoView()
      .find('button.select-trigger')
      .click();
    cy.contains('.select-option', /fixed interval|固定间隔/i).click({ force: true });

    let expectedInterval = '';
    cy.contains('.sub-setting-item', /auto update interval|自动更新间隔/i)
      .scrollIntoView()
      .find('input[type="number"]')
      .invoke('val')
      .then((currentValue) => {
        expectedInterval = String(currentValue) === '31' ? '30' : '31';
        cy.contains('.sub-setting-item', /auto update interval|自动更新间隔/i)
          .find('input[type="number"]')
          .invoke('val', expectedInterval)
          .trigger('input')
          .trigger('change');
        cy.wait('@saveSettings', { timeout: 5000 })
          .its('request.body.update_interval')
          .should('eq', expectedInterval);
      });

    cy.get('body').type('{esc}');
    cy.wait(500);
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');
    cy.contains('.sub-setting-item', /auto update interval|自动更新间隔/i)
      .scrollIntoView()
      .find('input[type="number"]')
      .should(($input) => {
        expect($input.val()).to.equal(expectedInterval);
      });
  });

  it('should handle multiple setting changes in sequence', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    // Change theme
    cy.contains(/general|常规/i).click({ force: true });
    cy.get('body').then(($body) => {
      if ($body.find(/light|亮色/i).length > 0) {
        cy.contains(/light|亮色/i).click({ force: true });
        cy.wait(1000);
      }
    });

    // Navigate to another tab
    cy.get('body').then(($body) => {
      if ($body.find(/feeds|订阅/i).length > 0) {
        cy.contains(/feeds|订阅/i).click({ force: true });
        cy.wait(500);

        // Just verify the tab is open (no number input on feeds tab)
        cy.contains(/feeds|订阅/i).should('exist');
      }
    });

    // Close and reopen
    cy.get('body').type('{esc}');
    cy.wait(500);

    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    // Verify settings modal is open
    cy.contains(/settings|设置/i).should('be.visible');
  });

  it('should save settings when switching between tabs', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    let expectedTheme = '';
    cy.contains(/general|常规/i).click({ force: true });
    cy.contains('.setting-item', /theme|主题/i)
      .find('button.select-trigger')
      .invoke('text')
      .then((currentTheme) => {
        const chooseDark = !/dark|暗色|深色/i.test(currentTheme);
        expectedTheme = chooseDark ? 'dark' : 'light';
        cy.contains('.setting-item', /theme|主题/i)
          .find('button.select-trigger')
          .click();
        cy.contains('.select-option', chooseDark ? /dark|暗色|深色/i : /light|亮色/i).click({
          force: true,
        });
        cy.contains('button', /^(Feeds|订阅)$/i).click({ force: true });
        cy.wait('@saveSettings', { timeout: 5000 })
          .its('request.body.theme')
          .should('eq', expectedTheme);
      });

    // Close settings
    cy.get('body').type('{esc}');

    // Reopen and verify the change was saved
    cy.wait(500);
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    cy.contains(/general|常规/i).click({ force: true });
    cy.contains(/dark|深色/i).should('exist');
  });

  it('should apply and persist interface typography without changing article content typography', () => {
    let settingsState: Record<string, string> = {
      language: 'en-US',
      theme: 'light',
      layout_mode: 'normal',
      ui_font_family: 'system',
      ui_font_size: '16',
      content_font_family: 'system',
      content_font_size: '16',
      content_line_height: '1.6',
      update_check_enabled: 'false',
      update_interval: '10',
      last_global_refresh: new Date().toISOString(),
    };

    const feed = {
      id: 1,
      url: 'https://example.com/feed.xml',
      title: 'Typography Test Feed',
      category: '',
      last_fetched_at: '2026-08-13T00:00:00Z',
    };
    const article = {
      id: 1,
      feed_id: 1,
      feed_title: feed.title,
      title: 'Typography Test Article',
      url: 'https://example.com/article',
      published_at: '2026-08-13T00:00:00Z',
      is_read: false,
      is_favorite: false,
      is_hidden: false,
      is_read_later: false,
    };

    cy.intercept('/api/**', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/settings', (req) => {
      req.reply({ statusCode: 200, body: settingsState });
    }).as('typographySettings');
    cy.intercept('POST', '/api/settings', (req) => {
      settingsState = { ...settingsState, ...req.body };
      req.reply({ statusCode: 200, body: { success: true } });
    }).as('saveTypographySettings');
    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: [feed] }).as('typographyFeeds');
    cy.intercept(
      { method: 'GET', pathname: '/api/articles' },
      { statusCode: 200, body: [article] }
    ).as('typographyArticles');
    cy.intercept('GET', '/api/tags', { statusCode: 200, body: [] });
    cy.intercept('GET', '/api/articles/unread-counts', {
      statusCode: 200,
      body: { total: 1, feed_counts: { 1: 1 } },
    });
    cy.intercept('GET', '/api/articles/filter-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/progress', { statusCode: 200, body: { is_running: false } });
    cy.intercept('GET', '/api/articles/content*', {
      statusCode: 200,
      body: { content: '<p>Independent article content</p>', cached: true },
    });

    cy.reload();
    cy.wait('@typographyFeeds');
    cy.wait('@typographyArticles');

    let initialArticleTitleSize = 0;
    cy.get('.article-card .article-title')
      .should('be.visible')
      .then(($title) => {
        initialArticleTitleSize = parseFloat(getComputedStyle($title[0]).fontSize);
      });

    cy.get('button[title="Settings"]').click();
    cy.contains('button', /^General$/).click();

    cy.get('.setting-item').then(($items) => {
      const labels = [...$items].map((item) => item.textContent || '');
      const languageIndex = labels.findIndex((label) => label.includes('Language'));
      const interfaceFontIndex = labels.findIndex((label) =>
        label.includes('Interface Font Family')
      );
      const interfaceSizeIndex = labels.findIndex((label) => label.includes('Interface Font Size'));

      expect(languageIndex).to.be.greaterThan(-1);
      expect(interfaceFontIndex).to.equal(languageIndex + 1);
      expect(interfaceSizeIndex).to.equal(interfaceFontIndex + 1);
    });

    let interfaceFontOptions: string[] = [];

    cy.contains('.setting-item', 'Interface Font Family').within(() => {
      cy.get('button.select-trigger').click();
    });
    cy.get('.select-option').then(($options) => {
      interfaceFontOptions = [...$options].map((option) => option.textContent?.trim() || '');
      expect(interfaceFontOptions).to.include('System Default');
    });
    cy.contains('.select-option', 'Default Serif').click({ force: true });
    cy.wait('@saveTypographySettings').its('request.body.ui_font_family').should('eq', 'serif');

    cy.contains('.setting-item', 'Interface Font Size').within(() => {
      cy.get('input[type="number"]').invoke('val', '20').trigger('input').trigger('change');
    });
    cy.wait('@saveTypographySettings').its('request.body.ui_font_size').should('eq', '20');

    cy.document().then((doc) => {
      expect(doc.documentElement.style.getPropertyValue('--ui-font-size')).to.equal('20px');
      expect(doc.documentElement.style.getPropertyValue('--ui-font-scale')).to.equal('1.25');
      expect(doc.documentElement.style.getPropertyValue('--ui-font-family')).to.contain('Georgia');
    });
    cy.get('.feed-item').should(($feedItem) => {
      expect(getComputedStyle($feedItem[0]).fontFamily).to.contain('Georgia');
    });
    cy.get('.article-card .article-title').should(($title) => {
      expect(parseFloat(getComputedStyle($title[0]).fontSize)).to.be.greaterThan(
        initialArticleTitleSize
      );
    });

    cy.get('body').type('{esc}');
    cy.get('.article-card').first().click();
    cy.get('.prose-content').should(($content) => {
      const style = getComputedStyle($content[0]);
      expect(style.fontSize).to.equal('16px');
      expect(style.fontFamily).not.to.contain('Georgia');
    });

    cy.get('button[title="Settings"]').click();
    cy.contains('button', /^General$/).click();
    cy.contains('.setting-item', 'Interface Font Family')
      .find('button.select-trigger')
      .should('contain.text', 'Default Serif');
    cy.contains('.setting-item', 'Interface Font Size')
      .find('input[type="number"]')
      .should('have.value', '20');

    cy.contains('button', /^Reading$/).click();
    cy.contains('.setting-item', 'Interface Font Family').should('not.exist');
    cy.contains('.setting-item', 'Interface Font Size').should('not.exist');
    cy.contains('.setting-item', 'Content Font Family').within(() => {
      cy.get('button.select-trigger').click();
    });
    cy.get('.select-option').then(($options) => {
      const contentFontOptions = [...$options].map((option) => option.textContent?.trim() || '');
      expect(contentFontOptions).to.deep.equal(interfaceFontOptions);
    });
    cy.contains('.select-option', 'System Default').click({ force: true });

    cy.get('body').type('{esc}');
    cy.reload();
    cy.wait('@typographyFeeds');
    cy.document().then((doc) => {
      expect(doc.documentElement.style.getPropertyValue('--ui-font-size')).to.equal('20px');
      expect(doc.documentElement.style.getPropertyValue('--ui-font-family')).to.contain('Georgia');
    });
  });

  it('should switch away from card layout without per-article settings requests', () => {
    let layoutMode = 'card';
    let settingsGetCount = 0;
    let settingsRequestsBeforeSwitch = 0;

    const settingsResponse = () => ({
      language: 'en-US',
      layout_mode: layoutMode,
      update_check_enabled: 'false',
      update_interval: '10',
      last_global_refresh: new Date().toISOString(),
    });

    const articles = Array.from({ length: 24 }, (_, index) => ({
      id: index + 1,
      feed_id: 1,
      feed_title: 'Test Feed',
      title: `Test Article ${index + 1}`,
      url: `https://example.com/articles/${index + 1}`,
      published_at: '2026-08-13T00:00:00Z',
      image_url: '',
      translated_title: '',
      is_read: false,
      is_favorite: false,
      is_hidden: false,
      is_read_later: false,
    }));

    // Reload with deterministic API state so the regression does not depend on local data.
    cy.intercept('/api/**', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: [] }).as('layoutFeeds');
    cy.intercept('GET', '/api/tags', { statusCode: 200, body: [] });
    cy.intercept('GET', '/api/saved-filters', { statusCode: 200, body: [] });
    cy.intercept(
      { method: 'GET', pathname: '/api/articles' },
      {
        statusCode: 200,
        body: articles,
      }
    ).as('layoutArticles');
    cy.intercept('GET', '/api/articles/unread-counts', {
      statusCode: 200,
      body: { total: 0, feeds: {}, categories: {} },
    });
    cy.intercept('GET', '/api/articles/filter-counts', { statusCode: 200, body: {} });
    cy.intercept('GET', '/api/progress', { statusCode: 200, body: { is_running: false } });
    cy.intercept('GET', '/api/settings', (req) => {
      settingsGetCount += 1;
      req.reply({ statusCode: 200, body: settingsResponse() });
    }).as('layoutSettings');
    cy.intercept('POST', '/api/settings', (req) => {
      layoutMode = req.body.layout_mode;
      req.reply({ statusCode: 200, body: { success: true } });
    }).as('saveLayoutSettings');

    const layoutSelector = () =>
      cy.contains('.setting-item', 'Article List Layout').find('button.select-trigger');

    const selectLayout = (label: string) => {
      layoutSelector().click();
      cy.contains('.select-option', label).click({ force: true });
    };

    cy.reload();
    cy.wait('@layoutFeeds');
    cy.wait('@layoutArticles');

    cy.get('button[title="Settings"]').click();
    cy.contains('button', /^Reading$/).click();
    layoutSelector().should('contain.text', 'Card');
    cy.wait(150);

    cy.then(() => {
      settingsRequestsBeforeSwitch = settingsGetCount;
    });

    selectLayout('Normal');
    cy.wait('@saveLayoutSettings').its('request.body.layout_mode').should('eq', 'normal');
    layoutSelector().should('contain.text', 'Normal');
    cy.then(() => {
      expect(settingsGetCount - settingsRequestsBeforeSwitch).to.be.at.most(5);
    });

    selectLayout('Compact');
    cy.wait('@saveLayoutSettings').its('request.body.layout_mode').should('eq', 'compact');
    layoutSelector().should('contain.text', 'Compact');

    selectLayout('Card');
    cy.wait('@saveLayoutSettings').its('request.body.layout_mode').should('eq', 'card');
    layoutSelector().should('contain.text', 'Card');

    selectLayout('Normal');
    cy.wait('@saveLayoutSettings').its('request.body.layout_mode').should('eq', 'normal');
    layoutSelector().should('contain.text', 'Normal');

    cy.get('body').type('{esc}');
    cy.reload();
    cy.wait('@layoutFeeds');
    cy.wait('@layoutArticles');
    cy.get('button[title="Settings"]').click();
    cy.contains('button', /^Reading$/).click();
    layoutSelector().should('contain.text', 'Normal');
    cy.then(() => {
      expect(layoutMode).to.eq('normal');
    });
  });

  it('updates the AI usage limit immediately and refreshes usage while visible', () => {
    let usage = 120;
    let savedLimit = '20000';
    let usageRequests = 0;

    cy.intercept('GET', '/api/ai/profiles', { statusCode: 200, body: [] });
    cy.intercept('GET', '/api/ai-usage', (req) => {
      usageRequests += 1;
      const limit = Number(savedLimit);
      req.reply({
        statusCode: 200,
        body: {
          usage,
          limit,
          limit_reached: limit > 0 && usage >= limit,
        },
      });
    }).as('getAIUsage');
    cy.intercept('POST', '/api/settings', (req) => {
      if (req.body.ai_usage_limit !== undefined) {
        savedLimit = String(req.body.ai_usage_limit);
      }
      req.reply({ statusCode: 200, body: req.body });
    }).as('saveAIUsageLimit');

    cy.clock();
    cy.get('button')
      .filter('[title="Settings"], [title="设置"]')
      .should('exist')
      .click({ force: true });
    cy.wait('@getSettings');
    cy.contains('button', /^AI$/i).click({ force: true });
    cy.wait('@getAIUsage');

    cy.get('[data-testid="ai-usage-status"]').should('contain.text', '120 / 20,000');
    cy.get('[data-testid="ai-usage-limit-input"]').clear().type('400');
    cy.get('[data-testid="ai-usage-status"]').should('contain.text', '120 / 400');

    cy.tick(500);
    cy.wait('@saveAIUsageLimit').its('request.body.ai_usage_limit').should('eq', '400');
    cy.wait('@getAIUsage');
    cy.get('[data-testid="ai-usage-status"]').should('contain.text', '120 / 400');

    cy.then(() => {
      usage = 300;
    });
    cy.tick(15_000);
    cy.wait('@getAIUsage');
    cy.get('[data-testid="ai-usage-status"]').should('contain.text', '300 / 400');

    let requestsBeforeUnmount = 0;
    cy.then(() => {
      requestsBeforeUnmount = usageRequests;
    });
    cy.contains('button', /^(General|常规)$/i).click({ force: true });
    cy.tick(30_000);
    cy.then(() => {
      expect(usageRequests).to.eq(requestsBeforeUnmount);
    });
  });
});
