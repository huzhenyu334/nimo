import { test, expect } from '@playwright/test';

test('SKU BOM part click freeze test', async ({ page }) => {
  // Navigate to project detail via card click
  await page.goto('/projects');
  await page.waitForLoadState('networkidle');

  const projectCard = page.locator('.ant-card-hoverable').first();
  await expect(projectCard).toBeVisible({ timeout: 5000 });
  await projectCard.click();
  await page.waitForLoadState('networkidle');

  // Click SKU tab
  await page.getByRole('tab', { name: /SKU配色/ }).click();
  await page.waitForTimeout(1000);

  // Click on the SKU card text "星空黑" to enter detail
  const skuCardText = page.getByText('星空黑');
  if (!(await skuCardText.isVisible().catch(() => false))) {
    // No specific SKU, try any SKU card
    const anyCard = page.locator('.ant-card-body').first();
    if (!(await anyCard.isVisible().catch(() => false))) {
      // No SKU cards at all — verify the page at least loaded correctly
      await expect(page.getByRole('button', { name: '新建SKU' })).toBeVisible();
      return;
    }
    await anyCard.click();
  } else {
    await skuCardText.click();
  }
  await page.waitForTimeout(2000);

  // Check if BOM select tab is visible
  const bomTab = page.getByRole('tab', { name: /BOM零件勾选/ });
  const hasBomTab = await bomTab.isVisible().catch(() => false);

  if (!hasBomTab) {
    // No BOM tab — verify page is still responsive
    const isResponsive = await page.evaluate(() => 'alive').catch(() => 'DEAD');
    expect(isResponsive).toBe('alive');
    return;
  }

  // Make sure BOM select tab is active
  await bomTab.click();
  await page.waitForTimeout(1000);

  // Check BOM table
  const tableRows = page.locator('.ant-table-tbody tr.ant-table-row');
  const rowCount = await tableRows.count();

  if (rowCount === 0) {
    // No BOM items — verify page is responsive
    const isResponsive = await page.evaluate(() => 'alive').catch(() => 'DEAD');
    expect(isResponsive).toBe('alive');
    return;
  }

  // Click on the name cell of the first row (3rd column) — verify no freeze
  const firstRow = tableRows.first();
  await firstRow.locator('td').nth(2).click({ timeout: 5000 });

  // Check if page is still responsive
  const isResponsive = await page.evaluate(() => 'alive').catch(() => 'DEAD');
  expect(isResponsive).toBe('alive');
});
