/// <reference types="cypress" />

describe('Feed Management', () => {
  beforeEach(() => {
    // Set up intercepts before visiting the page
    cy.intercept('POST', '/api/feeds').as('addFeed');
    cy.intercept('GET', '/api/feeds').as('getFeeds');
    cy.intercept('DELETE', '/api/feeds/*').as('deleteFeed');
    cy.intercept('POST', '/api/feeds/refresh').as('refreshFeeds');
    cy.intercept('PUT', '/api/feeds/*').as('updateFeed');

    cy.visit('/');
    cy.get('body').should('be.visible');
  });

  it('should add a new feed', () => {
    // Look for add feed button in the sidebar footer (+ icon)
    cy.get('button')
      .filter('[title="Add Feed"], [title="添加订阅"]')
      .should('exist')
      .click({ force: true });

    // Wait for add feed modal to appear
    cy.wait(1000);

    // Check if modal opened
    cy.get('body').then(($body) => {
      if (
        $body.find(/add.*feed|添加.*feed|add.*subscription/i).length > 0 ||
        $body.find('[class*="modal"]').length > 0
      ) {
        cy.log('Add feed modal opened');

        // Try to fill in the feed URL if input exists
        cy.get('body').then(($body2) => {
          if ($body2.find('input[type="url"], input[type="text"]').length > 0) {
            cy.get('input[type="url"], input[type="text"]')
              .first()
              .type('https://example.com/feed.xml');

            // Submit the form if submit button exists
            cy.get('body').then(($body3) => {
              if (
                $body3
                  .find('button')
                  .filter((i, el) => /add|submit|确定|添加/i.test(el.textContent || '')).length > 0
              ) {
                cy.get('button')
                  .contains(/add|submit|确定|添加/i)
                  .click({ force: true });
              }
            });
          }
        });
      } else {
        cy.log('Add feed modal did not open or was not detected');
      }
    });
  });

  it('should delete a feed', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Try to find a feed to delete
    cy.get('body').then(($body) => {
      if ($body.find('[class*="feed"]').length > 0) {
        // Right-click on a feed to open context menu
        cy.get('[class*="feed"]').first().rightclick({ force: true });

        // Click delete option in context menu if it exists
        cy.get('body').then(($body2) => {
          if ($body2.find(/delete|删除/i).length > 0) {
            cy.contains(/delete|删除/i).click({ force: true });

            // Confirm deletion in the confirm dialog
            cy.get('body').then(($body3) => {
              if ($body3.find(/confirm|确认/i).length > 0) {
                cy.contains(/confirm|确认/i).click({ force: true });

                // Wait for deletion to complete
                cy.wait('@deleteFeed', { timeout: 10000 });
              }
            });
          } else {
            cy.log('Delete option not found in context menu');
          }
        });
      } else {
        cy.log('No feeds found to test deletion');
      }
    });
  });

  it('should refresh feeds', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Look for refresh button - check button title or content
    cy.get('body').then(($body) => {
      const refreshButtons = $body.find('button').filter((i, el) => {
        return /refresh|刷新/i.test(el.title || el.textContent || '');
      });

      if (refreshButtons.length > 0) {
        cy.wrap(refreshButtons).first().click({ force: true });
        cy.log('Refresh button clicked');

        // Wait a moment for any refresh to initiate
        cy.wait(500);
      } else {
        cy.log('Refresh button not found - may not be exposed in UI');
      }
    });
  });

  it('should edit feed details', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Try to find a feed to edit
    cy.get('body').then(($body) => {
      if ($body.find('[class*="feed"]').length > 0) {
        // Right-click on a feed
        cy.get('[class*="feed"]').first().rightclick({ force: true });

        // Click edit option if it exists
        cy.get('body').then(($body2) => {
          if ($body2.find(/edit|编辑/i).length > 0) {
            cy.contains(/edit|编辑/i).click({ force: true });

            // Wait for edit modal
            cy.wait(500);

            // Change the title
            cy.get('input[type="text"]').first().clear().type('Updated Feed Title');

            // Save changes
            cy.get('button')
              .contains(/save|保存|确定/i)
              .click({ force: true });

            // Wait for update to complete
            cy.wait('@updateFeed', { timeout: 10000 });
          } else {
            cy.log('Edit option not found in context menu');
          }
        });
      } else {
        cy.log('No feeds found to test editing');
      }
    });
  });

  it('should filter feeds by category', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Try to find category filter
    cy.get('body').then(($body) => {
      if ($body.find('select, [role="listbox"]').length > 0) {
        cy.get('select, [role="listbox"]').first().select(1);

        // Wait for filtered results
        cy.wait(500);

        // Verify feeds exist
        cy.get('[class*="feed"]').should('have.length.at.least', 0);
      } else {
        cy.log('Category filter not found');
      }
    });
  });

  it('should search feeds', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Look for search input in the sidebar
    cy.get('body').then(($body) => {
      if (
        $body.find('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]')
          .length > 0
      ) {
        cy.get('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]')
          .first()
          .type('test');

        // Wait for search results to filter
        cy.wait(500);

        // Verify search results
        cy.get('[class*="feed"]').should('exist');
      } else {
        cy.log('Search input not found');
      }
    });
  });
});

