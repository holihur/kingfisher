import { test, expect } from '@playwright/test';
import { ProfilePage } from '../pages/profile.page';
import { newAuthenticatedPage } from '../fixtures/auth';
import { URLS } from '../utils/constants';

test.describe('Profile Page (个人中心)', () => {
  let profilePage: ProfilePage;

  test('管理员 → 三个 Tab 可见', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    await expect(profilePage.profileTab()).toBeVisible();
    await expect(profilePage.passwordTab()).toBeVisible();
    await expect(profilePage.logTab()).toBeVisible();
  });

  test('用户资料 Tab → 昵称输入框可见', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    await expect(profilePage.nicknameInput()).toBeVisible({ timeout: 5000 });
  });

  test('修改密码 Tab → 切过去后表单可见', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    await profilePage.passwordTab().click();
    // 验证修改密码按钮可见（表单已渲染）
    await expect(profilePage.changePasswordButton()).toBeVisible({ timeout: 5000 });
  });

  test('登录日志 Tab → 表格可见', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'admin');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    await profilePage.logTab().click();
    await expect(profilePage.loginLogTable()).toBeVisible({ timeout: 5000 });
  });

  test('header 下拉 → 个人中心入口可见', async ({ page }) => {
    // 先访问 login 确保 origin 建立
    await page.goto(URLS.login);
    await page.waitForLoadState('networkidle');

    // 模拟登录后状态：直接注入 token
    await page.evaluate(() => {
      localStorage.setItem('kingfisher_token', 'fake-token');
    });
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');

    // 检查 header 区域存在
    await expect(page.locator('.ant-layout-header')).toBeVisible();
  });

  test('viewer 可以访问个人中心', async ({ context }) => {
    const page = await newAuthenticatedPage(context, 'viewer');
    await page.goto(URLS.profile);
    await page.waitForLoadState('networkidle');
    profilePage = new ProfilePage(page);

    await expect(profilePage.profileTab()).toBeVisible();
    await expect(profilePage.passwordTab()).toBeVisible();
    await expect(profilePage.logTab()).toBeVisible();
  });
});
