import React, { useCallback, useEffect, useRef } from 'react';
import {
  BoldOutlined,
  ItalicOutlined,
  UnderlineOutlined,
  StrikethroughOutlined,
  OrderedListOutlined,
  UnorderedListOutlined,
  ClearOutlined,
  UndoOutlined,
  RedoOutlined,
} from '@ant-design/icons';
import { Button, Divider, Select, Tooltip } from 'antd';
import {
  $createParagraphNode,
  $getRoot,
  $getSelection,
  $isRangeSelection,
  EditorState,
  FORMAT_ELEMENT_COMMAND,
  FORMAT_TEXT_COMMAND,
  REDO_COMMAND,
  UNDO_COMMAND,
} from 'lexical';
import { $setBlocksType } from '@lexical/selection';
import { HeadingNode, QuoteNode, $createHeadingNode } from '@lexical/rich-text';
import { INSERT_ORDERED_LIST_COMMAND, INSERT_UNORDERED_LIST_COMMAND, ListNode, ListItemNode } from '@lexical/list';
import { LexicalComposer } from '@lexical/react/LexicalComposer';
import { ContentEditable } from '@lexical/react/LexicalContentEditable';
import { HistoryPlugin } from '@lexical/react/LexicalHistoryPlugin';
import { RichTextPlugin } from '@lexical/react/LexicalRichTextPlugin';
import { OnChangePlugin } from '@lexical/react/LexicalOnChangePlugin';
import { ListPlugin } from '@lexical/react/LexicalListPlugin';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { $generateHtmlFromNodes, $generateNodesFromDOM } from '@lexical/html';
import { useThemeToken } from '../hooks/useThemeToken';

interface LexicalEditorProps {
  value: string; // HTML 内容（编辑时导入，变更时导出）
  onChange: (html: string) => void;
  placeholder?: string;
  minHeight?: number;
}

/** 用 HTML 字符串初始化/重置 Lexical 根节点 */
function $setRootFromHTML(editor: Parameters<typeof $generateNodesFromDOM>[0], html: string) {
  const dom = new DOMParser().parseFromString(html || '<p></p>', 'text/html');
  const nodes = $generateNodesFromDOM(editor as never, dom);
  const root = $getRoot();
  root.clear();
  root.append(...nodes);
}

/** antd 风格工具栏 + 变更导出（放在 LexicalComposer 内以访问 editor） */
function Toolbar() {
  const [editor] = useLexicalComposerContext();
  const token = useThemeToken();

  const fmt = (type: Parameters<typeof editor.dispatchCommand>[0], value?: unknown) => (e: React.MouseEvent) => {
    e.preventDefault(); // 避免 blur 丢失选区
    editor.dispatchCommand(type as never, value as never);
  };

  return (
    <div className="lex-toolbar" style={{ borderBottom: `1px solid ${token.colorBorder}` }}>
      <Select
        size="small"
        defaultValue="paragraph"
        style={{ width: 84, marginRight: 8 }}
        options={[
          { value: 'paragraph', label: '段落' },
          { value: 'h1', label: '标题 1' },
          { value: 'h2', label: '标题 2' },
          { value: 'h3', label: '标题 3' },
        ]}
        onMouseDown={(e) => e.preventDefault()}
        onChange={(v) => {
          editor.update(() => {
            const selection = $getSelection();
            if (!$isRangeSelection(selection)) return;
            if (v === 'paragraph') {
              $setBlocksType(selection, () => $createParagraphNode());
            } else {
              const level = Number(v.replace('h', '')) as 1 | 2 | 3;
              $setBlocksType(selection, () => $createHeadingNode(`h${level}`));
            }
          });
        }}
      />
      <Divider orientation="vertical" />
      <Tooltip title="加粗">
        <Button type="text" size="small" icon={<BoldOutlined />} onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'bold')} />
      </Tooltip>
      <Tooltip title="斜体">
        <Button type="text" size="small" icon={<ItalicOutlined />} onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'italic')} />
      </Tooltip>
      <Tooltip title="下划线">
        <Button
          type="text"
          size="small"
          icon={<UnderlineOutlined />}
          onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'underline')}
        />
      </Tooltip>
      <Tooltip title="删除线">
        <Button
          type="text"
          size="small"
          icon={<StrikethroughOutlined />}
          onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'strikethrough')}
        />
      </Tooltip>
      <Divider orientation="vertical" />
      <Tooltip title="无序列表">
        <Button
          type="text"
          size="small"
          icon={<UnorderedListOutlined />}
          onMouseDown={fmt(INSERT_UNORDERED_LIST_COMMAND)}
        />
      </Tooltip>
      <Tooltip title="有序列表">
        <Button
          type="text"
          size="small"
          icon={<OrderedListOutlined />}
          onMouseDown={fmt(INSERT_ORDERED_LIST_COMMAND)}
        />
      </Tooltip>
      <Tooltip title="引用块">
        <Button type="text" size="small" onMouseDown={fmt(FORMAT_ELEMENT_COMMAND, 'quote')}>
          ❝
        </Button>
      </Tooltip>
      <Divider orientation="vertical" />
      <Tooltip title="撤销">
        <Button type="text" size="small" icon={<UndoOutlined />} onMouseDown={fmt(UNDO_COMMAND)} />
      </Tooltip>
      <Tooltip title="重做">
        <Button type="text" size="small" icon={<RedoOutlined />} onMouseDown={fmt(REDO_COMMAND)} />
      </Tooltip>
      <Tooltip title="清除格式">
        <Button type="text" size="small" icon={<ClearOutlined />} onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'clear')} />
      </Tooltip>
    </div>
  );
}

