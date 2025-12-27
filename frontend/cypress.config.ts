import { defineConfig } from 'cypress'
import { devServer } from '@cypress/vite-dev-server'

export default defineConfig({
  e2e: {
    // Use Vite dev server port in development, production port otherwise
    baseUrl: process.env.CI === 'true' ? 'http://localhost:34115' : 'http://localhost:9245',
    specPattern: 'cypress/e2e/**/*.cy.{js,jsx,ts,tsx}',
    supportFile: 'cypress/support/e2e.ts',
    videosFolder: 'cypress/videos',
    screenshotsFolder: 'cypress/screenshots',
    video: false,
    screenshotOnRunFailure: true,
    viewportWidth: 1280,
    viewportHeight: 720,
    defaultCommandTimeout: 10000,
    requestTimeout: 10000,
    responseTimeout: 10000,
    setupNodeEvents(on, config) {
      // Use Vite dev server for component testing
      on('dev-server:start', (options) => {
        return devServer({
          ...options,
          viteConfig: {},
        })
      })

      // Environment-specific configuration
      const isCI = process.env.CI === 'true'
      const devPort = isCI ? 34115 : 9245

      return {
        ...config,
        env: {
          ...config.env,
          isCI,
          backendUrl: process.env.BACKEND_URL || `http://localhost:${devPort}`,
        },
        // Adjust timeouts for CI
        defaultCommandTimeout: isCI ? 15000 : 10000,
        requestTimeout: isCI ? 15000 : 10000,
        responseTimeout: isCI ? 15000 : 10000,
      }
    },
  },
  component: {
    devServer: {
      framework: 'vue',
      bundler: 'vite',
    },
    specPattern: 'src/**/*.cy.{js,jsx,ts,tsx}',
    supportFile: 'cypress/support/component.ts',
    setupNodeEvents(on, config) {
      on('dev-server:start', (options) => {
        return devServer({
          ...options,
          viteConfig: {},
        })
      })
    },
  },
})
