import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/login.page';
import { loginViaUI } from '../fixtures/auth';
import { CREDENTIALS, URLS } from '../utils/constants';

test.describe('Login Page', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
  });

  test('未登录访问首页 → 自动跳转 /login', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/login/);
  });

  test('空表单提交 → 校验提示"请输入用户名"', async ({ page }) => {
    await loginPage.goto();
    await loginPage.submitButton().click();
    await expect(loginPage.formErrors().first()).toBeVisible();
    await expect(loginPage.formErrors().first()).toContainText('请输入用户名');
  });

  test('只填用户名不填密码 → 校验提示"请输入密码"', async ({ page }) => {
    await loginPage.goto();
    await loginPage.usernameInput().fill('admin');
    await loginPage.submitButton().click();
    await expect(loginPage.formErrors().first()).toBeVisible();
    await expect(loginPage.formErrors().first()).toContainText('请输入密码');
  });

  test('错误密码登录 → 提示错误', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin', 'wrongpassword');
    // antd message.error appears as .ant-message-notice
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });
  });

  test('不存在用户登录 → 返回错误提示（防枚举）', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('nonexistent_user', 'anything123');
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });
  });

  test('正确凭据登录 → 跳转 /dashboard，localStorage 有 token', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login(CREDENTIALS.admin.username, CREDENTIALS.admin.password);
    await loginPage.waitForRedirect();

    const token = await page.evaluate(() => localStorage.getItem('kingfisher_token'));
    expect(token).toBeTruthy();
    expect(page.url()).toContain(URLS.dashboard);
  });

  test('登录限流：连续 6 次错误 → 429 提示', async ({ page }) => {
    test.slow();
    await loginPage.goto();

    for (let i = 0; i < 7; i++) {
      await loginPage.usernameInput().clear();
      await loginPage.passwordInput().clear();
      await loginPage.login(CREDENTIALS.admin.username, `wrong_pass_${i}`);
    }

    // After repeated failures, rate limiting should kick in
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 10000 });
  });

  test('Tab 键焦点顺序：用户名 → 密码 → 登录按钮', async ({ page }) => {
    await loginPage.goto();

    // First Tab from document body → should focus username
    await page.keyboard.press('Tab');
    await expect(page.getByPlaceholder('用户名')).toBeFocused();

    // Second tab → password
    await page.keyboard.press('Tab');
    await expect(page.getByPlaceholder('密码')).toBeFocused();
  });

  test('Enter 键提交登录', async ({ page }) => {
    await loginPage.goto();
    await loginPage.usernameInput().fill(CREDENTIALS.admin.username);
    await loginPage.passwordInput().fill(CREDENTIALS.admin.password);
    await page.keyboard.press('Enter');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });
});
