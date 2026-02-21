
const puppeteer = require('puppeteer');

async function run() {
    const browser = await puppeteer.launch({
        headless: true,
        args: ['--no-sandbox', '--disable-setuid-sandbox'],
        defaultViewport: { width: 1920, height: 1080 }
    });
    const page = await browser.newPage();

    try {
        console.log('Navigating to login...');
        await page.goto('http://localhost:3000/auth/signin', { waitUntil: 'networkidle0' });

        console.log('Logging in...');
        await page.type('input[type="email"]', 'yogisyahroni766.ysr@gmail.com');
        await page.type('input[type="password"]', 'Namakamu766!!');

        await Promise.all([
            page.waitForNavigation({ waitUntil: 'networkidle0' }),
            page.click('button[type="submit"]')
        ]);

        console.log('Logged in successfully. Taking dashboard screenshot...');
        await page.screenshot({ path: 'dashboard_polish.png', fullPage: true });

        console.log('Hovering over sidebar item...');
        // Selector might need adjustment based on rendered output, trying a generic approach
        try {
            await page.waitForSelector('aside nav a[href="/connections"]', { timeout: 5000 });
            await page.hover('aside nav a[href="/connections"]');
            await page.screenshot({ path: 'dashboard_polish_hover.png' });
            console.log('Hover screenshot taken.');
        } catch (e) {
            console.log('Could not hover over connections link:', e.message);
        }

    } catch (error) {
        console.error('Error during verification:', error);
    } finally {
        await browser.close();
    }
}

run();
