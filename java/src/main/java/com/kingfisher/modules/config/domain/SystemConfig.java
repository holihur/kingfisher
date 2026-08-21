package com.kingfisher.modules.config.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class SystemConfig {

    private Long id;
    private String key;
    private String value;
    private String remark;
    @JsonProperty("is_public")
    private Boolean isPublic;
    private String version;
    private String render;
    @JsonProperty("render_options")
    private String renderOptions;
    @JsonProperty("group_id")
    private Long groupId;
    @JsonProperty("group_name")
    private String groupName;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
