import React, { useCallback, useMemo, useState } from 'react';
import { createPortal } from 'react-dom';
import { TextNode } from 'lexical';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import {
  LexicalTypeaheadMenuPlugin,
  MenuOption,
  useBasicTypeaheadTriggerMatch,
} from '@lexical/react/LexicalTypeaheadMenuPlugin';
import {
  insertHeading,
  insertParagraph,
  insertQuote,
  insertBulletList,
  insertNumberedList,
  insertCheckList,
  insertCodeBlock,
  insertCallout,
  insertToggle,
  insertTable,
  InsertFn,
} from './blocks';

interface SlashOptionItem {
  title: string;
  description?: string;
  icon: React.ReactElement;
  action: InsertFn;
}

const SLASH_ITEMS: SlashOptionItem[] = [
  {
    title: '文本',
    description: '普通段落',
    icon: <span className="lex-slash-icon-text">T</span>,
    action: insertParagraph,
  },
  {
    title: '标题 1',
    description: '大标题',
    icon: <span className="lex-slash-icon-text">H1</span>,
    action: insertHeading('h1'),
  },
  {
    title: '标题 2',
    description: '中等标题',
    icon: <span className="lex-slash-icon-text">H2</span>,
    action: insertHeading('h2'),
  },
  {
    title: '标题 3',
    description: '小标题',
    icon: <span className="lex-slash-icon-text">H3</span>,
    action: insertHeading('h3'),
  },
  {
    title: '待办事项',
    description: '带复选框的列表',
    icon: <span className="lex-slash-icon-text">☑</span>,
    action: insertCheckList,
  },
  {
    title: '无序列表',
    description: '圆点列表',
    icon: <span className="lex-slash-icon-text">•</span>,
    action: insertBulletList,
  },
  {
    title: '有序列表',
    description: '编号列表',
    icon: <span className="lex-slash-icon-text">1.</span>,
    action: insertNumberedList,
  },
  {
    title: '引用',
    description: '引用一段内容',
    icon: <span className="lex-slash-icon-text">❝</span>,
    action: insertQuote,
  },
  {
    title: '代码块',
    description: '带语言的代码片段',
    icon: <span className="lex-slash-icon-text">&lt;/&gt;</span>,
    action: insertCodeBlock,
  },
  {
    title: '提示框',
    description: '带图标的高亮块',
    icon: <span className="lex-slash-icon-text">💡</span>,
    action: insertCallout,
  },
  {
    title: '折叠块',
    description: '可折叠收起的内容',
    icon: <span className="lex-slash-icon-text">▾</span>,
    action: insertToggle,
  },
  {
    title: '表格',
    description: '3 行 3 列表格',
    icon: <span className="lex-slash-icon-text">▦</span>,
    action: insertTable(3, 3),
  },
];

/** 斜杠选项：在 MenuOption 基础上携带插入动作 */
class SlashOption extends MenuOption {
  title: string;
  description?: string;
  icon: React.ReactElement;
  action: InsertFn;

  constructor(key: string, item: SlashOptionItem) {
    super(key);
    this.title = item.title;
    this.description = item.description;
    this.icon = item.icon;
    this.action = item.action;
  }
}

/** 删除光标前的 "/查询" 文本（斜杠命令占位） */
function $removeSlashTrigger(textNode: TextNode) {
  const text = textNode.getTextContent();
  const slash = text.lastIndexOf('/');
  if (slash === -1) return;
  textNode.spliceText(0, slash + 1, '', true);
}

/** Notion 风格斜杠命令菜单：输入 "/" 弹出块选择 */
export function SlashMenuPlugin(): React.JSX.Element | null {
  const [editor] = useLexicalComposerContext();
  // onQueryChange 在菜单关闭时回调 null，需允许 null 并兜底
  const [query, setQuery] = useState<string | null>('');

  const options = useMemo<SlashOption[]>(() => {
    const q = (query || '').trim().toLowerCase();
    return SLASH_ITEMS.filter(
      (item) => !q || item.title.toLowerCase().includes(q) || (item.description || '').toLowerCase().includes(q)
    ).map((item, i) => new SlashOption(`${item.title}-${i}`, item));
  }, [query]);

  const onSelectOption = useCallback(
    (option: SlashOption, textNodeContainingQuery: TextNode | null, closeMenu: () => void) => {
      // 先移除 "/查询" 文本（独立 update，避免嵌套 editor.update）
      if (textNodeContainingQuery) {
        editor.update(() => $removeSlashTrigger(textNodeContainingQuery));
      }
      option.action(editor);
      closeMenu();
    },
    [editor]
  );

  const triggerFn = useBasicTypeaheadTriggerMatch('/', { minLength: 0 });

  return (
    <LexicalTypeaheadMenuPlugin<SlashOption>
      options={options}
      onQueryChange={setQuery}
      onSelectOption={onSelectOption}
      triggerFn={triggerFn}
      anchorClassName="lex-slash-anchor"
      menuRenderFn={(
        anchorRef,
        { selectedIndex, selectOptionAndCleanUp, setHighlightedIndex, options: menuOptions }
      ) =>
        anchorRef.current
          ? createPortal(
              <ul className="lex-slash-menu" role="listbox" aria-label="插入块">
                {menuOptions.map((option, i) => (
                  <li
                    key={option.key}
                    role="option"
                    aria-selected={selectedIndex === i}
                    className={`lex-slash-item${selectedIndex === i ? ' is-selected' : ''}`}
                    onMouseEnter={() => setHighlightedIndex(i)}
                    onClick={() => selectOptionAndCleanUp(option)}
                    ref={option.setRefElement}
                  >
                    <span className="lex-slash-icon">{option.icon}</span>
                    <span className="lex-slash-body">
                      <span className="lex-slash-title">{option.title}</span>
                      {option.description && <span className="lex-slash-desc">{option.description}</span>}
                    </span>
                  </li>
                ))}
              </ul>,
              anchorRef.current
            )
          : null
      }
    />
  );
}
