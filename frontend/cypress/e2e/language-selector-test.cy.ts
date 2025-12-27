/// <reference types="cypress" />

describe('Language Selector Test', () => {
  beforeEach(() => {
    cy.intercept('GET', '/api/settings').as('getSettings');
    cy.intercept('POST', '/api/settings').as('saveSettings');
    cy.intercept('GET', '/api/feeds').as('getFeeds');

    cy.visit('/');
    cy.get('body').should('be.visible');
  });

  it('should find all select elements in settings', () => {
    cy.wait('@getFeeds', { timeout: 10000 });

    // Open settings
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist').click({ force: true });
    cy.wait('@getSettings');

    // Navigate to general tab
    cy.contains(/general|常规/i).click({ force: true });

    // List all select elements and their options
    cy.get('select').should('exist').then(($selects) => {
      cy.log(`Found ${$selects.length} select elements`);

      $selects.each((index, select) => {
        const $select = Cypress.$(select);
        const options = $select.find('option').map((i, opt) => opt.textContent).get();
        cy.log(`Select ${index}: ${options.join(', ')}`);
      });
    });
  });
});
