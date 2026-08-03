import { test, expect } from '@playwright/test';
import { newAuthenticatedPage, loginViaUI } from '../fixtures/auth';
import { LayoutPage } from '../pages/layout.page';
import { CREDENTIALS } from '../utils/constants';

test.describe('Layout & Navigation', () => {
  test('侧边栏显示 Logo + 菜单树', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');
    const layout = new LayoutPage(page);

    await expect(layout.logo()).toBeVisible();
    await expect(layout.menuItem('Dashboard')).toBeVisible();
    await expect(layout.menuItem('系统管理')).toBeVisible();

    await context.close();
  });

  test('侧边栏折叠/展开', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');
    const layout = new LayoutPage(page);

    await expect(layout.logo()).toBeVisible();

    await layout.collapse();
    await expect(layout.collapsedLogo()).toBeVisible();

    await layout.expand();
    await expect(layout.logo()).toBeVisible();

    await context.close();
  });

  test('顶栏显示用户名 + 头像', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');
    const layout = new LayoutPage(page);

    await expect(layout.avatar()).toBeVisible();
    await expect(layout.headerUser()).toContainText(CREDENTIALS.admin.username);

    await context.close();
  });

  test('面包屑导航', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');
    const layout = new LayoutPage(page);

    await expect(layout.breadcrumb()).toContainText('首页');

    await context.close();
  });

  test('页面刷新后保持登录状态', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');

    await page.reload();
    await page.waitForLoadState('networkidle');

    // Should still be on dashboard, not redirected to login
    expect(page.url()).toContain('/dashboard');

    await context.close();
  });

  test('退出登录 → 跳转 /login', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');

    // Open user dropdown then click logout
    await page.locator('.ant-layout-header .ant-dropdown-trigger').click();
    await page.getByText('退出登录').click();
    await page.waitForURL('**/login');

    const token = await page.evaluate(() => localStorage.getItem('kingfisher_token'));
    expect(token).toBeNull();

    await context.close();
  });
});
