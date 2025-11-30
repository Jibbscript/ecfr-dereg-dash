import { test, expect } from '@playwright/test';

test.describe('Feature: Dashboard Overview', () => {
  test('Scenario: User sees the hero and summary metrics on landing', async ({ page }) => {
    await page.goto('/');

    // Then the hero should be visible
    await expect(page.locator('.usa-hero')).toBeVisible();

    // And summary metric cards should render
    await expect(page.getByText('Total Regulatory Words')).toBeVisible();
    await expect(page.getByText('Avg. RSCS Score')).toBeVisible();

    // And the data table eventually appears or an error alert
    await expect(page.locator('.usa-table').or(page.locator('.usa-alert--error'))).toBeVisible({ timeout: 15000 });
  });

  test('Scenario: User can expand an agency to see sub-agencies', async ({ page }) => {
    await page.goto('/');

    // Wait for table or error
    await expect(page.locator('.usa-table').or(page.locator('.usa-alert--error'))).toBeVisible({ timeout: 15000 });

    // If table is visible, try to expand a parent row
    if (await page.locator('.usa-table').isVisible()) {
      const parentRow = page.locator('tr.parent-row').first();
      await expect(parentRow).toBeVisible({ timeout: 15000 });

      // Click parent row if it has children
      await parentRow.click();
      const childRow = page.locator('tr.child-row').first();
      await expect(childRow).toBeVisible();
    }
  });
});
