package com.kingfisher.modules.doc.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

/**
 * 目录树中的文档叶子节点（轻量，不含正文）。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class DocTreeItem {
    private Long id;
    private String title;
    private String status;
    private String visibility;
}
