/* ============ Lexical editorState JSON → HTML 序列化器 ============ */

/**
 * 把 Lexical 的 serialized EditorState（`editor.getEditorState().toJSON()` 的输出）
 * 转成富文本 HTML，供 RichTextPreview / 公开页渲染。输出仍会经过 DOMPurify 清洗，
 * 因此这里只保证结构与现有 CSS（标题/引用/列表/待办/callout/toggle/表格/代码）对齐，
 * 不需要自证安全（文本一律转义）。
 *
 * 与编辑器共用同一份块语义：列表用 `__lexicallisttype`/`aria-checked` 标记待办
 * （编辑器导入端靠这些属性还原），代码用 `data-language` 标记语言。
 */

type SerializedNode = Record<string, any> & { type: string; children?: SerializedNode[] };

const escapeHtml = (s: unknown): string =>
  String(s ?? '').replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]!));

/** 文本格式位掩码（与 lexical 的 IS_* 常量一致） */
const IS_BOLD = 1;
const IS_ITALIC = 1 << 1;
const IS_STRIKETHROUGH = 1 << 2;
const IS_UNDERLINE = 1 << 3;
const IS_CODE = 1 << 4;
const IS_SUBSCRIPT = 1 << 5;
const IS_SUPERSCRIPT = 1 << 6;
const IS_HIGHLIGHT = 1 << 7;

/** 元素块上的对齐/缩进 → style 属性 */
function fmtAttr(node: SerializedNode): string {
  const styles: string[] = [];
  const format = node.format as string | undefined;
  if (format === 'left' || format === 'center' || format === 'right' || format === 'justify') {
    styles.push(`text-align: ${format}`);
  }
  if (typeof node.indent === 'number' && node.indent > 0) {
    styles.push(`margin-left: ${node.indent * 24}px`);
  }
  return styles.length ? ` style="${styles.join('; ')}"` : '';
}

/** 文本节点（含 code-highlight）：按 format 位掩码 + style 套标签 */
function serializeText(node: SerializedNode): string {
  let text = escapeHtml(node.text);
  const format = Number(node.format) || 0;
  const style = typeof node.style === 'string' && node.style.trim() ? node.style.trim() : '';

  if (style) text = `<span style="${escapeHtml(style)}">${text}</span>`;
  // 嵌套顺序固定：最外层到最后。Lexical 的 style 同时承载颜色/背景。
  if (format & IS_CODE) text = `<code>${text}</code>`;
  if (format & IS_SUPERSCRIPT) text = `<sup>${text}</sup>`;
  if (format & IS_SUBSCRIPT) text = `<sub>${text}</sub>`;
  if (format & IS_HIGHLIGHT) text = `<mark>${text}</mark>`;
  if (format & IS_STRIKETHROUGH) text = `<s>${text}</s>`;
  if (format & IS_UNDERLINE) text = `<u>${text}</u>`;
  if (format & IS_ITALIC) text = `<em>${text}</em>`;
  if (format & IS_BOLD) text = `<strong>${text}</strong>`;
  return text;
}

