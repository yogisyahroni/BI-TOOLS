import { test, expect } from '@playwright/test';

test.describe('E2E Verification: Security & Core Flows', () => {

    test('should complete full onboarding flow: Register -> Connect DB -> Create App', async ({ page }) => {
        // Enable console logging from browser
        page.on('console', msg => console.log(`BROWSER LOG: ${msg.text()}`));
        page.on('pageerror', err => console.log(`BROWSER ERROR: ${err.message}`));

        const timestamp = Date.now();
        const email = `verify${timestamp}@example.com`;
        const username = `verifyuser${timestamp}`;
        const password = 'TestPassword123!';
        const connName = `TestDB-${timestamp}`;
        const appName = `TestApp-${timestamp}`;

        // --- 1. REGISTRATION ---
        console.log(`Starting registration for ${email}`);
        await page.goto('/auth/register');
        await page.getByLabel('Full Name (Optional)').fill('Verification User');
        await page.getByLabel('Email Address').fill(email);
        await page.getByLabel('Username').fill(username);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByLabel('Confirm Password').fill(password);

        // Handle terms checkbox
        const termsCheckbox = page.getByLabel('I agree to the');
        if (await termsCheckbox.isVisible()) {
            await termsCheckbox.check();
        } else {
            // Fallback if label differs slightly
            await page.locator('input[type="checkbox"]').first().check();
        }

        await page.getByTestId('register-submit-btn').click();
        await expect(page.getByText('Registration Successful!')).toBeVisible({ timeout: 15000 });

        // Go to Sign In
        await page.getByRole('button', { name: /Go to Sign In/i }).click();

        // --- 2. LOGIN ---
        console.log('Logging in...');
        await expect(page).toHaveURL(/.*\/auth\/signin/);
        await page.getByLabel('Email Address').fill(email);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByTestId('signin-submit-btn').click();
        await expect(page.getByText('Sign In Successful!')).toBeVisible({ timeout: 15000 });
        // Wait for dashboard redirection
        await expect(page).toHaveURL(/\/dashboards|^\/$/, { timeout: 20000 });

        // --- 3. ADD CONNECTION ---
        console.log('Navigating to Connections...');
        await page.goto('/connections');
        // Wait for heading
        await expect(page.getByRole('heading', { name: 'Database Connections' })).toBeVisible();

        // Open Dialog
        await page.getByRole('button', { name: 'Add Connection' }).click();
        await expect(page.getByRole('dialog')).toBeVisible();

        // Fill Form
        await page.getByLabel('Display Name').fill(connName);

        // Handle Select for Type (Postgres is default, but let's be explicit if possible, or just default)
        // Select trigger usually has a value or placeholder
        // Check current value
        // await page.getByRole('combobox', { name: 'Type' }).click(); 
        // await page.getByRole('option', { name: 'PostgreSQL' }).click();

        await page.getByLabel('Database Name').fill('postgres'); // Default db
        await page.getByLabel('Host').fill('localhost');
        await page.getByLabel('Port').fill('5432');
        await page.getByLabel('Username').fill('postgres'); // Match backend expectation
        await page.getByLabel('Password').fill('postgres');

        // Test SSL Checkbox
        // Note: Backend default SSL is require if checked. Localhost usually disable.
        // Let's leave it unchecked for localhost test, but assert it exists
        const sslCheckbox = page.getByLabel('Use SSL Connection');
        await expect(sslCheckbox).toBeVisible();
        // await sslCheckbox.check(); // Don't check for localhost

        // Save
        await page.getByRole('button', { name: 'Save Connection' }).click();

        // Verify Toast & List
        await expect(page.getByText('Connection created successfully')).toBeVisible();
        await expect(page.getByRole('dialog')).toBeHidden();
        // Check list
        await expect(page.getByText(connName)).toBeVisible();

        // --- 4. CREATE APP ---
        console.log('Navigating to Apps...');
        await page.goto('/apps');
        await expect(page.getByRole('heading', { name: 'Data Apps' })).toBeVisible();

        // Open Dialog
        await page.getByRole('button', { name: 'Create App' }).click();

        // Fill Form
        await page.getByLabel('App Name').fill(appName);
        // Slug should auto-fill
        const slugInput = page.getByLabel('URL Slug');
        await expect(slugInput).not.toBeEmpty();

        // Create
        await page.getByRole('button', { name: 'Create App', exact: true }).click();

        // Verify Toast & Redirect
        await expect(page.getByText('App created successfully')).toBeVisible();
        await expect(page).toHaveURL(/\/apps\/builder\/.*/, { timeout: 15000 });

        console.log('Verification Complete!');
    });
});
