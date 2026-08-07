import { test as base, type Page } from '@playwright/test';

/** Extended test that collects console errors/warnings and fails the test if any are found. */
export const test = base.extend<{ page: Page }>({
  page: async ({ page }, use, testInfo) => {
    const issues: string[] = [];
    const handler = (msg: { type: () => string; text: () => string }) => {
      if (msg.type() === 'error' || msg.type() === 'warning') {
        issues.push(`[${msg.type()}] ${msg.text()}`);
      }
    };
    page.on('console', handler);

    await use(page);

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
