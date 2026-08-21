package com.kingfisher.modules.user.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.util.List;

/**
 * 管理员更新用户请求，与 Go UpdateUserReq 对齐。
 * 字段均为可选，null 表示不更新。
 */
@Data
public class UpdateUserRequest {

    private String email;
    private Integer status;

    @JsonProperty("role_ids")
    private List<Long> roleIds;

    @JsonProperty("dept_ids")
    private List<Long> deptIds;
}
