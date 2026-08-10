const ACCOUNTS_KEY = 'kingfisher_accounts';
const PWD_REMEMBER_KEY = 'kingfisher_remember_pwd';

export interface SavedAccount {
  username: string;
  password: string;
  lastLogin: number;
}

function encodePwd(plain: string): string {
  try { return btoa(plain); } catch { return ''; }
}
function decodePwd(encoded: string): string {
  try { return atob(encoded); } catch { return ''; }
}

export function loadAccounts(): SavedAccount[] {
  try {
    const raw = localStorage.getItem(ACCOUNTS_KEY);
    return raw ? (JSON.parse(raw) as SavedAccount[]) : [];
  } catch { return []; }
}

export function saveAccount(username: string, password: string, rememberAccount: boolean, rememberPwd: boolean): void {
  if (!rememberAccount) return;
  const accounts = loadAccounts().filter(a => a.username !== username);
  accounts.push({
    username,
    password: rememberPwd ? encodePwd(password) : '',
    lastLogin: Date.now(),
  });
  accounts.sort((a, b) => b.lastLogin - a.lastLogin);
  localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(accounts.slice(0, 10)));
}

export function getAccountPassword(username: string): string {
  const a = loadAccounts().find(a => a.username === username);
  return a ? decodePwd(a.password) : '';
}

export function removeAccount(username: string): void {
  const accounts = loadAccounts().filter(a => a.username !== username);
  localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(accounts));
}

// 记住密码勾选状态持久化：用户选择一次后持续生效，
// 避免重新登录时状态重置导致已保存的密码被覆盖清空。
export function loadRememberPwd(): boolean {
  try {
    return localStorage.getItem(PWD_REMEMBER_KEY) === '1';
  } catch { return false; }
}

export function saveRememberPwd(v: boolean): void {
  try {
    localStorage.setItem(PWD_REMEMBER_KEY, v ? '1' : '0');
  } catch { /* ignore */ }
}
