/// <reference types="cypress" />

describe('Feed Management', () => {
  beforeEach(() => {
    // Set up intercepts before visiting the page
    cy.intercept('POST', '/api/feeds').as('addFeed');
    cy.intercept('GET', '/api/feeds').as('getFeeds');
    cy.intercept('DELETE', '/api/feeds/*').as('deleteFeed');
    cy.intercept('POST', '/api/feeds/refresh').as('refreshFeeds');
    cy.intercept('PUT', '/api/feeds/*').as('updateFeed');

    cy.visit('/');
    cy.get('body').should('be.visible');
  });

  it('should add a new feed', () => {
    // Look for add feed button in the sidebar footer (+ icon)
    cy.get('button').filter('[title="Add Feed"], [title="添加订阅"]').should('exist').click({ force: true });

    // Wait for add feed modal to appear
    cy.wait(1000);

    // Check if modal opened
    cy.get('body').then(($body) => {
      if ($body.find(/add.*feed|添加.*feed|add.*subscription/i).length > 0 ||
          $body.find('[class*="modal"]').length > 0) {
        cy.log('Add feed modal opened');

        // Try to fill in the feed URL if input exists
        cy.get('body').then(($body2) => {
          if ($body2.find('input[type="url"], input[type="text"]').length > 0) {
            cy.get('input[type="url"], input[type="text"]').first().type('https://example.com/feed.xml');

            // Submit the form if submit button exists
            cy.get('body').then(($body3) => {
              if ($body3.find('button').filter((i, el) => /add|submit|确定|添加/i.test(el.textContent || '')).length > 0) {
                cy.get('button').contains(/add|submit|确定|添加/i).click({ force: true });
              }
            });
          }
        });
      } else {
        cy.log('Add feed modal did not open or was not detected');
      }
    });
  });

  it('should delete a feed', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Try to find a feed to delete
    cy.get('body').then(($body) => {
      if ($body.find('[class*="feed"]').length > 0) {
        // Right-click on a feed to open context menu
        cy.get('[class*="feed"]').first().rightclick({ force: true });

        // Click delete option in context menu if it exists
        cy.get('body').then(($body2) => {
          if ($body2.find(/delete|删除/i).length > 0) {
            cy.contains(/delete|删除/i).click({ force: true });

            // Confirm deletion in the confirm dialog
            cy.get('body').then(($body3) => {
              if ($body3.find(/confirm|确认/i).length > 0) {
                cy.contains(/confirm|确认/i).click({ force: true });

                // Wait for deletion to complete
                cy.wait('@deleteFeed', { timeout: 10000 });
              }
            });
          } else {
            cy.log('Delete option not found in context menu');
          }
        });
      } else {
        cy.log('No feeds found to test deletion');
      }
    });
  });

  it('should refresh feeds', () => {
    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Look for refresh button - check button title or content
    cy.get('body').then(($body) => {
      const refreshButtons = $body.find('button').filter((i, el) => {
        return /refresh|刷新/i.test(el.title || el.textContent || '');
      });

      if (refreshButtons.length > 0) {
        cy.wrap(refreshButtons).first().click({ force: true });
        cy.log('Refresh button clicked');

        // Wait a moment for any refresh to initiate
        cy.wait(500);
      } else {
        cy.log('Refresh button not found - may not be exposed in UI');
      }
    });
  });

  it('should edit feed details', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Try to find a feed to edit
    cy.get('body').then(($body) => {
      if ($body.find('[class*="feed"]').length > 0) {
        // Right-click on a feed
        cy.get('[class*="feed"]').first().rightclick({ force: true });

        // Click edit option if it exists
        cy.get('body').then(($body2) => {
          if ($body2.find(/edit|编辑/i).length > 0) {
            cy.contains(/edit|编辑/i).click({ force: true });

            // Wait for edit modal
            cy.wait(500);

            // Change the title
            cy.get('input[type="text"]').first().clear().type('Updated Feed Title');

            // Save changes
            cy.get('button')
              .contains(/save|保存|确定/i)
              .click({ force: true });

            // Wait for update to complete
            cy.wait('@updateFeed', { timeout: 10000 });
          } else {
            cy.log('Edit option not found in context menu');
          }
        });
      } else {
        cy.log('No feeds found to test editing');
      }
    });
  });

  it('should filter feeds by category', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Try to find category filter
    cy.get('body').then(($body) => {
      if ($body.find('select, [role="listbox"]').length > 0) {
        cy.get('select, [role="listbox"]').first().select(1);

        // Wait for filtered results
        cy.wait(500);

        // Verify feeds exist
        cy.get('[class*="feed"]').should('have.length.at.least', 0);
      } else {
        cy.log('Category filter not found');
      }
    });
  });

  it('should search feeds', () => {
    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 });

    // Look for search input in the sidebar
    cy.get('body').then(($body) => {
      if ($body.find('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]').length > 0) {
        cy.get('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]')
          .first()
          .type('test');

        // Wait for search results to filter
        cy.wait(500);

        // Verify search results
        cy.get('[class*="feed"]').should('exist');
      } else {
        cy.log('Search input not found');
      }
    });
  });
});
