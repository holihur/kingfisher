package com.kingfisher.modules.rbac.domain;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import java.time.LocalDateTime;

@Data
public class Role {
    private Long id;
    private String name;
    private String code;
    private String description;
    private Integer status;
    private Integer level;
    @JsonProperty("landing_page")
    private String landingPage;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
