package com.kingfisher.modules.user.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import lombok.Data;

/**
 * 修改密码请求，与 Go ChangePwdReq 对齐。
 */
@Data
public class ChangePasswordRequest {

    @NotBlank(message = "旧密码不能为空")
    @JsonProperty("old_password")
    private String oldPassword;

    @NotBlank(message = "新密码不能为空")
    @Size(min = 8, max = 64, message = "密码长度 8-64")
    @JsonProperty("new_password")
    private String newPassword;
}
