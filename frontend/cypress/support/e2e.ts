// ***********************************************************
// This example support/e2e.ts is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './commands';

// Alternatively you can use CommonJS syntax:
// require('./commands')

// Global setup before all tests
before(() => {
  cy.env(['isCI', 'backendUrl']).then(({ isCI, backendUrl }) => {
    // Log test environment info
    cy.log('Testing Environment:', isCI ? 'CI' : 'Local');
    cy.log('Backend URL:', backendUrl);

    // Check if backend is available
    cy.request({
      url: `${backendUrl}/api/feeds`,
      failOnStatusCode: false,
    }).then((response) => {
      if (response.status === 500 || response.status === 0) {
        cy.log('⚠️ Backend may not be running. Tests might fail.');
      } else {
        cy.log('✅ Backend is responding');
      }
    });
  });
});

// Cleanup after each test
afterEach(() => {
  // Clear localStorage between tests to prevent state leakage
  cy.clearCookies();
  cy.clearLocalStorage();
});

// Global cleanup after all tests
after(() => {
  cy.log('All tests completed');
});
