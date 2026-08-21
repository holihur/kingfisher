package com.kingfisher.modules.doc.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 文档领域实体，对应 documents 表。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Document {

    private Long id;
    @JsonProperty("dir_id")
    private Long dirId;
    private String title;
    private String content;
    @JsonProperty("owner_id")
    private Long ownerId;
    @JsonProperty("owner_name")
    private String ownerName;
    private String visibility;
    private String status;
    @JsonProperty("current_version")
    private Integer currentVersion;
    private Integer sort;
    @JsonProperty("published_at")
    private LocalDateTime publishedAt;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;
}
