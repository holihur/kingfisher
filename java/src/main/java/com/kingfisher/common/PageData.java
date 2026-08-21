package com.kingfisher.common;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Data;

/**
 * 分页响应包装，与 Go response.PageData 对齐。
 * 前端 DataTable 依赖 items/total/page/page_size 契约。
 */
@Data
@AllArgsConstructor
public class PageData<T> {
    private T items;
    private long total;
    private int page;
    @JsonProperty("page_size")
    private int pageSize;
}
