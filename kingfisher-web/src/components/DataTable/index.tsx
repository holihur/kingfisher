import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Button, Card, Empty, Form, Input, Select, Space, Table } from 'antd';
import { DownOutlined, UpOutlined } from '@ant-design/icons';
import type { TableProps } from 'antd';
import { useSearchParams } from 'react-router-dom';
import { useTableUrlQuery } from '../../hooks/useTableUrlQuery';
import { buildQueryParams } from '../../utils/query';

export type ColumnsType<T> = TableProps<T>['columns'];

export interface SearchField {
  name: string;
  label: string;
  type: 'text' | 'select';
  options?: { label: string; value: string | number }[];
}

interface DataTableProps<T = Record<string, unknown>> {
  columns: ColumnsType<T>;
  rowKey: string | ((r: T) => React.Key);
  /** 数据加载函数；不传则使用静态 dataSource */
  request?: (params: Record<string, unknown>) => Promise<{ items: T[]; total: number }>;
  /** 静态数据模式（无 request 时） */
  dataSource?: T[];
  /** 搜索字段；不传或空数组则不渲染搜索表单 */
  searchFields?: SearchField[];
  headerTitle?: React.ReactNode;
  /** 标题旁副标题 */
  headerSubtitle?: React.ReactNode;
  toolBarRender?: React.ReactNode;
  /** 额外固定查询参数（如字典 type_id） */
  tableParams?: Record<string, unknown>;
  /** 外部触发刷新：增删改后递增此值 */
  reloadKey?: number;
}

/**
 * 通用数据表格：搜索表单 + 分页 + URL 同步 + 重载。
 * 统一展示风格：骨架屏加载、友好空态、斑马纹、标题栏。
 */
export default function DataTable<T = Record<string, unknown>>({
  columns,
  rowKey,
  request,
  dataSource,
  searchFields = [],
  headerTitle,
  headerSubtitle,
  toolBarRender,
  tableParams,
  reloadKey = 0,
}: DataTableProps<T>) {
  const { urlParams, page, pageSize, syncFormFromUrl, onPageChange } = useTableUrlQuery();
  const [, setSearchParams] = useSearchParams();
  const [form] = Form.useForm();
  const [data, setData] = useState<T[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);

  const isStatic = !request;
  const hasSearch = searchFields.length > 0;
  // 默认只显示首字段（关键词），其余为高级搜索字段
  const baseFields = searchFields.slice(0, 1);
  const advancedFields = searchFields.slice(1);
  const hasAdvanced = advancedFields.length > 0;

  // 用 ref 持有 request，避免内联 request 每次渲染新引用导致 useEffect 无限 refetch
  const requestRef = useRef(request);
  requestRef.current = request;

  // 挂载时用 URL 反填搜索表单
  useEffect(() => {
    if (hasSearch) syncFormFromUrl(form);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 加载数据（依赖 URL 参数、刷新计数、额外参数）
  useEffect(() => {
    if (!requestRef.current) return;
    let cancelled = false;
    setLoading(true);
    // URL 参数去分页键后作为筛选条件，与分页合并
    const { page: _p, page_size: _ps, ...search } = urlParams;
    const loadParams = { current: page, pageSize, ...search, ...tableParams };
    requestRef.current(buildQueryParams(loadParams))
      .then((r) => {
        if (cancelled) return;
        setData(r.items);
        setTotal(r.total);
      })
      .catch(() => {
        if (cancelled) return;
        setData([]);
        setTotal(0);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, urlParams, tableParams, reloadKey]);

  const handleSearch = useCallback(
    (values: Record<string, unknown>) => {
      // 搜索时 URL 变化会触发 useEffect 重新加载
      const next = new URLSearchParams();
      Object.entries({ ...values, page: 1, page_size: pageSize }).forEach(([k, v]) => {
        if (v === undefined || v === null) return;
        const s = String(v);
        if (s !== '') next.set(k, s);
      });
      setSearchParams(next, { replace: true });
    },
    [pageSize, setSearchParams],
  );

  const handleReset = useCallback(() => {
    form.resetFields();
    setSearchParams(new URLSearchParams({ page: '1', page_size: String(pageSize) }), { replace: true });
  }, [form, pageSize, setSearchParams]);

  const pagination = useMemo(
    () =>
      isStatic
        ? false
        : {
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (t: number) => `共 ${t} 条`,
            onChange: (p: number, ps: number) => onPageChange(p, ps, form),
          },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [isStatic, page, pageSize, total, form],
  );

  return (
    <Card
      title={
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
          {/* 左侧：标题 + 副标题 */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <span style={{ fontSize: 16, fontWeight: 600 }}>{headerTitle}</span>
            {headerSubtitle ? (
              <span style={{ color: '#8c8c8c', fontSize: 13 }}>{headerSubtitle}</span>
            ) : null}
          </div>
          {/* 右侧：关键词搜索 + 搜索/重置/高级 + 操作按钮 */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            {hasSearch && (
              <Form form={form} layout="inline" onFinish={handleSearch} style={{ rowGap: 8 }}>
                {baseFields.map((f) => (
                  <Form.Item key={f.name} name={f.name} label={f.label} style={{ marginBottom: 0 }}>
                    {f.type === 'select' ? (
                      <Select style={{ width: 150 }} placeholder="请选择" allowClear options={f.options} />
                    ) : (
                      <Input placeholder="搜索" allowClear />
                    )}
                  </Form.Item>
                ))}
                <Form.Item style={{ marginBottom: 0 }}>
                  <Space>
                    <Button type="primary" htmlType="submit">
                      搜索
                    </Button>
                    <Button onClick={handleReset}>重置</Button>
                    {hasAdvanced && (
                      <Button
                        type="link"
                        size="small"
                        icon={advancedOpen ? <UpOutlined /> : <DownOutlined />}
                        onClick={() => setAdvancedOpen((o) => !o)}
                      >
                        {advancedOpen ? '收起' : '高级'}
                      </Button>
                    )}
                  </Space>
                </Form.Item>
              </Form>
            )}
            {toolBarRender}
          </div>
        </div>
      }
      styles={{ body: { paddingTop: 16 } }}
    >
      {/* 高级搜索区：点高级后在 head 与 body 之间展开高级字段 */}
      {hasSearch && advancedOpen && advancedFields.length > 0 && (
        <div style={{ padding: '12px 24px', borderTop: '1px solid #f0f0f0', background: '#fafafa' }}>
          <Form form={form} layout="inline" onFinish={handleSearch} style={{ rowGap: 8, flexWrap: 'wrap' }}>
            {advancedFields.map((f) => (
              <Form.Item key={f.name} name={f.name} label={f.label} style={{ marginBottom: 0 }}>
                {f.type === 'select' ? (
                  <Select style={{ width: 150 }} placeholder="请选择" allowClear options={f.options} />
                ) : (
                  <Input placeholder="搜索" allowClear />
                )}
              </Form.Item>
            ))}
          </Form>
        </div>
      )}
      <Table<T>
        columns={columns}
        rowKey={rowKey}
        dataSource={isStatic ? dataSource : data}
        loading={loading}
        pagination={pagination}
        scroll={{ x: 'max-content' }}
        size="middle"
        rowClassName={() => 'data-table-row'}
        locale={{
          emptyText: (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={hasSearch ? '暂无匹配数据，可尝试调整搜索条件' : '暂无数据'}
            />
          ),
        }}
      />
    </Card>
  );
}
