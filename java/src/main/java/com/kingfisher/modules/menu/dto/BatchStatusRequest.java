package com.kingfisher.modules.menu.dto;

import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.List;

/**
 * 批量更新状态请求体，与 Go batchStatusReq 对齐。
 */
@Data
public class BatchStatusRequest {

    @NotEmpty(message = "ids 不能为空")
    private List<Long> ids;

    @NotNull(message = "status 不能为空")
    private Integer status;
}
