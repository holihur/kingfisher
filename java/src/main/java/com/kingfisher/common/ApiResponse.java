package com.kingfisher.common;

import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * 统一响应体 {code, message, data}，与 Go core/response 对齐。
 * 分页场景 data 为 PageData，前端 DataTable 依赖此契约。
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
@JsonInclude(JsonInclude.Include.NON_NULL)
public class ApiResponse<T> {

    private int code;
    private String message;
    private T data;

    public static <T> ApiResponse<T> ok(T data) {
        return new ApiResponse<>(ErrorCode.SUCCESS, ErrorCode.message(ErrorCode.SUCCESS), data);
    }

    public static ApiResponse<Void> ok() {
        return new ApiResponse<>(ErrorCode.SUCCESS, ErrorCode.message(ErrorCode.SUCCESS), null);
    }

    public static <T> ApiResponse<T> error(int code) {
        return new ApiResponse<>(code, ErrorCode.message(code), null);
    }

    public static <T> ApiResponse<T> error(int code, String message) {
        return new ApiResponse<>(code, message, null);
    }

    /** 分页响应，data 为 PageData {items, total, page, page_size} */
    public static <T> ApiResponse<PageData<T>> page(T items, long total, int page, int pageSize) {
        return new ApiResponse<>(ErrorCode.SUCCESS, ErrorCode.message(ErrorCode.SUCCESS),
                new PageData<>(items, total, page, pageSize));
    }
}
