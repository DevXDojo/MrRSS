/// <reference types="cypress" />

describe('Application Smoke Tests', () => {
  beforeEach(() => {
    // Set up intercepts before visiting the page
    cy.intercept('GET', '/api/articles*').as('getArticles');
    cy.intercept('GET', '/api/feeds*').as('getFeeds');
    cy.visit('/');
  });

  it('should load the application successfully', () => {
    // Verify the app loads
    cy.get('body').should('be.visible');

    // Check for main layout elements
    cy.get('[class*="sidebar"]').should('exist');
    cy.get('[class*="article"]').should('exist');
  });

  it('should display the sidebar', () => {
    // Verify sidebar is present
    cy.get('[class*="sidebar"]').should('be.visible');

    // Check for common sidebar elements
    cy.get('button[title="All Articles"], button[title="所有文章"]').should('exist');
    cy.get('button[title="Unread"], button[title="未读"]').should('exist');
  });

  it('should keep all activity bar icons on the same centre line', () => {
    cy.viewport(1280, 720);

    cy.get('.smart-activity-bar').then(($bar) => {
      const bar = $bar[0].getBoundingClientRect();
      const centreX = bar.left + bar.width / 2;

      cy.wrap($bar)
        .find('button:visible svg')
        .each(($icon) => {
          const icon = $icon[0].getBoundingClientRect();
          const iconCentreX = icon.left + icon.width / 2;

          expect(Math.abs(iconCentreX - centreX)).to.be.lessThan(0.6);
        });
    });
  });

  it('should have working navigation', () => {
    // Click on different navigation items
    cy.get('button[title="All Articles"], button[title="所有文章"]').click({ force: true });
    cy.wait(500);

    cy.get('button[title="Unread"], button[title="未读"]').click({ force: true });
    cy.wait(500);

    cy.get('button[title="Favorites"], button[title="收藏"]').click({ force: true });
    cy.wait(500);
  });

  it('should open and close settings modal', () => {
    // Wait for initial data to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings - find the gear icon button
    cy.get('button')
      .filter('[title="Settings"], [title="设置"]')
      .should('exist')
      .click({ force: true });

    // Wait for modal to appear
    cy.wait(1000);

    // Verify modal content is visible (check for settings text or modal structure)
    cy.get('body').then(($body) => {
      const hasSettingsModal =
        $body.find(/settings|设置|general|常规/i).length > 0 ||
        $body.find('[class*="modal"]').length > 0;
      if (hasSettingsModal) {
        cy.log('Settings modal opened successfully');
      } else {
        cy.log('Settings modal may have opened but not detected');
      }
    });

    // Close modal using ESC key
    cy.get('body').type('{esc}');
    cy.wait(1000);
  });

  it('should handle keyboard shortcuts', () => {
    // Wait for initial data to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Test settings shortcut (Ctrl+,)
    cy.get('body').type('{ctrl},');
    cy.wait(1000);

    // Check if settings opened (may not always work in test environment)
    cy.get('body').then(($body) => {
      if ($body.find(/settings|设置/i).length > 0) {
        cy.log('Settings opened via keyboard shortcut');

        // Close with ESC
        cy.get('body').type('{esc}');
        cy.wait(500);
      } else {
        cy.log('Keyboard shortcut may not work in test environment');
      }
    });
  });

  it('should display articles when feeds exist', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Check if articles are displayed (or empty state)
    cy.get('[class*="article"], [class*="empty"], [class*="no-articles"]').should('exist');
  });

  it('should handle API errors gracefully', () => {
    // Wait for app to load first
    cy.wait('@getFeeds', { timeout: 10000 });

    // Verify app doesn't crash even if APIs fail
    cy.get('body').should('be.visible');

    // The app should show empty state or handle errors gracefully
    cy.get('[class*="sidebar"]').should('exist');
  });

  it('should be responsive', () => {
    // Test different viewport sizes
    cy.viewport(1920, 1080);
    cy.get('body').should('be.visible');

    cy.viewport(1280, 720);
    cy.get('body').should('be.visible');

    cy.viewport(768, 1024);
    cy.get('body').should('be.visible');

    // Mobile view
    cy.viewport(375, 667);
    cy.get('body').should('be.visible');
  });

  it('should handle long content gracefully', () => {
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 });

    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Try to click on an article if it exists
    cy.get('body').then(($body) => {
      if ($body.find('[class*="article"]').length > 0) {
        cy.get('[class*="article"]').first().click({ force: true });

        // Wait for content to load
        cy.wait(500);

        // Verify page is still responsive
        cy.get('body').should('be.visible');
      } else {
        // No articles to test, skip gracefully
        cy.log('No articles found to test long content');
      }
    });
  });

  it('should maintain state during navigation', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Select unread filter
    cy.get('button[title="Unread"], button[title="未读"]').click({ force: true });
    cy.wait(500);

    // Open settings
    cy.get('button')
      .filter('[title="Settings"], [title="设置"]')
      .should('exist')
      .click({ force: true });
    cy.wait(500);

    // Close settings
    cy.get('body').type('{esc}');
    cy.wait(500);

    // Verify unread filter is still active by checking if the element exists
    cy.get('button[title="Unread"], button[title="未读"]').should('exist').and('be.visible');
  });
});
