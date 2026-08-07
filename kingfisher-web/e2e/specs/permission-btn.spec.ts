import { test, expect } from '@playwright/test';
import { CREDENTIALS } from '../utils/constants';

async function loginAs(page, role: 'admin' | 'editor' | 'viewer') {
  const resp = await page.request.post('http://localhost:18080/api/v1/auth/login', {
    data: { username: CREDENTIALS[role].username, password: CREDENTIALS[role].password },
  });
  const body = await resp.json();
  await page.goto('/login');
  await page.evaluate(
    ({ t, r }) => { localStorage.setItem('kingfisher_token', t); localStorage.setItem('kingfisher_refresh', r); },
    { t: body.data.access_token, r: body.data.refresh_token },
  );
}

test('admin → 用户列表全部按钮可见', async ({ page }) => {
  await loginAs(page, 'admin');
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await expect(page.getByRole('button', { name: '新增用户' })).toBeVisible();
  await expect(page.locator('tr', { hasText: 'admin' }).getByText('编辑')).toBeVisible();
});

test('editor → 有新增编辑，无删除', async ({ page }) => {
  await loginAs(page, 'editor');
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await expect(page.getByRole('button', { name: '新增用户' })).toBeVisible();
  await expect(page.locator('tr', { hasText: 'admin' }).getByText('编辑')).toBeVisible();
  await expect(page.locator('tr', { hasText: 'admin' }).getByText('删除')).toHaveCount(0);
});

test('viewer → 无增删改按钮', async ({ page }) => {
  await loginAs(page, 'viewer');
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await expect(page.getByRole('button', { name: '新增用户' })).toHaveCount(0);
  await expect(page.locator('tr', { hasText: 'admin' }).getByText('编辑')).toHaveCount(0);
  await expect(page.locator('tr', { hasText: 'admin' }).getByText('删除')).toHaveCount(0);
});

test('viewer → 菜单管理无新增', async ({ page }) => {
  await loginAs(page, 'viewer');
  await page.goto('/system/menus');
  await page.waitForLoadState('networkidle');
  await expect(page.getByRole('button', { name: '新增根菜单' })).toHaveCount(0);
});

test('viewer → 配置无编辑', async ({ page }) => {
  await loginAs(page, 'viewer');
  await page.goto('/system/configs');
  await page.waitForLoadState('networkidle');
  await expect(page.locator('tr', { hasText: 'site_name' }).getByText('编辑')).toHaveCount(0);
});
