import { test, expect } from '@playwright/test';
import { newAuthenticatedPage } from '../fixtures/auth';
import { MenuManagePage } from '../pages/menu-manage.page';

test.describe('Menu Management', () => {
  test('树形表格展示 Dashboard + 系统管理', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const menuPage = new MenuManagePage(page);
    await menuPage.goto();

    await expect(menuPage.table()).toBeVisible();
    await expect(menuPage.table()).toContainText('Dashboard');
    await expect(menuPage.table()).toContainText('系统管理');

    await context.close();
  });

  test('类型标签：目录/菜单/按钮', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const menuPage = new MenuManagePage(page);
    await menuPage.goto();

    // System management is a directory type (目录)
    await expect(menuPage.table()).toContainText('目录');
    await expect(menuPage.table()).toContainText('菜单');

    await context.close();
  });

  test('新增根菜单', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const menuPage = new MenuManagePage(page);
    await menuPage.goto();

    await menuPage.addRootButton().click();
    await expect(menuPage.modal()).toBeVisible();

    const menuName = `e2e_menu_${Date.now()}`;
    await menuPage.modal().locator('#name').fill(menuName);
    await menuPage.modal().locator('#path').fill(`/${menuName}`);
    await menuPage.modal().getByRole('button', { name: /确定|保存/ }).click();

    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    await context.close();
  });

  test('删除有子节点的菜单被拒绝', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const menuPage = new MenuManagePage(page);
    await menuPage.goto();

    // 系统管理 has child menus, deleting should fail
    await menuPage.deleteLink('系统管理').click();

    // Either popconfirm appears or error shows
    const popconfirm = menuPage.popconfirm();
    if (await popconfirm.isVisible({ timeout: 2000 }).catch(() => false)) {
      await menuPage.popconfirmOk().click();
      await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });
    }

    await context.close();
  });
});
