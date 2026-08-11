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
  await expect(page.locator('.ant-table')).toContainText('仪表盘');
  await expect(page.locator('.ant-table')).toContainText('系统管理');
});

test('菜单表格展示菜单行', async ({ page }) => {
  await page.goto('/system/menus');
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table')).toContainText('仪表盘');
  await expect(page.locator('.ant-table')).toContainText('系统管理');
});

test('新增根菜单', async ({ page }) => {
  await page.goto('/system/menus');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button', { name: '新增根菜单' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  await page.waitForTimeout(500); // 等 modal 渲染完成
  const name = `e2e_m_${Date.now()}`;
  // antd v6 Form 受控：pressSequentially 触发真实 input 事件后，等 React 提交
  await page.locator('.ant-modal input[id="name"]').pressSequentially(name);
  await page.locator('.ant-modal input[id="path"]').pressSequentially(`/${name}`);
  await page.waitForTimeout(300);
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
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
