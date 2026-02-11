import { test, expect } from '@playwright/test';

test.describe('Project Management', () => {
  test('projects API returns 401 without auth', async ({ request }) => {
    const response = await request.get('/api/v1/projects');
    expect(response.status()).toBe(401);
  });

  test('projects page redirects to login without auth', async ({ page }) => {
    // Clear tokens
    await page.goto('/login');
    await page.evaluate(() => {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
    });

    await page.goto('/projects');
    await page.waitForURL(/\/login/, { timeout: 5000 });
    await expect(page).toHaveURL(/\/login/);
  });

  // Authenticated tests - require valid session
  test.describe('with authentication', () => {
    test.skip(() => true, 'Requires valid Feishu SSO session - run manually with storageState');

    test('project list page loads', async ({ page }) => {
      await page.goto('/projects');
      await page.waitForLoadState('networkidle');

      // Should show project list table
      await expect(page.locator('text=项目')).toBeVisible();
    });

    test('create project flow', async ({ page }) => {
      await page.goto('/projects');
      await page.waitForLoadState('networkidle');

      // Click create button
      await page.click('button:has-text("新建项目")');

      // Fill project form
      await page.fill('input[placeholder*="项目名称"]', 'E2E测试项目');
      await page.fill('textarea[placeholder*="描述"]', 'E2E自动化测试创建的项目');

      // Submit
      await page.click('button:has-text("确定")');

      // Verify project created
      await expect(page.locator('text=E2E测试项目')).toBeVisible({ timeout: 10000 });
    });

    test('view project detail', async ({ page }) => {
      await page.goto('/projects');
      await page.waitForLoadState('networkidle');

      // Click on project to view detail
      await page.click('text=E2E测试项目');

      // Should navigate to detail page
      await expect(page).toHaveURL(/\/projects\/.+/);
      await expect(page.locator('text=E2E测试项目')).toBeVisible();
    });

    test('delete project', async ({ page }) => {
      await page.goto('/projects');
      await page.waitForLoadState('networkidle');

      // Find and delete test project
      const row = page.locator('tr', { hasText: 'E2E测试项目' });
      await row.locator('button:has-text("删除")').click();

      // Confirm
      await page.click('button:has-text("确定")');

      // Verify removed
      await expect(page.locator('text=E2E测试项目')).not.toBeVisible({ timeout: 5000 });
    });
  });
});
