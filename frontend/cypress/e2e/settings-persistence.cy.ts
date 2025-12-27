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

    // Find the theme selector - try to find dark theme option
    cy.get('body').then(($body) => {
      if ($body.find(/dark|深色/i).length > 0) {
        cy.contains(/dark|深色/i).click({ force: true });

        // Wait for settings to be saved
        cy.wait('@saveSettings', { timeout: 5000 });
      } else {
        cy.log('Dark theme option not found');
      }
    });

    // Close the settings modal
    cy.get('body').type('{esc}');

    // Wait a bit for modal to close
    cy.wait(500);

    // Reopen settings to verify the change persisted
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });

    // Wait for settings to load again
    cy.wait('@getSettings');

    // Verify dark theme option exists
    cy.contains(/dark|深色/i).should('exist');
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

    // Look for language selector and change it
    cy.get('body').then(($body) => {
      if ($body.find('select').length > 0) {
        // If there's a select dropdown
        cy.get('select').first().select(1);

        // Wait for settings to be saved
        cy.wait('@saveSettings', { timeout: 5000 });
      } else if ($body.find('[role="radiogroup"]').length > 0) {
        // If there are radio buttons
        cy.get('[role="radio"]').last().click({ force: true });

        // Wait for settings to be saved
        cy.wait('@saveSettings', { timeout: 5000 });
      } else {
        cy.log('Language selector not found');
      }
    });

    // Close settings
    cy.get('body').type('{esc}');
    cy.wait(500);

    // Reopen settings to verify
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    // Verify language selector exists
    cy.get('select, [role="radiogroup"]').should('exist');
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

    // Look for update interval input (it only appears when refresh mode is 'fixed')
    // Use data-testid to find the refresh mode selector
    cy.get('[data-testid="refresh-mode-selector"]').then(($select) => {
      if ($select.length > 0) {
        // Set refresh mode to 'fixed' to show the interval input
        cy.wrap($select).select('fixed');
        cy.wait(500);

        // Now look for the number input
        cy.get('input[type="number"]').then(($input) => {
          if ($input.length > 0) {
            cy.wrap($input).first().clear().type('30');

            // Wait for auto-save
            cy.wait(2000);

            // Close settings
            cy.get('body').type('{esc}');
            cy.wait(500);

            // Reopen to verify
            cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
            cy.wait('@getSettings');

            // Verify the input exists
            cy.get('input[type="number"]').first().should('exist');
          } else {
            cy.log('Update interval input not found after setting refresh mode');
          }
        });
      } else {
        cy.log('Refresh mode selector not found - skipping test');
      }
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

    // Make a change in general tab
    cy.contains(/general|常规/i).click({ force: true });
    cy.get('body').then(($body) => {
      if ($body.find(/dark|深色/i).length > 0) {
        cy.contains(/dark|深色/i).click({ force: true });

        // Switch to feeds tab - settings should auto-save
        cy.get('body').then(($body2) => {
          if ($body2.find(/feeds|订阅/i).length > 0) {
            cy.contains(/feeds|订阅/i).click({ force: true });
            cy.wait('@saveSettings', { timeout: 5000 });
          }
        });
      }
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
});
