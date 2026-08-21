package com.kingfisher.modules.user.dto;

import jakarta.validation.constraints.NotEmpty;
import lombok.Data;

import java.util.List;

/**
 * 批量操作请求（批量删除等），与 Go BatchUserOp 对齐。
 */
@Data
public class BatchIdsRequest {

    @NotEmpty(message = "ids 不能为空")
    private List<Long> ids;
}
