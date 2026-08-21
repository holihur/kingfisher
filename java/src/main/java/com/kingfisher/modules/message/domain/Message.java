package com.kingfisher.modules.message.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 站内信消息领域实体，对应 messages 表。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Message {

    private Long id;
    @JsonProperty("sender_id")
    private Long senderId;
    @JsonProperty("sender_type")
    private String senderType;
    @JsonProperty("recipient_id")
    private Long recipientId;
    @JsonProperty("batch_id")
    private Long batchId;
    @JsonProperty("recipient_name")
    private String recipientName;
    private String title;
    private String content;
    private String status;
    @JsonProperty("is_read")
    private Boolean isRead;
    @JsonProperty("read_at")
    private LocalDateTime readAt;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
