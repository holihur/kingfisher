package com.kingfisher.modules.user.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.util.List;

/**
 * 管理员创建用户请求，与 Go CreateUserReq 对齐。
 */
@Data
public class CreateUserRequest {

    @NotBlank(message = "用户名不能为空")
    @Size(min = 3, max = 32, message = "用户名长度 3-32")
    private String username;

    @NotBlank(message = "密码不能为空")
    @Size(min = 8, max = 64, message = "密码长度 8-64")
    private String password;

    private String email;

    @JsonProperty("role_ids")
    private List<Long> roleIds;

    @JsonProperty("dept_ids")
    private List<Long> deptIds;
}
