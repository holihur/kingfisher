/** Menu utilities: flatten, find breadcrumb chain, etc. */

export interface MenuItem {
  id: number;
  name: string;
  path?: string;
  icon?: string;
  parent_id?: number;
  type?: number;
  status?: number;
  sort?: number;
  children?: MenuItem[];
}

/** Flatten a nested menu tree into a flat array. */
export function flatMenus(tree: MenuItem[]): MenuItem[] {
  const result: MenuItem[] = [];
  function walk(items: MenuItem[]) {
    for (const item of items) {
      result.push(item);
      if (item.children?.length) walk(item.children);
    }
  }
  walk(tree);
  return result;
}

/**
 * 匹配 pathname 对应的菜单路径：优先精确匹配，否则取最长前缀匹配
 * （如 /agent/123 → /agent，用于会话级子路由高亮父菜单）。
 */
export function matchMenuPath(tree: MenuItem[], pathname: string): string | null {
  let best: string | null = null;
  for (const m of flatMenus(tree)) {
    if (!m.path) continue;
    if (m.path === pathname) return m.path;
    if (pathname.startsWith(m.path + '/') && (best === null || m.path.length > best.length)) {
      best = m.path;
    }
  }
  return best;
}

/** Find the breadcrumb chain from root to the node matching the given path. */
export function findBreadcrumb(tree: MenuItem[], targetPath: string): MenuItem[] {
  // 先解析出匹配的菜单路径（含前缀匹配），再精确查找其面包屑链。
  const matched = matchMenuPath(tree, targetPath) || targetPath;
  function find(items: MenuItem[], chain: MenuItem[]): MenuItem[] | null {
    for (const item of items) {
      const next = [...chain, item];
      if (item.path === matched) return next;
      if (item.children?.length) {
        const found = find(item.children, next);
        if (found) return found;
      }
    }
    return null;
  }
  return find(tree, []) || [];
}
