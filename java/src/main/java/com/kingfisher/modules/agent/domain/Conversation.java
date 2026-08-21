package com.kingfisher.modules.agent.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * Agent 聊天会话领域实体，对应 agent_conversations 表。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Conversation {

    private Long id;
    @JsonProperty("user_id")
    private Long userId;
    private String title;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
