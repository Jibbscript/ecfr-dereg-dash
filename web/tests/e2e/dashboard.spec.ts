import { test, expect } from '@playwright/test';

test('homepage loads and displays agencies', async ({ page }) => {
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    await page.goto('/');
    await expect(page).toHaveTitle(/eCFR Dashboard/);

    // Check for dummy data
    await expect(page.getByText('Department of Agriculture')).toBeVisible();
    await expect(page.getByText('Department of Commerce')).toBeVisible();
});
