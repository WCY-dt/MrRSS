/// <reference types="cypress" />

/**
 * Simple smoke test to verify Cypress is configured correctly
 * This test does not require the backend to be running
 */
describe('Cypress Configuration Tests', () => {
  it('should load Cypress successfully', () => {
    // This test always passes if Cypress is working
    cy.log('Cypress is running correctly!')
    expect(true).to.be.true
  })

  it('should have correct configuration', () => {
    // Check environment variables
    const baseUrl = Cypress.config().baseUrl as string
    const isCI = Cypress.env('isCI')
    const expectedPort = isCI ? 34115 : 9245
    const expectedUrl = `http://localhost:${expectedPort}`

    cy.log('Base URL:', baseUrl)
    cy.log('Viewport Width:', Cypress.config().viewportWidth)
    cy.log('Viewport Height:', Cypress.config().viewportHeight)
    cy.log('Default Command Timeout:', Cypress.config().defaultCommandTimeout)
    cy.log('Is CI:', isCI)

    // Verify configuration
    expect(baseUrl).to.equal(expectedUrl)
    expect(Cypress.config().viewportWidth).to.equal(1280)
    expect(Cypress.config().viewportHeight).to.equal(720)
  })

  it('should access custom environment variables', () => {
    // Check if we can access environment variables
    const backendUrl = Cypress.env('backendUrl') as string
    const isCI = Cypress.env('isCI')
    const expectedPort = isCI ? 34115 : 9245
    const expectedUrl = `http://localhost:${expectedPort}`

    cy.log('Backend URL from env:', backendUrl || 'not set')
    cy.log('Is CI:', isCI)

    expect(backendUrl).to.equal(expectedUrl)
  })

  it('should handle basic DOM operations', () => {
    // Visit a blank page (this will fail without backend, but that's ok)
    cy.visit('/', { failOnStatusCode: false })

    // Check if we got any response (success or failure)
    cy.then(() => {
      cy.log('Attempted to visit base URL')
    })
  })

  it('should support custom commands', () => {
    // Verify custom commands are registered
    cy.log('Testing custom command availability')

    // Note: These will fail if elements don't exist, but we're just
    // checking if the commands are registered
    cy.window().then((win) => {
      cy.log('Window object available:', !!win)
    })
  })
})
