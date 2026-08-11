import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import Split from 'split.js';
import {
  App,
  Button,
  Checkbox,
  Dropdown,
  Empty,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Tree,
  Typography,
} from 'antd';
import type { TreeDataNode } from 'antd';
import {
  DeleteOutlined,
  EditOutlined,
  FileAddOutlined,
  FileTextOutlined,
  FolderAddOutlined,
  FolderOutlined,
  HistoryOutlined,
  MoreOutlined,
  PlusOutlined,
  SafetyOutlined,
} from '@ant-design/icons';
import ReactQuill from 'react-quill-new';
import 'react-quill-new/dist/quill.snow.css';
import { useThemeToken } from '../../hooks/useThemeToken';
import { useAuthStore } from '../../stores/auth';
import { docApi, docDirApi, DocDirNode, DocItem, DocVersion } from '../../api/doc';
import { roleApi } from '../../api/role';
import { formatTime } from '../../utils/format';
import RichTextPreview from '../../components/RichTextPreview';

const quillModules = {
  toolbar: [
    [{ header: [1, 2, 3, false] }],
    ['bold', 'italic', 'underline', 'strike'],
    [{ list: 'ordered' }, { list: 'bullet' }],
    ['blockquote', 'code-block'],
    [{ align: [] }],
    ['link', 'image'],
    ['clean'],
  ],
};

/** 目录节点 → antd Tree 数据（目录 + 目录下可见的文档叶子，key 带前缀区分） */
const toTreeData = (nodes: DocDirNode[]): TreeDataNode[] =>
  nodes.map((n) => {
    const children: TreeDataNode[] = [];
    if (n.children?.length) children.push(...toTreeData(n.children));
    if (n.docs?.length) {
      children.push(
        ...n.docs.map((d) => ({
          key: `doc-${d.id}`,
          title: d.title,
          isLeaf: true,
        }))
      );
    }
    return {
      key: `dir-${n.id}`,
      title: n.name,
      children: children.length ? children : undefined,
    };
  });

