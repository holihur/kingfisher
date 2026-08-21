package com.kingfisher.modules.config.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class SetConfigRequest {

    @NotNull(message = "value 不能为空")
    private String value;
    @JsonProperty("is_public")
    private Boolean isPublic;
    private String version;
    private String render;
    @JsonProperty("render_options")
    private String renderOptions;
    @JsonProperty("group_id")
    private Long groupId;
}
