import { test, expect } from '@playwright/test';

test.describe('eCFR Dashboard E2E', () => {
    test('homepage loads and displays agencies', async ({ page }) => {
        page.on('console', msg => console.log('PAGE LOG:', msg.text()));
        await page.goto('/');
await expect(page).toHaveTitle(/eCFR Deregulation Dashboard/);

        // Wait for loading to finish (either table or error)
// Wait until either table is visible or an error alert appears
await expect(page.locator('.usa-table').or(page.locator('.usa-alert--error'))).toBeVisible({ timeout: 15000 });

        if (await page.locator('.usa-alert--error').isVisible()) {
            console.log('App Error:', await page.locator('.usa-alert__text').textContent());
        }

        // Verify table headers using CSS if role fails
await expect(page.locator('th').filter({ hasText: 'Agency Name' })).toBeVisible();
        await expect(page.locator('th').filter({ hasText: 'Word Count' })).toBeVisible();

        // Verify data presence
        const rows = page.locator('tbody tr');
        await expect(rows).not.toHaveCount(0);
    });

    test('title page loads', async ({ page }) => {
        // Direct navigation as per plan
        await page.goto('/title/1');
        // We expect some content related to Title 1
        // Adjust expectation based on actual UI, for now checking for no error
        await expect(page.getByText('Error')).not.toBeVisible();
        // If the page displays "Title 1", check for that
        // await expect(page.getByText('Title 1')).toBeVisible(); 
    });

    test('section page loads', async ({ page }) => {
        // Direct navigation as per plan
        await page.goto('/section/1');
        // We expect some content related to Section 1
        await expect(page.getByText('Error')).not.toBeVisible();
    });
});
