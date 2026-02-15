import { test, expect } from '@playwright/test';

test.describe('Settings Page', () => {
    test('should allow saving AI configuration', async ({ page }) => {
        const timestamp = Date.now();
        const email = `settings_${timestamp}@example.com`;
        const username = `settings_${timestamp}`;
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

        // --- TEST: Settings ---
        await page.goto('/settings');

        // Verify Tabs
        await expect(page.getByRole('tab', { name: 'Database' })).toBeVisible();
        await expect(page.getByRole('tab', { name: 'AI Providers' })).toBeVisible();

        // Switch to AI Providers
        await page.getByRole('tab', { name: 'AI Providers' }).click();

        // Fill Form
        await page.getByLabel('API Key').fill('sk-test-key-12345');
        await page.getByLabel('Model').fill('gpt-4o-test');

        // Save
        await page.getByRole('button', { name: 'Save Configuration' }).click();

        // Verify Success Toast
        await expect(page.locator('li[data-sonner-toast]')).toContainText('AI Configuration saved');
    });
});
