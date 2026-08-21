package com.kingfisher.modules.user.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

/**
 * 登录请求，与 Go LoginReq 对齐。
 */
@Data
public class LoginRequest {
    @NotBlank(message = "用户名不能为空")
    private String username;

    @NotBlank(message = "密码不能为空")
    private String password;
}
