import React, { useState } from 'react';
import { Tag, Modal, Descriptions, Typography } from 'antd';
import DataTable, { SearchField } from '../../components/DataTable';
import { useThemeToken } from '../../hooks/useThemeToken';
import { auditApi } from '../../api/audit';
import { formatTime } from '../../utils/format';

const ACTION_VALUES: Record<string, { text: string; color?: string }> = {
  create: { text: '创建', color: 'green' },
  update: { text: '更新', color: 'blue' },
  delete: { text: '删除', color: 'red' },
  login: { text: '登录', color: 'cyan' },
  logout: { text: '退出', color: 'default' },
  register: { text: '注册', color: 'purple' },
};

interface AuditLogRow {
  id: number;
  username: string;
  action: string;
  resource: string;
  resource_id: number;
  detail: string;
  result: string;
  latency: number;
  message: string;
  ip: string;
  user_agent: string;
  created_at: string;
}

const searchFields: SearchField[] = [
  { name: 'q', label: '关键词', type: 'text' },
  {
    name: 'action',
    label: '操作',
    type: 'select',
    options: [
      { label: '创建', value: 'create' },
      { label: '更新', value: 'update' },
      { label: '删除', value: 'delete' },
      { label: '登录', value: 'login' },
      { label: '退出', value: 'logout' },
      { label: '注册', value: 'register' },
    ],
  },
  {
    name: 'resource',
    label: '资源',
    type: 'select',
    options: [
      { label: '用户', value: '用户' },
      { label: '角色', value: '角色' },
      { label: '菜单', value: '菜单' },
      { label: '系统配置', value: '系统配置' },
      { label: '字典类型', value: '字典类型' },
      { label: '字典条目', value: '字典条目' },
      { label: '站内信', value: '站内信' },
      { label: '消息模板', value: '消息模板' },
      { label: '周期任务', value: '周期任务' },
    ],
  },
  {
    name: 'result',
    label: '结果',
    type: 'select',
    options: [
      { label: '成功', value: 'success' },
      { label: '失败', value: 'failure' },
    ],
  },
];

