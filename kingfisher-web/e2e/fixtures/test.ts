import { test as base, type Page } from '@playwright/test';

/** Extended test that collects console errors/warnings and fails the test if any are found. */
export const test = base.extend<{ page: Page }>({
  // @ts-expect-error oxlint false positive: `use` is Playwright fixture API, not a React hook
  page: async ({ page }, _use, testInfo) => {
    // 每个页面加载前清掉「记住账户」的已存账户，
    // 避免登录页进入"选择账户"视图而找不到登录表单（导致 submitButton 超时）
    await page.addInitScript(() => localStorage.removeItem('kingfisher_accounts'));
    const issues: string[] = [];
    const handler = (msg: { type: () => string; text: () => string }) => {
      const text = msg.text();
      // 忽略「Failed to load resource: 4xx」—— 这是业务失败（如错误密码 400、限流 429），
      // 测试经常故意触发，不应当作 console 错误。只拦截真正的 JS 异常/未捕获错误。
      if (msg.type() === 'error' && /Failed to load resource/.test(text)) {
        return;
      }
      if (msg.type() === 'error' || msg.type() === 'warning') {
        issues.push(`[${msg.type()}] ${text}`);
      }
    };
    page.on('console', handler);

    await _use(page);

    if (issues.length > 0) {
      const lines = issues.slice(0, 20).join('\n');
      testInfo.errors.push({
        message: `Console errors/warnings in "${testInfo.title}":\n${lines}${issues.length > 20 ? `\n... and ${issues.length - 20} more` : ''}`,
      });
      // Throw to fail the test
      throw new Error(`Console errors/warnings in "${testInfo.title}":\n${lines}${issues.length > 20 ? `\n... and ${issues.length - 20} more` : ''}`);
    }
  },
});

export { expect } from '@playwright/test';
