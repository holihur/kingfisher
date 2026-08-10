import { test, expect } from '../fixtures/test';
import { CREDENTIALS, URLS } from '../utils/constants';

test('未登录访问 /dashboard → 跳转 /login', async ({ page }) => {
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/\/login/);
});

test('未登录访问 /system/users → /login 带 redirect', async ({ page }) => {
  await page.goto(URLS.users);
  await expect(page).toHaveURL(/\/login/);
  await expect(page).toHaveURL(/redirect=/);
});

test('redirect 参数：登录后跳回', async ({ page }) => {
  await page.evaluate(() => localStorage.clear());
  await page.goto(URLS.users);
  await page.waitForURL(/\/login/);
  // Login via API
  const resp = await page.request.post('http://localhost:18080/api/v1/auth/login', {
    data: { username: CREDENTIALS.admin.username, password: CREDENTIALS.admin.password },
  });
  const body = await resp.json();
  await page.evaluate(
    ({ t, r }) => { localStorage.setItem('kingfisher_token', t); localStorage.setItem('kingfisher_refresh', r); },
    { t: body.data.access_token, r: body.data.refresh_token },
  );
  // Navigate to the redirect target
  const url = new URL(page.url());
  const redirect = url.searchParams.get('redirect') || '/dashboard';
  await page.goto(redirect);
  expect(page.url()).toContain(URLS.users);
});

test('退出后访问 protected → /login', async ({ page }) => {
  const resp = await page.request.post('http://localhost:18080/api/v1/auth/login', {
    data: { username: CREDENTIALS.admin.username, password: CREDENTIALS.admin.password },
  });
  const body = await resp.json();
  await page.goto('/login');
  await page.evaluate(
    ({ t, r }) => { localStorage.setItem('kingfisher_token', t); localStorage.setItem('kingfisher_refresh', r); },
    { t: body.data.access_token, r: body.data.refresh_token },
  );
  await page.goto('/dashboard');
  await page.locator('.ant-layout-header').locator('[class*="dropdown"]').first().click();
  await page.getByText('退出登录').click();
  await page.waitForURL('**/login');
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/\/login/);
});
