const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 }
  });
  const page = await context.newPage();
  const baseUrl = 'http://localhost:4200';
  const screenshotsDir = path.join(__dirname, '..', 'screenshots');

  if (!fs.existsSync(screenshotsDir)) {
    fs.mkdirSync(screenshotsDir);
  }

  const pages = [
    { name: 'dashboard', path: '/' },
    { name: 'search', path: '/search' },
    { name: 'discover', path: '/discover' },
    { name: 'watches', path: '/watches' },
    { name: 'blacklist', path: '/blacklist' },
    { name: 'settings', path: '/settings' },
    { name: 'trawl', path: '/trawl' },
  ];

  // Wait for server to be ready
  let retries = 30;
  while (retries > 0) {
    try {
      const response = await page.goto(`${baseUrl}/health`);
      if (response.status() === 200) {
        console.log('Server is ready');
        break;
      }
    } catch (e) {
      console.log('Waiting for server...');
      await new Promise(resolve => setTimeout(resolve, 1000));
      retries--;
    }
  }

  if (retries === 0) {
    console.error('Server timed out');
    process.exit(1);
  }

  for (const p of pages) {
    console.log(`Capturing ${p.name}...`);
    try {
      await page.goto(`${baseUrl}${p.path}`, { waitUntil: 'load' });
      await page.waitForTimeout(2000); // Give it some time to render
      await page.screenshot({ path: path.join(screenshotsDir, `${p.name}.png`), fullPage: false });
    } catch (error) {
      console.error(`Failed to capture ${p.name}:`, error);
    }
  }

  await browser.close();
})();
