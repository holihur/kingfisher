import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './specs',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
  ],
  timeout: 30000,
  expect: { timeout: 10000 },

  use: {
    baseURL: 'http://localhost:15173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chrome',
      use: { ...devices['Desktop Chrome'], channel: 'chrome' },
    },
  ],

  webServer: [
    {
      command: 'bash scripts/e2e-server.sh',
      port: 18080,
      timeout: 60000,
      reuseExistingServer: !process.env.CI,
      cwd: '../../',
      env: { CONFIG_PATH: 'config/e2e.yaml' },
    },
    {
      command: 'npx vite --port 15173 --strictPort',
      port: 15173,
      timeout: 30000,
      reuseExistingServer: !process.env.CI,
      cwd: '../',
      env: { VITE_API_TARGET: 'http://localhost:18080' },
    },
  ],
});
