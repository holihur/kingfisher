import React, { useCallback, useEffect, useRef, useState } from 'react';
import type { EditorState } from 'lexical';
import { HeadingNode, QuoteNode } from '@lexical/rich-text';
import { CodeNode, CodeHighlightNode, registerCodeHighlighting } from '@lexical/code';
import { ListNode, ListItemNode } from '@lexical/list';
import { LinkNode } from '@lexical/link';
import { HorizontalRuleNode } from '@lexical/react/LexicalHorizontalRuleNode';
import { TableNode, TableRowNode, TableCellNode } from '@lexical/table';
import { LexicalComposer } from '@lexical/react/LexicalComposer';
import { ContentEditable } from '@lexical/react/LexicalContentEditable';
import { HistoryPlugin, createEmptyHistoryState } from '@lexical/react/LexicalHistoryPlugin';
import { RichTextPlugin } from '@lexical/react/LexicalRichTextPlugin';
import { OnChangePlugin } from '@lexical/react/LexicalOnChangePlugin';
import { ListPlugin } from '@lexical/react/LexicalListPlugin';
import { LinkPlugin } from '@lexical/react/LexicalLinkPlugin';
import { MarkdownShortcutPlugin, DEFAULT_TRANSFORMERS } from '@lexical/react/LexicalMarkdownShortcutPlugin';
import { CheckListPlugin } from '@lexical/react/LexicalCheckListPlugin';
import { TablePlugin } from '@lexical/react/LexicalTablePlugin';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { useThemeToken } from '../hooks/useThemeToken';
import { CalloutNode, ToggleNode } from './lexical/blocks';
import { ImageNode } from './lexical/ImageNode';
import { SlashMenuPlugin } from './lexical/SlashMenuPlugin';
import { FloatingFormatToolbar } from './lexical/FloatingFormatToolbar';
import { DraggableBlocks } from './lexical/DraggableBlocks';
import { ToolbarPlugin } from './lexical/ToolbarPlugin';
import { TableActionMenuPlugin } from './lexical/TableActionMenuPlugin';

interface LexicalEditorProps {
  value: string; // Lexical editorState JSON（编辑时导入，变更时导出）
  onChange: (json: string) => void;
  placeholder?: string;
  minHeight?: number;
}

/** 空文档的 serialized editorState（根下挂一个空段落，避免 parseEditorState 对空根抛错） */
const EMPTY_STATE_JSON =
  '{"root":{"type":"root","direction":null,"format":"","indent":0,"version":1,"children":[{"type":"paragraph","version":1,"children":[]}]}}';

/** 递归补齐序列化时被省略的默认字段：0.49 的 exportJSON 省略 direction(null)/indent(0)，
 *  import 端不回填，直接渲染会产出 dir="undefined" / calc(undefined * …)。 */
function normalizeElementDefaults(json: unknown): unknown {
  if (Array.isArray(json)) {
    json.forEach(normalizeElementDefaults);
    return json;
  }
  if (json && typeof json === 'object') {
    const node = json as Record<string, any>;
    if (Array.isArray(node.children)) {
      if (node.direction === undefined) node.direction = null;
      if (node.indent === undefined) node.indent = 0;
      node.children.forEach(normalizeElementDefaults);
    }
  }
  return json;
}

/** 校验 value 是否为合法 editorState JSON；非法（空/旧数据）时回退为空文档 */
const resolveInitialState = (value: string): string => {
  const trimmed = (value || '').trim();
  if (trimmed.startsWith('{')) {
    try {
      return JSON.stringify(normalizeElementDefaults(JSON.parse(trimmed)));
    } catch {
      /* 落到空文档兜底 */
    }
  }
  return EMPTY_STATE_JSON;
};

/** 代码语法高亮（@lexical/code 的 Prism 注册） */
function CodeHighlightPlugin() {
  const [editor] = useLexicalComposerContext();
  useEffect(() => registerCodeHighlighting(editor), [editor]);
  return null;
}

/** 变更 → 导出 editorState JSON 回调（内部组件才能拿 editor context） */
function StateExporter({ onJson }: { onJson: (json: string) => void }) {
  useLexicalComposerContext();
  const handleChange = useCallback(
    (editorState: EditorState) => {
      onJson(JSON.stringify(editorState.toJSON()));
    },
    [onJson]
  );
  return <OnChangePlugin onChange={handleChange} />;
}

/** 兜底错误边界（Lexical 装饰节点需要，含 onError 回调） */
class LexicalErrorBoundary extends React.Component<
  { children?: React.ReactNode; onError: (error: Error) => void },
  { error: unknown }
> {
  state = { error: null as unknown };
  static getDerivedStateFromError(error: unknown) {
    return { error };
  }
  componentDidCatch(error: unknown) {
    this.props.onError(error as Error);
    console.error('[Lexical] render error', error);
  }
  render() {
    if (this.state.error) return <div className="lex-error">编辑器渲染异常，请刷新重试</div>;
    return this.props.children;
  }
}

/** 代码高亮 token 的主题类名（供 registerCodeHighlighting 使用） */
const codeTheme: Record<string, string> = {
  atrule: 'lex-tok-atrule',
  attr: 'lex-tok-attr',
  boolean: 'lex-tok-boolean',
  builtin: 'lex-tok-builtin',
  cdata: 'lex-tok-cdata',
  char: 'lex-tok-char',
  class: 'lex-tok-class',
  'class-name': 'lex-tok-class-name',
  comment: 'lex-tok-comment',
  constant: 'lex-tok-constant',
  deleted: 'lex-tok-deleted',
  doctype: 'lex-tok-doctype',
  entity: 'lex-tok-entity',
  function: 'lex-tok-function',
  important: 'lex-tok-important',
  inserted: 'lex-tok-inserted',
  keyword: 'lex-tok-keyword',
  namespace: 'lex-tok-namespace',
  number: 'lex-tok-number',
  operator: 'lex-tok-operator',
  property: 'lex-tok-property',
  punctuation: 'lex-tok-punctuation',
  regex: 'lex-tok-regex',
  selector: 'lex-tok-selector',
  string: 'lex-tok-string',
  symbol: 'lex-tok-symbol',
  tag: 'lex-tok-tag',
  url: 'lex-tok-url',
  variable: 'lex-tok-variable',
};

