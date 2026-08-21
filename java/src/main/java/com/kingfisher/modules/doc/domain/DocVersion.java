package com.kingfisher.modules.doc.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;

/**
 * 文档版本历史领域实体，对应 doc_versions 表。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class DocVersion {

    private Long id;
    @JsonProperty("doc_id")
    private Long docId;
    @JsonProperty("version_no")
    private Integer versionNo;
    private String title;
    private String content;
    @JsonProperty("owner_id")
    private Long ownerId;
    @JsonProperty("owner_name")
    private String ownerName;
    private String note;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
}
