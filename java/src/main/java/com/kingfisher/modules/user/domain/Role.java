package com.kingfisher.modules.user.domain;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 角色实体，对应 roles 表。
 */
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
