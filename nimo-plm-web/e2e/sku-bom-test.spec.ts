import { test, expect } from '@playwright/test';

test('SKU BOM part click freeze test', async ({ page }) => {
  // Navigate to project detail directly
  await page.goto('/projects');
  await page.waitForLoadState('networkidle');

  // Click first project detail button
  const detailBtn = page.getByRole('button', { name: '详情' }).first();
  if (!(await detailBtn.isVisible())) { test.skip(); return; }
  await detailBtn.click();
  await page.waitForLoadState('networkidle');

  // Click SKU tab
  await page.getByRole('tab', { name: /SKU配色/ }).click();
  await page.waitForTimeout(1000);

  // Click on the SKU card text "星空黑" to enter detail
  const skuCardText = page.getByText('星空黑');
  if (!(await skuCardText.isVisible().catch(() => false))) {
    console.log('No "星空黑" SKU found, looking for any SKU card...');
    const anyCard = page.locator('.ant-card-body').first();
    if (!(await anyCard.isVisible().catch(() => false))) {
      console.log('No SKU cards found, skipping test');
      test.skip();
      return;
    }
    await anyCard.click();
  } else {
    await skuCardText.click();
  }
  await page.waitForTimeout(2000);

  // Take screenshot of SKU detail
  await page.screenshot({ path: '/tmp/sku-detail.png', fullPage: true });

  // Check if BOM select tab is visible
  const bomTab = page.getByRole('tab', { name: /BOM零件勾选/ });
  const hasBomTab = await bomTab.isVisible().catch(() => false);
  console.log('BOM select tab visible:', hasBomTab);

  if (!hasBomTab) {
    console.log('BOM select tab not found');
    return;
  }

  // Make sure BOM select tab is active
  await bomTab.click();
  await page.waitForTimeout(1000);

  // Check BOM table
  const tableRows = page.locator('.ant-table-tbody tr.ant-table-row');
  const rowCount = await tableRows.count();
  console.log('BOM table row count:', rowCount);

  if (rowCount === 0) {
    console.log('No BOM items - need SBOM data to test');
    await page.screenshot({ path: '/tmp/sku-bom-empty.png', fullPage: true });
    return;
  }

  // Log first row content
  const firstRow = tableRows.first();
  const firstRowText = await firstRow.textContent();
  console.log('First row text:', firstRowText);

  // Click on the name cell of the first row (3rd column)
  console.log('About to click on BOM item name...');

  // Use a 5s timeout to detect freeze
  const startTime = Date.now();
  try {
    await firstRow.locator('td').nth(2).click({ timeout: 5000 });
    const elapsed = Date.now() - startTime;
    console.log('Name click completed in', elapsed, 'ms');
  } catch (e) {
    const elapsed = Date.now() - startTime;
    console.log('Name click FAILED after', elapsed, 'ms:', (e as Error).message);
  }

  // Check if page is still responsive
  const isResponsive = await page.evaluate(() => {
    return 'alive';
  }).catch(() => 'DEAD');
  console.log('Page responsive:', isResponsive);

  // Now try clicking the checkbox
  const checkbox = firstRow.locator('.ant-checkbox-input');
  if (await checkbox.isVisible()) {
    console.log('About to click checkbox...');
    const checkStart = Date.now();
    try {
      await checkbox.click({ timeout: 10000 });
      const elapsed = Date.now() - checkStart;
      console.log('Checkbox click completed in', elapsed, 'ms');
    } catch (e) {
      const elapsed = Date.now() - checkStart;
      console.log('Checkbox click FAILED/TIMED OUT after', elapsed, 'ms:', (e as Error).message);
    }

    // Wait and check responsiveness
    await page.waitForTimeout(3000);
    const isResponsive2 = await page.evaluate(() => 'alive').catch(() => 'DEAD');
    console.log('Page responsive after checkbox:', isResponsive2);
  }

  // Monitor console errors
  const errors: string[] = [];
  page.on('console', msg => {
    if (msg.type() === 'error') errors.push(msg.text());
  });

  await page.screenshot({ path: '/tmp/sku-bom-after-click.png', fullPage: true });
  console.log('Console errors:', errors.length > 0 ? errors.join('; ') : 'none');
});
