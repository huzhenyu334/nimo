import { test, expect } from '@playwright/test';

test.use({ storageState: './e2e/.auth/storage-state.json' });

test.describe('Inventory Management', () => {
  test('inventory page loads', async ({ page }) => {
    await page.goto('/srm/inventory');
    await page.waitForTimeout(1500);
    const heading = page.locator('text=库存管理');
    await expect(heading.first()).toBeVisible();
  });

  test('inventory page has search input', async ({ page }) => {
    await page.goto('/srm/inventory');
    await page.waitForTimeout(1500);
    const searchInput = page.locator('input[placeholder*="搜索"]');
    await expect(searchInput.first()).toBeVisible();
  });

  test('inventory page has stock in button', async ({ page }) => {
    await page.goto('/srm/inventory');
    await page.waitForTimeout(1500);
    const stockInBtn = page.locator('button').filter({ hasText: '手动入库' });
    await expect(stockInBtn).toBeVisible();
  });

  test('inventory table renders', async ({ page }) => {
    await page.goto('/srm/inventory');
    await page.waitForTimeout(1500);
    const table = page.locator('.ant-table');
    await expect(table).toBeVisible();
  });

  test('stock in modal opens', async ({ page }) => {
    await page.goto('/srm/inventory');
    await page.waitForTimeout(1500);
    const stockInBtn = page.locator('button').filter({ hasText: '手动入库' });
    await stockInBtn.click();
    await page.waitForTimeout(500);
    const modal = page.locator('.ant-modal');
    await expect(modal).toBeVisible();
    const materialLabel = modal.locator('text=物料编码');
    await expect(materialLabel.first()).toBeVisible();
  });

  test('low stock filter exists as select option', async ({ page }) => {
    await page.goto('/srm/inventory');
    await page.waitForTimeout(1500);
    // The low stock filter is a Select with "全部" and "低库存" options
    const select = page.locator('.ant-select').first();
    await select.click();
    await page.waitForTimeout(300);
    const lowStockOption = page.locator('.ant-select-item-option').filter({ hasText: '低库存' });
    await expect(lowStockOption).toBeVisible();
  });
});

test.describe('Inventory API', () => {
  test('list inventory API returns valid response', async ({ page }) => {
    await page.goto('/srm/inventory');
    await page.waitForTimeout(1000);
    const result = await page.evaluate(async () => {
      const token = localStorage.getItem('access_token');
      const resp = await fetch('/api/v1/srm/inventory?page=1&page_size=5', {
        headers: { Authorization: `Bearer ${token}` },
      });
      return { status: resp.status, body: await resp.json() };
    });
    expect(result.status).toBe(200);
    expect(result.body).toHaveProperty('data');
  });
});
