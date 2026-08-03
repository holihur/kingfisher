import { test, expect } from '@playwright/test';
import { newAuthenticatedPage } from '../fixtures/auth';
import { UserListPage } from '../pages/user-list.page';
import { MenuManagePage } from '../pages/menu-manage.page';
import { ConfigManagePage } from '../pages/config-manage.page';

test.describe('Permission Button Visibility', () => {
  test('admin → 用户列表全部按钮可见', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await expect(userList.addButton()).toBeVisible();
    await expect(userList.editLink('admin')).toBeVisible();

    await context.close();
  });

  test('editor → 有新增编辑，无删除', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'editor');
    const userList = new UserListPage(page);
    await userList.goto();

    // Editor has user:create and user:update
    await expect(userList.addButton()).toBeVisible();
    await expect(userList.editLink('admin')).toBeVisible();
    // Editor does NOT have user:delete
    await expect(userList.deleteLink('admin')).toHaveCount(0);

    await context.close();
  });

  test('viewer → 用户列表无增删改按钮', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'viewer');
    const userList = new UserListPage(page);
    await userList.goto();

    // Viewer has no create/update/delete permissions
    await expect(userList.addButton()).toHaveCount(0);
    await expect(userList.editLink('admin')).toHaveCount(0);
    await expect(userList.deleteLink('admin')).toHaveCount(0);
    // Table should still show data
    await expect(userList.table()).toBeVisible();

    await context.close();
  });

  test('viewer → 菜单管理无增删改按钮', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'viewer');
    const menuPage = new MenuManagePage(page);
    await menuPage.goto();

    await expect(menuPage.addRootButton()).toHaveCount(0);

    await context.close();
  });

  test('viewer → 系统配置无编辑', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'viewer');
    const configPage = new ConfigManagePage(page);
    await configPage.goto();

    await expect(configPage.table()).toBeVisible();
    // Edit link should not exist for viewer
    await expect(configPage.editLink('site_name')).toHaveCount(0);

    await context.close();
  });
});
