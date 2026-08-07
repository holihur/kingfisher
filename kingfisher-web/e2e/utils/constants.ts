// Shared test data constants for E2E tests

export const CREDENTIALS = {
  admin:  { username: 'admin',  password: 'Abcd1234' },
  editor: { username: 'editor', password: 'Abcd1234' },
  viewer: { username: 'viewer', password: 'Abcd1234' },
} as const;

export const URLS = {
  login:     '/login',
  dashboard: '/dashboard',
  users:     '/system/users',
  menus:     '/system/menus',
  roles:     '/system/roles',
  configs:   '/system/configs',
  audit:     '/system/audit',
  dicts:     '/system/dicts',
  profile:   '/profile',
} as const;

export const ROLE_NAMES: Record<string, string> = {
  admin:  '超级管理员',
  editor: '编辑',
  viewer: '访客',
};
