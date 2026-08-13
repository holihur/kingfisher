import React, { useCallback, useEffect, useState } from 'react';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import type { HistoryState } from '@lexical/react/LexicalHistoryPlugin';
import {
  $getSelection,
  $isBlockElementNode,
  $isElementNode,
  $isRangeSelection,
  $isTextNode,
  CAN_REDO_COMMAND,
  CAN_UNDO_COMMAND,
  COMMAND_PRIORITY_LOW,
  FORMAT_ELEMENT_COMMAND,
  FORMAT_TEXT_COMMAND,
  REDO_COMMAND,
  SELECTION_CHANGE_COMMAND,
  UNDO_COMMAND,
  type ElementFormatType,
  type LexicalCommand,
  type LexicalNode,
  type RangeSelection,
} from 'lexical';
import { $isHeadingNode, $isQuoteNode } from '@lexical/rich-text';
import { $isListItemNode, $isListNode } from '@lexical/list';
import { $isCodeNode } from '@lexical/code';
import { $patchStyleText } from '@lexical/selection';
import { $isLinkNode } from '@lexical/link';
import { $createLinkNode } from '@lexical/link';
import { Button, ColorPicker, Divider, Dropdown, Input, Modal, Tooltip } from 'antd';
import {
  AlignCenterOutlined,
  AlignLeftOutlined,
  AlignRightOutlined,
  BgColorsOutlined,
  BoldOutlined,
  BulbOutlined,
  CheckSquareOutlined,
  ClearOutlined,
  CodeOutlined,
  DownOutlined,
  FontColorsOutlined,
  ItalicOutlined,
  LinkOutlined,
  MinusOutlined,
  PictureOutlined,
  OrderedListOutlined,
  RedoOutlined,
  StrikethroughOutlined,
  TableOutlined,
  UnderlineOutlined,
  UndoOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import {
  insertHeading,
  insertParagraph,
  insertQuote,
  insertCodeBlock,
  insertBulletList,
  insertNumberedList,
  insertCheckList,
  insertCallout,
  insertToggle,
  insertTable,
  insertHorizontalRule,
  insertImage,
  type InsertFn,
} from './blocks';

type BlockType = 'paragraph' | 'h1' | 'h2' | 'h3' | 'quote' | 'code';
type ListType = 'bullet' | 'number' | 'check' | null;

const BLOCK_OPTIONS: { key: BlockType; label: string; action: InsertFn }[] = [
  { key: 'paragraph', label: '正文', action: insertParagraph },
  { key: 'h1', label: '标题 1', action: insertHeading('h1') },
  { key: 'h2', label: '标题 2', action: insertHeading('h2') },
  { key: 'h3', label: '标题 3', action: insertHeading('h3') },
  { key: 'quote', label: '引用', action: insertQuote },
  { key: 'code', label: '代码块', action: insertCodeBlock },
];

const ALIGN_OPTIONS: { key: ElementFormatType; label: string; icon: React.ReactNode }[] = [
  { key: 'left', label: '左对齐', icon: <AlignLeftOutlined /> },
  { key: 'center', label: '居中', icon: <AlignCenterOutlined /> },
  { key: 'right', label: '右对齐', icon: <AlignRightOutlined /> },
  { key: 'justify', label: '两端对齐', icon: <AlignLeftOutlined /> },
];

const INSERT_OPTIONS: { key: string; label: string; icon: React.ReactNode; action: InsertFn }[] = [
  { key: 'image', label: '图片', icon: <PictureOutlined />, action: insertImage },
  { key: 'hr', label: '分割线', icon: <MinusOutlined />, action: insertHorizontalRule },
  { key: 'table', label: '表格', icon: <TableOutlined />, action: insertTable(3, 3) },
  { key: 'callout', label: '提示框', icon: <BulbOutlined />, action: insertCallout },
  {
    key: 'toggle',
    label: '折叠块',
    icon: <span style={{ fontSize: 13, fontWeight: 700 }}>▾</span>,
    action: insertToggle,
  },
];

interface ToolbarState {
  bold: boolean;
  italic: boolean;
  underline: boolean;
  strikethrough: boolean;
  code: boolean;
}

const EMPTY_FORMATS: ToolbarState = { bold: false, italic: false, underline: false, strikethrough: false, code: false };

/** 选区是否落在链接内（锚点或其父级为 LinkNode） */
function $isSelectionInLink(selection: RangeSelection): boolean {
  const anchor = selection.anchor.getNode();
  return !!anchor && ($isLinkNode(anchor) || $isLinkNode(anchor.getParent()));
}

/**
 * 常驻顶部工具栏（参照 Lexical playground 的 ActionToolbar）：
 * 撤销/重做、块类型下拉（正文/标题/引用/代码块）、加粗/斜体/下划线/删除线/行内代码、
 * 文字颜色/背景色、清除格式、链接、无序/有序/待办列表、对齐、插入（分割线/表格/提示框/折叠块）。
 * 按钮均带 active 状态；整个容器拦截 mousedown 防止选区/焦点丢失。
 */
export function ToolbarPlugin({ historyState }: { historyState: HistoryState }): React.JSX.Element {
  const [editor] = useLexicalComposerContext();
  const [canUndo, setCanUndo] = useState(false);
  const [canRedo, setCanRedo] = useState(false);
  const [formats, setFormats] = useState<ToolbarState>(EMPTY_FORMATS);
  const [blockType, setBlockType] = useState<BlockType>('paragraph');
  const [listType, setListType] = useState<ListType>(null);
  const [align, setAlign] = useState<ElementFormatType>('left');
  const [linkActive, setLinkActive] = useState(false);
  const [linkModal, setLinkModal] = useState(false);
  const [linkUrl, setLinkUrl] = useState('');

  const refreshUndoRedo = useCallback(() => {
    setCanUndo(historyState.undoStack.length > 0);
    setCanRedo(historyState.redoStack.length > 0);
  }, [historyState]);

  // 读取当前选区，刷新各按钮的 active 状态（在命令/更新上下文内调用，可用 $getSelection）
  const updateToolbar = useCallback(() => {
    const selection = $getSelection();
    if (!$isRangeSelection(selection)) {
      setFormats(EMPTY_FORMATS);
      setBlockType('paragraph');
      setListType(null);
      setAlign('left');
      setLinkActive(false);
      return;
    }
    setFormats({
      bold: selection.hasFormat('bold'),
      italic: selection.hasFormat('italic'),
      underline: selection.hasFormat('underline'),
      strikethrough: selection.hasFormat('strikethrough'),
      code: selection.hasFormat('code'),
    });
    setLinkActive($isSelectionInLink(selection));

    // 找锚点所在块元素（文本 → 元素 → 一直上溯到首个块）
    let node: LexicalNode | null = selection.anchor.getNode();
    while (node && !$isBlockElementNode(node)) node = node.getParent();
    // 块类型：列表项还要再看其父级列表
    if ($isListItemNode(node)) {
      const list = node.getParent();
      setBlockType('paragraph');
      setListType($isListNode(list) ? list.getListType() : null);
    } else if ($isListNode(node)) {
      setBlockType('paragraph');
      setListType(node.getListType());
    } else if ($isHeadingNode(node)) {
      setBlockType(node.getTag() as BlockType);
      setListType(null);
    } else if ($isQuoteNode(node)) {
      setBlockType('quote');
      setListType(null);
    } else if ($isCodeNode(node)) {
      setBlockType('code');
      setListType(null);
    } else {
      setBlockType('paragraph');
      setListType(null);
    }
    // 对齐：块元素的 formatType（start/end 是 LTR/RTL 默认，归一化为 left/right）
    const formatType = node && $isElementNode(node) ? node.getFormatType() : 'left';
    setAlign(formatType === 'start' ? 'left' : formatType === 'end' ? 'right' : formatType);
  }, []);

  useEffect(() => {
    // 选区变化时刷新按钮状态
    const unregisterSelection = editor.registerCommand(
      SELECTION_CHANGE_COMMAND,
      () => {
        updateToolbar();
        return false;
      },
      COMMAND_PRIORITY_LOW
    );
    // HistoryPlugin 在撤销栈变化时派发 CAN_UNDO/CAN_REDO，据此刷新按钮可用性
    const unregisterUndo = editor.registerCommand(
      CAN_UNDO_COMMAND,
      (payload) => {
        setCanUndo(payload);
        return false;
      },
      COMMAND_PRIORITY_LOW
    );
    const unregisterRedo = editor.registerCommand(
      CAN_REDO_COMMAND,
      (payload) => {
        setCanRedo(payload);
        return false;
      },
      COMMAND_PRIORITY_LOW
    );
    // 兜底：内容更新后延迟一拍读取历史栈（HistoryPlugin 的 applyChange 与本组件
    // 都是 update listener，且注册更晚，同步读会拿到推栈前的旧长度）
    const unregisterUpdate = editor.registerUpdateListener(() => {
      setTimeout(() => refreshUndoRedo(), 0);
    });
    // 首次挂载读一次初始状态
    editor.getEditorState().read(updateToolbar);
    return () => {
      unregisterSelection();
      unregisterUndo();
      unregisterRedo();
      unregisterUpdate();
    };
  }, [editor, updateToolbar, refreshUndoRedo]);

  /** 执行一个带 payload 的 Lexical 命令（mousedown 阶段触发，保留选区） */
  const runCommand =
    <T,>(command: LexicalCommand<T>, payload?: T) =>
    (e: React.MouseEvent) => {
      e.preventDefault();
      // dispatchCommand 对 void/TextFormatType 的 payload 签名不一致，这里统一断言
      editor.dispatchCommand(command as never, payload as never);
    };

  /** 执行块插入/转换动作 */
  const runInsert = (action: InsertFn) => (e: React.MouseEvent) => {
    e.preventDefault();
    action(editor);
  };

  /** 清除文字格式（加粗/颜色等）：遍历选区文本节点重置 format 与行内 style */
  const clearFormatting = (e: React.MouseEvent) => {
    e.preventDefault();
    editor.update(() => {
      const selection = $getSelection();
      if (!$isRangeSelection(selection) || selection.isCollapsed()) return;
      for (const node of selection.getNodes()) {
        if ($isTextNode(node)) {
          node.setFormat(0);
          node.setStyle('');
        }
      }
    });
  };

  /** 文字颜色 / 背景色。注意 key 必须是 kebab-case：Lexical 渲染用
   *  CSSStyleDeclaration.setProperty()，传 camelCase（backgroundColor）会静默无效。 */
  const applyColor = (styleKey: 'color' | 'background-color') => (hex: string) => {
    editor.update(() => {
      const selection = $getSelection();
      if ($isRangeSelection(selection)) $patchStyleText(selection, { [styleKey]: hex });
    });
  };

  const submitLink = () => {
    const url = linkUrl.trim();
    setLinkModal(false);
    setLinkUrl('');
    if (!url) return;
    editor.update(() => {
      const selection = $getSelection();
      if ($isRangeSelection(selection)) selection.insertNodes([$createLinkNode(url)]);
    });
  };

  const blockLabel = BLOCK_OPTIONS.find((o) => o.key === blockType)?.label ?? '正文';
  const alignIcon = ALIGN_OPTIONS.find((o) => o.key === align)?.icon ?? <AlignLeftOutlined />;

  return (
    <div className="lex-toolbar" onMouseDown={(e) => e.preventDefault()}>
      {/* 撤销 / 重做 */}
      <Tooltip title="撤销 (⌘Z)">
        <Button
          type="text"
          size="small"
          disabled={!canUndo}
          icon={<UndoOutlined />}
          onMouseDown={runCommand(UNDO_COMMAND)}
        />
      </Tooltip>
      <Tooltip title="重做 (⌘⇧Z)">
        <Button
          type="text"
          size="small"
          disabled={!canRedo}
          icon={<RedoOutlined />}
          onMouseDown={runCommand(REDO_COMMAND)}
        />
      </Tooltip>
      <Divider orientation="vertical" className="lex-toolbar-divider" />

      {/* 块类型 */}
      <Dropdown
        trigger={['click']}
        menu={{
          items: BLOCK_OPTIONS.map((o) => ({ key: o.key, label: o.label })),
          selectedKeys: [blockType],
          onClick: ({ key }) => BLOCK_OPTIONS.find((o) => o.key === key)?.action(editor),
        }}
      >
        <Button type="text" size="small" className="lex-toolbar-dropdown">
          {blockLabel} <DownOutlined style={{ fontSize: 10 }} />
        </Button>
      </Dropdown>
      <Divider orientation="vertical" className="lex-toolbar-divider" />

      {/* 文字格式 */}
      <Tooltip title="加粗 (⌘B)">
        <Button
          type="text"
          size="small"
          className={formats.bold ? 'is-active' : ''}
          icon={<BoldOutlined />}
          onMouseDown={runCommand(FORMAT_TEXT_COMMAND, 'bold')}
        />
      </Tooltip>
      <Tooltip title="斜体 (⌘I)">
        <Button
          type="text"
          size="small"
          className={formats.italic ? 'is-active' : ''}
          icon={<ItalicOutlined />}
          onMouseDown={runCommand(FORMAT_TEXT_COMMAND, 'italic')}
        />
      </Tooltip>
      <Tooltip title="下划线 (⌘U)">
        <Button
          type="text"
          size="small"
          className={formats.underline ? 'is-active' : ''}
          icon={<UnderlineOutlined />}
          onMouseDown={runCommand(FORMAT_TEXT_COMMAND, 'underline')}
        />
      </Tooltip>
      <Tooltip title="删除线">
        <Button
          type="text"
          size="small"
          className={formats.strikethrough ? 'is-active' : ''}
          icon={<StrikethroughOutlined />}
          onMouseDown={runCommand(FORMAT_TEXT_COMMAND, 'strikethrough')}
        />
      </Tooltip>
      <Tooltip title="行内代码">
        <Button
          type="text"
          size="small"
          className={formats.code ? 'is-active' : ''}
          icon={<CodeOutlined />}
          onMouseDown={runCommand(FORMAT_TEXT_COMMAND, 'code')}
        />
      </Tooltip>

      {/* 文字颜色 / 背景色 */}
      {/* antd 预设色点击只触发 onChange（不触发 onChangeComplete），用 onChange 应用颜色 */}
      <Tooltip title="文字颜色">
        <ColorPicker
          disabledAlpha
          size="small"
          presets={[
            {
              label: '推荐',
              colors: ['#e11d48', '#f97316', '#facc15', '#22c55e', '#0ea5e9', '#6366f1', '#a855f7', '#111827'],
            },
          ]}
          onChange={(c) => applyColor('color')(c.toHexString())}
        >
          <span className="lex-toolbar-color lex-toolbar-color-text" title="文字颜色">
            <FontColorsOutlined />
          </span>
        </ColorPicker>
      </Tooltip>
      <Tooltip title="背景色">
        <ColorPicker
          disabledAlpha
          size="small"
          presets={[
            {
              label: '推荐',
              colors: ['#fee2e2', '#ffedd5', '#fef9c3', '#dcfce7', '#e0f2fe', '#e0e7ff', '#f3e8ff', '#111827'],
            },
          ]}
          onChange={(c) => applyColor('background-color')(c.toHexString())}
        >
          <span className="lex-toolbar-color lex-toolbar-color-bg" title="背景色">
            <BgColorsOutlined />
          </span>
        </ColorPicker>
      </Tooltip>
      <Tooltip title="清除格式">
        <Button type="text" size="small" icon={<ClearOutlined />} onMouseDown={clearFormatting} />
      </Tooltip>
      <Divider orientation="vertical" className="lex-toolbar-divider" />

      {/* 链接 */}
      <Tooltip title={linkActive ? '编辑链接' : '插入链接'}>
        <Button
          type="text"
          size="small"
          className={linkActive ? 'is-active' : ''}
          icon={<LinkOutlined />}
          onMouseDown={(e) => {
            e.preventDefault();
            setLinkModal(true);
          }}
        />
      </Tooltip>
      <Divider orientation="vertical" className="lex-toolbar-divider" />

      {/* 列表 */}
      <Tooltip title="无序列表">
        <Button
          type="text"
          size="small"
          className={listType === 'bullet' ? 'is-active' : ''}
          icon={<UnorderedListOutlined />}
          onMouseDown={runInsert(insertBulletList)}
        />
      </Tooltip>
      <Tooltip title="有序列表">
        <Button
          type="text"
          size="small"
          className={listType === 'number' ? 'is-active' : ''}
          icon={<OrderedListOutlined />}
          onMouseDown={runInsert(insertNumberedList)}
        />
      </Tooltip>
      <Tooltip title="待办列表">
        <Button
          type="text"
          size="small"
          className={listType === 'check' ? 'is-active' : ''}
          icon={<CheckSquareOutlined />}
          onMouseDown={runInsert(insertCheckList)}
        />
      </Tooltip>

      {/* 对齐 */}
      <Dropdown
        trigger={['click']}
        menu={{
          items: ALIGN_OPTIONS.map((o) => ({ key: o.key, label: o.label, icon: o.icon })),
          selectedKeys: [align],
          onClick: ({ key }) => editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, key as ElementFormatType),
        }}
      >
        <Button type="text" size="small" className="lex-toolbar-dropdown">
          {alignIcon} <DownOutlined style={{ fontSize: 10 }} />
        </Button>
      </Dropdown>
      <Divider orientation="vertical" className="lex-toolbar-divider" />

      {/* 插入块 */}
      <Dropdown
        trigger={['click']}
        menu={{
          items: INSERT_OPTIONS.map((o) => ({ key: o.key, label: o.label, icon: o.icon })),
          onClick: ({ key }) => INSERT_OPTIONS.find((o) => o.key === key)?.action(editor),
        }}
      >
        <Button type="text" size="small" className="lex-toolbar-dropdown">
          <span style={{ fontSize: 13, fontWeight: 700 }}>＋</span> 插入 <DownOutlined style={{ fontSize: 10 }} />
        </Button>
      </Dropdown>

      {/* 插入链接弹窗 */}
      <Modal
        open={linkModal}
        title="插入链接"
        okText="确定"
        cancelText="取消"
        onOk={submitLink}
        onCancel={() => setLinkModal(false)}
        destroyOnHidden
      >
        <Input
          placeholder="https://example.com"
          value={linkUrl}
          onChange={(e) => setLinkUrl(e.target.value)}
          onPressEnter={submitLink}
          autoFocus
        />
      </Modal>
    </div>
  );
}

export default ToolbarPlugin;
