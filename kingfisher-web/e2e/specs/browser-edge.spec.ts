import { test, expect } from '@playwright/test';
import { newAuthenticatedPage } from '../fixtures/auth';
import { UserListPage } from '../pages/user-list.page';

test.describe('Browser Edge Cases', () => {
  test('ESC 关闭弹窗', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await userList.addButton().click();
    await expect(userList.modal()).toBeVisible();

    await page.keyboard.press('Escape');
    await expect(userList.modal()).not.toBeVisible();

    await context.close();
  });

  test('弹窗外点击关闭（点击遮罩）', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await userList.addButton().click();
    await expect(userList.modal()).toBeVisible();

    // Click the mask behind the modal
    const mask = page.locator('.ant-modal-mask');
    await mask.click({ position: { x: 10, y: 10 } });
    await expect(userList.modal()).not.toBeVisible();

    await context.close();
  });

  test('Mobile viewport (375x812) 布局不溢出', async ({ browser }) => {
    const context = await browser.newContext({
      viewport: { width: 375, height: 812 },
    });
    const page = await newAuthenticatedPage(context, 'admin');

    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    expect(bodyWidth).toBeLessThanOrEqual(375);

    await context.close();
  });

  test('页面刷新后保持状态不跳转登录', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    expect(page.url()).toContain('/dashboard');

    await page.reload();
    await page.waitForLoadState('networkidle');

    // Should stay on dashboard
    expect(page.url()).toContain('/dashboard');

    await context.close();
  });
});
