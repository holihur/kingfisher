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

test('角色列表显示 3 个角色', async ({ page }) => {
  await page.goto('/system/roles');
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table')).toBeVisible();
  await expect(page.locator('.ant-table')).toContainText('超级管理员');
  await expect(page.locator('.ant-table')).toContainText('编辑');
  await expect(page.locator('.ant-table')).toContainText('访客');
});

test('点击权限弹窗', async ({ page }) => {
  await page.goto('/system/roles');
  await page.waitForLoadState('networkidle');
  await page.locator('tr', { hasText: '超级管理员' }).getByText('权限').click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.locator('.ant-modal')).not.toBeVisible();
});

test('点击菜单弹窗', async ({ page }) => {
  await page.goto('/system/roles');
  await page.waitForLoadState('networkidle');
  await page.locator('tr', { hasText: '超级管理员' }).getByText('菜单').click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.locator('.ant-modal')).not.toBeVisible();
});

test('新增角色', async ({ page }) => {
  await page.goto('/system/roles');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button', { name: '新增角色' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  const name = `e2er_${Date.now()}`;
  await page.getByRole('textbox', { name: /角色名/ }).fill(name);
  await page.getByRole('textbox', { name: /编码/ }).fill(name);
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
  await expect(page.locator('.ant-modal')).not.toBeVisible({ timeout: 10000 });
});
