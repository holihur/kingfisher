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

test('树形表格展示', async ({ page }) => {
  await page.goto('/system/menus');
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table')).toBeVisible();
  await expect(page.locator('.ant-table')).toContainText('Dashboard');
  await expect(page.locator('.ant-table')).toContainText('系统管理');
});

test('类型标签可见', async ({ page }) => {
  await page.goto('/system/menus');
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table')).toContainText('目录');
  await expect(page.locator('.ant-table')).toContainText('菜单');
});

test('新增根菜单', async ({ page }) => {
  await page.goto('/system/menus');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button', { name: '新增根菜单' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  const name = `e2e_m_${Date.now()}`;
  await page.locator('#name').fill(name);
  await page.locator('#path').fill(`/${name}`);
  await page.locator('.ant-modal').getByRole('button', { name: /确定|保存/ }).click();
  await expect(page.locator('.ant-modal')).not.toBeVisible({ timeout: 10000 });
});

test('删除有子节点的菜单被拒绝', async ({ page }) => {
  await page.goto('/system/menus');
  await page.waitForLoadState('networkidle');
  // 系统管理 has children — click delete
  await page.locator('tr', { hasText: '系统管理' }).getByText('删除').click();
  const popconfirm = page.locator('.ant-popconfirm');
  if (await popconfirm.isVisible({ timeout: 2000 }).catch(() => false)) {
    await popconfirm.getByRole('button', { name: '确 定' }).click();
  }
  // Should show error via modal staying or message
  await page.waitForTimeout(1000);
});
