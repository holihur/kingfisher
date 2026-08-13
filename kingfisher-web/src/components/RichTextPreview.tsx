import React from 'react';
import DOMPurify from 'dompurify';
import { serializeEditorState } from './lexical/toHtml';

// Notion 块用到的自定义属性：提示框图标、代码块语言标记、待办列表类型/勾选状态
const ALLOWED_DATA_ATTRS = ['data-icon', 'data-language', 'data-highlight-language', '__lexicallisttype', 'aria-checked'];

/** 纯文本转义（旧数据非 HTML 兜底） */
const escapeText = (s: string): string =>
  s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]!));

/**
 * 文档内容 → 可渲染 HTML：
 * - 新版内容为 Lexical editorState JSON → 经 toHtml 序列化
 * - 旧数据（HTML/纯文本，非 JSON）→ 含 `<` 当 HTML 清洗渲染（开发期历史数据兜底），否则纯文本转义
 */
const renderContent = (content: string): string => {
  const raw = content || '';
  const trimmed = raw.trim();
  if (trimmed.startsWith('{')) {
    try {
      return serializeEditorState(JSON.parse(trimmed));
    } catch {
      /* 落到兜底 */
    }
  }
  // 旧数据兜底：宁可当纯文本转义，绝不当 HTML 直接注入
  return trimmed.includes('<') ? trimmed : escapeText(trimmed);
};

/**
 * 富文本预览共用部件：所有 dangerouslySetInnerHTML 渲染富文本必须走这里，
 * 内部用 DOMPurify 清洗 HTML，杜绝 XSS。全站富文本预览统一入口。
 * 内容为 Lexical editorState JSON，经序列化器转 HTML 后渲染。
 * 包装类默认 rt-preview，与编辑器共用块样式（callout/toggle/table/checklist 等）。
 */
const RichTextPreview: React.FC<{ content: string; className?: string }> = ({ content, className }) => {
  const html = DOMPurify.sanitize(renderContent(content || ''), {
    ADD_ATTR: ALLOWED_DATA_ATTRS,
  });
  return <div className={className || 'rt-preview'} dangerouslySetInnerHTML={{ __html: html }} />;
};

export default RichTextPreview;
