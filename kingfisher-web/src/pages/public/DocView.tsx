import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Empty, Spin, Tree, Typography } from 'antd';
import type { TreeDataNode } from 'antd';
import { FileTextOutlined, FolderOpenOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons';
import { useThemeToken } from '../../hooks/useThemeToken';
import { publicDocApi, DocItem, DocDirNode } from '../../api/doc';
import RichTextPreview from '../../components/RichTextPreview';
import DocOutline from '../../components/DocOutline';

/** 公开目录树 → antd Tree（目录 + 文档叶子，key 带前缀区分） */
const toTreeData = (nodes: DocDirNode[]): TreeDataNode[] =>
  (nodes || []).map((n) => {
    const children: TreeDataNode[] = [];
    if (n.children?.length) children.push(...toTreeData(n.children));
    if (n.docs?.length) {
      children.push(
        ...n.docs.map((d: any) => ({
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

/** 扁平化目录树：找目录节点（用于右侧展示该目录的文档列表） */
const findDir = (nodes: any[], dirId: number): any | null => {
  for (const n of nodes || []) {
    if (n.id === dirId) return n;
    const found = findDir(n.children, dirId);
    if (found) return found;
  }
  return null;
};

/**
 * 公开文档站（无需登录）：左侧可折叠公开目录树，右侧显示文档内容 / 目录下文档列表。
 * /docs/public  → 目录导航；/docs/public/:id → 文档阅读（含右侧悬浮大纲）。
 * 只展示 shared 目录及其 published+shared 文档（由后端 /public/docs/tree 过滤）。
 */
const PublicDocView: React.FC = () => {
  const token = useThemeToken();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const docId = id ? Number(id) : null;

  const [tree, setTree] = useState<any[]>([]);
  const [treeLoading, setTreeLoading] = useState(true);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [collapsed, setCollapsed] = useState(false);

  // 文档内容 / 当前目录
  const [loading, setLoading] = useState(docId !== null);
  const [doc, setDoc] = useState<DocItem | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [selectedDirId, setSelectedDirId] = useState<number | null>(null);
  const outlineTargetRef = useRef<HTMLDivElement>(null);

  // 加载公开目录树
  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setTreeLoading(true);
      try {
        const r = await publicDocApi.getTree();
        if (!cancelled) setTree((r.data as any[]) || []);
      } catch {
        /* 已提示 */
      } finally {
        if (!cancelled) setTreeLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  // 树加载后默认展开全部目录
  useEffect(() => {
    const collect = (nodes: any[]): React.Key[] =>
      (nodes || []).flatMap((n) => [`dir-${n.id}`, ...collect(n.children)]);
    setExpandedKeys((prev) => Array.from(new Set([...prev, ...collect(tree)])));
  }, [tree]);

  // 加载文档内容（带 id 时）
  useEffect(() => {
    if (docId === null) {
      setDoc(null);
      setLoading(false);
      setNotFound(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setNotFound(false);
    const load = async () => {
      try {
        const r = await publicDocApi.get(docId);
        if (!cancelled) {
          setDoc(r.data as DocItem);
          setSelectedDirId((r.data as DocItem).dir_id);
        }
      } catch {
        if (!cancelled) setNotFound(true);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [docId]);

  // 当前目录节点（用于无 id 时展示文档列表）
  const currentDir = useMemo(() => findDir(tree, selectedDirId ?? -1), [tree, selectedDirId]);

  const treeData = useMemo(() => toTreeData(tree), [tree]);

  // 树节点点击：目录 → 展示该目录文档列表；文档 → 跳转阅读
  const onSelect = (keys: React.Key[]) => {
    if (!keys.length) return;
    const key = String(keys[0]);
    if (key.startsWith('doc-')) {
      navigate(`/docs/public/${key.replace('doc-', '')}`);
    } else {
      setSelectedDirId(Number(key.replace('dir-', '')));
      navigate('/docs/public', { replace: true });
    }
  };

  // 当前目录下文档列表（无 id 时右侧展示）
  const dirDocs = currentDir?.docs || [];

  return (
    <div
      style={{
        minHeight: '100vh',
        background: token.colorBgLayout,
        display: 'flex',
        padding: 16,
        gap: 16,
        alignItems: 'flex-start',
      }}
    >
      {/* 左侧：公开目录树（可折叠） */}
      <aside
        style={{
          flex: 'none',
          width: collapsed ? 0 : 240,
          overflow: 'hidden',
          background: token.colorBgContainer,
          borderRadius: token.borderRadiusLG,
          padding: collapsed ? 0 : 12,
          transition: 'width 0.2s',
          position: 'sticky',
          top: 16,
          maxHeight: 'calc(100vh - 32px)',
          overflowY: 'auto',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
          <span style={{ fontWeight: 600 }}>文档目录</span>
          <span
            onClick={() => setCollapsed(true)}
            title="收起目录"
            style={{ cursor: 'pointer', color: token.colorTextTertiary, display: 'flex', alignItems: 'center' }}
          >
            <MenuFoldOutlined />
          </span>
        </div>
        {collapsed ? null : treeLoading ? (
          <div style={{ textAlign: 'center', padding: 24 }}>
            <Spin />
          </div>
        ) : tree.length === 0 ? (
          <Empty description="暂无公开文档" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Tree
            treeData={treeData}
            showLine
            blockNode
            expandedKeys={expandedKeys}
            onExpand={setExpandedKeys}
            selectedKeys={[docId ? `doc-${docId}` : selectedDirId ? `dir-${selectedDirId}` : '']}
            onSelect={onSelect}
          />
        )}
      </aside>

      {/* 左栏已折叠：展开入口 */}
      {collapsed && (
        <span
          onClick={() => setCollapsed(false)}
          title="展开目录"
          style={{
            cursor: 'pointer',
            color: token.colorTextTertiary,
            background: token.colorBgContainer,
            borderRadius: 6,
            padding: '8px 6px',
            display: 'flex',
            alignItems: 'center',
            boxShadow: '0 1px 4px rgba(0,0,0,0.06)',
          }}
        >
          <MenuUnfoldOutlined />
        </span>
      )}

      {/* 右侧：文档内容 / 目录文档列表 */}
      <main
        style={{
          flex: 1,
          minWidth: 0,
          maxWidth: 860,
          background: token.colorBgContainer,
          borderRadius: token.borderRadiusLG,
          padding: 40,
          boxShadow: '0 1px 4px rgba(0,0,0,0.06)',
        }}
      >
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}>
            <Spin size="large" />
          </div>
        ) : notFound || (docId !== null && !doc) ? (
          <Empty description="文档不存在或未公开" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ padding: 40 }} />
        ) : docId !== null && doc ? (
          <>
            <Typography.Title level={2} style={{ marginTop: 0, marginBottom: 8 }}>
              {doc.title}
            </Typography.Title>
            <Typography.Text type="secondary" style={{ fontSize: 13, display: 'block', marginBottom: 24 }}>
              最后更新：{new Date(doc.updated_at).toLocaleString('zh-CN')}
            </Typography.Text>
            <div ref={outlineTargetRef}>
              <RichTextPreview content={doc.content} />
            </div>
            <DocOutline content={doc.content} targetRef={outlineTargetRef} />
          </>
        ) : selectedDirId ? (
          <>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 }}>
              <FolderOpenOutlined style={{ color: token.colorPrimary }} />
              <Typography.Title level={4} style={{ margin: 0 }}>
                {currentDir?.name || '目录'}
              </Typography.Title>
            </div>
            {dirDocs.length === 0 ? (
              <Empty description="该目录暂无公开文档" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ padding: 40 }} />
            ) : (
              <ul style={{ listStyle: 'none', margin: 0, padding: 0 }}>
                {dirDocs.map((d: any) => (
                  <li key={d.id} style={{ borderBottom: `1px solid ${token.colorSplit}` }}>
                    <a
                      href={`/docs/public/${d.id}`}
                      onClick={(e) => {
                        e.preventDefault();
                        navigate(`/docs/public/${d.id}`);
                      }}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        padding: '12px 4px',
                        color: token.colorText,
                        textDecoration: 'none',
                      }}
                    >
                      <FileTextOutlined style={{ color: token.colorTextTertiary }} />
                      <span>{d.title}</span>
                    </a>
                  </li>
                ))}
              </ul>
            )}
          </>
        ) : (
          <Empty description="请从左侧选择文档" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ padding: 40 }} />
        )}
      </main>
    </div>
  );
};

export default PublicDocView;
