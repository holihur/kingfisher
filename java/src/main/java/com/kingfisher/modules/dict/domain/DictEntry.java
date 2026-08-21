package com.kingfisher.modules.dict.domain;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import java.time.LocalDateTime;

@Data
public class DictEntry {
    private Long id;
    @JsonProperty("type_id")
    private Long typeId;
    private String label;
    private String value;
    private Integer sort;
    private Integer status;
    private String remark;
    private String version;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