/**
 * Lexical 富文本编辑器，Notion 风格：斜杠命令、Markdown 快捷键、块手柄拖拽、
 * 待办/提示框/折叠块/表格/代码高亮、顶部工具栏 + 选中浮动格式栏。
 * 内容以 Lexical editorState JSON 存取（无损往返；预览由 RichTextPreview 经
 * toHtml 序列化渲染）。粘贴外部 HTML 仍走内置 importDOM 转换。
 */
const LexicalEditor: React.FC<LexicalEditorProps> = ({ value, onChange, placeholder, minHeight = 300 }) => {
  const token = useThemeToken();
  // value 变化（打开另一文档/新建）时用 key 重建编辑器，从新 HTML 导入
  const valueRef = useRef(value);
  const [instanceKey, setInstanceKey] = React.useState(0);
  const [paperEl, setPaperEl] = useState<HTMLDivElement | null>(null);
  // 共享历史栈：HistoryPlugin 与顶部工具栏（撤销/重做按钮可用性）共用同一实例。
  // 依赖 instanceKey：切换文档重建编辑器时清空旧文档的历史栈，避免跨文档撤销。
  // eslint-disable-next-line react-hooks/exhaustive-deps -- instanceKey 是刻意的重建信号，createEmptyHistoryState 本身无依赖
  const historyState = React.useMemo(() => createEmptyHistoryState(), [instanceKey]);
  useEffect(() => {
    if (value !== valueRef.current) {
      valueRef.current = value;
      setInstanceKey((k) => k + 1);
    }
  }, [value]);

  const initialConfig = {
    namespace: 'doc-editor',
    onError: (error: Error) => console.error('[Lexical]', error),
    nodes: [
      HeadingNode,
      QuoteNode,
      ListNode,
      ListItemNode,
      CodeNode,
      CodeHighlightNode,
      LinkNode,
      HorizontalRuleNode,
      TableNode,
      TableRowNode,
      TableCellNode,
      CalloutNode,
      ToggleNode,
      ImageNode,
    ],
    theme: {
      code: 'lex-code',
      codeHighlight: codeTheme,
      // 下划线/删除线在 0.49 不再渲染为 <u>/<s> 标签，而是走 theme.text 的 CSS 类；
      // 必须提供类名否则格式位已置但视觉无效果。underlineStrikethrough 处理两者叠加。
      text: {
        underline: 'lex-text-underline',
        strikethrough: 'lex-text-strikethrough',
        underlineStrikethrough: 'lex-text-underline-strikethrough',
      },
      // 待办 checkbox 的可见/可点击区域（由 ListItemNode 渲染为 li 上的主题类）
      list: {
        listitemChecked: 'lex-checklist-item lex-checklist-item-checked',
        listitemUnchecked: 'lex-checklist-item',
      },
    },
    // editorState JSON 字符串 → parseEditorState → setEditorState。
    // 不再走 HTML 往返，自定义节点（callout/toggle 等）靠 exportJSON/importJSON 无损还原；
    // 待办 checkbox 由 ListPlugin 主题类渲染，无需 html.export 覆盖（跨应用复制粘贴除外）。
    editorState: resolveInitialState(value),
  };

  const handleJson = useCallback(
    (json: string) => {
      if (json !== valueRef.current) {
        valueRef.current = json;
        onChange(json);
      }
    },
    [onChange]
  );

  return (
    <div
      className="lex-editor"
      style={{
        ['--lex-primary' as string]: token.colorPrimary,
        ['--lex-border' as string]: token.colorBorder,
        ['--lex-bg' as string]: token.colorBgContainer,
        ['--lex-text' as string]: token.colorText,
        ['--lex-placeholder' as string]: token.colorTextPlaceholder,
        ['--lex-radius' as string]: `${token.borderRadiusLG}px`,
        ['--lex-bg-secondary' as string]: token.colorBgLayout,
      }}
    >
      <LexicalComposer key={instanceKey} initialConfig={initialConfig}>
        <div className="lex-content">
          {/* 常驻顶部工具栏（参照 Lexical playground） */}
          <ToolbarPlugin historyState={historyState} />
          <div className="lex-paper" ref={(el) => setPaperEl(el)}>
            <RichTextPlugin
              contentEditable={<ContentEditable className="lex-editable" style={{ minHeight }} />}
              placeholder={<div className="lex-placeholder">{placeholder || '请输入内容…'}</div>}
              ErrorBoundary={LexicalErrorBoundary as never}
            />
          </div>
        </div>
        <HistoryPlugin externalHistoryState={historyState} />
        <ListPlugin />
        <LinkPlugin />
        <CodeHighlightPlugin />
        <MarkdownShortcutPlugin transformers={DEFAULT_TRANSFORMERS} />
        <CheckListPlugin />
        <TablePlugin hasHorizontalScroll />
        <SlashMenuPlugin />
        <FloatingFormatToolbar />
        <TableActionMenuPlugin />
        {paperEl && <DraggableBlocks anchorElem={paperEl} />}
        <StateExporter onJson={handleJson} />
      </LexicalComposer>
    </div>
  );
};

export default LexicalEditor;
