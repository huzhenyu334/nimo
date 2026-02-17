import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Helper: get auth token from storage state
function getAuthToken(): string {
  const statePath = path.join(__dirname, '.auth', 'storage-state.json');
  const state = JSON.parse(fs.readFileSync(statePath, 'utf-8'));
  for (const origin of state.origins || []) {
    for (const item of origin.localStorage || []) {
      if (item.name === 'access_token') return item.value;
    }
  }
  return '';
}

// Helper: navigate to first project detail page
async function navigateToProjectDetail(page: any) {
  await page.goto('/projects');
  await page.waitForLoadState('networkidle');
  const projectCard = page.locator('.ant-card-hoverable').first();
  await expect(projectCard).toBeVisible({ timeout: 5000 });
  await projectCard.click();
  await page.waitForLoadState('networkidle');
}

test.describe('SKU Management', () => {
  test('SKU API returns 401 without auth', async ({ request }) => {
    const response = await request.get('/api/v1/projects/fake-id/skus', {
      headers: { Authorization: '' },
    });
    expect(response.status()).toBe(401);
  });

  test('SKU list API works with auth', async ({ request }) => {
    const token = getAuthToken();
    expect(token).toBeTruthy();

    const projectsRes = await request.get('/api/v1/projects', {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(projectsRes.ok()).toBeTruthy();
    const projectsData = await projectsRes.json();
    const projects = projectsData.data?.items || projectsData.data || [];
    expect(projects.length).toBeGreaterThan(0);

    const projectId = projects[0].id;
    const skuRes = await request.get(`/api/v1/projects/${projectId}/skus`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(skuRes.ok()).toBeTruthy();
    const skuData = await skuRes.json();
    expect(skuData.data).toBeDefined();
  });

  test('SKU CRUD flow via API', async ({ request }) => {
    const token = getAuthToken();
    expect(token).toBeTruthy();
    const headers = { Authorization: `Bearer ${token}` };

    // Get a project
    const projectsRes = await request.get('/api/v1/projects', { headers });
    const projectsData = await projectsRes.json();
    const projects = projectsData.data?.items || projectsData.data || [];
    expect(projects.length).toBeGreaterThan(0);
    const projectId = projects[0].id;

    // Create SKU
    const createRes = await request.post(`/api/v1/projects/${projectId}/skus`, {
      headers,
      data: { name: 'E2E测试SKU', code: 'E2E-SKU-001', description: 'playwright测试用' },
    });
    expect(createRes.ok()).toBeTruthy();
    const created = await createRes.json();
    const skuId = created.data.id;
    expect(skuId).toBeTruthy();
    expect(created.data.name).toBe('E2E测试SKU');

    // List and verify
    const listRes = await request.get(`/api/v1/projects/${projectId}/skus`, { headers });
    const listData = await listRes.json();
    const items = listData.data?.items || [];
    expect(items.some((s: any) => s.id === skuId)).toBeTruthy();

    // Update SKU
    const updateRes = await request.put(`/api/v1/projects/${projectId}/skus/${skuId}`, {
      headers,
      data: { name: 'E2E测试SKU-更新' },
    });
    expect(updateRes.ok()).toBeTruthy();

    // BOM items API
    const saveBomRes = await request.put(`/api/v1/projects/${projectId}/skus/${skuId}/bom-items`, {
      headers, data: [],
    });
    expect(saveBomRes.ok()).toBeTruthy();

    const getBomRes = await request.get(`/api/v1/projects/${projectId}/skus/${skuId}/bom-items`, { headers });
    expect(getBomRes.ok()).toBeTruthy();

    // CMF API
    const saveCmfRes = await request.put(`/api/v1/projects/${projectId}/skus/${skuId}/cmf`, {
      headers, data: [],
    });
    expect(saveCmfRes.ok()).toBeTruthy();

    const getCmfRes = await request.get(`/api/v1/projects/${projectId}/skus/${skuId}/cmf`, { headers });
    expect(getCmfRes.ok()).toBeTruthy();

    // Full BOM
    const fullBomRes = await request.get(`/api/v1/projects/${projectId}/skus/${skuId}/full-bom`, { headers });
    expect(fullBomRes.ok()).toBeTruthy();

    // Delete SKU
    const deleteRes = await request.delete(`/api/v1/projects/${projectId}/skus/${skuId}`, { headers });
    expect(deleteRes.ok()).toBeTruthy();
  });

  test('project detail page has SKU tab', async ({ page }) => {
    await navigateToProjectDetail(page);
    const skuTab = page.getByRole('tab', { name: /SKU配色/ });
    await expect(skuTab).toBeVisible();
  });

  test('SKU tab shows list and create button', async ({ page }) => {
    await navigateToProjectDetail(page);

    await page.getByRole('tab', { name: /SKU配色/ }).click();
    await page.waitForTimeout(500);

    await expect(page.getByRole('button', { name: '新建SKU' })).toBeVisible();
    await expect(page.getByText('配色方案 / SKU')).toBeVisible();
  });

  test('create and delete SKU via UI', async ({ page }) => {
    await navigateToProjectDetail(page);

    // Click SKU tab
    await page.getByRole('tab', { name: /SKU配色/ }).click();
    await page.waitForTimeout(500);

    // Click create button
    await page.getByRole('button', { name: '新建SKU' }).click();

    // Wait for modal to appear
    await expect(page.getByText('新建SKU').first()).toBeVisible();

    // Fill form (simplified: only name field in new modal)
    await page.getByLabel('名称').fill('Playwright测试色');

    // Submit and wait for modal to close
    await page.getByRole('button', { name: '确 定' }).click();

    // Wait for the modal to disappear (success) with longer timeout
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10000 });

    // Wait for the SKU card to appear
    await expect(page.getByText('Playwright测试色')).toBeVisible({ timeout: 5000 });

    // Delete the created SKU
    const deleteBtn = page.locator('.anticon-delete').last();
    await deleteBtn.click();
    await page.waitForTimeout(300);

    // Confirm deletion in popconfirm
    await page.getByRole('button', { name: '确 定' }).click();
    await page.waitForTimeout(1500);
  });
});
