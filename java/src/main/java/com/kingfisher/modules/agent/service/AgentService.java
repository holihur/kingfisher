package com.kingfisher.modules.agent.service;

import com.kingfisher.common.ErrorCode;
import com.kingfisher.modules.agent.domain.Conversation;
import com.kingfisher.modules.agent.mapper.AgentMapper;
import com.kingfisher.modules.config.service.ConfigService;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

@Service
public class AgentService {

    private final AgentMapper agentMapper;
    private final LlmClient llmClient;
    private final ConfigService configService;

    @Value("${kingfisher.agent.enabled:true}")
    private boolean enabled;

    @Value("${kingfisher.agent.system-prompt:}")
    private String systemPromptFallback;

    private static final String SYSTEM_FALLBACK = "你是 Kingfisher 后台管理系统的 AI 助手，可以通过调用系统接口帮助用户查询和操作系统。请始终用中文回答，简洁、准确。";

    public AgentService(AgentMapper agentMapper, LlmClient llmClient, ConfigService configService) {
        this.agentMapper = agentMapper;
        this.llmClient = llmClient;
        this.configService = configService;
    }

    public void checkEnabled() {
        if (!enabled) throw new BizException(ErrorCode.ERR_AGENT_DISABLED, "Agent 未启用");
    }

    public List<Conversation> listConversations(Long userId) {
        checkEnabled();
        return agentMapper.listConversations(userId);
    }

    public Conversation getConversation(Long id, Long userId) {
        checkEnabled();
        Conversation c = agentMapper.findConversationById(id, userId);
        if (c == null) throw new BizException(ErrorCode.ERR_AGENT_CONVERSATION_NF, "会话不存在");
        return c;
    }

    public Conversation createConversation(Long userId, String title) {
        checkEnabled();
        if (title == null || title.isBlank()) title = "新会话";
        Conversation c = new Conversation();
        c.setUserId(userId);
        c.setTitle(title);
        agentMapper.insertConversation(c);
        return c;
    }

    public void deleteConversation(Long id, Long userId) {
        getConversation(id, userId);
        agentMapper.deleteConversation(id, userId);
        agentMapper.deleteMessagesByConversation(id);
    }

    public List<com.kingfisher.modules.agent.domain.AgentMessage> listMessages(Long conversationId, Long userId) {
        checkEnabled();
        getConversation(conversationId, userId);
        return agentMapper.listMessages(conversationId, userId);
    }

    public Map<String,Object> chat(Long conversationId, Long userId, String content, String userToken) {
        checkEnabled();
        Conversation conv = getConversation(conversationId, userId);
        // 追加用户消息
        com.kingfisher.modules.agent.domain.AgentMessage userMsg = new com.kingfisher.modules.agent.domain.AgentMessage();
        userMsg.setConversationId(conversationId);
        userMsg.setRole("user");
        userMsg.setContent(content);
        agentMapper.insertMessage(userMsg);

        // 首条消息更新标题
        if ("新会话".equals(conv.getTitle()) || conv.getTitle() == null || conv.getTitle().isBlank()) {
            String title = content.length() > 20 ? content.substring(0, 20) : content;
            agentMapper.updateConversationTitle(conversationId, title);
        }

        // 获取 API Key
        String apiKey = resolveApiKey();
        if (apiKey == null || apiKey.isBlank()) throw new BizException(ErrorCode.ERR_AGENT_NO_API_KEY, "未配置 LLM API Key");

        // 系统提示词
        String system = resolveSystemPrompt();

        // 构建历史消息
        List<Map<String,Object>> messages = agentMapper.listMessages(conversationId, userId).stream()
                .map(m -> Map.<String,Object>of("role", m.getRole(), "content", m.getContent() != null ? m.getContent() : ""))
                .toList();

        // 真调 DeepSeek
        String reply;
        try {
            reply = llmClient.chatSync(apiKey, system, messages);
        } catch (Exception e) {
            throw new BizException(ErrorCode.ERR_AGENT_LLM_ERROR, "LLM 调用失败: " + e.getMessage());
        }

        // 持久化 assistant 消息
        com.kingfisher.modules.agent.domain.AgentMessage asstMsg = new com.kingfisher.modules.agent.domain.AgentMessage();
        asstMsg.setConversationId(conversationId);
        asstMsg.setRole("assistant");
        asstMsg.setContent(reply);
        agentMapper.insertMessage(asstMsg);

        return Map.of("id", conversationId, "role", "assistant", "content", reply);
    }

    private String resolveApiKey() {
        // 1. system_configs.llm_api_key
        try {
            var cfg = configService.getByKey("llm_api_key");
            if (cfg != null && cfg.getValue() != null && !cfg.getValue().isBlank()) return cfg.getValue();
        } catch (Exception ignored) {}
        // 2. 环境变量 DEEPSEEK_API_KEY
        String env = System.getenv("DEEPSEEK_API_KEY");
        if (env != null && !env.isBlank()) return env;
        env = System.getProperty("DEEPSEEK_API_KEY");
        if (env != null && !env.isBlank()) return env;
        // 3. 配置文件
        env = System.getenv("LLM_API_KEY");
        return env;
    }

    private String resolveSystemPrompt() {
        try {
            var cfg = configService.getByKey("agent_system_prompt");
            if (cfg != null && cfg.getValue() != null && !cfg.getValue().isBlank()) return cfg.getValue();
        } catch (Exception ignored) {}
        if (systemPromptFallback != null && !systemPromptFallback.isBlank()) return systemPromptFallback;
        return SYSTEM_FALLBACK;
    }

    public static class BizException extends RuntimeException {
        private final int code;
        public BizException(int code, String msg) { super(msg); this.code = code; }
        public int getCode() { return code; }
    }
}
