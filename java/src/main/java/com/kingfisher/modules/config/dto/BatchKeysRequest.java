package com.kingfisher.modules.config.dto;

import jakarta.validation.constraints.NotEmpty;
import lombok.Data;

import java.util.List;

@Data
public class BatchKeysRequest {

    @NotEmpty(message = "keys 不能为空")
    private List<String> keys;
}
