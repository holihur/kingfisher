import type { Page, Locator } from '@playwright/test';

/**
 * Page Object Model for the Profile (个人中心) page.
 */
export class ProfilePage {
  constructor(readonly page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto('/profile');
    await this.page.waitForLoadState('networkidle');
  }

  // ---- Tabs ----
  profileTab(): Locator {
    return this.page.locator('.ant-tabs-tab').filter({ hasText: '用户资料' });
  }
  passwordTab(): Locator {
    return this.page.locator('.ant-tabs-tab').filter({ hasText: '修改密码' });
  }
  logTab(): Locator {
    return this.page.locator('.ant-tabs-tab').filter({ hasText: '登录日志' });
  }

  // ---- Profile form ----
  nicknameInput(): Locator {
    return this.page.getByPlaceholder('设置显示昵称');
  }
  emailInput(): Locator {
    return this.page.getByPlaceholder('设置邮箱');
  }
  saveProfileButton(): Locator {
    return this.page.getByRole('button', { name: '保存' });
  }
  uploadButton(): Locator {
    return this.page.getByRole('button', { name: '上传头像' });
  }
  avatarImage(): Locator {
    return this.page.locator('img[alt="avatar"]');
  }

  // ---- Password form ----
  oldPasswordInput(): Locator {
    return this.page.locator('.ant-form-item').filter({ hasText: '旧密码' }).locator('input');
  }
  newPasswordInput(): Locator {
    return this.page.locator('.ant-form-item').filter({ hasText: '新密码' }).locator('input').first();
  }
  confirmPasswordInput(): Locator {
    return this.page.locator('.ant-form-item').filter({ hasText: '确认新密码' }).locator('input');
  }
  changePasswordButton(): Locator {
    return this.page.getByRole('button', { name: '修改密码' });
  }

  // ---- Login logs table ----
  loginLogTable(): Locator {
    return this.page.locator('.ant-table');
  }
  logRows(): Locator {
    return this.page.locator('.ant-table-tbody tr');
  }

  // ---- Header dropdown ----
  async openUserDropdown(): Promise<void> {
    await this.page.locator('.ant-layout-header [class*="dropdown"]').first().click();
  }
  profileMenuEntry(): Locator {
    return this.page.getByText('个人中心');
  }
}
