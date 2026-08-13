import React, { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { Dropdown, InputNumber, Modal, Select, Space } from 'antd';
import {
  $createParagraphNode,
  $getNearestNodeFromDOMNode,
  $getNodeByKey,
  $getSelection,
  type LexicalNode,
} from 'lexical';
import {
  $deleteTableColumn,
  $getNodeTriplet,
  $getTableNodeFromLexicalNodeOrThrow,
  $getTableRowIndexFromTableCellNode,
  $getTableColumnIndexFromTableCellNode,
  $getTableCellNodeFromLexicalNode,
  $insertTableColumnAtNode,
  $insertTableRowAtNode,
  $isTableCellNode,
  $isTableRowNode,
  $isTableSelection,
  $mergeCells,
  $removeTableRowAtIndex,
  $setTableColumnIsHeader,
  $setTableRowIsHeader,
  $unmergeCellNode,
  TableCellNode,
} from '@lexical/table';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import {
  AlignCenterOutlined,
  AlignLeftOutlined,
  AlignRightOutlined,
  ClearOutlined,
  ColumnHeightOutlined,
  DeleteColumnOutlined,
  DeleteRowOutlined,
  DeleteOutlined,
  DownOutlined,
  InsertRowAboveOutlined,
  InsertRowLeftOutlined,
  MergeCellsOutlined,
  SplitCellsOutlined,
} from '@ant-design/icons';

/** 表格顶部悬浮的操作按钮 + 菜单（增删行/列、清空、表头、删除表格） */
interface TableMenuState {
  top: number;
  left: number;
  cellKey: string;
}

export function TableActionMenuPlugin(): React.JSX.Element | null {
  const [editor] = useLexicalComposerContext();
  const [menu, setMenu] = useState<TableMenuState | null>(null);
  // 下拉打开期间不移除按钮（避免鼠标移到菜单上时按钮/菜单一起消失）
  const openRef = useRef(false);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    // 用 document 级 capture 监听：编辑器插入块等结构化变更会重建 contentEditable 根元素，
    // 绑定在 getRootElement() 上的 listener 会失效；capture 阶段每次都能拿到真实 target。
    const handleMouseOver = (e: MouseEvent) => {
      const root = editor.getRootElement();
      const target = e.target as HTMLElement;
      if (!root || !root.contains(target)) return;
      const tableEl = target.closest('table');
      if (!tableEl) {
        if (!openRef.current) setMenu(null);
        return;
      }
      const cellEl = target.closest('td, th');
      const cellKey = editor.getEditorState().read(
        () => {
          // $getNearestNodeFromDOMNode 内部 $getNodeFromDOMNode 依赖 active editor，
          // read 需传 { editor } 提供该上下文（否则报 "Unable to find an active editor"）
          const start = cellEl ? $getNearestNodeFromDOMNode(cellEl, editor.getEditorState()) : null;
          let n: LexicalNode | null = start;
          while (n && !$getTableCellNodeFromLexicalNode(n)) n = n.getParent();
          return n ? n.getKey() : null;
        },
        { editor }
      );
      if (!cellKey) {
        if (!openRef.current) setMenu(null);
        return;
      }
      const rect = tableEl.getBoundingClientRect();
      setMenu((prev) => {
        // 同一表格不刷新坐标（避免悬停内部元素时按钮抖动）
        if (prev && prev.cellKey === cellKey) return prev;
        return { top: rect.top, left: rect.left, cellKey };
      });
    };
    document.addEventListener('mouseover', handleMouseOver, true);
    return () => document.removeEventListener('mouseover', handleMouseOver, true);
  }, [editor]);

  /** 对悬浮单元格执行编辑操作 */
  const runWithCell = (action: (cell: TableCellNode) => void) => {
    if (!menu) return;
    setOpen(false);
    editor.update(() => {
      const node = $getNodeByKey(menu.cellKey);
      const cell = node && $getTableCellNodeFromLexicalNode(node);
      if (cell) action(cell as TableCellNode);
    });
  };

  // 行/列多数量操作的弹窗状态：op = 插入/删除 行/列；dir = 方向（插行上/下，插列左/右）
  type CountOp = 'insert-row' | 'insert-col' | 'del-row' | 'del-col';
  type CountDir = 'before' | 'after';
  const [countOp, setCountOp] = useState<CountOp | null>(null);
  const [countDir, setCountDir] = useState<CountDir>('after');
  const [count, setCount] = useState(1);

  const openCountOp = (op: CountOp) => {
    setOpen(false);
    setCount(1);
    setCountDir(op === 'insert-row' || op === 'insert-col' ? 'after' : 'after');
    setCountOp(op);
  };

  /** 确认执行：单次 update 内循环，用持久的 table 引用 + 固定行列号操作 */
  const submitCountOp = () => {
    if (!countOp || !menu) return;
    const n = Math.max(1, Math.min(count, 20));
    const op = countOp;
    const dir = countDir;
    const cellKey = menu.cellKey;
    setCountOp(null);
    editor.update(() => {
      const firstNode = $getNodeByKey(cellKey);
      const firstCell = firstNode && $getTableCellNodeFromLexicalNode(firstNode);
      if (!firstCell) return;
      const [, , table] = $getNodeTriplet(firstCell);
      const rowIndex = $getTableRowIndexFromTableCellNode(firstCell);
      const colIndex = $getTableColumnIndexFromTableCellNode(firstCell);

      const firstRowCell = (): TableCellNode | null => {
        const firstRow = table.getFirstChild();
        return firstRow && $isTableRowNode(firstRow) ? (firstRow.getFirstChild() as TableCellNode | null) : null;
      };

      for (let i = 0; i < n; i++) {
        if (op === 'insert-row') {
          const c = firstRowCell();
          if (c) $insertTableRowAtNode(c, dir === 'after');
        } else if (op === 'insert-col') {
          const c = firstRowCell();
          if (c) $insertTableColumnAtNode(c, dir === 'after', false);
        } else if (op === 'del-row') {
          if (table.getChildrenSize() <= 1) break;
          // 删固定 index：下方行上移补位 → 连续删 N 行
          $removeTableRowAtIndex(table, Math.min(rowIndex, table.getChildrenSize() - 1));
        } else if (op === 'del-col') {
          const firstRow = table.getFirstChild();
          const cellCount = firstRow ? (firstRow as any).getChildrenSize?.() || 0 : 0;
          if (cellCount <= 1) break;
          $deleteTableColumn(table, Math.min(colIndex, cellCount - 1));
        }
      }
    });
  };

  // 当前是否有 TableSelection（拖选多个格子）——决定「合并单元格」可用性
  const hasMultiCellSelection = editor.getEditorState().read(() => {
    const sel = $getSelection();
    return $isTableSelection(sel) && sel.getNodes().filter($isTableCellNode).length >= 2;
  });

  /** 合并选中单元格（先拖选区域再点此操作） */
  const mergeCells = () => {
    setOpen(false);
    editor.update(() => {
      const sel = $getSelection();
      if (!$isTableSelection(sel)) return;
      const cells = sel.getNodes().filter($isTableCellNode);
      if (cells.length >= 2) $mergeCells(cells);
    });
  };

  /** 单元格对齐：对当前单元格内段落应用 FORMAT_ELEMENT_COMMAND */
  const alignCell = (align: 'left' | 'center' | 'right') => () => {
    runWithCell((cell) => {
      cell.getChildren().forEach((child) => {
        if ((child as any).setFormat) (child as any).setFormat(align);
      });
    });
  };

  const items = menu
    ? [
        {
          key: 'merge',
          label: '合并单元格',
          icon: <MergeCellsOutlined />,
          disabled: !hasMultiCellSelection,
          onClick: mergeCells,
        },
        {
          key: 'unmerge',
          label: '拆分单元格',
          icon: <SplitCellsOutlined />,
          // 当前格子是合并态（colSpan/rowSpan > 1）才可拆
          onClick: () =>
            runWithCell((cell) => {
              if (cell.getColSpan() > 1 || cell.getRowSpan() > 1) $unmergeCellNode(cell);
            }),
        },
        { type: 'divider' as const },
        {
          key: 'insert-row',
          label: '插入行…',
          icon: <InsertRowAboveOutlined />,
          onClick: () => openCountOp('insert-row'),
        },
        {
          key: 'insert-col',
          label: '插入列…',
          icon: <InsertRowLeftOutlined />,
          onClick: () => openCountOp('insert-col'),
        },
        { type: 'divider' as const },
        {
          key: 'delete-row',
          label: '删除行…',
          icon: <DeleteRowOutlined />,
          onClick: () => openCountOp('del-row'),
        },
        {
          key: 'delete-col',
          label: '删除列…',
          icon: <DeleteColumnOutlined />,
          onClick: () => openCountOp('del-col'),
        },
        {
          key: 'clear',
          label: '清空单元格',
          icon: <ClearOutlined />,
          onClick: () =>
            runWithCell((cell) => {
              const [cellNode] = $getNodeTriplet(cell);
              cellNode.getChildren().forEach((c) => c.remove());
              cellNode.append($createParagraphNode());
            }),
        },
        {
          key: 'header-row',
          label: '设为表头行',
          icon: <ColumnHeightOutlined />,
          onClick: () =>
            runWithCell((cell) => {
              const [, , table] = $getNodeTriplet(cell);
              const rowIndex = $getTableRowIndexFromTableCellNode(cell);
              const isHeader = Boolean(cell.getHeaderStyles() & 1); // TableCellHeaderStates.ROW
              $setTableRowIsHeader(table, rowIndex, !isHeader);
            }),
        },
        {
          key: 'header-col',
          label: '设为表头列',
          icon: <ColumnHeightOutlined />,
          onClick: () =>
            runWithCell((cell) => {
              const [, , table] = $getNodeTriplet(cell);
              const colIndex = $getTableColumnIndexFromTableCellNode(cell);
              const isHeader = Boolean(cell.getHeaderStyles() & 2); // TableCellHeaderStates.COLUMN
              $setTableColumnIsHeader(table, colIndex, !isHeader);
            }),
        },
        { type: 'divider' as const },
        {
          key: 'align',
          label: '单元格对齐',
          icon: <AlignLeftOutlined />,
          children: [
            { key: 'align-left', label: '左对齐', icon: <AlignLeftOutlined />, onClick: alignCell('left') },
            { key: 'align-center', label: '居中', icon: <AlignCenterOutlined />, onClick: alignCell('center') },
            { key: 'align-right', label: '右对齐', icon: <AlignRightOutlined />, onClick: alignCell('right') },
          ],
        },
        { type: 'divider' as const },
        {
          key: 'delete-table',
          label: '删除表格',
          danger: true,
          icon: <DeleteOutlined />,
          onClick: () =>
            runWithCell((cell) => {
              $getTableNodeFromLexicalNodeOrThrow(cell).remove();
              setMenu(null);
            }),
        },
      ]
    : [];

  if (!menu) return null;

  const countOpLabels: Record<CountOp, string> = {
    'insert-row': '插入行',
    'insert-col': '插入列',
    'del-row': '删除行',
    'del-col': '删除列',
  };
  // 是否行操作（决定单位「行/列」与方向选项）
  const isRowOp = countOp === 'insert-row' || countOp === 'del-row';

  return createPortal(
    <>
      <Dropdown
        menu={{ items }}
        trigger={['click']}
        open={open}
        onOpenChange={(o) => {
          openRef.current = o;
          setOpen(o);
        }}
      >
        <div
          className="lex-table-menu"
          style={{ top: menu.top - 26, left: menu.left + 4 }}
          onMouseDown={(e) => e.preventDefault()}
          title="表格操作"
        >
          <DownOutlined />
        </div>
      </Dropdown>
      {/* 行/列数量操作弹窗：插入/删除几行几列（插入时选方向） */}
      <Modal
        open={countOp !== null}
        title={countOp ? countOpLabels[countOp] : ''}
        okText="确定"
        cancelText="取消"
        onOk={submitCountOp}
        onCancel={() => setCountOp(null)}
        destroyOnHidden
      >
        <Space>
          <InputNumber
            min={1}
            max={20}
            value={count}
            onChange={(v) => setCount(Number(v) || 1)}
            style={{ width: 90 }}
          />
          <span style={{ color: 'rgba(0,0,0,0.45)' }}>{isRowOp ? '行' : '列'}</span>
          {countOp === 'insert-row' || countOp === 'insert-col' ? (
            <Select
              value={countDir}
              onChange={setCountDir}
              style={{ width: 110 }}
              options={
                countOp === 'insert-row'
                  ? [
                      { value: 'before', label: '在上方' },
                      { value: 'after', label: '在下方' },
                    ]
                  : [
                      { value: 'before', label: '在左侧' },
                      { value: 'after', label: '在右侧' },
                    ]
              }
            />
          ) : null}
        </Space>
      </Modal>
    </>,
    document.body
  );
}

export default TableActionMenuPlugin;
