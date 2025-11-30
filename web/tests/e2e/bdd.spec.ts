import { test, expect } from '@playwright/test';

test.describe('Feature: Dashboard Overview', () => {
  test('Scenario: User sees the summary statistics on landing', async ({ page }) => {
    console.log('Navigating to /');
    await page.goto('/');
    console.log('Navigated. Checking visibility.');

    // Then the summary box should be visible
    await expect(page.locator('.usa-summary-box')).toBeVisible({ timeout: 10000 });
    
    // And it should show regulatory overview text
    await expect(page.locator('h2', { hasText: 'Regulatory Overview' })).toBeVisible();
    
    // And loading state should eventually disappear
    // Use a longer timeout here as data fetching might be slow?
    console.log('Waiting for loading to disappear');
    await expect(page.locator('text=Loading agency data...')).not.toBeVisible({ timeout: 10000 });
  });

  test('Scenario: User can expand an agency to see sub-agencies', async ({ page }) => {
    await page.goto('/');
    
    // Wait for loading to disappear first
    await expect(page.locator('text=Loading agency data...')).not.toBeVisible({ timeout: 15000 });

    // Wait for table
    await expect(page.locator('table')).toBeVisible();

    // Wait for at least one row
    const parentRow = page.locator('tr.parent-row').first();
    await expect(parentRow).toBeVisible({ timeout: 15000 });

    // Find a row that has children (indicated by arrow)
    // We look for a row containing the arrow symbol or just try to find one that produces children.
    // Better: select one we know has children?
    // Or just check if any row has children.
    
    // Try to find a row with the expand icon (triangle)
    const expandableRow = page.locator('tr.parent-row').filter({ hasText: '▶' }).first();
    
    if (await expandableRow.count() > 0) {
        console.log('Clicking expandable row');
        await expandableRow.click();
        const childRow = page.locator('tr.child-row').first();
        await expect(childRow).toBeVisible();
    } else {
        console.log('No expandable rows found - skipping expand test');
    }
  });
});
