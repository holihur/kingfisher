package com.kingfisher.modules.worktask.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 工作任务领域实体，对应 tasks 表。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class WorkTask {

    private Long id;
    private String title;
    private String description;
    @JsonProperty("owner_id")
    private Long ownerId;
    @JsonProperty("department_id")
    private Long departmentId;
    private String status;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
