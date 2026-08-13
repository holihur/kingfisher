import React, { useCallback, useEffect, useState } from 'react';
import { $getSelection, $isRangeSelection, FORMAT_TEXT_COMMAND } from 'lexical';
import { CodeNode } from '@lexical/code';
import { $createLinkNode } from '@lexical/link';
import { $getNearestNodeOfType } from '@lexical/utils';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { Button, Divider, Input, Modal, Select, Tooltip } from 'antd';
import {
  BoldOutlined,
  ItalicOutlined,
  UnderlineOutlined,
  StrikethroughOutlined,
  CodeOutlined,
  LinkOutlined,
} from '@ant-design/icons';
import { CODE_LANGUAGES } from './blocks';

const TOOLBAR_WIDTH = 300;
const TOOLBAR_HEIGHT = 36;

/** Notion 风格浮动格式栏：选中文字时出现在选区上方；光标在代码块内时切换语言 */
export function FloatingFormatToolbar(): React.JSX.Element | null {
  const [editor] = useLexicalComposerContext();
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null);
  const [active, setActive] = useState<{
    bold?: boolean;
    italic?: boolean;
    underline?: boolean;
    strikethrough?: boolean;
    code?: boolean;
  }>({});
  const [inCode, setInCode] = useState(false);
  const [codeLang, setCodeLang] = useState('');
  const [linkModal, setLinkModal] = useState(false);
  const [linkUrl, setLinkUrl] = useState('');

  const update = useCallback(() => {
    const root = editor.getRootElement();
    const domSel = window.getSelection?.();
    const inEditor = !!root && !!domSel && domSel.rangeCount > 0 && root.contains(domSel.anchorNode);
    if (!inEditor || !domSel || domSel.isCollapsed) {
      setPos(null);
      setInCode(false);
      return;
    }
    const rect = domSel.getRangeAt(0).getBoundingClientRect();
    if (rect.width === 0 && rect.height === 0) {
      setPos(null);
      return;
    }
    // 计算浮动栏位置（默认显示在选区上方，空间不足时翻转到下方）
    const vw = window.innerWidth;
    const top = rect.top - TOOLBAR_HEIGHT - 10 < 8 ? rect.bottom + 10 : rect.top - TOOLBAR_HEIGHT - 10;
    const left = Math.max(8, Math.min(rect.left + rect.width / 2 - TOOLBAR_WIDTH / 2, vw - TOOLBAR_WIDTH - 8));
    setPos({ top, left });

    editor.getEditorState().read(() => {
      const selection = $getSelection();
      if (!$isRangeSelection(selection)) {
        setInCode(false);
        return;
      }
      const anchorNode = selection.anchor.getNode();
      const codeNode = anchorNode ? $getNearestNodeOfType(anchorNode, CodeNode) : null;
      setInCode(!!codeNode);
      if (codeNode) {
        setCodeLang(codeNode.getLanguage() || 'plaintext');
      } else {
        setActive({
          bold: selection.hasFormat('bold'),
          italic: selection.hasFormat('italic'),
          underline: selection.hasFormat('underline'),
          strikethrough: selection.hasFormat('strikethrough'),
          code: selection.hasFormat('code'),
        });
      }
    });
  }, [editor]);

  useEffect(() => {
    update();
    const unload = editor.registerUpdateListener(update);
    document.addEventListener('selectionchange', update);
    window.addEventListener('scroll', update, true);
    return () => {
      unload();
      document.removeEventListener('selectionchange', update);
      window.removeEventListener('scroll', update, true);
    };
  }, [editor, update]);

  const fmt = (type: Parameters<typeof editor.dispatchCommand>[0], value?: unknown) => (e: React.MouseEvent) => {
    e.preventDefault();
    editor.dispatchCommand(type as never, value as never);
  };

  const changeLang = (lang: string) => {
    editor.update(() => {
      const selection = $getSelection();
      if (!$isRangeSelection(selection)) return;
      const codeNode = $getNearestNodeOfType(selection.anchor.getNode(), CodeNode);
      codeNode?.setLanguage(lang);
    });
  };

  const submitLink = () => {
    const url = linkUrl.trim();
    if (!url) {
      setLinkModal(false);
      return;
    }
    editor.update(() => {
      const selection = $getSelection();
      if (!$isRangeSelection(selection)) return;
      const link = $createLinkNode(url);
      selection.insertNodes([link]);
    });
    setLinkModal(false);
    setLinkUrl('');
  };

  if (!pos) return null;

  return (
    <>
      <div
        className="lex-float-toolbar"
        style={{ top: pos.top, left: pos.left }}
        onMouseDown={(e) => e.preventDefault()}
      >
        {inCode ? (
          <Select
            size="small"
            value={codeLang}
            onChange={changeLang}
            options={CODE_LANGUAGES}
            style={{ width: 200 }}
            popupMatchSelectWidth={false}
          />
        ) : (
          <>
            <Tooltip title="加粗">
              <Button
                type="text"
                size="small"
                className={active.bold ? 'is-active' : ''}
                icon={<BoldOutlined />}
                onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'bold')}
              />
            </Tooltip>
            <Tooltip title="斜体">
              <Button
                type="text"
                size="small"
                className={active.italic ? 'is-active' : ''}
                icon={<ItalicOutlined />}
                onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'italic')}
              />
            </Tooltip>
            <Tooltip title="下划线">
              <Button
                type="text"
                size="small"
                className={active.underline ? 'is-active' : ''}
                icon={<UnderlineOutlined />}
                onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'underline')}
              />
            </Tooltip>
            <Tooltip title="删除线">
              <Button
                type="text"
                size="small"
                className={active.strikethrough ? 'is-active' : ''}
                icon={<StrikethroughOutlined />}
                onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'strikethrough')}
              />
            </Tooltip>
            <Divider orientation="vertical" />
            <Tooltip title="行内代码">
              <Button
                type="text"
                size="small"
                className={active.code ? 'is-active' : ''}
                icon={<CodeOutlined />}
                onMouseDown={fmt(FORMAT_TEXT_COMMAND, 'code')}
              />
            </Tooltip>
            <Tooltip title="插入链接">
              <Button
                type="text"
                size="small"
                icon={<LinkOutlined />}
                onMouseDown={(e) => {
                  e.preventDefault();
                  setLinkModal(true);
                }}
              />
            </Tooltip>
          </>
        )}
      </div>
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
    </>
  );
}
