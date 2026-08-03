import { useAuthStore } from '../stores/auth';

/** Check if the current user has a specific permission code. */
export function hasPerm(code: string): boolean {
  return useAuthStore.getState().permissions.includes(code);
}

/** Check if the current user has any of the given permission codes. */
export function hasAnyPerm(...codes: string[]): boolean {
  const perms = useAuthStore.getState().permissions;
  return codes.some((c) => perms.includes(c));
}
