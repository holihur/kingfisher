import React, { useEffect, useState } from 'react';
import { Button, Modal, Form, Input, Tree, Tabs, Checkbox, Row, Col, message, Popconfirm, Badge } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import ProTable, { ProColumns } from '@ant-design/pro-table';
import { useAuthStore } from '../../stores/auth';
import { roleApi } from '../../api/role';
import { menuApi } from '../../api/menu';
import { useTableUrlQuery } from '../../hooks/useTableUrlQuery';
import { buildQueryParams } from '../../utils/query';

const RoleList: React.FC = () => {
  const { urlParams, page, pageSize, actionRef, formRef, syncFormFromUrl, onSearch, onReset, onPageChange } = useTableUrlQuery();
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Record<string, unknown> | null>(null);
  const [permModal, setPermModal] = useState<{ open: boolean; role: Record<string, unknown> | null }>({
    open: false,
    role: null,
  });
  const [menuModal, setMenuModal] = useState<{ open: boolean; role: Record<string, unknown> | null }>({
    open: false,
    role: null,
  });
  const [allPerms, setAllPerms] = useState<Record<string, unknown>[]>([]);
  const [selectedPerms, setSelectedPerms] = useState<number[]>([]);
  const [allMenus, setAllMenus] = useState<Record<string, unknown>[]>([]);
  const [selectedMenus, setSelectedMenus] = useState<number[]>([]);
  const [form] = Form.useForm<Record<string, unknown>>();
  const perms = useAuthStore((s) => s.permissions);

  // 挂载时用 URL 反填搜索表单
  useEffect(() => {
    syncFormFromUrl();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);


  const columns: ProColumns[] = [
    { title: 'ID', dataIndex: 'id', width: 80, search: false },
    {
      title: '关键词',
      dataIndex: 'q',
      hideInTable: true,
      search: { transform: (v) => ({ q: v }) },
    },
    { title: '角色名', dataIndex: 'name', search: false },
    { title: '编码', dataIndex: 'code', search: false },
    { title: '描述', dataIndex: 'description', ellipsis: true, search: false },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      valueEnum: { 1: { text: '启用' }, 0: { text: '禁用' } },
      render: (_, r) => (
        <Badge status={(r.status as number) === 1 ? 'success' : 'error'} text={(r.status as number) === 1 ? '启用' : '禁用'} />
      ),
    },
    {
      title: '操作',
      valueType: 'option',
      render: (_, r) => [
        perms.includes('role:update') ? (
          <a
            key="perm"
            onClick={async () => {
              const p = await roleApi.getAllPermissions();
              setAllPerms((p.data as Record<string, unknown>[]) || []);
              const rp = await roleApi.getPermissions(r.id as number);
              setSelectedPerms(
                ((rp.data as Record<string, unknown>[]) || []).map((i: Record<string, unknown>) => i.id as number)
              );
              setPermModal({ open: true, role: r as Record<string, unknown> });
            }}
          >
            权限
          </a>
        ) : null,
        perms.includes('role:update') ? (
          <a
            key="menu"
            onClick={async () => {
              const m = await menuApi.getTree();
              setAllMenus((m.data as Record<string, unknown>[]) || []);
              const rm = await roleApi.getMenus(r.id as number);
              setSelectedMenus(
                ((rm.data as Record<string, unknown>[]) || []).map((i: Record<string, unknown>) => i.id as number)
              );
              setMenuModal({ open: true, role: r as Record<string, unknown> });
            }}
          >
            菜单
          </a>
        ) : null,
        perms.includes('role:update') ? (
          <a
            key="ed"
            onClick={() => {
              setEditing(r as Record<string, unknown>);
              setModalOpen(true);
            }}
          >
            编辑
          </a>
        ) : null,
        perms.includes('role:delete') ? (
          <Popconfirm
            key="del"
            title="确认删除？"
            onConfirm={async () => {
              await roleApi.delete(r.id as number);
              message.success('已删除');
              actionRef.current?.reload();
            }}
          >
            <a style={{ color: 'red' }}>删除</a>
          </Popconfirm>
        ) : null,
      ],
    },
  ];

  const handleSubmit = async () => {
    const v = await form.validateFields();
    if (editing?.id) {
      await roleApi.update(editing.id as number, v);
    } else {
      await roleApi.create(v);
    }
    message.success('保存成功');
    setModalOpen(false);
    setEditing(null);
    actionRef.current?.reload();
  };

  const grouped = (allPerms || []).reduce(
    (acc: Record<string, Record<string, unknown>[]>, p: Record<string, unknown>) => {
      const k = p.resource as string;
      if (!acc[k]) acc[k] = [];
      acc[k].push(p);
      return acc;
    },
    {}
  );

  // 递归转换菜单树：给每个节点补上 key/title，否则 Tree 组件显示 ---
  const toTreeData = (items: Record<string, unknown>[]): Record<string, unknown>[] =>
    items.map((m) => ({
      ...m,
      key: m.id as number,
      title: m.name as string,
      children: m.children ? toTreeData(m.children as Record<string, unknown>[]) : undefined,
    }));

  return (
    <>
      <ProTable
        columns={columns}
        actionRef={actionRef}
        formRef={formRef}
        params={urlParams}
        request={async (params) => {
          const r = await roleApi.getList(buildQueryParams(params));
          const data = r.data as Record<string, unknown>;
          return {
            data: (data.items as Record<string, unknown>[]) || [],
            total: (data.total as number) || 0,
            success: true,
          };
        }}
        rowKey="id"
        onSubmit={onSearch}
        onReset={onReset}
        search={{ labelWidth: 'auto' }}
        pagination={{ current: page, pageSize, showSizeChanger: true, onChange: onPageChange }}
        headerTitle="角色管理"
        toolBarRender={() => [
          perms.includes('role:create') ? (
            <Button
              key="add"
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                setModalOpen(true);
              }}
            >
              新增角色
            </Button>
          ) : null,
        ]}
      />
      <Modal
        title={editing ? '编辑角色' : '新增角色'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
        }}
        afterOpenChange={(open) => {
          if (open && editing) {
            form.setFieldsValue(editing as never);
          } else if (open && !editing) {
            form.resetFields();
          }
        }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true }]}>
            <Input disabled={!!editing?.id} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title={`分配权限 — ${permModal.role?.name || ''}`}
        open={permModal.open}
        width={700}
        onOk={async () => {
          await roleApi.assignPermissions(permModal.role!.id as number, selectedPerms);
          message.success('权限已更新');
          setPermModal({ open: false, role: null });
        }}
        onCancel={() => setPermModal({ open: false, role: null })}
      >
        <Checkbox.Group value={selectedPerms} onChange={(v) => setSelectedPerms(v as number[])} style={{ width: '100%' }}>
          <Tabs
            items={Object.entries(grouped).map(([res, ps]) => ({
              key: res,
              label: { user: '用户', menu: '菜单', role: '角色', config: '配置', audit: '审计' }[res] || res,
              children: (
                <Row gutter={[16, 8]}>
                  {ps.map((p) => (
                    <Col span={12} key={p.id as number}>
                      <Checkbox value={p.id}>{p.name as string}</Checkbox>
                    </Col>
                  ))}
                </Row>
              ),
            }))}
          />
        </Checkbox.Group>
      </Modal>
      <Modal
        title={`分配菜单 — ${menuModal.role?.name || ''}`}
        open={menuModal.open}
        onOk={async () => {
          await roleApi.assignMenus(menuModal.role!.id as number, selectedMenus);
          message.success('菜单已更新');
          setMenuModal({ open: false, role: null });
        }}
        onCancel={() => setMenuModal({ open: false, role: null })}
      >
        <Tree
          checkable
          defaultExpandAll
          key={menuModal.role?.id as number}
          checkedKeys={selectedMenus}
          onCheck={(keys) => {
            const arr = Array.isArray(keys) ? keys : (keys as { checked: React.Key[] }).checked;
            setSelectedMenus(arr as number[]);
          }}
          treeData={toTreeData(allMenus as Record<string, unknown>[]) as Record<string, unknown>[]}
        />
      </Modal>
    </>
  );
};

export default RoleList;
