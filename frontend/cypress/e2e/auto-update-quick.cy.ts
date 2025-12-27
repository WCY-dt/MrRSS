/// <reference types="cypress" />

/**
 * Quick test for auto-update feature
 * This file focuses on the most important scenarios that are hard to test manually
 */

describe('Auto Update - Quick Validation', () => {
  beforeEach(() => {
    // Basic intercepts
    cy.intercept('GET', '/api/feeds', { statusCode: 200, body: [] }).as('getFeeds');
    cy.intercept('GET', '/api/articles', { statusCode: 200, body: [] }).as('getArticles');
    cy.intercept('GET', '/api/version', {
      statusCode: 200,
      body: { version: '1.3.12' },
    }).as('getVersion');
  });

  it('Scenario 1: Auto-update DISABLED - Should show dialog on startup', () => {
    // Mock settings with auto_update = false
    cy.intercept('GET', '/api/settings', {
      statusCode: 200,
      body: {
        auto_update: false,
        theme: 'light',
        language: 'en-US',
      },
    }).as('getSettings');

    // Mock update available
    cy.intercept('GET', '/api/check-updates', {
      statusCode: 200,
      body: {
        has_update: true,
        current_version: '1.3.12',
        latest_version: '1.3.13',
        download_url: 'https://example.com/update.exe',
      },
    }).as('checkUpdates');

    cy.visit('/');

    // Wait for app to load
    cy.wait('@getFeeds');

    // Wait for automatic update check (happens after 3 seconds)
    cy.wait(4000);

    // ✅ Verify: Update dialog should be shown
    cy.contains(/update available|有可用更新/i, { timeout: 5000 })
      .should('be.visible')
      .screenshot('dialog-when-auto-update-disabled');

    // ✅ Verify: Both buttons should be visible
    cy.contains(/not now|暂不更新/i).should('be.visible');
    cy.contains(/update now|立即更新/i).should('be.visible');

    cy.log('✅ Test Passed: Dialog shown when auto-update is disabled');
  });

  it('Scenario 2: Auto-update ENABLED - Should NOT show dialog', () => {
    // Mock settings with auto_update = true
    cy.intercept('GET', '/api/settings', {
      statusCode: 200,
      body: {
        auto_update: true,
        theme: 'light',
        language: 'en-US',
      },
    }).as('getSettings');

    // Mock update available
    cy.intercept('GET', '/api/check-updates', {
      statusCode: 200,
      body: {
        has_update: true,
        current_version: '1.3.12',
        latest_version: '1.3.13',
        download_url: 'https://example.com/update.exe',
      },
    }).as('checkUpdates');

    // Mock download and install endpoints
    cy.intercept('POST', '/api/download-update', {
      statusCode: 200,
      body: { success: true, file_path: '/path/to/update.exe' },
    }).as('downloadUpdate');

    cy.intercept('POST', '/api/install-update', {
      statusCode: 200,
      body: { success: true },
    }).as('installUpdate');

    cy.visit('/');

    // Wait for app to load
    cy.wait('@getFeeds');

    // Wait for automatic update check and install
    cy.wait(5000);

    // ✅ Verify: Update dialog should NOT be shown
    cy.contains(/update available|有可用更新/i).should('not.exist');

    // ✅ Verify: Download was triggered automatically
    cy.wait('@downloadUpdate', { timeout: 10000 });

    cy.log('✅ Test Passed: No dialog shown, auto-download triggered when auto-update is enabled');
  });

  it('Scenario 3: User clicks "Not Now" - Dialog should close', () => {
    // Mock settings with auto_update = false
    cy.intercept('GET', '/api/settings', {
      statusCode: 200,
      body: {
        auto_update: false,
        theme: 'light',
        language: 'en-US',
      },
    }).as('getSettings');

    // Mock update available
    cy.intercept('GET', '/api/check-updates', {
      statusCode: 200,
      body: {
        has_update: true,
        current_version: '1.3.12',
        latest_version: '1.3.13',
        download_url: 'https://example.com/update.exe',
      },
    }).as('checkUpdates');

    cy.visit('/');
    cy.wait('@getFeeds');

    // Wait for update dialog
    cy.wait(4000);
    cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

    // Click "Not Now"
    cy.contains(/not now|暂不更新/i).click();

    // ✅ Verify: Dialog should close
    cy.contains(/update available|有可用更新/i).should('not.exist');

    cy.log('✅ Test Passed: Dialog closes when clicking "Not Now"');
  });

  it('Scenario 4: User clicks "Update Now" - Should trigger download', () => {
    // Mock settings with auto_update = false
    cy.intercept('GET', '/api/settings', {
      statusCode: 200,
      body: {
        auto_update: false,
        theme: 'light',
        language: 'en-US',
      },
    }).as('getSettings');

    // Mock update available
    cy.intercept('GET', '/api/check-updates', {
      statusCode: 200,
      body: {
        has_update: true,
        current_version: '1.3.12',
        latest_version: '1.3.13',
        download_url: 'https://example.com/update.exe',
      },
    }).as('checkUpdates');

    // Mock download endpoint
    cy.intercept('POST', '/api/download-update', {
      statusCode: 200,
      body: { success: true, file_path: '/path/to/update.exe' },
    }).as('downloadUpdate');

    cy.visit('/');
    cy.wait('@getFeeds');

    // Wait for update dialog
    cy.wait(4000);
    cy.contains(/update available|有可用更新/i, { timeout: 5000 }).should('be.visible');

    // Click "Update Now"
    cy.contains(/update now|立即更新/i).click();

    // ✅ Verify: Download should be triggered
    cy.wait('@downloadUpdate', { timeout: 10000 });

    // ✅ Verify: Progress indicator should appear
    cy.contains(/downloading|正在下载/i).should('be.visible');

    cy.log('✅ Test Passed: Download triggered when clicking "Update Now"');
  });

  it('Scenario 5: Toggle auto-update setting in UI', () => {
    cy.intercept('GET', '/api/settings', {
      fixture: 'settings.json',
    }).as('getSettings');

    cy.intercept('POST', '/api/settings', { statusCode: 200 }).as('saveSettings');

    cy.visit('/');
    cy.wait('@getFeeds');

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

    // Navigate to General tab
    cy.contains(/general|常规/i).click({ force: true });

    // Wait for settings to load
    cy.wait('@getSettings');

    // Auto-update setting is now in the Application section, below "startup on boot"
    // Find "startup on boot" first, then find "auto update" below it
    cy.contains(/startup on boot|开机自启动/i)
      .parent()
      .parent()
      .next()
      .contains(/auto update|自动更新/i)
      .should('exist')
      .parent()
      .find('input[type="checkbox"]')
      .as('autoUpdateToggle')
      .screenshot('auto-update-toggle');

    // Get initial state
    cy.get('@autoUpdateToggle').then(($checkbox) => {
      const initialState = $checkbox.prop('checked');
      cy.log(`Initial state: ${initialState}`);

      // Click to toggle
      cy.get('@autoUpdateToggle').click();

      // Wait for auto-save
      cy.wait('@saveSettings', { timeout: 5000 });

      // Verify state changed
      cy.get('@autoUpdateToggle').should('not.have.prop', 'checked', initialState);

      cy.log('✅ Test Passed: Auto-update toggle works correctly');
    });
  });

  it('Scenario 6: No update available - No dialog shown', () => {
    // Mock settings with auto_update = false
    cy.intercept('GET', '/api/settings', {
      statusCode: 200,
      body: {
        auto_update: false,
        theme: 'light',
        language: 'en-US',
      },
    }).as('getSettings');

    // Mock NO update available
    cy.intercept('GET', '/api/check-updates', {
      statusCode: 200,
      body: {
        has_update: false,
        current_version: '1.3.12',
        latest_version: '1.3.12',
      },
    }).as('checkUpdates');

    cy.visit('/');
    cy.wait('@getFeeds');

    // Wait for automatic update check
    cy.wait(4000);

    // ✅ Verify: Update dialog should NOT be shown
    cy.contains(/update available|有可用更新/i).should('not.exist');

    cy.log('✅ Test Passed: No dialog when no update available');
  });
});
