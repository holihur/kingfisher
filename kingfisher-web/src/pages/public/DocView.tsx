import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Spin, Typography, Empty } from 'antd';
import { useThemeToken } from '../../hooks/useThemeToken';
import { publicDocApi, DocItem } from '../../api/doc';
import RichTextPreview from '../../components/RichTextPreview';

/**
 * 公开文档预览页（无需登录）。
 * 访问 /docs/public/:id 即可阅读已发布+共享的文档，无后台侧边栏/编辑按钮。
 */
const PublicDocView: React.FC = () => {
  const token = useThemeToken();
  const { id } = useParams<{ id: string }>();
  const [loading, setLoading] = useState(true);
  const [doc, setDoc] = useState<DocItem | null>(null);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      if (!id) return;
      setLoading(true);
      try {
        const r = await publicDocApi.get(Number(id));
        if (!cancelled) setDoc(r.data as DocItem);
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
  }, [id]);

  return (
    <div
      style={{
        minHeight: '100vh',
        background: token.colorBgLayout,
        display: 'flex',
        justifyContent: 'center',
        padding: '40px 16px',
      }}
    >
      <div
        style={{
          width: 860,
          maxWidth: '100%',
          background: token.colorBgContainer,
          borderRadius: token.borderRadiusLG,
          padding: 40,
          boxShadow: '0 1px 4px rgba(0,0,0,0.06)',
          alignSelf: 'flex-start',
        }}
      >
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 80 }}>
            <Spin size="large" />
          </div>
        ) : notFound || !doc ? (
          <Empty description="文档不存在或未公开" image={Empty.PRESENTED_IMAGE_SIMPLE} style={{ padding: 40 }} />
        ) : (
          <>
            <Typography.Title level={2} style={{ marginTop: 0, marginBottom: 24 }}>
              {doc.title}
            </Typography.Title>
            <RichTextPreview content={doc.content} />
          </>
        )}
      </div>
    </div>
  );
};

export default PublicDocView;
