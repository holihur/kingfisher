import { test, expect } from '@playwright/test';
import { newAuthenticatedPage, loginViaUI } from '../fixtures/auth';
import { URLS } from '../utils/constants';

test.describe('Route Guard & Logout', () => {
  test('未登录访问 /dashboard → 跳转 /login', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('未登录访问 /system/users → 跳转 /login 带 redirect', async ({ page }) => {
    await page.goto(URLS.users);
    await expect(page).toHaveURL(/\/login/);
    await expect(page).toHaveURL(/redirect=/);
  });

  test('redirect 参数：登录后跳回原始目标页面', async ({ page }) => {
    await page.evaluate(() => localStorage.clear());
    await page.goto(URLS.users);
    await page.waitForURL(/\/login/);

    await loginViaUI(page, 'admin');

    // Should redirect back to the originally requested page
    expect(page.url()).toContain(URLS.users);
  });

  test('退出后访问受保护页面 → 跳转 /login', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');

    // Logout
    await page.locator('.ant-layout-header .ant-dropdown-trigger').click();
    await page.getByText('退出登录').click();
    await page.waitForURL('**/login');

    // Navigate to protected page
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);

    await context.close();
  });

  test('多个 browser context 各自独立登录', async ({ browser }) => {
    const ctx1 = await browser.newContext();
    const page1 = await newAuthenticatedPage(ctx1, 'admin');
    await page1.goto('/dashboard');

    const ctx2 = await browser.newContext();
    const page2 = await newAuthenticatedPage(ctx2, 'viewer');
    await page2.goto('/dashboard');

    // Both should be authenticated with different tokens
    const token1 = await page1.evaluate(() => localStorage.getItem('kingfisher_token'));
    const token2 = await page2.evaluate(() => localStorage.getItem('kingfisher_token'));
    expect(token1).toBeTruthy();
    expect(token2).toBeTruthy();
    expect(token1).not.toBe(token2);

    // Logout context 1 → should not affect context 2
    await page1.locator('.ant-layout-header .ant-dropdown-trigger').click();
    await page1.getByText('退出登录').click();
    await page1.waitForURL('**/login');

    const token1After = await page1.evaluate(() => localStorage.getItem('kingfisher_token'));
    expect(token1After).toBeNull();

    const token2After = await page2.evaluate(() => localStorage.getItem('kingfisher_token'));
    expect(token2After).toBeTruthy();

    await ctx1.close();
    await ctx2.close();
  });
});