const DocManage: React.FC = () => {
  const { message, modal } = App.useApp();
  const token = useThemeToken();
  const perms = useAuthStore((s) => s.permissions);
  const userInfo = useAuthStore((s) => s.userInfo) as Record<string, unknown> | null;

  const [tree, setTree] = useState<DocDirNode[]>([]);
  const [treeLoading, setTreeLoading] = useState(false);
  // 目录树展开的 key（默认全展开；defaultExpandAll 对异步数据无效，需受控）
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);

  // 目录 modal
  const [dirModal, setDirModal] = useState<{
    open: boolean;
    mode: 'create' | 'rename';
    parentId: number;
    dir?: DocDirNode;
  }>({
    open: false,
    mode: 'create',
    parentId: 0,
  });
  const [dirName, setDirName] = useState('');

  // 授权 modal
  const [authModal, setAuthModal] = useState<{
    open: boolean;
    dir?: DocDirNode;
    roleIds: number[];
    allRoles: { id: number; name: string }[];
  }>({
    open: false,
    roleIds: [],
    allRoles: [],
  });

  // 右侧文档区
  const [selectedDirId, setSelectedDirId] = useState<number | null>(null);
  // 目录树选中节点 key（dir-<id> | doc-<id>），驱动高亮
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [docs, setDocs] = useState<DocItem[]>([]);
  const [docsTotal, setDocsTotal] = useState(0);
  const [docsPage, setDocsPage] = useState(1);
  const [docsLoading, setDocsLoading] = useState(false);
  const [docsRefresh, setDocsRefresh] = useState(0);

  // 编辑器状态
  const [editing, setEditing] = useState<DocItem | null>(null); // null = 列表视图
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const [editVisibility, setEditVisibility] = useState<'shared' | 'private'>('shared');
  const [editNote, setEditNote] = useState('');

  // 版本历史 modal
  const [verModal, setVerModal] = useState<{
    open: boolean;
    doc?: DocItem;
    versions: DocVersion[];
    loading: boolean;
    preview?: DocVersion;
  }>({
    open: false,
    versions: [],
    loading: false,
  });

  const canCreate = perms.includes('doc:create');
  const canUpdate = perms.includes('doc:update');
  const canDelete = perms.includes('doc:delete');

  // —— 目录树 ——
  const loadTree = useCallback(async () => {
    setTreeLoading(true);
    try {
      const r = await docDirApi.getTree();
      setTree((r.data as DocDirNode[]) || []);
    } catch {
      /* request 拦截器已提示 */
    } finally {
      setTreeLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadTree();
  }, [loadTree]);

  // 树加载后默认展开全部目录（defaultExpandAll 对异步数据不生效）
  useEffect(() => {
    const collect = (nodes: DocDirNode[]): React.Key[] =>
      nodes.flatMap((n) => [`dir-${n.id}`, ...collect(n.children || [])]);
    setExpandedKeys((prev) => Array.from(new Set([...prev, ...collect(tree)])));
  }, [tree]);

  // —— 文档列表 ——
  useEffect(() => {
    if (!selectedDirId) {
      setDocs([]);
      setDocsTotal(0);
      return;
    }
    setDocsLoading(true);
    docApi
      .list(selectedDirId, { page: docsPage, page_size: 10, sort: '-updated_at' })
      .then((r) => {
        const d = r.data as Record<string, unknown>;
        setDocs((d.items as DocItem[]) || []);
        setDocsTotal((d.total as number) || 0);
      })
      .catch(() => {})
      .finally(() => setDocsLoading(false));
  }, [selectedDirId, docsPage, docsRefresh]);

  // split.js：左右（目录树 / 内容）可拖拽调宽
  const containerRef = useRef<HTMLDivElement>(null);
  const leftPaneRef = useRef<HTMLDivElement>(null);
  const rightPaneRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const container = containerRef.current;
    const left = leftPaneRef.current;
    const right = rightPaneRef.current;
    if (!container || !left || !right) return;
    const instance = Split([left, right], {
      sizes: [20, 80],
      minSize: [160, 400],
      gutterSize: 6,
      cursor: 'col-resize',
      gutter: (_index, _direction) => {
        const g = document.createElement('div');
        g.className = 'doc-split-gutter';
        return g;
      },
    });
    return () => instance.destroy();
  }, []);

  const openDirModal = (mode: 'create' | 'rename', parentId: number, dir?: DocDirNode) => {
    setDirName(dir ? dir.name : '');
    setDirModal({ open: true, mode, parentId, dir });
  };

  const submitDirModal = async () => {
    const name = dirName.trim();
    if (!name) {
      message.warning('请输入目录名称');
      return;
    }
    try {
      if (dirModal.mode === 'create') {
        await docDirApi.create({ parent_id: dirModal.parentId, name });
        message.success('目录已创建');
      } else if (dirModal.dir) {
        await docDirApi.update(dirModal.dir.id, { name });
        message.success('目录已重命名');
      }
      setDirModal((s) => ({ ...s, open: false }));
      void loadTree();
    } catch {
      /* 已提示 */
    }
  };

  const handleDeleteDir = async (dir: DocDirNode) => {
    modal.confirm({
      title: `删除目录「${dir.name}」？`,
      content: '目录下有子目录或文档时将无法删除。',
      okText: '删除',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await docDirApi.delete(dir.id);
          message.success('目录已删除');
          if (selectedDirId === dir.id) {
            setSelectedDirId(null);
            setSelectedKey(null);
          }
          void loadTree();
        } catch {
          /* 已提示 */
        }
      },
    });
  };

  const openAuthModal = async (dir: DocDirNode) => {
    try {
      const [rolesR, grantedR] = await Promise.all([
        roleApi.getList({ page: 1, page_size: 100 }),
        docDirApi.getRoles(dir.id),
      ]);
      const roles = ((rolesR.data as Record<string, unknown>).items as { id: number; name: string }[]) || [];
      setAuthModal({ open: true, dir, roleIds: (grantedR.data as number[]) || [], allRoles: roles });
    } catch {
      /* 已提示 */
    }
  };

  const submitAuth = async () => {
    if (!authModal.dir) return;
    try {
      await docDirApi.setRoles(authModal.dir.id, authModal.roleIds);
      message.success('可见角色已更新');
      setAuthModal((s) => ({ ...s, open: false }));
      void loadTree();
    } catch {
      /* 已提示 */
    }
  };

  // —— 文档 ——
  const openDoc = async (doc: DocItem) => {
    try {
      const r = await docApi.getById(doc.id);
      const d = r.data as DocItem;
      setEditing(d);
      setSelectedKey(`doc-${d.id}`);
      setSelectedDirId(d.dir_id);
      setEditTitle(d.title);
      setEditContent(d.content);
      setEditVisibility(d.visibility);
      setEditNote('');
    } catch {
      /* 已提示 */
    }
  };

  // 从目录树叶子节点打开文档
  const openDocById = async (docId: number) => {
    await openDoc({ id: docId } as DocItem);
  };

  const createDoc = () => {
    if (!selectedDirId) return;
    setEditing({
      id: 0,
      dir_id: selectedDirId,
      title: '',
      content: '',
      owner_id: 0,
      visibility: 'shared',
      status: 'draft',
      current_version: 0,
      created_at: '',
      updated_at: '',
    });
    setEditTitle('');
    setEditContent('');
    setEditVisibility('shared');
    setEditNote('');
  };

  const saveDoc = async () => {
    if (!editTitle.trim()) {
      message.warning('请输入文档标题');
      return;
    }
    try {
      if (editing && editing.id > 0) {
        await docApi.update(editing.id, { title: editTitle, content: editContent, note: editNote });
        message.success('已保存，新版本已生成');
      } else {
        await docApi.create({
          dir_id: selectedDirId as number,
          title: editTitle,
          content: editContent,
          visibility: editVisibility,
          note: editNote,
        });
        message.success('文档已创建');
      }
      setDocsRefresh((k) => k + 1);
      if (editing && editing.id > 0) await openDoc({ ...editing, title: editTitle, content: editContent });
      else {
        setEditing(null);
        setDocsRefresh((k) => k + 1);
      }
    } catch {
      /* 已提示 */
    }
  };

  const togglePublish = async () => {
    if (!editing || editing.id <= 0) return;
    try {
      if (editing.status === 'published') {
        await docApi.unpublish(editing.id);
        message.success('已撤稿');
      } else {
        await docApi.publish(editing.id);
        message.success('已发布');
      }
      const r = await docApi.getById(editing.id);
      setEditing(r.data as DocItem);
      setDocsRefresh((k) => k + 1);
    } catch {
      /* 已提示 */
    }
  };

  const deleteDoc = async (doc: DocItem) => {
    modal.confirm({
      title: `删除文档「${doc.title}」？`,
      content: '文档及其全部版本将一并删除，不可恢复。',
      okText: '删除',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await docApi.delete(doc.id);
          message.success('文档已删除');
          setEditing(null);
          setDocsRefresh((k) => k + 1);
        } catch {
          /* 已提示 */
        }
      },
    });
  };

  const openVersions = async (doc: DocItem) => {
    setVerModal({ open: true, doc, versions: [], loading: true });
    try {
      const r = await docApi.versions(doc.id);
      setVerModal((s) => ({ ...s, versions: (r.data as DocVersion[]) || [], loading: false }));
    } catch {
      setVerModal((s) => ({ ...s, loading: false }));
    }
  };

  const viewVersion = async (doc: DocItem, no: number) => {
    try {
      const r = await docApi.getVersion(doc.id, no);
      setVerModal((s) => ({ ...s, preview: r.data as DocVersion }));
    } catch {
      /* 已提示 */
    }
  };

  const restoreVersion = async (doc: DocItem, no: number) => {
    modal.confirm({
      title: `还原到版本 ${no}？`,
      content: '将追加一个新版本并覆盖当前内容（历史仍保留）。',
      okText: '还原',
      onOk: async () => {
        try {
          await docApi.restore(doc.id, no);
          message.success('已还原');
          setVerModal((s) => ({ ...s, open: false }));
          setDocsRefresh((k) => k + 1);
        } catch {
          /* 已提示 */
        }
      },
    });
  };

  // —— 渲染 ——
  const treeData = useMemo(() => toTreeData(tree), [tree]);

  // 目录操作：hover 显示「更多」按钮 → Dropdown 下拉菜单（按权限动态出项）
  const dirActions = (dir: DocDirNode) => (
    <Dropdown
      trigger={['click']}
      menu={{
        items: [
          ...(canCreate ? [{ key: 'create', icon: <FolderAddOutlined />, label: '新建子目录' }] : []),
          ...(canUpdate ? [{ key: 'rename', icon: <EditOutlined />, label: '重命名' }] : []),
          ...(canUpdate ? [{ key: 'roles', icon: <SafetyOutlined />, label: '可见角色' }] : []),
          ...(canDelete ? [{ key: 'delete', icon: <DeleteOutlined />, label: '删除', danger: true }] : []),
        ],
        onClick: ({ key, domEvent }) => {
          domEvent.stopPropagation();
          if (key === 'create') openDirModal('create', dir.id);
          else if (key === 'rename') openDirModal('rename', dir.parent_id, dir);
          else if (key === 'roles') void openAuthModal(dir);
          else if (key === 'delete') handleDeleteDir(dir);
        },
      }}
    >
      <Button size="small" type="text" icon={<MoreOutlined />} title="操作" onClick={(e) => e.stopPropagation()} />
    </Dropdown>
  );

  // id → 目录 扁平映射（antd v6 Tree titleRender 只传 node，用 key 反查）
  const dirById = useMemo(() => {
    const m = new Map<number, DocDirNode>();
    const walk = (ns: DocDirNode[]) => {
      for (const n of ns) {
        m.set(n.id, n);
        if (n.children) walk(n.children);
      }
    };
    walk(tree);
    return m;
  }, [tree]);

  // 目录树节点渲染：目录（文件夹 + 更多菜单）/ 文档叶子（文件图标，点击打开）
  const dirTitleRender = (data: any) => {
    const key = String((data as { key: unknown }).key);
    if (key.startsWith('doc-')) {
      return (
        <span
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            width: '100%',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          <FileTextOutlined style={{ marginRight: 6, color: token.colorTextTertiary }} />
          {data?.title}
        </span>
      );
    }
    const dir = dirById.get(Number(key.replace('dir-', '')));
    if (!dir) return data?.title;
    return (
      <span
        style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'space-between', width: '100%', gap: 4 }}
      >
        {/* 目录名不换行，超长省略号截断（flex:1 占满剩余空间，minWidth:0 允许收缩） */}
        <span style={{ flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          <FolderOutlined style={{ marginRight: 6, color: token.colorPrimary }} />
          {data?.title}
        </span>
        {/* 「更多」按钮常驻，保证鼠标能移到弹出菜单（避免 hover 门控导致菜单关闭） */}
        {dirActions(dir)}
      </span>
    );
  };

  const docColumns = [
    { title: '标题', dataIndex: 'title', key: 'title', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (s: string) => (s === 'published' ? <Tag color="green">已发布</Tag> : <Tag color="orange">草稿</Tag>),
    },
    {
      title: '可见性',
      dataIndex: 'visibility',
      key: 'visibility',
      width: 90,
      render: (v: string) => (v === 'private' ? <Tag>私有</Tag> : <Tag color="blue">共享</Tag>),
    },
    { title: '作者', dataIndex: 'owner_name', key: 'owner_name', width: 100 },
    { title: '更新于', dataIndex: 'updated_at', key: 'updated_at', width: 150, render: (v: string) => formatTime(v) },
    {
      title: '操作',
      key: 'actions',
      width: 160,
      render: (_: unknown, doc: DocItem) => (
        <Space size={4}>
          <Button size="small" type="link" onClick={() => void openDoc(doc)}>
            打开
          </Button>
          {canUpdate && (
            <Button size="small" type="link" icon={<HistoryOutlined />} onClick={() => void openVersions(doc)}>
              版本
            </Button>
          )}
          {canDelete && (
            <Button size="small" type="link" danger onClick={() => deleteDoc(doc)}>
              删除
            </Button>
          )}
        </Space>
      ),
    },
  ];

  const isAuthor = editing ? Number((userInfo as Record<string, unknown> | null)?.id) === editing.owner_id : false;
  const showEditor = canUpdate; // 可读文档下：private 仅作者/admin 可读（非作者拿不到），shared 有 update 权限者可编

  return (
    // split.js 容器：display:flex + 左右两个 pane（宽度由拖拽控制，见上方 useEffect）
    <div ref={containerRef} style={{ display: 'flex', alignItems: 'flex-start' }}>
      {/* 左侧：目录树 */}
      <div
        ref={leftPaneRef}
        style={{
          minWidth: 0,
          background: token.colorBgContainer,
          borderRadius: token.borderRadiusLG,
          padding: 12,
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
          <span style={{ fontWeight: 600 }}>文档目录</span>
          {canCreate && (
            <Button size="small" type="primary" ghost icon={<PlusOutlined />} onClick={() => openDirModal('create', 0)}>
              新增
            </Button>
          )}
        </div>
        {tree.length === 0 && !treeLoading ? (
          <Empty description="暂无目录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Tree
            treeData={treeData}
            showLine
            expandedKeys={expandedKeys}
            onExpand={setExpandedKeys}
            selectedKeys={selectedKey ? [selectedKey] : []}
            onSelect={(keys) => {
              if (keys.length === 0) return;
              const key = String(keys[0]);
              if (key.startsWith('doc-')) {
                // 点击文档叶子 → 直接在右侧打开
                void openDocById(Number(key.replace('doc-', '')));
              } else {
                // 点击目录 → 右侧显示该目录文档列表
                setSelectedKey(key);
                setSelectedDirId(Number(key.replace('dir-', '')));
                setDocsPage(1);
                setEditing(null);
              }
            }}
            titleRender={dirTitleRender}
          />
        )}
      </div>

      {/* 右侧：列表 / 编辑器 */}
      <div
        ref={rightPaneRef}
        style={{
          minWidth: 0,
          background: token.colorBgContainer,
          borderRadius: token.borderRadiusLG,
          padding: 16,
        }}
      >
        {editing ? (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <Button
                onClick={() => {
                  setEditing(null);
                  if (selectedDirId) setSelectedKey(`dir-${selectedDirId}`);
                }}
              >
                返回列表
              </Button>
              <Typography.Text strong style={{ fontSize: 15 }}>
                {editing.id > 0 ? '编辑文档' : '新建文档'}
              </Typography.Text>
              {editing.id > 0 &&
                (editing.status === 'published' ? <Tag color="green">已发布</Tag> : <Tag color="orange">草稿</Tag>)}
            </div>
            {showEditor ? (
              <>
                <Space>
                  <Input
                    placeholder="文档标题"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    style={{ width: 360 }}
                  />
                  {editing.id === 0 && (
                    <Select
                      value={editVisibility}
                      onChange={setEditVisibility}
                      style={{ width: 100 }}
                      options={[
                        { value: 'shared', label: '共享' },
                        { value: 'private', label: '私有' },
                      ]}
                    />
                  )}
                  <Input
                    placeholder="变更说明（可选）"
                    value={editNote}
                    onChange={(e) => setEditNote(e.target.value)}
                    style={{ width: 240 }}
                  />
                </Space>
                <ReactQuill
                  theme="snow"
                  value={editContent}
                  onChange={setEditContent}
                  modules={quillModules}
                  style={{ minHeight: 320 }}
                />
                <Space>
                  <Button type="primary" onClick={() => void saveDoc()}>
                    保存
                  </Button>
                  {editing.id > 0 && (
                    <>
                      <Button onClick={() => void togglePublish()}>
                        {editing.status === 'published' ? '撤稿' : '发布'}
                      </Button>
                      <Button icon={<HistoryOutlined />} onClick={() => void openVersions(editing)}>
                        版本历史
                      </Button>
                    </>
                  )}
                </Space>
              </>
            ) : (
              <>
                <Typography.Title level={4} style={{ marginTop: 0 }}>
                  {editing.title}
                </Typography.Title>
                <RichTextPreview content={editing.content} />
                {isAuthor && (
                  <Button
                    icon={<HistoryOutlined />}
                    style={{ marginTop: 12 }}
                    onClick={() => void openVersions(editing)}
                  >
                    版本历史
                  </Button>
                )}
              </>
            )}
          </Space>
        ) : !selectedDirId ? (
          <Empty description="请在左侧选择一个目录" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ marginTop: 80 }} />
        ) : (
          <>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
              <Typography.Text strong>目录文档</Typography.Text>
              {canCreate && (
                <Button type="primary" icon={<FileAddOutlined />} onClick={createDoc}>
                  新建文档
                </Button>
              )}
            </div>
            <Table<DocItem>
              rowKey="id"
              columns={docColumns}
              dataSource={docs}
              loading={docsLoading}
              pagination={{
                current: docsPage,
                pageSize: 10,
                total: docsTotal,
                onChange: setDocsPage,
                showSizeChanger: false,
              }}
              size="small"
            />
          </>
        )}
      </div>

      {/* 目录创建/重命名 Modal */}
      <Modal
        open={dirModal.open}
        title={dirModal.mode === 'create' ? (dirModal.parentId ? '新建子目录' : '新建根目录') : '重命名目录'}
        onOk={() => void submitDirModal()}
        onCancel={() => setDirModal((s) => ({ ...s, open: false }))}
        destroyOnHidden
      >
        <Input
          placeholder="目录名称"
          value={dirName}
          onChange={(e) => setDirName(e.target.value)}
          onPressEnter={() => void submitDirModal()}
        />
      </Modal>

      {/* 目录可见角色 Modal */}
      <Modal
        open={authModal.open}
        title={`「${authModal.dir?.name || ''}」可见角色`}
        okText="保存"
        onOk={() => void submitAuth()}
        onCancel={() => setAuthModal((s) => ({ ...s, open: false }))}
      >
        <Typography.Paragraph type="secondary" style={{ fontSize: 12 }}>
          仅勾选的角色可见该目录（含子目录）。不勾选任何角色 = 目录对所有人隐藏（仅管理员可见）。
        </Typography.Paragraph>
        <Checkbox.Group
          value={authModal.roleIds}
          onChange={(vals) => setAuthModal((s) => ({ ...s, roleIds: vals as number[] }))}
          options={authModal.allRoles.map((r) => ({ label: r.name, value: r.id }))}
        />
      </Modal>

      {/* 版本历史 Modal */}
      <Modal
        open={verModal.open}
        title={`版本历史${verModal.preview ? ` - 版本 ${verModal.preview.version_no}` : ''}`}
        width={720}
        footer={null}
        onCancel={() => setVerModal((s) => ({ ...s, open: false, preview: undefined }))}
      >
        {verModal.preview ? (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Space>
              <Button size="small" onClick={() => setVerModal((s) => ({ ...s, preview: undefined }))}>
                返回列表
              </Button>
              <Typography.Text strong>{verModal.preview.title}</Typography.Text>
              {verModal.preview.note && <Tag>{verModal.preview.note}</Tag>}
            </Space>
            <div style={{ border: '1px solid ' + token.colorSplit, borderRadius: 6, padding: 12, minHeight: 200 }}>
              <RichTextPreview content={verModal.preview.content} />
            </div>
          </Space>
        ) : (
          <Table<DocVersion>
            rowKey="version_no"
            size="small"
            loading={verModal.loading}
            pagination={false}
            dataSource={verModal.versions}
            columns={[
              { title: '版本', dataIndex: 'version_no', key: 'version_no', width: 60 },
              { title: '标题', dataIndex: 'title', key: 'title', ellipsis: true },
              { title: '说明', dataIndex: 'note', key: 'note', ellipsis: true },
              { title: '作者', dataIndex: 'owner_name', key: 'owner_name', width: 90 },
              {
                title: '时间',
                dataIndex: 'created_at',
                key: 'created_at',
                width: 160,
                render: (v: string) => formatTime(v),
              },
              {
                title: '操作',
                key: 'actions',
                width: 140,
                render: (_: unknown, v: DocVersion) =>
                  verModal.doc ? (
                    <Space size={2}>
                      <Button size="small" type="link" onClick={() => void viewVersion(verModal.doc!, v.version_no)}>
                        查看
                      </Button>
                      {isAuthor && (
                        <Popconfirm
                          title={`还原到版本 ${v.version_no}？`}
                          onConfirm={() => void restoreVersion(verModal.doc!, v.version_no)}
                        >
                          <Button size="small" type="link">
                            还原
                          </Button>
                        </Popconfirm>
                      )}
                    </Space>
                  ) : null,
              },
            ]}
          />
        )}
      </Modal>
    </div>
  );
};

export default DocManage;
