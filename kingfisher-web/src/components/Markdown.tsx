import { memo, useMemo } from 'react';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { theme } from 'antd';

// marked 全局配置：开启 GFM 表/删除线；编码不转义中文等。
marked.setOptions({
  gfm: true,
  breaks: true,
});

/** Markdown 渲染组件：marked 解析 + DOMPurify 消毒（防 XSS）。 */
function Markdown({ content }: { content: string }) {
  const { token } = theme.useToken();

  const html = useMemo(() => {
    if (!content) return '';
    const raw = marked.parse(content, { async: false }) as string;
    return DOMPurify.sanitize(raw);
  }, [content]);

  // 组件级样式：与 antd 主题色联动；禁用全局 CSS 污染（scoped 选择器前缀）。
  return (
    <div
      className="agent-markdown"
      style={{
        lineHeight: 1.7,
        fontSize: 14,
        wordBreak: 'break-word',
        color: token.colorText,
      }}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

export default memo(Markdown);
