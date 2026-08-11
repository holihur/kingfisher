import React from 'react';
import DOMPurify from 'dompurify';

/**
 * 富文本预览共用部件：所有 dangerouslySetInnerHTML 渲染富文本必须走这里，
 * 内部用 DOMPurify 清洗 HTML，杜绝 XSS。全站富文本预览统一入口。
 */
const RichTextPreview: React.FC<{ content: string; className?: string }> = ({ content, className }) => {
  return <div className={className} dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(content || '') }} />;
};

export default RichTextPreview;
