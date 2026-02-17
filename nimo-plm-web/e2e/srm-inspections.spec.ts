import { test, expect } from '@playwright/test';

test.use({ storageState: './e2e/.auth/storage-state.json' });

test.describe('Inspections Enhanced', () => {
  test('inspections page loads', async ({ page }) => {
    await page.goto('/srm/inspections');
    await page.waitForTimeout(1500);
    const heading = page.locator('text=来料检验');
    await expect(heading.first()).toBeVisible();
  });

  test('inspections page has refresh button', async ({ page }) => {
    await page.goto('/srm/inspections');
    await page.waitForTimeout(1500);
    const refreshBtn = page.locator('button').filter({ hasText: '刷新' });
    await expect(refreshBtn).toBeVisible();
  });

  test('inspections table renders', async ({ page }) => {
    await page.goto('/srm/inspections');
    await page.waitForTimeout(1500);
    const table = page.locator('.ant-table');
    await expect(table).toBeVisible();
  });

  test('inspections page has status filter', async ({ page }) => {
    await page.goto('/srm/inspections');
    await page.waitForTimeout(1000);
    const statusSelect = page.locator('.ant-select');
    await expect(statusSelect.first()).toBeVisible();
  });

  test('inspections page has result filter', async ({ page }) => {
    await page.goto('/srm/inspections');
    await page.waitForTimeout(1000);
    // Second select is result filter
    const selects = page.locator('.ant-select');
    await expect(selects.nth(1)).toBeVisible();
  });
});

test.describe('Inspections API', () => {
  test('list inspections API returns valid response', async ({ page }) => {
    await page.goto('/srm/inspections');
    await page.waitForTimeout(1000);
    const result = await page.evaluate(async () => {
      const token = localStorage.getItem('access_token');
      const resp = await fetch('/api/v1/srm/inspections?page=1&page_size=5', {
        headers: { Authorization: `Bearer ${token}` },
      });
      return { status: resp.status, body: await resp.json() };
    });
    expect(result.status).toBe(200);
    expect(result.body).toHaveProperty('data');
  });

  test('create-from-po API endpoint exists', async ({ page }) => {
    await page.goto('/srm/inspections');
    await page.waitForTimeout(1000);
    const result = await page.evaluate(async () => {
      const token = localStorage.getItem('access_token');
      const resp = await fetch('/api/v1/srm/inspections/from-po', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ po_id: 'nonexistent' }),
      });
      return { status: resp.status };
    });
    // Should get 400 or 500, not 404 (route exists)
    expect([400, 500]).toContain(result.status);
  });
});
