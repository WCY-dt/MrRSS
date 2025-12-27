/// <reference types="cypress" />

/**
 * Simple verification test - checks if the application loads
 * This test works with wails3 dev mode (localhost:9245)
 */
describe('Application Loading Test', () => {
  const baseUrl = Cypress.config().baseUrl as string
  const isCI = Cypress.env('isCI')

  it('should connect to the application', () => {
    cy.log('Testing URL:', baseUrl)
    cy.log('Environment:', isCI ? 'CI' : 'Local Development')

    // Visit the application
    cy.visit('/')

    // The app should load (check for body element)
    cy.get('body').should('exist')
  })

  it('should display the main app container', () => {
    cy.visit('/')

    // Wait for the app to mount
    cy.get('#app', { timeout: 10000 }).should('exist')
  })

  it('should have Vue app mounted', () => {
    cy.visit('/')

    // Check if Vue app is mounted (the #app div should have content)
    cy.get('#app', { timeout: 10000 }).should('not.be.empty')
  })
})
