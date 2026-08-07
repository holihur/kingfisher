import { type Page, type BrowserContext, request } from '@playwright/test';
import { CREDENTIALS } from '../utils/constants';

/**
 * Get valid tokens by calling the login API.
 */
async function getTokens(role: 'admin' | 'editor' | 'viewer') {
  const creds = CREDENTIALS[role];
  const ctx = await request.newContext({ baseURL: 'http://localhost:18080' });
  const resp = await ctx.post('/api/v1/auth/login', {
    data: { username: creds.username, password: creds.password },
  });
  const body = await resp.json();
  await ctx.dispose();
  if (body.code !== 0) {
    throw new Error(`Login failed for ${role}: ${JSON.stringify(body)}`);
  }
  return {
    token: body.data.access_token as string,
    refresh: body.data.refresh_token as string,
  };
}

/**
 * Create an authenticated page. Navigates to /, sets localStorage, then navigates to dashboard.
 * Simpler than addInitScript — no race condition, trivially testable.
 */
export async function newAuthenticatedPage(
  context: BrowserContext,
  role: 'admin' | 'editor' | 'viewer',
): Promise<Page> {
  const { token, refresh } = await getTokens(role);
  const page = await context.newPage();

  // Navigate to login to establish origin
  await page.goto('/login');
  // Inject tokens directly
  await page.evaluate(
    ({ t, r }) => {
      localStorage.setItem('kingfisher_token', t);
      localStorage.setItem('kingfisher_refresh', r);
    },
    { t: token, r: refresh },
  );

  return page;
}

/**
 * UI-based login.
 */
export async function loginViaUI(page: Page, role: 'admin' | 'editor' | 'viewer'): Promise<void> {
  const creds = CREDENTIALS[role];
  await page.goto('/login');
  await page.getByPlaceholder('用户名').fill(creds.username);
  await page.getByPlaceholder('密码').fill(creds.password);
  await page.getByRole('button', { name: '登 录' }).click();
  await page.waitForURL('**/dashboard');
}

/**
 * Logout via UI.
 */
export async function logoutViaUI(page: Page): Promise<void> {
  await page.locator('.ant-layout-header').locator('[class*="dropdown"]').first().click();
  await page.getByText('退出登录').click();
  await page.waitForURL('**/login');
}
