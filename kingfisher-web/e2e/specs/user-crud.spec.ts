import { test, expect } from '../fixtures/test';
import { CREDENTIALS } from '../utils/constants';

test.beforeEach(async ({ page }) => {
  const resp = await page.request.post('http://localhost:18080/api/v1/auth/login', {
    data: { username: CREDENTIALS.admin.username, password: CREDENTIALS.admin.password },
  });
  const body = await resp.json();
  await page.goto('/login');
  await page.evaluate(
    ({ t, r }) => { localStorage.setItem('kingfisher_token', t); localStorage.setItem('kingfisher_refresh', r); },
    { t: body.data.access_token, r: body.data.refresh_token },
  );
});

test('用户列表展示列', async ({ page }) => {
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table')).toBeVisible();
  await expect(page.locator('.ant-table-thead')).toContainText('用户');
  await expect(page.locator('.ant-table-thead')).toContainText('邮箱');
});

test('搜索过滤 admin', async ({ page }) => {
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await page.getByPlaceholder('搜索').fill('admin');
  await page.getByRole('button', { name: /搜\s*索/ }).click();
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table-row').first()).toContainText('admin');
});

test('搜索无结果', async ({ page }) => {
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await page.getByPlaceholder('搜索').fill('zzz_nosuchuser');
  await page.getByRole('button', { name: /搜\s*索/ }).click();
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-empty')).toBeVisible({ timeout: 5000 });
});

test('新增用户 - 表单校验', async ({ page }) => {
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button', { name: '新增用户' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
  await expect(page.locator('.ant-modal')).toContainText(/请/);
});

test('新增用户 - 成功创建', async ({ page }) => {
  const newUser = `e2e_${Date.now()}`;
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button', { name: '新增用户' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  // 等 Modal 打开动画 + afterOpenChange(resetFields) 完成，避免后续 fill 被清空
  await page.waitForTimeout(400);
  await page.locator('#username').fill(newUser);
  await page.locator('#password').fill('Abcd1234');
  await page.locator('#email').fill(`${newUser}@test.com`);
  // 选择角色（多选，必填）——选"访客"，然后关闭下拉避免遮挡确定按钮
  await page.locator('#role_ids').click();
  await page.locator('.ant-select-dropdown').getByText('访客', { exact: true }).click();
  await page.keyboard.press('Escape');
  await page.waitForTimeout(300);
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
  await expect(page.locator('.ant-modal')).not.toBeVisible({ timeout: 10000 });
  await page.getByPlaceholder('搜索').fill(newUser);
  await page.getByRole('button', { name: /搜\s*索/ }).click();
  await page.waitForLoadState('networkidle');
  await expect(page.locator('.ant-table')).toContainText(newUser);
});

test('编辑用户 - 用户名 disabled', async ({ page }) => {
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await page.getByPlaceholder('搜索').fill('editor');
  await page.getByRole('button', { name: /搜\s*索/ }).click();
  await page.waitForLoadState('networkidle');
  await page.locator('tr', { hasText: 'editor' }).locator('a', { hasText: '编辑' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  // 等 Modal 打开动画 + afterOpenChange(setFieldsValue) 完成
  await page.waitForTimeout(400);
  await expect(page.locator('#username')).toBeDisabled();
  await page.locator('#email').clear();
  await page.locator('#email').fill('editor2@test.com');
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
  await expect(page.locator('.ant-modal')).not.toBeVisible({ timeout: 10000 });
});

test('删除用户', async ({ page }) => {
  const newUser = `e2edel_${Date.now()}`;
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  // Create
  await page.getByRole('button', { name: '新增用户' }).click();
  // 等 Modal 打开动画 + afterOpenChange(resetFields) 完成，避免后续 fill 被清空
  await page.waitForTimeout(400);
  await page.locator('#username').fill(newUser);
  await page.locator('#password').fill('Abcd1234');
  // 选择角色（多选，必填）——选"访客"，然后关闭下拉避免遮挡确定按钮
  await page.locator('#role_ids').click();
  await page.locator('.ant-select-dropdown').getByText('访客', { exact: true }).click();
  await page.keyboard.press('Escape');
  await page.waitForTimeout(300);
  await page.locator('.ant-modal').getByRole('button', { name: /确\s*定|保存/ }).click();
  await expect(page.locator('.ant-modal')).not.toBeVisible({ timeout: 10000 });
  // Search
  await page.getByPlaceholder('搜索').fill(newUser);
  await page.getByRole('button', { name: /搜\s*索/ }).click();
  await page.waitForLoadState('networkidle');
  // Delete
  await page.locator('.ant-spin').waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
  await page.locator('tr', { hasText: newUser }).getByText('删除').click();
  await page.locator('.ant-popconfirm').getByRole('button', { name: '确 定' }).click();
  await expect(page.locator('.ant-table')).not.toContainText(newUser, { timeout: 10000 });
});