const AuditLogList: React.FC = () => {
  const token = useThemeToken();
  const [detail, setDetail] = useState<AuditLogRow | null>(null);

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '用户', dataIndex: 'username', width: 120 },
    {
      title: '操作',
      dataIndex: 'action',
      width: 90,
      render: (_: unknown, r: AuditLogRow) => {
        const v = ACTION_VALUES[r.action];
        return v ? <Tag color={v.color}>{v.text}</Tag> : <Tag>{r.action}</Tag>;
      },
    },
    { title: '资源', dataIndex: 'resource', width: 110, render: (_: unknown, r: AuditLogRow) => <Tag>{r.resource}</Tag> },
    { title: '资源ID', dataIndex: 'resource_id', width: 80 },
    {
      title: '结果',
      dataIndex: 'result',
      width: 80,
      render: (_: unknown, r: AuditLogRow) =>
        r.result === 'success' ? <Tag color="green">成功</Tag> : <Tag color="red">失败</Tag>,
    },
    {
      title: '耗时',
      dataIndex: 'latency',
      width: 80,
      render: (_: unknown, r: AuditLogRow) => (r.latency ? <span>{r.latency}ms</span> : <span style={{ color: token.colorTextTertiary }}>-</span>),
    },
    {
      title: '操作详情',
      key: 'detail',
      render: (_: unknown, r: AuditLogRow) =>
        r.detail ? (
          <a onClick={() => setDetail(r)}>查看</a>
        ) : (
          <span style={{ color: token.colorTextTertiary }}>-</span>
        ),
    },
    { title: 'IP', dataIndex: 'ip', width: 130, render: (_: unknown, r: AuditLogRow) => <code style={{ background: token.colorFillAlter, padding: '2px 6px', borderRadius: 4 }}>{r.ip}</code> },
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 160,
      render: (v: unknown) => formatTime(v),
    },
  ];

  return (
    <>
      <DataTable<AuditLogRow>
        columns={columns}
        rowKey="id"
        request={async (params) => {
          const r = await auditApi.getList(params);
          const data = r.data as Record<string, unknown>;
          return {
            items: (data.items as AuditLogRow[]) || [],
            total: (data.total as number) || 0,
          };
        }}
        searchFields={searchFields}
        headerTitle="审计日志"
      />

      <Modal
        title={`操作详情 #${detail?.id ?? ''}`}
        open={!!detail}
        onCancel={() => setDetail(null)}
        footer={null}
        width={560}
      >
        {detail && (
          <Descriptions column={1} size="small" bordered>
            <Descriptions.Item label="用户">{detail.username}</Descriptions.Item>
            <Descriptions.Item label="操作">
              {ACTION_VALUES[detail.action]?.text || detail.action}
            </Descriptions.Item>
            <Descriptions.Item label="资源">
              {detail.resource}{detail.resource_id ? ` #${detail.resource_id}` : ''}
            </Descriptions.Item>
            <Descriptions.Item label="结果">
              {detail.result === 'success' ? <Tag color="green">成功</Tag> : <Tag color="red">失败</Tag>}
              {detail.message ? `（${detail.message}）` : ''}
            </Descriptions.Item>
            <Descriptions.Item label="耗时">
              {detail.latency ? `${detail.latency}ms` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="IP">{detail.ip}</Descriptions.Item>
            <Descriptions.Item label="UserAgent">
              <Typography.Text style={{ fontSize: 12 }} ellipsis={{ tooltip: detail.user_agent }}>
                {detail.user_agent}
              </Typography.Text>
            </Descriptions.Item>
            <Descriptions.Item label="时间">{formatTime(detail.created_at)}</Descriptions.Item>
            <Descriptions.Item label="详情">
              {renderDetail(detail.detail, token)}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </>
  );
};

// renderDetail 渲染审计详情：
// - diff 结构（字段 → {old,new}）→ 友好的 旧值→新值 展示
// - 普通 JSON → 格式化展示
function renderDetail(raw: string, token: ReturnType<typeof useThemeToken>): React.ReactNode {
  if (!raw) return <span style={{ color: token.colorTextTertiary }}>-</span>;
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return <pre style={{ margin: 0, fontSize: 12 }}>{raw}</pre>;
  }
  // 检测是否为 diff 结构：对象且至少一个值含 old/new
  const isDiff =
    parsed && typeof parsed === 'object' &&
    Object.values(parsed as Record<string, unknown>).some(
      (v) => v && typeof v === 'object' && 'old' in (v as object) && 'new' in (v as object),
    );
  if (isDiff) {
    const entries = Object.entries(parsed as Record<string, { old: unknown; new: unknown }>);
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        {entries.map(([field, d]) => (
          <div key={field} style={{ fontSize: 12, lineHeight: 1.6 }}>
            <b>{field}</b>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 2 }}>
              <span style={{
                background: token.colorErrorBg, color: token.colorError, padding: '1px 8px', borderRadius: 4,
                textDecoration: 'line-through', opacity: 0.85,
              }}>
                {fmtVal(d.old)}
              </span>
              <span style={{ color: token.colorTextTertiary }}>→</span>
              <span style={{ background: token.colorSuccessBg, color: token.colorSuccess, padding: '1px 8px', borderRadius: 4 }}>
                {fmtVal(d.new)}
              </span>
            </div>
          </div>
        ))}
      </div>
    );
  }
  return (
    <pre style={{ margin: 0, padding: 12, background: token.colorFillAlter, borderRadius: 6, fontSize: 12, maxHeight: 240, overflow: 'auto', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
      {JSON.stringify(parsed, null, 2)}
    </pre>
  );
}

function fmtVal(v: unknown): string {
  if (Array.isArray(v)) return v.join(', ');
  if (v === null || v === undefined) return '空';
  if (typeof v === 'object') return JSON.stringify(v);
  return String(v);
}

export default AuditLogList;
