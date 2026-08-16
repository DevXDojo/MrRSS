/// <reference types="cypress" />

describe('Auto Update Feature', () => {
  const mockUpdateInfo = {
    has_update: true,
    current_version: '1.3.12',
    latest_version: '1.3.13',
    download_url: 'https://github.com/DevXDojo/MrRSS/releases/download/v1.3.13/MrRSS-1.3.13-windows-amd64-installer.exe',
    asset_name: 'MrRSS-1.3.13-windows-amd64-installer.exe',
  };

  const mockNoUpdateInfo = {
    has_update: false,
    current_version: '1.3.12',
    latest_version: '1.3.12',
  };

  beforeEach(() => {
    // Set up intercepts for all API calls
    cy.intercept('GET', '/api/settings', {
      fixture: 'settings.json',
    }).as('getSettings');

    cy.intercept('POST', '/api/settings', { statusCode: 200 }).as('saveSettings');

    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: [] }).as('getFeeds');

    cy.intercept('GET', '/api/articles', { statusCode: 200, body: [] }).as('getArticles');

    // Mock version endpoint
    cy.intercept('GET', '/api/version', {
      statusCode: 200,
      body: { version: '1.3.12' },
    }).as('getVersion');

    cy.visit('/');

    // Wait for the app to be fully loaded
    cy.get('body').should('be.visible');
  });

  describe('Auto Update Setting UI', () => {
    it('should display auto update toggle in settings', () => {
      cy.wait('@getFeeds', { timeout: 10000 });

      // Open settings modal
      cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

      // Wait for settings modal to be visible
      cy.contains(/settings|设置/i).should('be.visible');

      // Navigate to General tab
      cy.contains(/general|常规/i).click({ force: true });

      // Verify auto update toggle exists (in Application section)
      cy.contains(/auto update|自动更新/i).should('be.visible');

      // Verify the toggle/checkbox exists
      cy.get('input[type="checkbox"]').should('exist');
    });

    it('should toggle auto update setting', () => {
      cy.wait('@getFeeds', { timeout: 10000 });

      // Open settings modal
      cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

      cy.contains(/general|常规/i).click({ force: true });
      cy.wait('@getSettings');

      // Find the auto update toggle (below "startup on boot")
      cy.contains(/startup on boot|开机自启动/i)
        .parent()
        .parent()
        .next()
        .contains(/auto update|自动更新/i)
        .parent()
        .find('input[type="checkbox"]')
        .as('autoUpdateToggle');

      // Get initial state
      cy.get('@autoUpdateToggle').then(($checkbox) => {
        const initialState = $checkbox.prop('checked');

        // Click the toggle to change state
        cy.get('@autoUpdateToggle').click();

        // Wait for settings to be saved (auto-save)
        cy.wait('@saveSettings', { timeout: 5000 });

        // Verify the state changed
        cy.get('@autoUpdateToggle').should('not.have.prop', 'checked', initialState);
      });
    });

    it('should persist auto update setting after closing and reopening settings', () => {
      cy.wait('@getFeeds', { timeout: 10000 });

      // Open settings and enable auto update
      cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

      cy.contains(/general|常规/i).click({ force: true });
      cy.wait('@getSettings');

      // Find and toggle the auto update setting
      cy.contains(/startup on boot|开机自启动/i)
        .parent()
        .parent()
        .next()
        .contains(/auto update|自动更新/i)
        .parent()
        .find('input[type="checkbox"]')
        .as('autoUpdateToggle')
        .click();

      cy.wait('@saveSettings', { timeout: 5000 });

      // Close settings modal
      cy.get('body').type('{esc}');
      cy.wait(500);

      // Reopen settings
      cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
      cy.wait('@getSettings');

      // Navigate to general tab
      cy.contains(/general|常规/i).click({ force: true });

      // Verify the setting is still there
      cy.contains(/auto update|自动更新/i).should('be.visible');
      cy.get('input[type="checkbox"]').should('exist');
    });
  });

  describe('Update Available Dialog', () => {
    it('should show update dialog when update is available and auto-update is disabled', () => {
      // Mock settings with auto_update disabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: false,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check API to return update available
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for automatic update check (happens after 3 seconds)
      cy.wait(4000);

      // Verify update dialog is shown
      cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

      // Verify dialog content
      cy.contains(/1\.3\.13/).should('be.visible');
      cy.contains(/not now|暂不更新/i).should('be.visible');
      cy.contains(/update now|立即更新/i).should('be.visible');
    });

    it('should not show dialog when auto-update is enabled', () => {
      // Mock settings with auto_update enabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: true,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check and download/install APIs
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      cy.intercept('POST', '/api/download-update', {
        statusCode: 200,
        body: { success: true, file_path: '/path/to/update.exe' },
      }).as('downloadUpdate');

      cy.intercept('POST', '/api/install-update', {
        statusCode: 200,
        body: { success: true },
      }).as('installUpdate');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for automatic update check and download
      cy.wait(5000);

      // Verify dialog is NOT shown
      cy.contains(/update available|有可用更新/i).should('not.exist');

      // Verify download was triggered
      cy.wait('@downloadUpdate', { timeout: 10000 });
    });

    it('should close update dialog when clicking "Not Now"', () => {
      // Mock settings with auto_update disabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: false,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check API
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for update dialog to appear
      cy.wait(4000);
      cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

      // Click "Not Now" button
      cy.contains(/not now|暂不更新/i).click();

      // Verify dialog is closed
      cy.contains(/update available|有可用更新/i).should('not.exist');
    });

    it('should trigger update when clicking "Update Now"', () => {
      // Mock settings with auto_update disabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: false,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check and download/install APIs
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      cy.intercept('POST', '/api/download-update', {
        statusCode: 200,
        body: { success: true, file_path: '/path/to/update.exe' },
      }).as('downloadUpdate');

      cy.intercept('POST', '/api/install-update', {
        statusCode: 200,
        body: { success: true },
      }).as('installUpdate');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for update dialog to appear
      cy.wait(4000);
      cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

      // Click "Update Now" button
      cy.contains(/update now|立即更新/i).click();

      // Verify download is triggered
      cy.wait('@downloadUpdate', { timeout: 10000 });

      // Verify installing message appears
      cy.contains(/installing|正在安装/i).should('be.visible');
    });

    it('should not show dialog when no update is available', () => {
      // Mock settings with auto_update disabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: false,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check API to return no update
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockNoUpdateInfo,
      }).as('checkUpdates');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for automatic update check
      cy.wait(4000);

      // Verify update dialog is NOT shown
      cy.contains(/update available|有可用更新/i).should('not.exist');
    });
  });

  describe('Manual Update Check from About Tab', () => {
    it('should allow manual update check from About tab', () => {
      cy.wait('@getFeeds', { timeout: 10000 });

      // Open settings modal
      cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

      // Navigate to About tab
      cy.contains(/about|关于/i).click({ force: true });

      // Mock update check API
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      // Click check for updates button
      cy.contains(/check for updates|检查更新/i).click();

      // Wait for update check to complete
      cy.wait('@checkUpdates');

      // Verify update information is displayed
      cy.contains(/1\.3\.13/).should('be.visible');
      cy.contains(/update available|有可用更新/i).should('be.visible');
    });

    it('should show up to date message when no update available', () => {
      cy.wait('@getFeeds', { timeout: 10000 });

      // Open settings modal
      cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

      // Navigate to About tab
      cy.contains(/about|关于/i).click({ force: true });

      // Mock update check API to return no update
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockNoUpdateInfo,
      }).as('checkUpdates');

      // Click check for updates button
      cy.contains(/check for updates|检查更新/i).click();

      // Wait for update check to complete
      cy.wait('@checkUpdates');

      // Verify up to date message is displayed
      cy.contains(/up to date|最新版本/i).should('be.visible');
    });
  });

  describe('Update Progress Display', () => {
    it('should show download progress in dialog', () => {
      // Mock settings with auto_update disabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: false,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check API
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      // Mock slow download to show progress
      cy.intercept('POST', '/api/download-update', {
        statusCode: 200,
        body: { success: true, file_path: '/path/to/update.exe' },
        delay: 2000,
      }).as('downloadUpdate');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for update dialog to appear
      cy.wait(4000);
      cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

      // Click "Update Now" button
      cy.contains(/update now|立即更新/i).click();

      // Verify downloading message appears
      cy.contains(/downloading|正在下载/i).should('be.visible');

      // Verify progress bar exists
      cy.get('.bg-accent.h-full.transition-all').should('exist');
    });

    it('should show installing message after download completes', () => {
      // Mock settings with auto_update disabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: false,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check API
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      // Mock download and install
      cy.intercept('POST', '/api/download-update', {
        statusCode: 200,
        body: { success: true, file_path: '/path/to/update.exe' },
      }).as('downloadUpdate');

      cy.intercept('POST', '/api/install-update', {
        statusCode: 200,
        body: { success: true },
        delay: 1000,
      }).as('installUpdate');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for update dialog to appear
      cy.wait(4000);
      cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

      // Click "Update Now" button
      cy.contains(/update now|立即更新/i).click();

      // Wait for download to complete
      cy.wait('@downloadUpdate');

      // Verify installing message appears
      cy.contains(/installing|正在安装/i).should('be.visible');

      // Verify gear icon (spinner) is visible
      cy.get('.animate-spin').should('exist');
    });
  });

  describe('Error Handling', () => {
    it('should handle download failure gracefully', () => {
      // Mock settings with auto_update disabled
      cy.intercept('GET', '/api/settings', {
        statusCode: 200,
        body: {
          auto_update: false,
          theme: 'light',
          language: 'en-US',
          update_interval: '30',
        },
      }).as('getSettings');

      // Mock update check API
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 200,
        body: mockUpdateInfo,
      }).as('checkUpdates');

      // Mock download failure
      cy.intercept('POST', '/api/download-update', {
        statusCode: 500,
        body: { error: 'Download failed' },
      }).as('downloadUpdate');

      cy.visit('/');
      cy.wait('@getFeeds');

      // Wait for update dialog to appear
      cy.wait(4000);
      cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

      // Click "Update Now" button
      cy.contains(/update now|立即更新/i).click();

      // Wait for download attempt
      cy.wait('@downloadUpdate');

      // Verify error message is shown (via toast notification)
      cy.get('.toast-container').should('exist');
    });

    it('should handle update check failure gracefully', () => {
      cy.wait('@getFeeds', { timeout: 10000 });

      // Open settings modal
      cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

      // Navigate to About tab
      cy.contains(/about|关于/i).click({ force: true });

      // Mock update check failure
      cy.intercept('GET', '/api/check-updates', {
        statusCode: 500,
        body: { error: 'Network error' },
      }).as('checkUpdates');

      // Click check for updates button
      cy.contains(/check for updates|检查更新/i).click();

      // Wait for update check attempt
      cy.wait('@checkUpdates');

      // Verify error is handled gracefully (no crash)
      cy.get('body').should('be.visible');
    });
  });
});
