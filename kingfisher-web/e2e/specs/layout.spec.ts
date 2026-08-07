import { test, expect } from '@playwright/test';
import { CREDENTIALS } from '../utils/constants';

test.describe('Layout & Navigation', () => {
  test.beforeEach(async ({ page }) => {
    // Login via API and inject token into localStorage
    const resp = await page.request.post('http://localhost:18080/api/v1/auth/login', {
      data: { username: CREDENTIALS.admin.username, password: CREDENTIALS.admin.password },
    });
    const body = await resp.json();
    const { access_token, refresh_token } = body.data;

    await page.goto('/login');
    await page.evaluate(
      ({ t, r }) => {
        localStorage.setItem('kingfisher_token', t);
        localStorage.setItem('kingfisher_refresh', r);
      },
      { t: access_token, r: refresh_token },
    );
    await page.goto('/dashboard');
  });

  test('侧边栏显示 Logo', async ({ page }) => {
    await expect(page.locator('.ant-layout-sider').getByText('Kingfisher')).toBeVisible();
  });

  test('菜单树可见', async ({ page }) => {
    await expect(page.locator('.ant-menu-root').getByText('Dashboard')).toBeVisible();
    await expect(page.locator('.ant-menu-root').getByText('系统管理')).toBeVisible();
  });

  test('侧边栏折叠/展开', async ({ page }) => {
    await expect(page.locator('.ant-layout-sider').getByText('Kingfisher')).toBeVisible();
    await page.locator('.ant-layout-sider-trigger').click();
    await page.waitForTimeout(400);
    await expect(page.locator('.ant-layout-sider').getByText('K')).toBeVisible();
    await page.locator('.ant-layout-sider-trigger').click();
    await page.waitForTimeout(400);
    await expect(page.locator('.ant-layout-sider').getByText('Kingfisher')).toBeVisible();
  });

  test('顶栏显示用户名', async ({ page }) => {
    await expect(page.locator('.ant-layout-header').getByText(CREDENTIALS.admin.username)).toBeVisible();
  });

  test('顶栏有头像', async ({ page }) => {
    await expect(page.locator('.ant-layout-header').locator('.ant-avatar')).toBeVisible();
  });

  test('面包屑导航', async ({ page }) => {
    await expect(page.locator('.ant-breadcrumb')).toContainText('首页');
  });

  test('页面刷新后保持登录', async ({ page }) => {
    await page.reload();
    await page.waitForLoadState('networkidle');
    expect(page.url()).toContain('/dashboard');
  });

  test('退出登录 → 跳转 /login', async ({ page }) => {
    await page.locator('.ant-layout-header').locator('[class*="dropdown"]').first().click();
    await page.getByText('退出登录').click();
    await page.waitForURL('**/login');
    const token = await page.evaluate(() => localStorage.getItem('kingfisher_token'));
    expect(token).toBeNull();
  });
});
