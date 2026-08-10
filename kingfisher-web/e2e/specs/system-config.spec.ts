import { test, expect } from '../fixtures/test';
import { CREDENTIALS } from '../utils/constants';

test.beforeEach(async ({ page }) => {
  const resp = await page.request.post('http://localhost:18080/api/v1/auth/login', {
    data: { username: CREDENTIALS.admin.username, password: CREDENTIALS.admin.password },
  });
  const body = await resp.json();
  await page.goto('/login');
  await page.evaluate(
    ({ t, r }) => { localStorage.setItem('kingfisher_token', t); localStorage.setItem('kingfisher_refresh', r); },
    { t: body.data.access_token, r: body.data.refresh_token },
  );
});

test('配置列表显示配置项', async ({ page }) => {
  await page.goto('/system/configs');
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table')).toBeVisible();
  await expect(page.locator('.ant-table')).toContainText('site_name');
  await expect(page.locator('.ant-table')).toContainText('max_login_attempts');
});

test('编辑 site_name', async ({ page }) => {
  await page.goto('/system/configs');
  await page.waitForLoadState('networkidle');
  await page.locator('tr', { hasText: 'site_name' }).getByText('编辑').click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  await page.locator('#value').clear();
  await page.locator('#value').fill('E2E Test');
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
  await expect(page.locator('.ant-modal')).not.toBeVisible({ timeout: 10000 });
  // Restore
  await page.locator('tr', { hasText: 'site_name' }).getByText('编辑').click();
  await page.locator('#value').clear();
  await page.locator('#value').fill('Kingfisher Admin');
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
});
