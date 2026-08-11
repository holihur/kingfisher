import { test, expect } from '../fixtures/test';
import { CREDENTIALS } from '../utils/constants';

async function loginAdmin(page) {
  const resp = await page.request.post('http://localhost:18080/api/v1/auth/login', {
    data: { username: CREDENTIALS.admin.username, password: CREDENTIALS.admin.password },
  });
  const body = await resp.json();
  await page.goto('/login');
  await page.evaluate(
    ({ t, r }) => { localStorage.setItem('kingfisher_token', t); localStorage.setItem('kingfisher_refresh', r); },
    { t: body.data.access_token, r: body.data.refresh_token },
  );
}

test('ESC 关闭弹窗', async ({ page }) => {
  await loginAdmin(page);
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button', { name: '新增用户' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.locator('.ant-modal')).not.toBeVisible();
});

test('遮罩点击关闭弹窗', async ({ page }) => {
  await loginAdmin(page);
  await page.goto('/system/users');
  await page.waitForLoadState('networkidle');
  await page.getByRole('button', { name: '新增用户' }).click();
  await expect(page.locator('.ant-modal')).toBeVisible();
  await page.waitForTimeout(600); // 等 antd modal fade 动画结束，mask 才可点击
  await page.locator('.ant-modal-mask').click({ position: { x: 10, y: 10 }, force: true });
  await expect(page.locator('.ant-modal')).not.toBeVisible();
});

test('Mobile viewport 不溢出', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 });
  await loginAdmin(page);
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
  expect(bodyWidth).toBeLessThanOrEqual(375);
});

test('页面刷新保持登录', async ({ page }) => {
  await loginAdmin(page);
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');
  await page.reload();
  await page.waitForLoadState('networkidle');
  expect(page.url()).toContain('/dashboard');
});
