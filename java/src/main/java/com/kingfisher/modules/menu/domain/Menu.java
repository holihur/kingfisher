package com.kingfisher.modules.menu.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

/**
 * 菜单领域实体，对应 menus 表。
 * 与 Go extends/menu/domain.Menu 1:1 对齐。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Menu {

    private Long id;
    @JsonProperty("parent_id")
    private Long parentId;
    private String name;
    private String path;
    private String component;
    private String icon;
    private Integer sort;
    private Integer type;
    private String permission;
    private Integer status;
    private String version;
    private List<Menu> children;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
