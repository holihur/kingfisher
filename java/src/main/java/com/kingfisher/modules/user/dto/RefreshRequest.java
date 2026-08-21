package com.kingfisher.modules.user.dto;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.annotation.JsonProperty;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

/**
 * 刷新请求，兼容 refresh_token / refreshToken 两种写法
 */
@Data
public class RefreshRequest {
    @NotBlank(message = "refresh_token 不能为空")
    @JsonAlias({"refresh_token", "refreshToken"})
    @JsonProperty("refresh_token")
    private String refreshToken;
}
