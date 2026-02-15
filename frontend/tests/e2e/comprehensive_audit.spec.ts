
import { test, expect } from '@playwright/test';

test.describe('Comprehensive Browser Console Audit', () => {
    let consoleErrors: string[] = [];

    // Global listener for console errors
    test.beforeEach(async ({ page }) => {
        consoleErrors = [];
        page.on('console', msg => {
            if (msg.type() === 'error') {
                const text = msg.text();
                // Check if the error is actually critical. 
                // Some 404s or network errors might happen if external resources are blocked in test env.
                consoleErrors.push(`[${page.url()}] ${text}`);
                console.log(`BROWSER ERROR DETECTED: ${text}`);
            }
        });
        page.on('pageerror', err => {
            consoleErrors.push(`[${page.url()}] ${err.message}`);
            console.log(`PAGE ERROR DETECTED: ${err.message}`);
        });
    });

    test('should navigate through core pages without console errors', async ({ page }) => {
        const timestamp = Date.now();
        const verifyEmail = `verify_audit_${timestamp}@example.com`;
        const username = `audit_user_${timestamp}`;
        const password = 'TestPassword123!';

        console.log(`Registering user: ${verifyEmail}`);
        await page.goto('/auth/register');

        // Use getByLabel to match verification.spec.ts reliability
        await page.getByLabel('Full Name (Optional)').fill('Audit User');
        await page.getByLabel('Email Address').fill(verifyEmail);
        await page.getByLabel('Username').fill(username);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByLabel('Confirm Password').fill(password);

        // Handle terms checkbox
        const termsCheckbox = page.getByLabel('I agree to the');
        if (await termsCheckbox.isVisible()) {
            await termsCheckbox.check();
        } else {
            await page.locator('input[type="checkbox"]').first().check();
        }

        await page.getByTestId('register-submit-btn').click();

        // Expect successful registration redirect or toast
        await expect(page.getByText('Registration Successful!')).toBeVisible({ timeout: 15000 });
        await page.getByRole('button', { name: /Go to Sign In/i }).click();

        // Login
        console.log('Logging in...');
        await page.getByLabel('Email Address').fill(verifyEmail);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByTestId('signin-submit-btn').click();

        // Wait for dashboard redirection
        await expect(page).toHaveURL(/\/dashboards|^\/$/, { timeout: 30000 });
        console.log('Login successful. Starting route audit...');

        // 2. Define Routes to Audit
        const routes = [
            '/dashboards',
            '/connections',
            '/apps',
            '/analytics',
            '/settings',
            '/settings/profile',
        ];

        // 3. Visit each route
        for (const route of routes) {
            console.log(`Navigating to ${route}...`);
            await page.goto(route);
            await page.waitForLoadState('networkidle'); // Wait for network to settle

            // Allow some time for delayed errors
            await page.waitForTimeout(1500);
        }

        // 4. Assert No Errors
        if (consoleErrors.length > 0) {
            console.error('Audit Failed with Console Errors:', consoleErrors);
            // Fail the test if any errors were found
            // expect(consoleErrors).toHaveLength(0); 
            // NOTE: We log them but don't strictly fail the test yet to allow collecting all data.
            // The user wants "tidak boleh ada error", so we should ideally fail.
            // But let's fail at the end.
        } else {
            console.log('Audit Passed: No console errors detected.');
        }

        expect(consoleErrors).toHaveLength(0);
    });
});
