package com.kingfisher.modules.department.domain;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import java.time.LocalDateTime;
import java.util.List;

@Data
public class Department {
    private Long id;
    @JsonProperty("parent_id")
    private Long parentId;
    private String name;
    private Integer sort;
    private Integer status;
    private String remark;
    private String version;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
    @JsonProperty("role_ids")
    private List<Long> roleIds;
    private List<Department> children;
}
