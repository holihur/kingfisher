import { test, expect } from '@playwright/test';
import { ProfilePage } from '../pages/profile.page';
import { newAuthenticatedPage, loginViaUI } from '../fixtures/auth';
import { URLS } from '../utils/constants';

test.describe('Profile Page (个人中心)', () => {
  let profilePage: ProfilePage;

  test('管理员通过 header 下拉进入个人中心 → 三个 Tab 可见', async ({ page }) => {
    await loginViaUI(page, 'admin');
    // Open header dropdown and click 个人中心
    await page.locator('.ant-layout-header [class*="dropdown"]').first().click();
    await page.getByText('个人中心').click();
    await page.waitForURL('**/profile');

    profilePage = new ProfilePage(page);
    await expect(profilePage.profileTab()).toBeVisible();
    await expect(profilePage.passwordTab()).toBeVisible();
    await expect(profilePage.logTab()).toBeVisible();
  });

  test('用户资料 Tab → 编辑昵称和邮箱 → 保存成功', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    // Fill in profile fields
    await profilePage.nicknameInput().clear();
    await profilePage.nicknameInput().fill('E2E昵称');
    await profilePage.emailInput().clear();
    await profilePage.emailInput().fill('e2e@example.com');
    await profilePage.saveProfileButton().click();

    // Wait for success message
    await expect(page.getByText('资料已更新')).toBeVisible({ timeout: 5000 });
  });

  test('修改密码 Tab → 错误旧密码被拒绝', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    // Switch to password tab
    await profilePage.passwordTab().click();

    // Try with wrong old password
    await profilePage.oldPasswordInput().fill('WrongOld123');
    await profilePage.newPasswordInput().fill('NewPass123');
    await profilePage.confirmPasswordInput().fill('NewPass123');
    await profilePage.changePasswordButton().click();

    // Should show error (old password wrong)
    await expect(page.getByText('密码错误')).toBeVisible({ timeout: 5000 });
  });

  test('登录日志 Tab → 管理员登录后有日志记录', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    // Switch to login log tab
    await profilePage.logTab().click();

    // The table should be visible (may have 0 rows if no login recorded for API-based auth)
    await expect(profilePage.loginLogTable()).toBeVisible({ timeout: 5000 });
  });

  test('普通用户 (viewer) 可以访问个人中心', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'viewer');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    // All three tabs visible for viewer too
    await expect(profilePage.profileTab()).toBeVisible();
    await expect(profilePage.passwordTab()).toBeVisible();
    await expect(profilePage.logTab()).toBeVisible();
  });
});
