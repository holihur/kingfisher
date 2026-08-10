// 账户记忆：仅记住用户名，不保存任何密码（安全）。
// 历史版本曾用 base64 保存密码（kingfisher_remember_pwd），现已移除；
// 登录时会清理 localStorage 中残留的密码数据。

const ACCOUNTS_KEY = 'kingfisher_accounts';
const PWD_REMEMBER_KEY = 'kingfisher_remember_pwd';

export interface SavedAccount {
  username: string;
  password: string; // 恒为空字符串，不保存密码
  lastLogin: number;
}

export function loadAccounts(): SavedAccount[] {
  try {
    const raw = localStorage.getItem(ACCOUNTS_KEY);
    return raw ? (JSON.parse(raw) as SavedAccount[]) : [];
  } catch { return []; }
}

export function saveAccount(username: string): void {
  const accounts = loadAccounts().filter(a => a.username !== username);
  accounts.push({
    username,
    password: '', // 永不保存密码
    lastLogin: Date.now(),
  });
  accounts.sort((a, b) => b.lastLogin - a.lastLogin);
  localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(accounts.slice(0, 10)));
}

export function getAccountPassword(_username: string): string {
  return ''; // 不再保存/读取密码
}

export function removeAccount(username: string): void {
  const accounts = loadAccounts().filter(a => a.username !== username);
  localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(accounts));
}

// 清理历史残留的密码数据（旧的 base64 密码 + 记住密码标记），安全原因。
export function purgeStoredPasswords(): void {
  try {
    const accounts = loadAccounts().map(a => ({ ...a, password: '' }));
    localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(accounts));
    localStorage.removeItem(PWD_REMEMBER_KEY);
  } catch { /* ignore */ }
}
