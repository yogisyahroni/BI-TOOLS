
import puppeteer from 'puppeteer';
import fs from 'fs';
import path from 'path';

async function run() {
    const browser = await puppeteer.launch({
        headless: true,
        args: ['--no-sandbox', '--disable-setuid-sandbox'],
        defaultViewport: { width: 1920, height: 1080 }
    });
    const page = await browser.newPage();

    try {
        // Login
        await page.goto('http://localhost:3000/auth/signin', { waitUntil: 'networkidle0' });
        await page.type('input[type="email"]', 'yogisyahroni766.ysr@gmail.com');
        await page.type('input[type="password"]', 'Namakamu766!!');

        await Promise.all([
            page.waitForNavigation({ waitUntil: 'networkidle0' }),
            page.click('button[type="submit"]')
        ]);

        console.log('Logged in successfully.');

        // Take Dashboard Screenshot
        await page.screenshot({ path: 'dashboard_polish.png', fullPage: true });
        console.log('Dashboard screenshot taken.');

        // Hover over a sidebar item to check hover state
        await page.hover('aside nav a[href="/connections"]');
        await page.screenshot({ path: 'dashboard_polish_hover.png' });
        console.log('Hover screenshot taken.');

    } catch (error) {
        console.error('Error during verification:', error);
    } finally {
        await browser.close();
    }
}

run();
