import React from 'react';
import { useAuthStore } from '../stores/auth';

const PermissionBtn: React.FC<{ code: string; children: React.ReactNode }> = ({ code, children }) => {
  const perms = useAuthStore((s) => s.permissions);
  return perms.includes(code) ? <>{children}</> : null;
};

export default PermissionBtn;
