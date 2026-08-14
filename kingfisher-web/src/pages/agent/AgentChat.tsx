import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { App, Avatar, Button, Flex, Input, Spin, Tooltip, theme } from 'antd';
import {
  DeleteOutlined, FileTextOutlined, FontSizeOutlined, MenuFoldOutlined,
  MenuUnfoldOutlined, PlusOutlined, SendOutlined, StopOutlined, RobotOutlined, UserOutlined,
} from '@ant-design/icons';
import Split from 'split.js';
import Markdown from '../../components/Markdown';
import { agentApi, type AgentSSEEvent } from '../../api/agent';
import { useAuthStore } from '../../stores/auth';

interface Conversation {
  id: number;
  user_id: number;
  title: string;
  created_at: string;
  updated_at: string;
}

interface ChatMessage {
  id: number;
  conversation_id?: number;
  role: 'user' | 'assistant' | 'tool';
  content: string;
  created_at?: string;
  /** 本地乐观消息（尚未落库） */
  local?: boolean;
  /** 工具调用信息 */
  toolName?: string;
}

const shorten = (s: string, n = 120): string => (s.length > n ? s.slice(0, n) + '…' : s);

/** Agent 聊天页：左侧会话列表 + 右侧聊天气泡 + 流式输入。 */
export default function AgentChat() {
  const { message: msg } = App.useApp();
  const { token: themeToken } = theme.useToken();
  const userInfo = useAuthStore((s) => s.userInfo) as Record<string, unknown> | null;
  const navigate = useNavigate();
  // URL 中的会话 id（/agent/:conversationId），支持深层链接直达指定会话。
  const { conversationId: urlConvId } = useParams();
  const urlId = urlConvId ? Number(urlConvId) : null;

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeId, setActiveId] = useState<number | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [streaming, setStreaming] = useState(false);
  const [loadingList, setLoadingList] = useState(true);
  const [loadingMsgs, setLoadingMsgs] = useState(false);
  // 当前会话对象（标题等）
  const activeConversation = conversations.find((c) => c.id === activeId);
  // 消息渲染模式：markdown（默认）| plain（纯文本）；本地记忆
  const VIEW_KEY = 'agent:view-mode';
  const [viewMode, setViewMode] = useState<'markdown' | 'plain'>(() => {
    try {
      const v = localStorage.getItem(VIEW_KEY);
      return v === 'plain' ? 'plain' : 'markdown';
    } catch {
      return 'markdown';
    }
  });
  useEffect(() => {
    try {
      localStorage.setItem(VIEW_KEY, viewMode);
    } catch {
      /* 忽略 */
    }
  }, [viewMode]);
  // 会话列表折叠状态（默认折叠；本地记忆）
  const SIDER_COLLAPSED_KEY = 'agent:conv-collapsed';
  const [collapsed, setCollapsed] = useState<boolean>(() => {
    try {
      const v = localStorage.getItem(SIDER_COLLAPSED_KEY);
      return v === null ? true : v === '1';
    } catch {
      return true;
    }
  });
  // 折叠状态写入 localStorage
  useEffect(() => {
    try {
      localStorage.setItem(SIDER_COLLAPSED_KEY, collapsed ? '1' : '0');
    } catch {
      /* 忽略 */
    }
  }, [collapsed]);

  // split.js 两栏实例（会话列表 | 聊天区）；始终渲染两个 pane，初始化一次。
  const listPaneRef = useRef<HTMLDivElement>(null);
  const chatPaneRef = useRef<HTMLDivElement>(null);
  const splitRef = useRef<ReturnType<typeof Split> | null>(null);
  useEffect(() => {
    const list = listPaneRef.current;
    const chat = chatPaneRef.current;
    if (!list || !chat) return;
    const instance = Split([list, chat], {
      sizes: [26, 74],
      minSize: [0, 420], // 会话列表可拖到 0 完全合上
      gutterSize: 6,
      cursor: 'col-resize',
      gutter: () => {
        const g = document.createElement('div');
        g.className = 'doc-split-gutter';
        return g;
      },
      onDragEnd: (sizes) => setCollapsed(sizes[0] < 12),
    });
    splitRef.current = instance;
    return () => {
      instance.destroy();
      splitRef.current = null;
    };
  }, []);

  // 依据折叠状态重算两栏占比
  useEffect(() => {
    splitRef.current?.setSizes(collapsed ? [0, 100] : [26, 74]);
  }, [collapsed]);

  const abortRef = useRef<{ abort: () => void } | null>(null);
  const bodyRef = useRef<HTMLDivElement>(null);
  // 标记最近一轮 assistant 文本已累积完成（收到 tool_result 后置 true），
  // 下一轮 text_delta 到来时需新建 assistant 气泡，避免追加到工具提示行后面。
  const assistantDoneRef = useRef(false);

  // 拉取会话列表；URL 指定会话优先，否则默认第一个/新建。
  const loadConversations = useCallback(async () => {
    setLoadingList(true);
    try {
      const r = await agentApi.listConversations();
      const list = (r.data as Conversation[]) || [];
      setConversations(list);
      // 无会话时自动新建
      if (list.length === 0) {
        const c = await agentApi.createConversation();
        setConversations([c.data as Conversation]);
        setActiveId((c.data as Conversation).id);
        navigate(`/agent/${(c.data as Conversation).id}`, { replace: true });
      } else if (urlId && list.some((c) => c.id === urlId)) {
        // URL 指定会话存在 → 打开它
        setActiveId(urlId);
      } else if (!activeId) {
        setActiveId(list[0].id);
        navigate(`/agent/${list[0].id}`, { replace: true });
      }
    } catch {
      msg.error('加载会话失败');
    } finally {
      setLoadingList(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeId, urlId, msg]);

  useEffect(() => {
    loadConversations();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // URL 变化（前进/后退或深层链接）→ 同步当前会话；若 id 非法则回退列表第一项
  useEffect(() => {
    if (urlId && urlId !== activeId) {
      if (conversations.some((c) => c.id === urlId)) {
        setActiveId(urlId);
      } else {
        // 非本用户会话或不存在：清掉消息并提示
        setMessages([]);
        if (conversations.length) navigate(`/agent/${conversations[0].id}`, { replace: true });
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlId]);

  // 切换/首次会话：加载消息
  useEffect(() => {
    if (!activeId) return;
    setLoadingMsgs(true);
    agentApi
      .listMessages(activeId)
      .then((r) => setMessages((r.data as ChatMessage[]) || []))
      .catch(() => msg.error('加载消息失败'))
      .finally(() => setLoadingMsgs(false));
  }, [activeId, msg]);

  // 新消息时自动滚动到底部
  useEffect(() => {
    bodyRef.current?.scrollTo({ top: bodyRef.current.scrollHeight, behavior: 'smooth' });
  }, [messages]);

  const switchConversation = (id: number) => {
    if (id === activeId) return;
    abortRef.current?.abort();
    abortRef.current = null;
    setStreaming(false);
    setActiveId(id);
    navigate(`/agent/${id}`);
  };

  const createConversation = async () => {
    try {
      const r = await agentApi.createConversation();
      const c = r.data as Conversation;
      setConversations((prev) => [c, ...prev]);
      setActiveId(c.id);
      setMessages([]);
      navigate(`/agent/${c.id}`);
    } catch {
      msg.error('创建会话失败');
    }
  };

  const removeConversation = async (id: number) => {
    try {
      await agentApi.deleteConversation(id);
      const next = conversations.filter((c) => c.id !== id);
      setConversations(next);
      if (activeId === id) {
        abortRef.current?.abort();
        abortRef.current = null;
        setStreaming(false);
        setActiveId(next.length ? next[0].id : null);
        setMessages([]);
        // 删除当前会话后 URL 跟随跳转
        if (next.length) navigate(`/agent/${next[0].id}`);
      }
    } catch {
      msg.error('删除会话失败');
    }
  };

  const send = () => {
    const text = input.trim();
    if (!text || !activeId || streaming) return;
    setInput('');

    // 本地乐观追加用户消息 + 空的 assistant 消息用于流式累积
    setMessages((prev) => [
      ...prev,
      { id: -Date.now(), role: 'user', content: text, local: true },
      { id: -Date.now() - 1, role: 'assistant', content: '', local: true },
    ]);
    setStreaming(true);
    assistantDoneRef.current = false;

    const handle = agentApi.chatStream(activeId, text, (evt: AgentSSEEvent) => {
      switch (evt.type) {
        case 'text_delta': {
          setMessages((prev) => {
            const copy = [...prev];
            const last = copy[copy.length - 1];
            // 若上一轮 assistant 已完成且最后一条不是 assistant，则新建气泡；
            // 否则累积到现有 assistant 气泡。
            if (last && last.role === 'assistant' && !assistantDoneRef.current) {
              copy[copy.length - 1] = { ...last, content: last.content + (evt.delta || '') };
            } else {
              assistantDoneRef.current = false;
              copy.push({ id: -Date.now() - 3, role: 'assistant', content: evt.delta || '', local: true });
            }
            return copy;
          });
          break;
        }
        case 'tool_use': {
          setMessages((prev) => [
            ...prev,
            { id: -Date.now() - 2, role: 'tool', content: '', local: true, toolName: evt.tool },
          ]);
          break;
        }
        case 'tool_result': {
          // 工具结果到达 → 当前这一轮 assistant 文本已结束，下一轮 delta 需新气泡。
          assistantDoneRef.current = true;
          setMessages((prev) => {
            const copy = [...prev];
            for (let i = copy.length - 1; i >= 0; i--) {
              if (copy[i].role === 'tool') {
                copy[i] = { ...copy[i], content: shorten(evt.message || '') };
                break;
              }
            }
            return copy;
          });
          break;
        }
        case 'done': {
          // 用最终完整回复覆盖/追加 assistant 消息
          if (evt.content) {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last && last.role === 'assistant' && last.local) {
                copy[copy.length - 1] = { ...last, content: evt.content as string };
              } else {
                copy.push({ id: -Date.now() - 4, role: 'assistant', content: evt.content as string, local: true });
              }
              return copy;
            });
          }
          setStreaming(false);
          abortRef.current = null;
          assistantDoneRef.current = false;
          loadConversations();
          break;
        }
        case 'error': {
          setMessages((prev) => prev.filter((m) => !(m.local && m.role === 'assistant' && !m.content)));
          setStreaming(false);
          abortRef.current = null;
          assistantDoneRef.current = false;
          msg.error(evt.message || '对话失败');
          break;
        }
        case 'start':
          break;
      }
    });
    abortRef.current = handle;
  };

  const stopStreaming = () => {
    abortRef.current?.abort();
    abortRef.current = null;
    setStreaming(false);
  };

  const renderBubble = (m: ChatMessage) => {
    if (m.role === 'tool') {
      return (
        <Flex justify="center" style={{ margin: '4px 0' }}>
          <div
            style={{
              fontSize: 12, color: themeToken.colorTextTertiary, background: themeToken.colorFillTertiary,
              padding: '2px 10px', borderRadius: 999,
            }}
          >
            🔧 调用工具 {m.toolName || 'call_api'}{m.content ? ` · ${m.content}` : '…'}
          </div>
        </Flex>
      );
    }
    const isUser = m.role === 'user';
    return (
      <Flex justify={isUser ? 'flex-end' : 'flex-start'} gap={8} style={{ margin: '10px 0' }}>
        {!isUser && (
          <Avatar icon={<RobotOutlined />} style={{ background: themeToken.colorPrimary, flexShrink: 0 }} />
        )}
        <div
          style={{
            maxWidth: '76%',
            background: isUser ? themeToken.colorPrimary : themeToken.colorFillSecondary,
            color: isUser ? '#fff' : themeToken.colorText,
            padding: '10px 14px',
            borderRadius: isUser ? '14px 14px 2px 14px' : '14px 14px 14px 2px',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            lineHeight: 1.7,
            fontSize: 14,
          }}
        >
          {isUser ? (
            (m.content || '')
          ) : m.content ? (
            viewMode === 'markdown' ? (
              // 流式进行中直接 pre-wrap（避免每次 delta 重复解析 markdown + 闪烁）；
              // 完成后切 Markdown 渲染。
              streaming && m.local ? (
                m.content
              ) : (
                <Markdown content={m.content} />
              )
            ) : (
              m.content
            )
          ) : streaming && m.role === 'assistant' && m.local ? (
            '思考中…'
          ) : (
            ''
          )}
        </div>
        {isUser && (
          <Avatar
            src={(userInfo?.avatar as string) || undefined}
            icon={<UserOutlined />}
            style={{ flexShrink: 0, background: themeToken.colorPrimary }}
          />
        )}
      </Flex>
    );
  };

  // 输入区：圆角容器内嵌无边框多行输入 + 圆形发送/停止按钮。
  // centered（欢迎态）：输入框固定宽度、整行水平居中；底部固定态撑满外层。
  const renderInputArea = (centered: boolean) => (
    <Flex
      align="flex-end"
      gap={8}
      justify={centered ? 'center' : 'flex-start'}
      style={centered ? undefined : { padding: 12 }}
    >
      <div
        style={{
          // 欢迎态输入框固定宽度居中；底部态 flex:1 撑满
          flex: centered ? '0 1 480px' : 1,
          minWidth: 0,
          background: themeToken.colorBgContainer,
          border: `1px solid ${themeToken.colorBorder}`,
          borderRadius: 12,
          padding: '6px 8px',
          display: 'flex',
          alignItems: 'flex-end',
          transition: 'border-color 0.2s',
        }}
      >
        <Input.TextArea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onPressEnter={(e) => {
            if (!e.shiftKey && !streaming) { e.preventDefault(); send(); }
          }}
          placeholder="输入问题，Enter 发送，Shift+Enter 换行"
          autoSize={{ minRows: 1, maxRows: 5 }}
          variant="borderless"
          disabled={streaming || !activeId}
          style={{ padding: '4px 4px', fontSize: 14 }}
        />
      </div>
      <Tooltip title={streaming ? '停止生成' : (input.trim() ? '发送 (Enter)' : '请输入内容')}>
        <Button
          type="primary"
          shape="circle"
          icon={streaming ? <StopOutlined /> : <SendOutlined />}
          onClick={streaming ? stopStreaming : send}
          disabled={(!activeId) || (!streaming && !input.trim())}
          style={{ flexShrink: 0 }}
        />
      </Tooltip>
    </Flex>
  );

  return (
    // 左右两栏：左侧（工具栏 + 会话列表） | 右侧（聊天消息区 + 底部发送区），宽度由 split 拖拽控制。
    <div style={{ display: 'flex', height: 'calc(100vh - 64px - 32px - 32px)', alignItems: 'stretch' }}>
      {/* 左侧：工具栏 + 会话列表（纵向） */}
      <div
        ref={listPaneRef}
        style={{
          minWidth: 0,
          overflow: 'hidden',
          background: themeToken.colorBgContainer,
          borderRadius: themeToken.borderRadiusLG,
          padding: collapsed ? 0 : 12,
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {/* 工具栏：新建 / 会话列表展开收起 / 渲染切换 */}
        {!collapsed && (
          <Flex align="center" gap={6} style={{ marginBottom: 8 }}>
            <Tooltip title="新建会话">
              <Button type="primary" icon={<PlusOutlined />} onClick={createConversation} />
            </Tooltip>
            <Flex align="center" gap={4} style={{ marginLeft: 'auto' }}>
              <Tooltip title="收起会话列表">
                <Button size="small" icon={<MenuFoldOutlined />} onClick={() => setCollapsed(true)} />
              </Tooltip>
              <Tooltip title={viewMode === 'markdown' ? '切换为纯文本' : '切换为 Markdown'}>
                <Button
                  size="small"
                  icon={viewMode === 'markdown' ? <FontSizeOutlined /> : <FileTextOutlined />}
                  onClick={() => setViewMode(viewMode === 'markdown' ? 'plain' : 'markdown')}
                />
              </Tooltip>
            </Flex>
          </Flex>
        )}
        {/* 会话列表（flex:1 撑满） */}
        <div style={{ flex: 1, overflowY: 'auto', minHeight: 0 }}>
          {loadingList ? (
            <Flex justify="center" style={{ padding: 24 }}><Spin /></Flex>
          ) : (
            /* 会话列表（antd List 已弃用，改纯 div 渲染避免弃用告警） */
            <Flex vertical>
              {conversations.map((c) => (
                <Flex
                  key={c.id}
                  align="center"
                  justify="center"
                  gap={4}
                  onClick={() => switchConversation(c.id)}
                  style={{
                    cursor: 'pointer',
                    borderRadius: 8,
                    // 左右留白：内容不贴边；标题居中显示
                    padding: '10px 20px',
                    background: c.id === activeId ? themeToken.colorPrimaryBg : 'transparent',
                  }}
                >
                  <span
                    style={{
                      fontSize: 13,
                      fontWeight: c.id === activeId ? 600 : 400,
                      flex: 1,
                      textAlign: 'center',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {c.title || '新会话'}
                  </span>
                  <Tooltip title="删除会话">
                    <DeleteOutlined
                      onClick={(e) => { e.stopPropagation(); removeConversation(c.id); }}
                      style={{ color: themeToken.colorTextTertiary, flexShrink: 0 }}
                    />
                  </Tooltip>
                </Flex>
              ))}
            </Flex>
          )}
        </div>
      </div>

      {/* 右侧：聊天消息区 + 底部发送区 */}
      <div ref={chatPaneRef} style={{ minWidth: 0, background: themeToken.colorBgLayout, borderRadius: themeToken.borderRadiusLG, overflow: 'hidden', display: 'flex', flexDirection: 'column', position: 'relative' }}>
        {/* 会话列表已合上：浮动展开图标（悬浮在聊天区左上角） */}
        {collapsed && (
          <Button
            type="primary"
            shape="circle"
            icon={<MenuUnfoldOutlined />}
            onClick={() => setCollapsed(false)}
            title="展开会话列表"
            style={{
              position: 'absolute',
              top: 12,
              left: 12,
              zIndex: 10,
              boxShadow: themeToken.boxShadow,
            }}
          />
        )}
        {/* 聊天区顶部：当前会话标题（灰底，与聊天区一致；下方留白与消息区隔开） */}
        <div style={{ display: 'flex', alignItems: 'center', padding: '8px 16px', marginBottom: 4, background: 'transparent' }}>
          <span style={{ fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', flex: 1 }}>
            {activeConversation?.title || '新会话'}
          </span>
        </div>
        <div
          ref={bodyRef}
          style={{
            flex: 1, overflowY: 'auto', padding: '8px 24px',
            background: themeToken.colorBgLayout, minHeight: 0,
          }}
        >
          {loadingMsgs ? (
            <Flex justify="center" style={{ padding: 40 }}><Spin /></Flex>
          ) : messages.length === 0 ? (
            /* 无会话/空消息：提示语上下左右居中 */
            <Flex vertical align="center" justify="center" gap={12} style={{ height: '100%' }}>
              <div style={{ fontSize: 18, fontWeight: 600 }}>向 Agent 提问</div>
              <div style={{ color: themeToken.colorTextTertiary }}>了解或操作系统</div>
            </Flex>
          ) : (
            /* 消息列居中收窄（仿 DeepSeek）：不撑满全宽，气泡在窄列内左右对齐 */
            <div style={{ maxWidth: 760, margin: '0 auto' }}>
              {messages.map((m) => <div key={m.id}>{renderBubble(m)}</div>)}
            </div>
          )}
        </div>
        {/* 底部发送区（居中窄列，与消息列对齐） */}
        <div style={{ maxWidth: 760, width: '100%', margin: '0 auto' }}>
          {renderInputArea(false)}
        </div>
      </div>
    </div>
  );
}
