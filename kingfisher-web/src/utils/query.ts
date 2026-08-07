/**
 * 把 ProTable request 收到的 params（表单值 + 分页 + 额外参数）
 * 转成后端列表查询 DSL 的参数：
 *   - page / page_size
 *   - q        关键词（模糊搜索可搜索字段）
 *   - filter   JSON 结构化过滤（{字段: 值}，值为裸值 = 等值）
 */
export function buildQueryParams(p: Record<string, unknown>): Record<string, unknown> {
  const params: Record<string, unknown> = {
    page: (p.current as number) || 1,
    page_size: (p.pageSize as number) || 20,
  };

  const q = (p.q as string) || (p.keyword as string) || '';
  if (q) params.q = q;

  const filter: Record<string, unknown> = {};
  const skip = new Set(['current', 'pageSize', 'q', 'keyword', 'page', 'page_size', 'sort']);
  Object.entries(p).forEach(([k, v]) => {
    if (skip.has(k)) return;
    if (v === undefined || v === null || v === '') return;
    filter[k] = v;
  });
  if (Object.keys(filter).length > 0) {
    params.filter = JSON.stringify(filter);
  }

  return params;
}
