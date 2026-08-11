import type { Page, Locator } from '@playwright/test';

/**
 * Page Object Model for the Login page.
 */
export class LoginPage {
  constructor(readonly page: Page) {}

  async goto(): Promise<void> {
    // 页面加载前清掉「记住账户」的已存账户，避免登录页进入"选择账户"视图而找不到登录表单
    await this.page.addInitScript(() => localStorage.removeItem('kingfisher_accounts'));
    await this.page.goto('/login');
    await this.page.waitForLoadState('networkidle');
  }

  usernameInput(): Locator {
    return this.page.getByPlaceholder('用户名');
  }

  passwordInput(): Locator {
    return this.page.getByPlaceholder('密码');
  }

  submitButton(): Locator {
    // antd v6 会自动在两个中文之间插空格（"登 录"），用正则容忍
    return this.page.getByRole('button', { name: /登\s*录/ });
  }

  registerLink(): Locator {
    return this.page.getByText('去注册');
  }

  rememberAccountCheckbox(): Locator {
    return this.page.getByText('记住账户');
  }

  rememberPwdCheckbox(): Locator {
    return this.page.getByText('记住密码');
  }

  formErrors(): Locator {
    return this.page.locator('.ant-form-item-explain-error');
  }

  title(): Locator {
    return this.page.locator('h2');
  }

  subtitle(): Locator {
    return this.page.locator('p').filter({ hasText: '后台管理系统' });
  }

  async login(username: string, password: string): Promise<void> {
    await this.usernameInput().fill(username);
    await this.passwordInput().fill(password);
    await this.submitButton().click();
  }

  async waitForRedirect(): Promise<void> {
    await this.page.waitForURL('**/dashboard');
  }
}
