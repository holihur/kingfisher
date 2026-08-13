import React, { useRef } from 'react';
import { DraggableBlockPlugin_EXPERIMENTAL } from '@lexical/react/LexicalDraggableBlockPlugin';

/**
 * Notion 风格块手柄：悬停块左侧显示拖拽手柄，拖拽可上下移动块位置。
 * 手柄与插入指示线的定位/显隐由官方 DraggableBlockPlugin_EXPERIMENTAL 管理，
 * 我们只提供元素与样式（手柄默认透明，插件在悬停块时将其置为可见）。
 */
export function DraggableBlocks({ anchorElem }: { anchorElem: HTMLElement }): React.JSX.Element {
  const menuRef = useRef<HTMLDivElement>(null);
  const targetLineRef = useRef<HTMLDivElement>(null);

  return (
    <DraggableBlockPlugin_EXPERIMENTAL
      anchorElem={anchorElem}
      menuRef={menuRef}
      targetLineRef={targetLineRef}
      menuComponent={
        <div ref={menuRef} className="lex-block-handle" draggable>
          <span className="lex-block-handle-icon">⠿</span>
        </div>
      }
      targetLineComponent={<div ref={targetLineRef} className="lex-block-target-line" />}
      isOnMenu={(el) => el.classList?.contains('lex-block-handle') || el.classList?.contains('lex-block-handle-icon')}
    />
  );
}
