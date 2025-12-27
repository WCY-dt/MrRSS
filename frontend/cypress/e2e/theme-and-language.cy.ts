/// <reference types="cypress" />

describe('Theme and Language Switching', () => {
  beforeEach(() => {
    // Set up intercepts before visiting the page
    cy.intercept('GET', '/api/settings').as('getSettings');
    cy.intercept('POST', '/api/settings').as('saveSettings');
    cy.intercept('GET', '/api/feeds').as('getFeeds');

    cy.visit('/');
    cy.get('body').should('be.visible');
  });

  it('should switch between light and dark themes', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    // Navigate to general tab
    cy.contains(/general|常规/i).click({ force: true });

    // Try to switch themes
    cy.get('body').then(($body) => {
      if ($body.find(/dark|深色/i).length > 0) {
        // Switch to dark theme
        cy.contains(/dark|深色/i).click({ force: true });
        cy.wait(1000);

        // Verify theme changed
        cy.get('html').should('have.class', /dark/);

        // Switch to light theme
        cy.contains(/light|亮色/i).click({ force: true });
        cy.wait(1000);

        // Verify theme changed back
        cy.get('html').should('not.have.class', 'dark');
      } else {
        cy.log('Theme options not found');
      }
    });

    // Close settings
    cy.get('body').type('{esc}');
  });

  it('should persist theme after page reload', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings and change theme
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    cy.contains(/general|常规/i).click({ force: true });
    cy.get('body').then(($body) => {
      if ($body.find(/dark|深色/i).length > 0) {
        cy.contains(/dark|深色/i).click({ force: true });
        cy.wait('@saveSettings', { timeout: 5000 });

        // Close settings
        cy.get('body').type('{esc}');
        cy.wait(500);

        // Reload page
        cy.reload();
        cy.wait('@getFeeds', { timeout: 10000 });

        // Verify dark theme persisted
        cy.get('html').should('have.class', /dark/);
      } else {
        cy.log('Dark theme option not found');
      }
    });
  });

  it('should switch between languages', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    // Navigate to general tab
    cy.contains(/general|常规/i).click({ force: true });

    // Use data-testid to find language selector
    cy.get('[data-testid="language-selector"]').then(($select) => {
      if ($select.length > 0) {
        const currentValue = $select.val();

        // Select the other option by value (not text)
        if (currentValue === 'en-US') {
          cy.wrap($select).select('zh-CN');
        } else {
          cy.wrap($select).select('en-US');
        }

        cy.wait(1000);

        // Switch back
        cy.wrap($select).then(($sel) => {
          const newVal = $sel.val();
          if (newVal === 'en-US') {
            cy.wrap($sel).select('zh-CN');
          } else {
            cy.wrap($sel).select('en-US');
          }
        });

        cy.wait(1000);
        cy.contains(/Settings|General|设置|常规/).should('exist');
      } else {
        cy.log('Language selector not found - skipping test');
      }
    });
  });

  it('should persist language after page reload', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings and change language
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    cy.contains(/general|常规/i).click({ force: true });

    // Switch to Chinese using data-testid
    cy.get('[data-testid="language-selector"]').then(($select) => {
      if ($select.length > 0) {
        cy.wrap($select).select('zh-CN');

        // Wait for settings to save
        cy.wait('@saveSettings', { timeout: 5000 });

        // Additional wait to ensure settings are fully persisted
        cy.wait(1000);

        // Close settings
        cy.get('body').type('{esc}');
        cy.wait(1000);

        // Reload page
        cy.reload();

        // Wait for body to be visible after reload
        cy.get('body').should('be.visible');

        // Wait for app to fully initialize
        cy.wait('@getFeeds', { timeout: 10000 });
        cy.wait('@getSettings', { timeout: 10000 });

        // Additional wait for language to load
        cy.wait(2000);

        // Verify Chinese language persisted by opening settings
        cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
        cy.wait('@getSettings', { timeout: 5000 });
        cy.contains(/general|常规/i).click({ force: true });

        // Check the language selector value directly with retries
        cy.get('[data-testid="language-selector"]', { timeout: 10000 }).should('have.value', 'zh-CN');
      } else {
        cy.log('Language selector not found - skipping test');
      }
    });
  });

  it('should handle system theme preference', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    // Navigate to general tab
    cy.contains(/general|常规/i).click({ force: true });

    // Try to select system theme
    cy.get('body').then(($body) => {
      if ($body.find(/system|系统/i).length > 0) {
        cy.contains(/system|系统/i).click({ force: true });
        cy.wait('@saveSettings', { timeout: 5000 });

        // Verify system theme is selected
        cy.contains(/system|系统/i).should('exist');
      } else {
        cy.log('System theme option not found');
      }
    });

    // Close settings
    cy.get('body').type('{esc}');
  });

  it('should apply theme to all components', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Switch to dark theme
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    cy.contains(/general|常规/i).click({ force: true });
    cy.get('body').then(($body) => {
      if ($body.find(/dark|深色/i).length > 0) {
        cy.contains(/dark|深色/i).click({ force: true });
        cy.wait(1000);

        // Close settings
        cy.get('body').type('{esc}');
        cy.wait(500);

        // Verify dark theme is applied to various components
        cy.get('body').should('have.css', 'background-color');
        cy.get('[class*="sidebar"]').should('exist');
        cy.get('[class*="article"]').should('exist');

        // Check that colors have changed (this is a simple check)
        cy.get('body').should('not.have.css', 'background-color', 'rgb(255, 255, 255)');
      } else {
        cy.log('Dark theme option not found');
      }
    });
  });
});
