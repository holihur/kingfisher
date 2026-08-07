import { useCallback, useMemo, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { ActionType } from '@ant-design/pro-table';
import type { ProFormInstance } from '@ant-design/pro-form';

/**
 * 把表格的搜索/分页状态同步到 URL query 参数。
 *
 * - 搜索提交/重置、翻页时自动写入 URL，刷新或分享链接可保持相同视图。
 * - 挂载时用 URL 反填搜索表单（syncFormFromUrl）。
 *
 * 用法见各列表页：ProTable 传 params={urlParams}、pagination 由 page/pageSize 受控，
 * search 的 onSearch/onReset 调 updateUrl。
 */
export function useTableUrlQuery() {
  const [searchParams, setSearchParams] = useSearchParams();
  const actionRef = useRef<ActionType>(null);
  const formRef = useRef<ProFormInstance | undefined>(undefined);

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

  /** 用当前 URL 反填搜索表单（不含分页参数）。 */
  const syncFormFromUrl = useCallback(() => {
    const { page: _p, page_size: _ps, ...search } = urlParams;
    if (Object.keys(search).length > 0) {
      formRef.current?.setFieldsValue(search);
    }
  }, [urlParams]);

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
    (p: number, ps: number) => {
      // 保留当前搜索条件，只更新分页
      const current = formRef.current?.getFieldsValue?.() || {};
      updateUrl({ ...current, page: p, page_size: ps });
    },
    [updateUrl],
  );

  return {
    urlParams,
    page,
    pageSize,
    actionRef,
    formRef,
    syncFormFromUrl,
    updateUrl,
    onSearch,
    onReset,
    onPageChange,
  };
}
