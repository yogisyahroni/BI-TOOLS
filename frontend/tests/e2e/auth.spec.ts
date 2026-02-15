import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {

    test('should register a new user and login successfully', async ({ page }) => {
        const timestamp = Date.now();
        const email = `testuser${timestamp}@example.com`;
        const username = `user${timestamp}`;
        const password = 'TestPassword123!';

        // --- REGISTRATION ---
        await page.goto('/auth/register');

        // Check elements
        await expect(page.getByLabel('Email Address')).toBeVisible();

        // Fill form
        await page.getByLabel('Full Name (Optional)').fill('Test User');
        await page.getByLabel('Email Address').fill(email);
        await page.getByLabel('Username').fill(username);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByLabel('Confirm Password').fill(password);

        // Agree terms (sometimes label click works better for checkbox)
        await page.getByLabel('I agree to the').check();

        // Submit
        await page.getByTestId('register-submit-btn').click();

        // Expect success (wait for success message)
        await expect(page.getByText('Registration Successful!')).toBeVisible({ timeout: 10000 });

        // Click Go to Sign In
        await page.getByRole('button', { name: /Go to Sign In/i }).click();

        // --- LOGIN ---
        await expect(page).toHaveURL(/.*\/auth\/signin/);

        await page.getByLabel('Email Address').fill(email);
        await page.getByLabel('Password', { exact: true }).fill(password);

        // Submit
        await page.getByTestId('signin-submit-btn').click();

        // Expect redirect to dashboard or success toast
        // The Signin page redirects to callbackUrl (default /dashboards)
        // It shows "Sign In Successful!" then redirects.
        await expect(page.getByText('Sign In Successful!')).toBeVisible({ timeout: 10000 });

        // Ideally verify we land on dashboard, but might need to wait for redirect
        // check URL eventually changes
        await expect(page).toHaveURL(/\/dashboards|^\/$/, { timeout: 15000 });
    });

    test('should show error for invalid credentials', async ({ page }) => {
        await page.goto('/auth/signin');

        await page.getByLabel('Email Address').fill('wrong@example.com');
        await page.getByLabel('Password', { exact: true }).fill('wrongpass');
        await page.getByTestId('signin-submit-btn').click();

        // Expect toast error (Sonner toast might be hard to catch by text, but let's try)
        // Or check if we represent error on UI
        // signIn component says: toast.error('Invalid email...')
        // We can check for a toast or the text anywhere
        await expect(page.getByText(/Invalid email or password/i)).toBeVisible({ timeout: 10000 });
    });
});
