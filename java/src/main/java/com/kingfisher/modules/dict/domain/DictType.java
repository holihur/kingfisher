package com.kingfisher.modules.dict.domain;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import java.time.LocalDateTime;

@Data
public class DictType {
    private Long id;
    private String code;
    private String name;
    @JsonProperty("is_public")
    private Boolean isPublic;
    private Integer status;
    private String remark;
    private String version;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
