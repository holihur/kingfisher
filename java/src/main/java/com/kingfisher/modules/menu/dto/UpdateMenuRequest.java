package com.kingfisher.modules.menu.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

/**
 * 更新菜单请求体，所有字段可选（部分更新），与 Go updateMenuReq 对齐。
 */
@Data
public class UpdateMenuRequest {

    private String name;
    private String icon;
    private String path;
    private String component;
    private Integer sort;
    @JsonProperty("parent_id")
    private Long parentId;
    private Integer status;
    private String version;
}
