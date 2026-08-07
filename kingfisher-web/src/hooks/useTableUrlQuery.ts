import { useCallback, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { FormInstance } from 'antd';

/**
 * 把表格的搜索/分页状态同步到 URL query 参数，并提供数据重载能力。
 *
 * - 搜索提交/重置、翻页时自动写入 URL，刷新或分享链接可保持相同视图。
 * - reload() 触发一次重新加载（配合 DataTable 的 refreshKey）。
 */
export function useTableUrlQuery() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [refreshKey, setRefreshKey] = useState(0);

  /** URL query 全部参数（字符串形式）。 */
  const urlParams = useMemo(() => {
    const p: Record<string, string> = {};
    searchParams.forEach((v, k) => {
      p[k] = v;
    });
    return p;
  }, [searchParams]);

  const page = Number(urlParams.page) || 1;
  const pageSize = Number(urlParams.page_size) || 20;

  /** 触发重新加载。 */
  const reload = useCallback(() => setRefreshKey((k) => k + 1), []);

  /** 用当前 URL 反填搜索表单（不含分页参数）。 */
  const syncFormFromUrl = useCallback(
    (form?: FormInstance) => {
      const { page: _p, page_size: _ps, ...search } = urlParams;
      if (Object.keys(search).length > 0) {
        form?.setFieldsValue(search);
      }
    },
    [urlParams],
  );

  /** 合并写入一组 key 到 URL；空值/undefined 会被移除。 */
  const updateUrl = useCallback(
    (values: Record<string, unknown>) => {
      const next = new URLSearchParams(searchParams);
      Object.entries(values).forEach(([k, v]) => {
        if (v === undefined || v === null || v === '') {
          next.delete(k);
        } else {
          next.set(k, String(v));
        }
      });
      setSearchParams(next, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  const onSearch = useCallback(
    (values: Record<string, unknown>) => {
      updateUrl({ ...values, page: 1 });
    },
    [updateUrl],
  );

  const onReset = useCallback(() => {
    // 清空所有查询参数（回到第一页）
    setSearchParams(new URLSearchParams(), { replace: true });
  }, [setSearchParams]);

  const onPageChange = useCallback(
    (p: number, ps: number, form?: FormInstance) => {
      // 保留当前搜索条件，只更新分页
      const current = form?.getFieldsValue() || {};
      updateUrl({ ...current, page: p, page_size: ps });
    },
    [updateUrl],
  );

  return {
    urlParams,
    page,
    pageSize,
    refreshKey,
    reload,
    syncFormFromUrl,
    updateUrl,
    onSearch,
    onReset,
    onPageChange,
  };
}
