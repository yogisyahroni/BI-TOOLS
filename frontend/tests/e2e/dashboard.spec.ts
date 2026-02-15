import { test, expect } from '@playwright/test';

test.describe('Dashboard Management', () => {
    test('should create a new dashboard successfully', async ({ page }) => {
        const timestamp = Date.now();
        const email = `dash_${timestamp}@example.com`;
        const username = `dash_${timestamp}`;
        const password = 'TestPassword123!';

        // --- SETUP: Register & Login ---
        await page.goto('/auth/register');
        await page.getByLabel('Email Address').fill(email);
        await page.getByLabel('Username').fill(username);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByLabel('Confirm Password').fill(password);
        await page.getByLabel('I agree to the').check();
        await page.getByTestId('register-submit-btn').click();
        await page.getByRole('button', { name: /Go to Sign In/i }).click();

        await page.getByLabel('Email Address').fill(email);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByTestId('signin-submit-btn').click();
        await expect(page).toHaveURL(/\/(dashboards)?$/, { timeout: 15000 });

        // --- TEST: Create Dashboard ---
        // Ensure we are on /dashboards (login might redirect there)
        await page.goto('/dashboards');

        // Open Dialog
        // The "Create New Dashboard" button might be visible on empty state or header
        // We look for the button with text
        await page.getByRole('button', { name: /Create New Dashboard/i }).first().click();

        // Fill Dialog
        await expect(page.getByRole('dialog')).toBeVisible();
        await page.getByLabel('Name').fill(`My Test Dashboard ${timestamp}`);
        await page.getByLabel('Description').fill('Created via E2E test');

        // Submit (Click "Create Manually")
        await page.getByRole('button', { name: 'Create Manually' }).click();

        // Verify it appears in the list or is selected
        // The code selects it active.
        // We verify the header title changes
        await expect(page.getByRole('heading', { level: 2 }).first()).toContainText(`My Test Dashboard ${timestamp}`);
    });
});
