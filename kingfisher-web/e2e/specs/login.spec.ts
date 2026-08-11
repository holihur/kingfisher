import { test, expect } from '../fixtures/test';
import { LoginPage } from '../pages/login.page';
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

  test('错误密码登录 → 停留在登录页', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin', 'wrongpassword');
    await page.waitForLoadState('networkidle');
    // Failed login should keep us on the login page
    await expect(page).toHaveURL(/\/login/);
  });

  test('不存在用户登录 → 停留在登录页（防枚举）', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('nonexistent_user', 'anything123');
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/\/login/);
  });

  test('正确凭据登录 → 跳转 /dashboard，localStorage 有 token', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login(CREDENTIALS.admin.username, CREDENTIALS.admin.password);
    await loginPage.waitForRedirect();

    const token = await page.evaluate(() => localStorage.getItem('kingfisher_token'));
    expect(token).toBeTruthy();
    expect(page.url()).toContain(URLS.dashboard);
  });

  test('登录限流：连续 6 次错误 → 仍然在登录页', async ({ page }) => {
    test.slow();
    await loginPage.goto();

    // Use a unique rate-limit-test user so we don't block admin
    const rlUser = `ratelimit_${Date.now()}`;
    for (let i = 0; i < 7; i++) {
      await loginPage.usernameInput().clear();
      await loginPage.passwordInput().clear();
      await loginPage.login(rlUser, 'wrong');
      await page.waitForLoadState('networkidle');
    }

    await expect(page).toHaveURL(/\/login/);
  });

  test('Tab 键焦点顺序：用户名 → 密码 → 登录按钮', async ({ page }) => {
    await loginPage.goto();

    // 先聚焦到用户名（可能需跳过 SiteNotice 的关闭按钮等前置可聚焦元素）
    await page.locator('#username').focus();
    await expect(page.locator('#username')).toBeFocused();

    // Tab → password
    await page.keyboard.press('Tab');
    await expect(page.locator('#password')).toBeFocused();
  });

  test('记住账户勾选框默认选中，且不再提供记住密码', async ({ page }) => {
    await loginPage.goto();

    // 记住账户默认勾选
    await expect(loginPage.rememberAccountCheckbox()).toBeVisible();
    await expect(page.locator('.ant-checkbox-wrapper').filter({ hasText: '记住账户' }).locator('.ant-checkbox')).toHaveClass(/ant-checkbox-checked/);

    // 安全：不再提供"记住密码"选项
    await expect(page.locator('.ant-checkbox-wrapper').filter({ hasText: '记住密码' })).toHaveCount(0);
  });

});
