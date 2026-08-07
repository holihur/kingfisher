/**
 * 通用展示格式化工具。
 */

/** 时间 → 'YYYY/MM/DD HH:mm'；空值 → '-' */
export function formatTime(v: unknown): string {
  if (!v) return '-';
  const d = new Date(v as string);
  if (Number.isNaN(d.getTime())) return String(v);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** 相对时间：刚刚 / N 分钟前 / N 小时前 / N 天前 / 具体时间 */
export function formatRelativeTime(v: unknown): string {
  if (!v) return '-';
  const d = new Date(v as string);
  if (Number.isNaN(d.getTime())) return String(v);
  const diff = Date.now() - d.getTime();
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;
  if (diff < minute) return '刚刚';
  if (diff < hour) return `${Math.floor(diff / minute)} 分钟前`;
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`;
  if (diff < 7 * day) return `${Math.floor(diff / day)} 天前`;
  return formatTime(v);
}

/** 空值兜底 '-' */
export function orDash(v: unknown): string {
  return v === undefined || v === null || v === '' ? '-' : String(v);
}
