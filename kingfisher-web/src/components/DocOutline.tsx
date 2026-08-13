import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { DoubleRightOutlined } from '@ant-design/icons';

interface OutlineItem {
  level: 1 | 2;
  text: string;
}

/** 从 Lexical editorState JSON 提取 h1/h2 大纲（编辑态与查看态共用同一份 JSON） */
function extractOutline(stateJson: string): OutlineItem[] {
  const trimmed = (stateJson || '').trim();
  if (!trimmed.startsWith('{')) return [];
  try {
    const state = JSON.parse(trimmed) as { root?: { children?: unknown[] } };
    const items: OutlineItem[] = [];
    const walk = (node: unknown) => {
      if (!node) return;
      if (Array.isArray(node)) {
        node.forEach(walk);
        return;
      }
      const n = node as { type?: string; tag?: string; children?: unknown[] };
      if (n.type === 'heading' && (n.tag === 'h1' || n.tag === 'h2')) {
        const text = (n.children || [])
          .map((c) => ((c as { text?: string }).text ?? ''))
          .join('')
          .trim();
        if (text) items.push({ level: n.tag === 'h1' ? 1 : 2, text });
      }
      if (Array.isArray(n.children)) n.children.forEach(walk);
    };
    walk(state.root);
    return items;
  } catch {
    return [];
  }
}

/** 锚点 id：`#h-{index}`（index 为该标题在大纲中的顺序，与渲染 DOM 的 h1/h2 一一对应） */
const anchorId = (index: number) => `h-${index}`;

interface DocOutlineProps {
  content: string; // editorState JSON
  targetRef: React.RefObject<HTMLElement | null>; // 包含渲染后标题的容器
  /** float=右侧悬浮面板（查看态/公开页）；inline=编辑态右侧布局列（sticky 跟随） */
  mode?: 'float' | 'inline';
  /** 折叠状态变化回调（inline 模式下父组件据此收窄 split 分栏，让编辑器占满剩余） */
  onCollapseChange?: (collapsed: boolean) => void;
  /** 初始折叠态（默认展开）；父组件「默认折叠」时传 true 保持一致 */
  defaultCollapsed?: boolean;
}

/**
 * 文档右侧大纲：列出 h1/h2 标题，点击平滑滚动并写入 `#h-n` 锚点（可直接分享直达），
 * 带锚点访问时自动定位，滚动时高亮当前标题，支持折叠。
 * content 是 Lexical JSON，渲染 DOM 的 h1/h2 与 JSON 提取顺序一致。
 */
const DocOutline: React.FC<DocOutlineProps> = ({
  content,
  targetRef,
  mode = 'float',
  onCollapseChange,
  defaultCollapsed = false,
}) => {
  const items = useMemo(() => extractOutline(content), [content]);
  const [active, setActive] = useState(-1);
  const [collapsed, setCollapsed] = useState(defaultCollapsed);
  const activeRef = useRef(active);
  activeRef.current = active;
  // 记录 URL 锚点是否已应用过：避免编辑态 content 每次变化（items 重算）导致反复滚动
  const hashAppliedRef = useRef(false);

  const setCollapsedBoth = (c: boolean) => {
    setCollapsed(c);
    onCollapseChange?.(c);
  };

  // 滚动到指定标题（index 越界安全忽略）。
  // 用 window.scrollBy 直接滚窗，绕开 scrollIntoView 对中间滚动容器的对齐干扰；
  // 文档不够长时浏览器会自然钳制在最大滚动位置。
  const jump = useCallback(
    (index: number) => {
      const container = targetRef.current;
      if (!container) return;
      const el = container.querySelectorAll<HTMLElement>('h1, h2')[index];
      if (!el) return;
      const rect = el.getBoundingClientRect();
      window.scrollBy({ top: rect.top - 76, behavior: 'smooth' });
    },
    [targetRef]
  );

  // 解析 URL 锚点并滚动：支持 #h-n 直达
  const scrollToHash = useCallback(
    (hash: string) => {
      const m = hash.match(/^#h-(\d+)$/);
      if (!m) return;
      const index = Number(m[1]);
      if (index >= 0 && index < items.length) jump(index);
    },
    [items.length, jump]
  );

  // hashchange（前进/后退/手动改锚点）时重新应用锚点；用 ref 拿最新 scrollToHash
  const scrollToHashRef = useRef(scrollToHash);
  scrollToHashRef.current = scrollToHash;
  useEffect(() => {
    const onHash = () => scrollToHashRef.current(window.location.hash);
    window.addEventListener('hashchange', onHash);
    return () => window.removeEventListener('hashchange', onHash);
  }, []);

  // URL 锚点直达：标题渲染完成后应用一次（编辑器/接口加载是异步的，标题未就绪时等下次）。
  // 只跑一次，编辑态每次击键 items 重算不会重新触发滚动。
  useEffect(() => {
    if (hashAppliedRef.current || items.length === 0) return;
    const container = targetRef.current;
    if (!container || !container.querySelector('h1, h2')) return;
    hashAppliedRef.current = true;
    scrollToHash(window.location.hash);
  }, [items, scrollToHash, targetRef]);

  // 滚动高亮：标记视口顶部附近（header 以下约 110px）的当前标题
  useEffect(() => {
    if (items.length === 0) return;
    const update = () => {
      const container = targetRef.current;
      if (!container) return;
      const headings = container.querySelectorAll<HTMLElement>('h1, h2');
      let current = -1;
      for (let i = 0; i < headings.length; i++) {
        if (headings[i].getBoundingClientRect().top <= 110) current = i;
        else break;
      }
      if (current !== activeRef.current) setActive(current);
    };
    update();
    window.addEventListener('scroll', update, true);
    return () => window.removeEventListener('scroll', update, true);
  }, [items, targetRef]);

  if (items.length === 0) return null;

  // 折叠态：缩成竖排小标签（inline 模式用 sticky 占位，float 模式 fixed 悬浮）
  if (collapsed) {
    const cls = mode === 'inline' ? 'doc-outline-tab doc-outline-tab-inline' : 'doc-outline-tab';
    return (
      <button type="button" className={cls} onClick={() => setCollapsedBoth(false)} title="展开大纲">
        大纲
      </button>
    );
  }

  const rootCls = mode === 'inline' ? 'doc-outline doc-outline-inline' : 'doc-outline';

  const onAnchorClick = (e: React.MouseEvent, index: number) => {
    e.preventDefault();
    jump(index);
    // 写入锚点便于分享直达（replaceState 不污染前进/后退历史）
    window.history.replaceState(null, '', `#${anchorId(index)}`);
    setActive(index);
  };

  return (
    <nav className={rootCls} aria-label="文档大纲">
      <div className="doc-outline-header">
        <span className="doc-outline-title">大纲</span>
        <button
          type="button"
          className="doc-outline-toggle"
          onClick={() => setCollapsedBoth(true)}
          title="折叠大纲"
          aria-label="折叠大纲"
        >
          <DoubleRightOutlined />
        </button>
      </div>
      <ul>
        {items.map((item, i) => (
          <li key={i} className={i === active ? 'is-active' : ''}>
            <a
              href={`#${anchorId(i)}`}
              className={`doc-outline-link level-${item.level}`}
              onClick={(e) => onAnchorClick(e, i)}
              title={item.text}
            >
              {item.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
};

export default DocOutline;