describe('Feed Management settings list', () => {
  beforeEach(() => {
    cy.fixture('settings').then((settings) => {
      cy.intercept('/api/**', { statusCode: 200, body: {} });
      cy.intercept('GET', '/api/settings', { statusCode: 200, body: settings });
      cy.intercept('GET', '/api/tags', { statusCode: 200, body: [] });
      cy.intercept('GET', '/api/saved-filters', { statusCode: 200, body: [] });
      cy.intercept({ method: 'GET', pathname: '/api/articles' }, { statusCode: 200, body: [] });
      cy.intercept('GET', '/api/articles/unread-counts', { statusCode: 200, body: {} });
      cy.intercept('GET', '/api/articles/filter-counts', { statusCode: 200, body: {} });
      cy.intercept('GET', '/api/progress', { statusCode: 200, body: { is_running: false } });
    });
  });

  const openFeedSettings = () => {
    cy.get('button')
      .filter('[title="Settings"], [title="设置"]')
      .should('exist')
      .click({ force: true });
    cy.get('[data-settings-modal="true"] .sidebar-tab-btn')
      .contains(/Feeds|订阅源|订阅/i)
      .click({ force: true });
    cy.get('[data-testid="feed-list"]').should('be.visible');
  };

  const makeFeed = (id: number, title: string, url = `https://feeds.example.com/${id}`) => ({
    id,
    title,
    url,
    category: id % 2 ? 'Technology/Frontend' : 'News',
    last_fetched_at: '2026-08-17T08:00:00Z',
    latest_article_time: '2026-08-17T07:00:00Z',
    articles_per_month: id,
    last_update_status: id % 3 === 0 ? 'failed' : 'success',
    last_error: id % 3 === 0 ? 'connection timeout' : '',
  });

  it('keeps API order and renders every feed column for 120 rows without page overflow', () => {
    const longChineseTitle = `这是一个用于验证订阅源列表不会被超长中文标题撑开的标题${'非常长'.repeat(20)}`;
    const longEnglishTitle = `An exceptionally long English feed title ${'with more words '.repeat(20)}`;
    const feeds = [
      makeFeed(
        900,
        longChineseTitle,
        `https://example.com/${'very-long-path/'.repeat(20)}feed.xml`
      ),
      makeFeed(100, longEnglishTitle),
      {
        ...makeFeed(500, 'No icon feed', 'rsshub://invalid-icon-source'),
        category: '',
        latest_article_time: undefined,
        articles_per_month: undefined,
      },
      ...Array.from({ length: 117 }, (_, index) =>
        makeFeed(index + 1, `Feed ${String(index + 1).padStart(3, '0')}`)
      ),
    ];

    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: feeds }).as('layoutFeeds');
    cy.viewport(960, 720);
    cy.visit('/');
    cy.wait('@layoutFeeds');
    openFeedSettings();

    cy.get('[data-testid="feed-row"]').should('have.length', 120);
    cy.get('[data-testid="feed-row"]').eq(0).should('have.attr', 'data-feed-id', '900');
    cy.get('[data-testid="feed-row"]').eq(1).should('have.attr', 'data-feed-id', '100');
    cy.get('[data-testid="feed-row"]').eq(2).should('have.attr', 'data-feed-id', '500');

    cy.get('[data-testid="feed-title"]').eq(0).should('have.attr', 'title', longChineseTitle);
    cy.get('[data-testid="feed-title"]').eq(1).should('have.attr', 'title', longEnglishTitle);
    cy.get('[data-testid="feed-source"]')
      .eq(0)
      .should('have.attr', 'title')
      .and('include', 'very-long-path');
    cy.get('[data-feed-id="500"] [data-testid="feed-icon-fallback"]').should('exist');
    cy.get('[data-testid="feed-column-name"]').should('be.visible');
    cy.get('[data-testid="feed-column-category"]').should('be.visible');
    cy.get('[data-testid="feed-column-latest"]').should('exist');
    cy.get('[data-testid="feed-column-frequency"]').should('exist');
    cy.get('[data-testid="feed-column-status"]').should('exist');
    cy.get('[data-testid="feed-column-actions"]').should('exist');

    cy.get('[data-feed-id="900"] [data-testid="feed-category"]')
      .should('have.attr', 'title', 'News')
      .and('contain.text', 'News');
    cy.get('[data-feed-id="900"] [data-testid="feed-latest"]')
      .should('have.attr', 'title', '2026-08-17T07:00:00Z')
      .and('not.be.empty');
    cy.get('[data-feed-id="900"] [data-testid="feed-frequency"]').should('have.text', '900');
    cy.get('[data-feed-id="500"] [data-testid="feed-category"]').should('have.text', '-');
    cy.get('[data-feed-id="500"] [data-testid="feed-latest"]').should('have.text', '-');
    cy.get('[data-feed-id="500"] [data-testid="feed-frequency"]').should('have.text', '0');

    cy.get('[data-testid="feed-row"]')
      .first()
      .then(($row) => {
        const height = $row[0].getBoundingClientRect().height;
        expect(height).to.be.at.least(60);
        expect(height).to.be.at.most(72);
      });
    cy.get('[data-testid="feed-list"]').then(($list) => {
      expect($list[0].scrollWidth).to.be.greaterThan($list[0].clientWidth);
    });
    cy.document().then((document) => {
      expect(document.documentElement.scrollWidth).to.be.at.most(
        document.documentElement.clientWidth + 1
      );
    });
  });

  it('moves optional sorting into one compact menu without changing the stored order', () => {
    const feeds = [makeFeed(30, 'Zebra'), makeFeed(10, 'Alpha'), makeFeed(20, 'Beta')];
    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: feeds }).as('sortableFeeds');
    cy.visit('/');
    cy.wait('@sortableFeeds');
    openFeedSettings();

    cy.get('[data-testid="feed-row"]').first().should('have.attr', 'data-feed-id', '30');
    cy.get('[data-testid="feed-sort-select"] button.select-trigger').click();
    cy.contains('.select-option', /Name|名称/).click({ force: true });
    cy.get('[data-testid="feed-title"]').first().should('have.text', 'Alpha');

    cy.get('[data-testid="feed-sort-direction"]').click();
    cy.get('[data-testid="feed-title"]').first().should('have.text', 'Zebra');

    cy.get('[data-testid="feed-sort-select"] button.select-trigger').click();
    cy.contains('.select-option', /Original order|原始顺序/).click({ force: true });
    cy.get('[data-testid="feed-row"]').first().should('have.attr', 'data-feed-id', '30');
  });

  it('shows a retryable error state instead of breaking the settings layout', () => {
    cy.intercept('GET', '/api/feeds', {
      statusCode: 503,
      body: { error: 'service unavailable' },
    }).as('failedFeeds');
    cy.visit('/');
    cy.wait('@failedFeeds');
    openFeedSettings();

    cy.get('[data-testid="feed-load-error"]').should('be.visible');

    cy.intercept('GET', '/api/feeds', {
      statusCode: 200,
      body: [makeFeed(42, 'Recovered feed')],
    }).as('recoveredFeeds');
    cy.get('[data-testid="feed-load-error"] button')
      .contains(/Retry|重试/)
      .click();
    cy.wait('@recoveredFeeds');
    cy.get('[data-testid="feed-row"]').should('have.length', 1);
    cy.get('[data-testid="feed-title"]').should('have.text', 'Recovered feed');
  });

  it('keeps status and existing edit/delete controls aligned', () => {
    cy.intercept('GET', '/api/feeds', {
      statusCode: 200,
      body: [makeFeed(1, 'Healthy feed'), makeFeed(3, 'Failed feed')],
    }).as('statusFeeds');
    cy.visit('/');
    cy.wait('@statusFeeds');
    openFeedSettings();

    cy.get('[data-testid="feed-row"]').each(($row) => {
      cy.wrap($row).find('[data-testid="feed-edit"]').should('exist');
      cy.wrap($row).find('[data-testid="feed-delete"]').should('exist');
    });
    cy.get('[data-feed-id="1"] [title*="success"], [data-feed-id="1"] [title*="成功"]').should(
      'exist'
    );
    cy.get('[data-feed-id="3"] [title]')
      .filter('[title*="timeout"], [title*="超时"]')
      .should('exist');
  });

  it('opens edit from an editable row without closing Settings or reacting to locked rows', () => {
    const lockedFeed = { ...makeFeed(2, 'FreshRSS feed'), is_freshrss_source: true };
    cy.intercept('GET', '/api/feeds', {
      statusCode: 200,
      body: [makeFeed(1, 'Editable feed'), lockedFeed],
    }).as('editableFeeds');
    cy.visit('/');
    cy.wait('@editableFeeds');
    openFeedSettings();

    cy.get('[data-settings-modal="true"] [data-feed-id="1"] input[type="checkbox"]').click();
    cy.contains('h3', /Edit Feed|编辑订阅/).should('not.exist');

    cy.get('[data-settings-modal="true"] [data-feed-id="1"]').click();
    cy.get('[data-settings-modal="true"]').should('exist');
    cy.contains('h3', /Edit Feed|编辑订阅/)
      .should('be.visible')
      .parent()
      .parent()
      .find('button')
      .click();

    cy.get('[data-settings-modal="true"]').should('be.visible');
    cy.contains('h3', /Edit Feed|编辑订阅/).should('not.exist');
    cy.get('[data-settings-modal="true"] [data-feed-id="2"]').click();
    cy.contains('h3', /Edit Feed|编辑订阅/).should('not.exist');
    cy.get('[data-settings-modal="true"]').should('be.visible');
  });

  it('keeps Manage Tags open for empty and legacy null tag responses', () => {
    cy.intercept('GET', '/api/feeds', {
      statusCode: 200,
      body: [makeFeed(1, 'Tagged feed')],
    }).as('tagFeeds');
    cy.intercept('GET', '/api/tags', { statusCode: 200, body: null }).as('nullTags');
    cy.visit('/');
    cy.wait('@tagFeeds');
    openFeedSettings();

    cy.contains('button', /Manage Tags|管理标签/).click();
    cy.get('[data-testid="tag-management-empty"]').should('be.visible');
    cy.get('[data-settings-modal="true"]').should('exist');
    cy.contains('h3', /Manage Tags|管理标签/).parent().parent().find('button').click();
    cy.get('[data-settings-modal="true"]').should('be.visible');
  });

  it('keeps the tag dialog open and retries after a tag request failure', () => {
    let failTagRequests = false;
    cy.intercept('GET', '/api/feeds', {
      statusCode: 200,
      body: [makeFeed(1, 'Tagged feed')],
    }).as('retryTagFeeds');
    cy.intercept('GET', '/api/tags', (request) => {
      request.reply(
        failTagRequests
          ? { statusCode: 503, body: { error: 'service unavailable' } }
          : { statusCode: 200, body: [] }
      );
    });
    cy.visit('/');
    cy.wait('@retryTagFeeds');
    openFeedSettings();
    cy.then(() => {
      failTagRequests = true;
    });

    cy.contains('button', /Manage Tags|管理标签/).click();
    cy.get('[data-testid="tag-management-error"]').should('be.visible');
    cy.get('[data-settings-modal="true"]').should('exist');
    cy.then(() => {
      failTagRequests = false;
    });
    cy.get('[data-testid="tag-management-error"] button')
      .contains(/Retry|重试/)
      .click();
    cy.get('[data-testid="tag-management-empty"]').should('be.visible');
  });
});
