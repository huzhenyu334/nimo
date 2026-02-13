import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const SCREENSHOTS_DIR = path.join(process.cwd(), 'screenshots');

test.beforeAll(() => {
  fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });
});

test.describe('SRM Kanban UI', () => {
  test('kanban page loads with authenticated user', async ({ page }) => {
    await page.goto('/srm/kanban');
    // Should NOT redirect to login — storageState has JWT
    await page.waitForURL('**/srm/kanban', { timeout: 10000 });
    // The page should show the project selector
    await expect(page.locator('.ant-select').first()).toBeVisible({ timeout: 10000 });
  });

  test('kanban board renders 8 columns after selecting project', async ({ page }) => {
    await page.goto('/srm/kanban');
    await page.waitForURL('**/srm/kanban', { timeout: 10000 });

    // Select the first available project in the dropdown
    const projectSelect = page.locator('.ant-select').first();
    await projectSelect.click();
    // Wait for dropdown options to appear and pick the first one
    const firstOption = page.locator('.ant-select-item-option').first();
    await firstOption.waitFor({ timeout: 10000 });
    await firstOption.click();

    // Wait for kanban columns to render (each column has a Badge with count)
    const columnHeaders = page.locator('div').filter({ has: page.locator('.ant-badge') }).filter({ hasText: /寻源中|报价中|待下单|已下单|已发货|已收货|检验中|已通过/ });
    await expect(columnHeaders.first()).toBeVisible({ timeout: 15000 });

    // Verify all 8 column labels exist
    const expectedLabels = ['寻源中', '报价中', '待下单', '已下单', '已发货', '已收货', '检验中', '已通过'];
    for (const label of expectedLabels) {
      await expect(page.getByText(label, { exact: true }).first()).toBeVisible();
    }

    // Screenshot: kanban with data
    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'kanban-board.png'), fullPage: false });
  });

  test('kanban cards have category Tags', async ({ page }) => {
    await page.goto('/srm/kanban');
    await page.waitForURL('**/srm/kanban', { timeout: 10000 });

    // Select project
    const projectSelect = page.locator('.ant-select').first();
    await projectSelect.click();
    await page.locator('.ant-select-item-option').first().click();

    // Wait for cards to render — cards contain .ant-tag elements for categories
    const tags = page.locator('.ant-tag');
    await tags.first().waitFor({ timeout: 15000 });
    const tagCount = await tags.count();
    expect(tagCount).toBeGreaterThan(0);
  });

  test('page height fits viewport (no body-level scrollbar)', async ({ page }) => {
    await page.goto('/srm/kanban');
    await page.waitForURL('**/srm/kanban', { timeout: 10000 });

    // Select project
    const projectSelect = page.locator('.ant-select').first();
    await projectSelect.click();
    await page.locator('.ant-select-item-option').first().click();

    // Wait for columns to render
    await page.getByText('寻源中', { exact: true }).first().waitFor({ timeout: 15000 });

    // Check that the document body doesn't overflow vertically
    const bodyOverflows = await page.evaluate(() => {
      return document.documentElement.scrollHeight > window.innerHeight + 10;
    });
    expect(bodyOverflows).toBe(false);
  });

  test('category multi-select filter exists', async ({ page }) => {
    await page.goto('/srm/kanban');
    await page.waitForURL('**/srm/kanban', { timeout: 10000 });

    // The category filter is a multi-select with default values pre-selected (全部需求, 电子EBOM, etc.)
    const categoryFilter = page.locator('.ant-select-multiple');
    await expect(categoryFilter).toBeVisible({ timeout: 10000 });
    // Verify it has selected tags (default filters)
    const selectedTags = categoryFilter.locator('.ant-select-selection-item');
    await expect(selectedTags.first()).toBeVisible();

    // Screenshot: initial state
    await page.screenshot({ path: path.join(SCREENSHOTS_DIR, 'kanban-initial.png'), fullPage: false });
  });
});
