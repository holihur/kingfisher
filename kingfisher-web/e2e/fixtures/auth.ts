import { type Page, type BrowserContext, request } from '@playwright/test';
import { CREDENTIALS } from '../utils/constants';

/**
 * Authenticate by calling the login API, then inject tokens via
 * addInitScript so localStorage is populated BEFORE any page navigation.
 * This avoids the redirect-to-/login race condition.
 */
async function getTokens(role: 'admin' | 'editor' | 'viewer') {
  const creds = CREDENTIALS[role];
  const apiContext = await request.newContext({ baseURL: 'http://localhost:18080' });
  const resp = await apiContext.post('/api/v1/auth/login', {
    data: { username: creds.username, password: creds.password },
  });
  const body = await resp.json();
  await apiContext.dispose();

  if (body.code !== 0) {
    throw new Error(`Login API failed for ${role}: ${JSON.stringify(body)}`);
  }
  return {
    token: body.data.access_token as string,
    refresh: body.data.refresh_token as string,
  };
}

/**
 * Create a new page that is already authenticated as the given role.
 * Uses addInitScript to inject tokens into localStorage before any navigation.
 */
export async function newAuthenticatedPage(
  context: BrowserContext,
  role: 'admin' | 'editor' | 'viewer',
): Promise<Page> {
  const { token, refresh } = await getTokens(role);

  // Inject BEFORE any page loads — otherwise AuthGuard redirects to /login
  await context.addInitScript(
    (tokens) => {
      localStorage.setItem('kingfisher_token', tokens.token);
      localStorage.setItem('kingfisher_refresh', tokens.refresh);
    },
    { token, refresh },
  );

  return context.newPage();
}

/**
 * UI-based login. Navigates to /login, fills the form, submits.
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
 * Logout via UI — click header dropdown → 退出登录.
 */
export async function logoutViaUI(page: Page): Promise<void> {
  // Open user dropdown in header — click on the avatar + username area
  await page.locator('.ant-layout-header .ant-dropdown-trigger').click();
  await page.getByText('退出登录').click();
  await page.waitForURL('**/login');
}
