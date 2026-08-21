package com.kingfisher.modules.doc.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

/**
 * 文档目录领域实体，对应 doc_directories 表。
 * 与 Go extends/doc/domain.DocDirectory 1:1 对齐。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class DocDirectory {
    private Long id;
    @JsonProperty("parent_id")
    private Long parentId;
    private String name;
    private Integer sort;
    private Integer status;
    private String version;
    private String visibility;
    @JsonProperty("docs")
    private List<DocTreeItem> docs;
    private List<DocDirectory> children;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
