package com.kingfisher.modules.agent.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * Agent 会话消息领域实体，对应 agent_messages 表。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class AgentMessage {

    private Long id;
    @JsonProperty("conversation_id")
    private Long conversationId;
    private String role;
    private String content;
    @JsonProperty("tool_calls")
    private String toolCalls;
    @JsonProperty("tool_result")
    private String toolResult;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
}
