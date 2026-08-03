import { test, expect } from '@playwright/test';
import { newAuthenticatedPage } from '../fixtures/auth';
import { RoleListPage } from '../pages/role-list.page';

test.describe('Role & Permission Management', () => {
  test('角色列表显示 3 个角色', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const rolePage = new RoleListPage(page);
    await rolePage.goto();

    await expect(rolePage.table()).toBeVisible();
    await expect(rolePage.table()).toContainText('超级管理员');
    await expect(rolePage.table()).toContainText('编辑');
    await expect(rolePage.table()).toContainText('访客');
    await expect(rolePage.tableRows()).toHaveCount(3);

    await context.close();
  });

  test('点击权限 → 弹窗显示权限分配', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const rolePage = new RoleListPage(page);
    await rolePage.goto();

    await rolePage.permissionsLink('超级管理员').click();
    await expect(rolePage.modal()).toBeVisible();
    // Modal should contain permission-related content
    await expect(rolePage.modal()).toContainText(/user|用户|menu|菜单/);

    await page.keyboard.press('Escape');
    await expect(rolePage.modal()).not.toBeVisible();

    await context.close();
  });

  test('点击菜单 → 弹窗显示菜单树分配', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const rolePage = new RoleListPage(page);
    await rolePage.goto();

    await rolePage.menusLink('超级管理员').click();
    await expect(rolePage.modal()).toBeVisible();
    // Modal should show menu tree
    await expect(rolePage.modal()).toContainText(/Dashboard|系统管理/);

    await page.keyboard.press('Escape');
    await expect(rolePage.modal()).not.toBeVisible();

    await context.close();
  });

  test('新增角色', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const rolePage = new RoleListPage(page);
    await rolePage.goto();

    await rolePage.addButton().click();
    await expect(rolePage.modal()).toBeVisible();

    const roleName = `e2e_role_${Date.now()}`;
    await rolePage.modal().locator('#name').fill(roleName);
    await rolePage.modal().locator('#code').fill(`e2e_${Date.now()}`);
    await rolePage.modal().getByRole('button', { name: /确定|保存/ }).click();

    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    await context.close();
  });
});