/** 序列化一个元素节点，返回 HTML（开标签 + 子节点 + 闭标签） */
function serializeElement(node: SerializedNode): string {
  const children = (node.children || []).map(serializeNode).join('');
  const fmt = fmtAttr(node);

  switch (node.type) {
    case 'paragraph':
      return `<p${fmt}>${children}</p>`;
    case 'heading': {
      const tag = /^h[1-6]$/.test(String(node.tag)) ? String(node.tag) : 'h2';
      return `<${tag}${fmt}>${children}</${tag}>`;
    }
    case 'quote':
      return `<blockquote${fmt}>${children}</blockquote>`;
    case 'code': {
      // 代码块子节点是 code-highlight / text / linebreak，直接拼成文本
      const code = (node.children || [])
        .map((c) => (c.type === 'linebreak' ? '\n' : c.text ?? ''))
        .join('');
      const lang = node.language ? ` data-language="${escapeHtml(node.language)}"` : '';
      return `<pre${fmt}><code${lang}>${escapeHtml(code)}</code></pre>`;
    }
    case 'list': {
      const listType = node.listType;
      const isCheck = listType === 'check';
      const tag = listType === 'number' ? 'ol' : 'ul';
      const checkAttr = isCheck ? ' __lexicallisttype="check"' : '';
      const startAttr = tag === 'ol' && node.start ? ` start="${Number(node.start)}"` : '';
      return `<${tag}${checkAttr}${startAttr}${fmt}>${children}</${tag}>`;
    }
    case 'listitem': {
      // 待办：导出 checkbox input + aria-checked（与编辑器导出/导入契约一致）
      let inner = children;
      if (node.checked !== undefined && node.checked !== null) {
        const checked = Boolean(node.checked);
        const input = `<input type="checkbox"${checked ? ' checked' : ''} disabled>`;
        inner = input + inner;
      }
      const valueAttr = typeof node.value === 'number' ? ` value="${node.value}"` : '';
      const checkedAttr =
        node.checked !== undefined && node.checked !== null ? ` aria-checked="${node.checked ? 'true' : 'false'}"` : '';
      return `<li${valueAttr}${checkedAttr}${fmt}>${inner}</li>`;
    }
    case 'link': {
      const href = escapeHtml(node.url ?? '');
      const rel = node.rel ? ` rel="${escapeHtml(node.rel)}"` : ' rel="noreferrer"';
      const target = node.target ? ` target="${escapeHtml(node.target)}"` : ' target="_blank"';
      return `<a href="${href}"${rel}${target}>${children}</a>`;
    }
    case 'table':
      return `<table${fmt}><tbody>${children}</tbody></table>`;
    case 'tablerow':
      return `<tr>${children}</tr>`;
    case 'tablecell': {
      const isHeader = Number(node.headerState || 0) !== 0;
      const tag = isHeader ? 'th' : 'td';
      const attrs: string[] = [];
      if (Number(node.colSpan) > 1) attrs.push(`colspan="${Number(node.colSpan)}"`);
      if (Number(node.rowSpan) > 1) attrs.push(`rowspan="${Number(node.rowSpan)}"`);
      if (node.backgroundColor) attrs.push(`style="background-color: ${escapeHtml(node.backgroundColor)}"`);
      return `<${tag}${attrs.length ? ' ' + attrs.join(' ') : ''}>${children}</${tag}>`;
    }
    case 'horizontalrule':
      return `<hr />`;
    case 'linebreak':
      return `<br />`;
    case 'image': {
      // 图片：src 非空才输出 <img>（未上传占位不预览）；文本转义避免属性注入
      const src = node.src ? escapeHtml(node.src) : '';
      if (!src) return '';
      const alt = node.altText ? ` alt="${escapeHtml(node.altText)}"` : '';
      return `<img src="${src}"${alt} style="max-width:100%;border-radius:8px" />`;
    }
    case 'callout':
      return `<div class="notion-callout" data-icon="${escapeHtml(node.icon || '💡')}">${children}</div>`;
    case 'toggle':
      return `<div class="notion-toggle">${children}</div>`;
    default:
      // 未知元素节点：降级渲染子节点，保持内容可见（不输出危险标签）
      return children;
  }
}

/** 递归序列化单个节点 */
export function serializeNode(node: SerializedNode): string {
  if (!node || typeof node.type !== 'string') return '';
  if (node.type === 'text' || node.type === 'code-highlight') return serializeText(node);
  if (node.type === 'root') return (node.children || []).map(serializeNode).join('');
  if (node.type === 'linebreak') return '<br />';
  return serializeElement(node);
}

/**
 * 序列化整个 editorState JSON。
 * 非法/非对象输入返回空串，交给调用方兜底（预览组件会把非 JSON 当旧数据/纯文本处理）。
 */
export function serializeEditorState(state: unknown): string {
  if (!state || typeof state !== 'object') return '';
  const root = (state as { root?: SerializedNode }).root;
  if (!root) return '';
  return (root.children || []).map(serializeNode).join('');
}
