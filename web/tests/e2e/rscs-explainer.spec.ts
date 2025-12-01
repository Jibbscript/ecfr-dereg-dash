import { test, expect } from '@playwright/test'

// These tests verify the shared RSCS Explainer modal works from multiple triggers

test.describe('RSCS Explainer interactions', () => {
  test('opens from table header info button and closes with ESC', async ({ page }) => {
    await page.goto('/')

    // Wait for table
    await expect(page.locator('.usa-table').or(page.locator('.usa-alert--error'))).toBeVisible({ timeout: 15000 })

    if (await page.locator('.usa-table').isVisible()) {
      const info = page.locator('th', { hasText: 'RSCS per 1K' }).locator('button.info-icon')
      await expect(info).toBeVisible()
      await info.click()

      // Dialog heading should be visible
      await expect(page.getByRole('heading', { name: /Understanding the RSCS Metric/i })).toBeVisible()

      // ESC closes
      await page.keyboard.press('Escape')
      await expect(page.getByRole('heading', { name: /Understanding the RSCS Metric/i })).not.toBeVisible()
    }
  })

  test('opens from header nav "About RSCS"', async ({ page }) => {
    await page.goto('/')
    const nav = page.locator('#about-rscs-trigger')
    await expect(nav).toBeVisible()
    await nav.click()
    await expect(page.getByRole('heading', { name: /Understanding the RSCS Metric/i })).toBeVisible()
  })

  test('opens from MetricCard info icon', async ({ page }) => {
    await page.goto('/')
    // Wait for metric card to render
    await expect(page.getByText('Avg. RSCS Score')).toBeVisible()
    // Click the first information icon near the label text
    const card = page.locator('.metric-card .usa-card__heading', { hasText: 'Avg. RSCS Score' })
    await card.locator('button').click()
    await expect(page.getByRole('heading', { name: /Understanding the RSCS Metric/i })).toBeVisible()
  })
})
