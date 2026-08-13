import {
  $createParagraphNode,
  $createTextNode,
  $getSelection,
  $isRangeSelection,
  DOMConversionMap,
  ElementNode,
  LexicalEditor,
  LexicalNode,
  SerializedElementNode,
  Spread,
  TextNode,
} from 'lexical';
import { $setBlocksType } from '@lexical/selection';
import { $getNearestBlockElementAncestorOrThrow } from '@lexical/utils';
import { $createHeadingNode, $createQuoteNode } from '@lexical/rich-text';
import { $createCodeNode, CODE_LANGUAGE_MAP } from '@lexical/code';
import { $createListNode } from '@lexical/list';
import { $createTableNodeWithDimensions } from '@lexical/table';
import { $createHorizontalRuleNode } from '@lexical/react/LexicalHorizontalRuleNode';
import { $createImageNode } from './ImageNode';

/* ============ 提示框 Callout ============ */

export type SerializedCalloutNode = Spread<{ type: 'callout'; icon: string }, SerializedElementNode>;

/** 提示框：<div class="notion-callout" data-icon="💡">…子块…</div>，子块直接挂在节点上 */
export class CalloutNode extends ElementNode {
  __icon: string;

  constructor(icon = '💡', key?: string) {
    super(key);
    this.__icon = icon;
  }

  static getType(): string {
    return 'callout';
  }

  static clone(node: CalloutNode): CalloutNode {
    return new CalloutNode(node.__icon, node.__key);
  }

  getIcon(): string {
    return this.getLatest().__icon;
  }

  setIcon(icon: string) {
    const writable = this.getWritable();
    writable.__icon = icon;
  }

  createDOM(): HTMLElement {
    const el = document.createElement('div');
    el.className = 'notion-callout';
    el.setAttribute('data-icon', this.getIcon());
    return el;
  }

  updateDOM(prev: CalloutNode, dom: HTMLElement): boolean {
    const icon = this.getIcon();
    if (prev.getIcon() !== icon) {
      dom.setAttribute('data-icon', icon);
    }
    return false;
  }

  static importDOM(): DOMConversionMap | null {
    return {
      div: (domNode: HTMLElement) =>
        domNode.classList?.contains('notion-callout')
          ? {
              conversion: (d: HTMLElement) => ({ node: $createCalloutNode(d.getAttribute('data-icon') || '💡') }),
              priority: 1,
            }
          : null,
    };
  }

  static importJSON(serialized: SerializedCalloutNode): CalloutNode {
    return $createCalloutNode(serialized.icon).updateFromJSON(serialized) as CalloutNode;
  }

  exportJSON(): SerializedCalloutNode {
    return { ...super.exportJSON(), type: 'callout', icon: this.getIcon() };
  }
}

export function $createCalloutNode(icon = '💡'): CalloutNode {
  return new CalloutNode(icon);
}

export function $isCalloutNode(node: LexicalNode | null | undefined): node is CalloutNode {
  return node instanceof CalloutNode;
}

/* ============ 折叠块 Toggle ============ */

export type SerializedToggleNode = Spread<{ type: 'toggle' }, SerializedElementNode>;

/**
 * 折叠块：<div class="notion-toggle">…子块…</div>，第一个子块为标题。
 * 点击标题行（firstElementChild）折叠/展开内容，折叠状态仅作用于当前 DOM，
 * 不持久化到 HTML（保存/导出总是展开态）。
 */
export class ToggleNode extends ElementNode {
  constructor(key?: string) {
    super(key);
  }

  static getType(): string {
    return 'toggle';
  }

  static clone(node: ToggleNode): ToggleNode {
    return new ToggleNode(node.__key);
  }

  createDOM(): HTMLElement {
    const el = document.createElement('div');
    el.className = 'notion-toggle';
    // 仅点击标题行最左侧的三角区（约 26px）切换折叠，点文字正常编辑
    el.addEventListener('click', (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target !== el.firstElementChild) return;
      const rect = target.getBoundingClientRect();
      if (e.clientX - rect.left < 26) {
        el.classList.toggle('notion-toggle-collapsed');
      }
    });
    return el;
  }

  updateDOM(): boolean {
    return false;
  }

  static importDOM(): DOMConversionMap | null {
    return {
      div: (domNode: HTMLElement) =>
        domNode.classList?.contains('notion-toggle')
          ? { conversion: () => ({ node: $createToggleNode() }), priority: 1 }
          : null,
      // 兼容外部粘贴/历史数据里的 <details><summary>…</summary><div>…</div></details>
      // （否则 details/summary 会被当内联元素提升到根，导致 root.splice 崩溃）
      details: (domNode: HTMLElement) =>
        domNode.querySelector(':scope > summary') ? { conversion: $convertToggleDetailsElement, priority: 1 } : null,
    };
  }

  static importJSON(serialized: SerializedToggleNode): ToggleNode {
    return $createToggleNode().updateFromJSON(serialized) as ToggleNode;
  }

  exportJSON(): SerializedToggleNode {
    return { ...super.exportJSON(), type: 'toggle' };
  }
}

