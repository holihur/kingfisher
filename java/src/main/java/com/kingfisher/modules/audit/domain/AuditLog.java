package com.kingfisher.modules.audit.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 审计日志领域实体，对应 audit_logs 表。
 * 与 Go extends/audit/domain.AuditLog 1:1 对齐。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class AuditLog {

    private Long id;
    @JsonProperty("user_id")
    private Long userId;
    private String username;
    private String action;
    private String resource;
    @JsonProperty("resource_id")
    private Long resourceId;
    /** 操作详情（JSON 字符串） */
    private String detail;
    /** 操作结果：success | failure */
    private String result;
    /** 操作耗时（毫秒） */
    private Long latency;
    private String message;
    private String ip;
    @JsonProperty("user_agent")
    private String userAgent;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
}
