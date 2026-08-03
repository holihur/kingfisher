/** Centralized token access. Single source of truth for auth storage keys. */

const KEYS = {
  token: 'kingfisher_token',
  refresh: 'kingfisher_refresh',
} as const;

export function getToken(): string | null {
  return localStorage.getItem(KEYS.token);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(KEYS.refresh);
}

export function setTokens(access: string, refresh: string): void {
  localStorage.setItem(KEYS.token, access);
  localStorage.setItem(KEYS.refresh, refresh);
}

export function clearTokens(): void {
  localStorage.removeItem(KEYS.token);
  localStorage.removeItem(KEYS.refresh);
}

export function hasToken(): boolean {
  return !!localStorage.getItem(KEYS.token);
}
