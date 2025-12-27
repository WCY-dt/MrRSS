/// <reference types="cypress" />

describe('Quick Smoke Test', () => {
  it('should load the application', () => {
    cy.visit('/');
    cy.get('body').should('be.visible');
  });

  it('should find settings button', () => {
    cy.visit('/');
    cy.get('button').filter('[title="Settings"], [title="设置"]').should('exist');
  });

  it('should find add feed button', () => {
    cy.visit('/');
    cy.get('button').filter('[title="Add Feed"], [title="添加订阅"]').should('exist');
  });
});
