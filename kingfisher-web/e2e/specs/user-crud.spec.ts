import { test, expect } from '@playwright/test';
import { newAuthenticatedPage } from '../fixtures/auth';
import { UserListPage } from '../pages/user-list.page';

test.describe('User CRUD', () => {
  test('用户列表展示列', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await expect(userList.table()).toBeVisible();
    const headers = userList.table().locator('.ant-table-thead');
    await expect(headers).toContainText('用户名');
    await expect(headers).toContainText('邮箱');
    await expect(headers).toContainText('状态');

    await context.close();
  });

  test('搜索过滤 admin', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await userList.search('admin');
    const rows = userList.tableRows();
    await expect(rows.first()).toContainText('admin');

    await context.close();
  });

  test('搜索无结果 → 空状态', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await userList.search('zzz_no_such_user_xyz');
    // Table should show empty state
    await expect(userList.table()).toContainText(/没有找到|暂无数据/, { timeout: 5000 });

    await context.close();
  });

  test('重置搜索', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await userList.search('admin');
    await userList.resetSearch();
    const rows = userList.tableRows();
    await expect(rows).not.toHaveCount(0);

    await context.close();
  });

  test('新增用户 - 空提交表单校验', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await userList.addButton().click();
    await expect(userList.modal()).toBeVisible();
    await expect(userList.modalTitle()).toContainText('新增用户');

    // Submit empty form
    await userList.modalSubmitButton().click();
    // Should show validation errors
    await expect(userList.modal()).toContainText(/请输入|必填|不能为空/);

    await context.close();
  });

  test('新增用户 - 成功创建', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    const newUser = `e2e_test_${Date.now()}`;
    await userList.addButton().click();
    await expect(userList.modal()).toBeVisible();

    await userList.usernameField().fill(newUser);
    await userList.passwordField().fill('Abcd1234');
    await userList.emailField().fill(`${newUser}@test.com`);
    await userList.modalSubmitButton().click();

    // Wait for success
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    // Search for the new user
    await userList.search(newUser);
    await expect(userList.table()).toContainText(newUser);

    await context.close();
  });

  test('编辑用户 - 用户名 disabled', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    await userList.search('admin');
    await userList.editLink('admin').click();

    await expect(userList.modal()).toBeVisible();
    await expect(userList.modalTitle()).toContainText('编辑用户');
    await expect(userList.usernameField()).toBeDisabled();

    await userList.emailField().clear();
    await userList.emailField().fill('admin_updated@example.com');
    await userList.modalSubmitButton().click();
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    await context.close();
  });

  test('删除用户 - Popconfirm 确认后删除', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const userList = new UserListPage(page);
    await userList.goto();

    // Create a user to delete
    const newUser = `e2e_del_${Date.now()}`;
    await userList.addButton().click();
    await userList.usernameField().fill(newUser);
    await userList.passwordField().fill('Abcd1234');
    await userList.modalSubmitButton().click();
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    // Search and delete
    await userList.search(newUser);
    await expect(userList.table()).toContainText(newUser);

    await userList.deleteLink(newUser).click();
    await expect(userList.popconfirm()).toBeVisible();
    await userList.popconfirmOk().click();
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    await context.close();
  });
});
