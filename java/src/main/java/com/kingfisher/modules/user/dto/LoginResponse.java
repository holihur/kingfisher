package com.kingfisher.modules.user.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.kingfisher.modules.user.domain.User;
import lombok.AllArgsConstructor;
import lombok.Data;

/**
 * 登录响应，与 Go LoginResp 对齐：access_token / refresh_token / user / landing_page
 */
@Data
@AllArgsConstructor
public class LoginResponse {

    @JsonProperty("access_token")
    private String accessToken;

    @JsonProperty("refresh_token")
    private String refreshToken;

    private User user;

    @JsonProperty("landing_page")
    private String landingPage;
}
