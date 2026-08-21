package com.kingfisher.modules.menu.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

/**
 * 创建菜单请求体，与 Go handler 直接绑定 domain.Menu 对齐。
 */
@Data
public class CreateMenuRequest {

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
}
