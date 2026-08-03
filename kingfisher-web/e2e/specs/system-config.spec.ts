import { test, expect } from '@playwright/test';
import { newAuthenticatedPage } from '../fixtures/auth';
import { ConfigManagePage } from '../pages/config-manage.page';

test.describe('System Config', () => {
  test('配置列表显示 5 个配置项', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const configPage = new ConfigManagePage(page);
    await configPage.goto();

    await expect(configPage.table()).toBeVisible();
    await expect(configPage.table()).toContainText('site_name');
    await expect(configPage.table()).toContainText('max_login_attempts');
    await expect(configPage.tableRows()).toHaveCount(5);

    await context.close();
  });

  test('编辑 site_name 配置值', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await newAuthenticatedPage(context, 'admin');
    const configPage = new ConfigManagePage(page);
    await configPage.goto();

    // Edit site_name
    await configPage.editLink('site_name').click();
    await expect(configPage.modal()).toBeVisible();

    const valueInput = configPage.modal().locator('#value');
    await valueInput.clear();
    await valueInput.fill('Kingfisher E2E Test');
    await configPage.modal().getByRole('button', { name: /确定|保存/ }).click();
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    // Restore original
    await configPage.editLink('site_name').click();
    await valueInput.clear();
    await valueInput.fill('Kingfisher Admin');
    await configPage.modal().getByRole('button', { name: /确定|保存/ }).click();
    await expect(page.locator('.ant-message-notice')).toBeVisible({ timeout: 5000 });

    await context.close();
  });
});
