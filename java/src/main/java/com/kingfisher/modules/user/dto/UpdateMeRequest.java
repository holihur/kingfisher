package com.kingfisher.modules.user.dto;

import lombok.Data;

/**
 * 当前用户更新个人资料请求，与 Go UpdateMeReq 对齐。
 */
@Data
public class UpdateMeRequest {
    private String email;
    private String nickname;
    private String avatar;
}
