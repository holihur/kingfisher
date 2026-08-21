package com.kingfisher.modules.config.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ConfigGroupRequest {

    @NotBlank(message = "name 不能为空")
    private String name;
    private Integer sort;
}
