package com.kingfisher.modules.rbac.domain;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import java.time.LocalDateTime;

@Data
public class Permission {
    private Long id;
    private String name;
    private String code;
    private String resource;
    private String action;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
