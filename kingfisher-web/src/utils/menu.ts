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

/** Find the breadcrumb chain from root to the node matching the given path. */
export function findBreadcrumb(tree: MenuItem[], targetPath: string): MenuItem[] {
  function find(items: MenuItem[], chain: MenuItem[]): MenuItem[] | null {
    for (const item of items) {
      const next = [...chain, item];
      if (item.path === targetPath) return next;
      if (item.children?.length) {
        const found = find(item.children, next);
        if (found) return found;
      }
    }
    return null;
  }
  return find(tree, []) || [];
}