export function $createToggleNode(): ToggleNode {
  return new ToggleNode();
}

/** <details> 转换：summary 文本 → 第一个标题子块，其余内容 → 后续子块 */
function $convertToggleDetailsElement(domNode: HTMLElement) {
  const node = $createToggleNode();
  const summary = domNode.querySelector(':scope > summary');
  const titleText = summary?.textContent?.trim() || '';
  return {
    node,
    after: (children: LexicalNode[]) => {
      // details/summary 被当内联元素，summary 文本会被提升为根级 TextNode，
      // 此处重组为 [标题段落, ...内容]（去掉与标题重复的文本节点）。
      const title = $createParagraphNode();
      if (titleText) {
        title.append($createTextNode(titleText));
      }
      const rest = children.filter((c) => !(c instanceof TextNode && c.getTextContent().trim() === titleText));
      return [title, ...rest];
    },
  };
}

export function $isToggleNode(node: LexicalNode | null | undefined): node is ToggleNode {
  return node instanceof ToggleNode;
}

/* ============ 代码块语言 ============ */

/** 可选的代码语言列表（@lexical/code 的映射 + 纯文本） */
export const CODE_LANGUAGES: { value: string; label: string }[] = [
  { value: 'javascript', label: 'JavaScript' },
  { value: 'typescript', label: 'TypeScript' },
  { value: 'python', label: 'Python' },
  { value: 'go', label: 'Go' },
  { value: 'java', label: 'Java' },
  { value: 'json', label: 'JSON' },
  { value: 'html', label: 'HTML' },
  { value: 'css', label: 'CSS' },
  { value: 'sql', label: 'SQL' },
  { value: 'bash', label: 'Bash' },
  { value: 'plaintext', label: '纯文本' },
  ...Object.entries(CODE_LANGUAGE_MAP)
    .filter(
      ([code]) =>
        !['javascript', 'typescript', 'python', 'go', 'java', 'json', 'html', 'css', 'sql', 'bash'].includes(code)
    )
    .map(([code, label]) => ({ value: code, label: String(label) })),
];

/* ============ 插入辅助 ============ */

export type InsertFn = (editor: LexicalEditor) => void;

/** 用新块替换当前光标所在块，并把光标放入新块 */
function replaceCurrentBlock(editor: LexicalEditor, build: () => ElementNode) {
  editor.update(() => {
    const selection = $getSelection();
    if (!$isRangeSelection(selection)) return;
    const block = $getNearestBlockElementAncestorOrThrow(selection.anchor.getNode());
    const newNode = build();
    block.replace(newNode);
    newNode.selectStart();
  });
}

/** 将当前块转换为指定类型（标题/引用/列表/代码等，复用 Lexical 的 $setBlocksType） */
function convertBlock(editor: LexicalEditor, build: () => ElementNode) {
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      $setBlocksType(selection, build);
    }
  });
}

export const insertHeading =
  (level: 'h1' | 'h2' | 'h3'): InsertFn =>
  (editor) =>
    convertBlock(editor, () => $createHeadingNode(level));

export const insertParagraph: InsertFn = (editor) => convertBlock(editor, () => $createParagraphNode());

export const insertQuote: InsertFn = (editor) => convertBlock(editor, () => $createQuoteNode());

export const insertBulletList: InsertFn = (editor) => convertBlock(editor, () => $createListNode('bullet'));

export const insertNumberedList: InsertFn = (editor) => convertBlock(editor, () => $createListNode('number'));

export const insertCheckList: InsertFn = (editor) => {
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      $setBlocksType(selection, () => $createListNode('check'));
    }
  });
};

export const insertCodeBlock: InsertFn = (editor) => {
  convertBlock(editor, () => $createCodeNode('javascript'));
};

export const insertCallout: InsertFn = (editor) =>
  replaceCurrentBlock(editor, () => {
    const node = $createCalloutNode('💡');
    node.append($createParagraphNode());
    return node;
  });

export const insertToggle: InsertFn = (editor) =>
  replaceCurrentBlock(editor, () => {
    const node = $createToggleNode();
    node.append($createParagraphNode());
    node.append($createParagraphNode());
    return node;
  });

export const insertTable =
  (rows = 3, columns = 3): InsertFn =>
  (editor) => {
    editor.update(() => {
      const selection = $getSelection();
      if (!$isRangeSelection(selection)) return;
      const table = $createTableNodeWithDimensions(rows, columns, true);
      selection.insertNodes([table]);
    });
  };

export const insertHorizontalRule: InsertFn = (editor) => {
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      selection.insertNodes([$createHorizontalRuleNode()]);
    }
  });
};

export const insertImage: InsertFn = (editor) => {
  editor.update(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      // 空 src：decorator 渲染为上传占位，点按选文件
      selection.insertNodes([$createImageNode('')]);
    }
  });
};
