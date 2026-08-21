package com.kingfisher.modules.template.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 模版领域实体，对应 templates 表。
 * 与 Go extends/template/domain.Template 1:1 对齐。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Template {
    private Long id;
    private String name;
    private String code;
    @JsonProperty("template_type")
    private String templateType;
    private String title;
    private String content;
    private Integer status;
    private String remark;
    private String version;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
