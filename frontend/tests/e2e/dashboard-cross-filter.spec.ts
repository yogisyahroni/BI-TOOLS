import { test, expect } from '@playwright/test';

test.describe('Dashboard Cross-Filtering', () => {
    test.beforeEach(async ({ page }) => {
        // Mock Login
        await page.goto('/auth/login');
        // Setup mock for dashboard access
        // We will mock the API responses so login might be bypassed if we mock /api/auth/session or similar
        // But the app might check cookie. 
        // Let's try to assume we can just go to the page if we mock the data requests.
        // But useDashboard checks authentication usually.
    });

    test('should apply filter when clicking on a chart bar', async ({ page }) => {
        // Mock Dashboard Response
        await page.route('**/api/dashboards/test-id', async route => {
            await route.fulfill({
                json: {
                    id: 'test-id',
                    title: 'Cross Filter Test',
                    layout: [],
                    cards: [
                        {
                            id: 'card-1',
                            title: 'Source Bar Chart',
                            type: 'visualization',
                            visualizationConfig: { type: 'bar', xAxis: 'category', yAxis: ['value'] },
                            query: { sql: 'SELECT * FROM source' },
                            position: { x: 0, y: 0, w: 6, h: 4 }
                        },
                        {
                            id: 'card-2',
                            title: 'Target Line Chart',
                            type: 'visualization',
                            visualizationConfig: { type: 'line', xAxis: 'category', yAxis: ['value'] },
                            query: { sql: 'SELECT * FROM target' },
                            position: { x: 6, y: 0, w: 6, h: 4 }
                        }
                    ]
                }
            });
        });

        // Mock Query Execution
        await page.route('**/api/queries/execute', async route => {
            const request = route.request();
            const postData = request.postDataJSON();

            // Check if cross-filters are applied in the request body
            // This is how we verify the "Backend" part of the integration
            if (postData.crossFilters && postData.crossFilters.length > 0) {
                // Return filtered data
                await route.fulfill({
                    json: {
                        data: [{ category: 'A', value: 10 }] // Filtered
                    }
                });
                return;
            }

            // Return full data
            await route.fulfill({
                json: {
                    data: [
                        { category: 'A', value: 10 },
                        { category: 'B', value: 20 },
                        { category: 'C', value: 30 }
                    ]
                }
            });
        });

        // Navigate to dashboard
        // We need to bypass login or reuse auth state. 
        // For now, assuming the test runner has auth state or we mock the auth check.
        // If getting 401, this test will fail.
        // But structurally this is the test we want.

        // Attempt to login first (copy-paste basic login from dashboard.spec.ts)
        const timestamp = Date.now();
        const email = `test_${timestamp}@example.com`;
        const password = 'Password123!';

        // Mock Auth API to return success? 
        // Or just use the UI if the backend is running. 
        // ERROR: The backend might NOT be running in this environment during `npx playwright test`.
        // The previous command `npm run build` does not start the server.
        // `npx playwright test` usually expects the app to be running at localhost:3000 (webServer config in playwright.config.ts).

    });
});
