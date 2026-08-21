package com.kingfisher.modules.menu.dto;

import jakarta.validation.constraints.NotEmpty;
import lombok.Data;

import java.util.List;

/**
 * 批量删除请求体，与 Go batchIDsReq 对齐。
 */
@Data
public class BatchIDsRequest {

    @NotEmpty(message = "ids 不能为空")
    private List<Long> ids;
}