/** 变更 → 导出 HTML 回调（内部组件才能拿 editor） */
function HtmlExporter({ onHtml }: { onHtml: (html: string) => void }) {
  const [editor] = useLexicalComposerContext();
  const handleChange = useCallback(
    (editorState: EditorState) => {
      editorState.read(() => {
        const html = $generateHtmlFromNodes(editor);
        onHtml(html);
      });
    },
    [editor, onHtml]
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

/**
 * Lexical 富文本编辑器，antd 风格，兼容项目 HTML 存储。
 * 编辑内容以 HTML 导入/导出（与后端 content 字段、XSS 管线、RichTextPreview 兼容）。
 */
const LexicalEditor: React.FC<LexicalEditorProps> = ({ value, onChange, placeholder, minHeight = 300 }) => {
  const token = useThemeToken();
  // value 变化（打开另一文档/新建）时用 key 重建编辑器，从新 HTML 导入
  const valueRef = useRef(value);
  const [instanceKey, setInstanceKey] = React.useState(0);
  useEffect(() => {
    if (value !== valueRef.current) {
      valueRef.current = value;
      setInstanceKey((k) => k + 1);
    }
  }, [value]);

  const initialConfig = {
    namespace: 'doc-editor',
    onError: (error: Error) => console.error('[Lexical]', error),
    nodes: [HeadingNode, QuoteNode, ListNode, ListItemNode],
    // HTML → Lexical（打开文档时导入）
    editorState: (editor: Parameters<typeof $generateNodesFromDOM>[0]) => {
      $setRootFromHTML(editor, valueRef.current);
    },
  };

  const handleHtml = useCallback(
    (html: string) => {
      if (html !== valueRef.current) {
        valueRef.current = html;
        onChange(html);
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
      }}
    >
      <LexicalComposer key={instanceKey} initialConfig={initialConfig}>
        <Toolbar />
        <div className="lex-content">
          <RichTextPlugin
            contentEditable={<ContentEditable className="lex-editable" style={{ minHeight }} />}
            placeholder={<div className="lex-placeholder">{placeholder || '请输入内容…'}</div>}
            ErrorBoundary={LexicalErrorBoundary as never}
          />
        </div>
        <HistoryPlugin />
        <ListPlugin />
        <HtmlExporter onHtml={handleHtml} />
      </LexicalComposer>
    </div>
  );
};

export default LexicalEditor;
