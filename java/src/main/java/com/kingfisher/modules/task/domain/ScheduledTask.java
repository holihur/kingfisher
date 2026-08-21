package com.kingfisher.modules.task.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 周期任务配置领域实体，对应 scheduled_tasks 表。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class ScheduledTask {

    private Long id;
    private String name;
    @JsonProperty("task_type")
    private String taskType;
    @JsonProperty("cron_spec")
    private String cronSpec;
    private String payload;
    private Integer enabled;
    private String remark;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
