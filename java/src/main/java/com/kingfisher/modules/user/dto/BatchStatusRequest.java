package com.kingfisher.modules.user.dto;

import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.List;

/**
 * 批量更新用户状态请求，与 Go BatchStatusReq 对齐。
 */
@Data
public class BatchStatusRequest {

    @NotEmpty(message = "ids 不能为空")
    private List<Long> ids;

    @NotNull(message = "status 不能为空")
    private Integer status;
}
