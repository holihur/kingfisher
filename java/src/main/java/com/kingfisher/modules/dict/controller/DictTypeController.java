package com.kingfisher.modules.dict.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.dict.domain.DictType;
import com.kingfisher.modules.dict.service.DictTypeService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/dict-types")
public class DictTypeController {

    private final DictTypeService service;
    private static final Defs DEFS = Defs.of(
            "code", new Field("code", FieldType.STRING, true, true),
            "name", new Field("name", FieldType.STRING, true, true),
            "remark", new Field("remark", FieldType.STRING, true, false),
            "is_public", new Field("is_public", FieldType.BOOL, false, true),
            "status", new Field("status", FieldType.INT, false, true),
            "created_at", new Field("created_at", FieldType.TIME, false, true)
    );

    public DictTypeController(DictTypeService service) { this.service = service; }

    @RequirePerm("dict:list")
    @GetMapping
    public ResponseEntity<ApiResponse<Map<String,Object>>> list(HttpServletRequest request) {
        try {
            Query q = Query.parse(request, DEFS);
            List<DictType> items = service.list(q);
            long total = service.count(q);
            Map<String,Object> data = Map.of("items", items, "total", total, "page", q.getPage(), "page_size", q.getPageSize());
            return ResponseEntity.ok(ApiResponse.ok(data));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("dict:list")
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<DictType>> getById(@PathVariable Long id) {
        DictType t = service.getById(id);
        if (t == null) return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_DICT_TYPE_NOT_FOUND)).body(ApiResponse.error(ErrorCode.ERR_DICT_TYPE_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(t));
    }

    @RequirePerm("dict:create")
    @PostMapping
    public ResponseEntity<ApiResponse<DictType>> create(@RequestBody Map<String,Object> body) {
        try {
            String code = (String) body.get("code");
            String name = (String) body.get("name");
            Boolean isPublic = body.get("is_public") != null ? (Boolean) body.get("is_public") : false;
            Integer status = body.get("status") != null ? ((Number) body.get("status")).intValue() : 1;
            String remark = (String) body.getOrDefault("remark", "");
            String version = (String) body.getOrDefault("version", "");
            DictType t = service.create(code, name, isPublic, status, remark, version);
            return ResponseEntity.ok(ApiResponse.ok(t));
        } catch (IllegalArgumentException e) {
            int code = e.getMessage().contains("编码已存在") ? ErrorCode.ERR_DICT_TYPE_CODE_EXISTS : ErrorCode.ERR_INVALID_PARAM;
            return ResponseEntity.status(ErrorCode.httpStatus(code)).body(ApiResponse.error(code, e.getMessage()));
        }
    }

    @RequirePerm("dict:update")
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> update(@PathVariable Long id, @RequestBody Map<String,Object> body) {
        try {
            String code = (String) body.get("code");
            String name = (String) body.get("name");
            Boolean isPublic = body.get("is_public") != null ? (Boolean) body.get("is_public") : false;
            Integer status = body.get("status") != null ? ((Number) body.get("status")).intValue() : 1;
            String remark = (String) body.getOrDefault("remark", "");
            String version = (String) body.getOrDefault("version", "");
            service.update(id, code, name, isPublic, status, remark, version);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            int code = e.getMessage().contains("编码已存在") ? ErrorCode.ERR_DICT_TYPE_CODE_EXISTS : ErrorCode.ERR_INVALID_PARAM;
            return ResponseEntity.status(ErrorCode.httpStatus(code)).body(ApiResponse.error(code, e.getMessage()));
        }
    }

    @RequirePerm("dict:delete")
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable Long id) {
        try {
            service.delete(id);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            int code = e.getMessage().contains("存在条目") ? ErrorCode.ERR_DICT_TYPE_HAS_ENTRIES : ErrorCode.ERR_DICT_TYPE_NOT_FOUND;
            return ResponseEntity.status(ErrorCode.httpStatus(code)).body(ApiResponse.error(code, e.getMessage()));
        }
    }

    @RequirePerm("dict:delete")
    @PostMapping("/batch-delete")
    public ResponseEntity<ApiResponse<Void>> batchDelete(@RequestBody Map<String, List<Long>> body) {
        service.batchDelete(body.get("ids"));
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("dict:update")
    @PostMapping("/batch-status")
    public ResponseEntity<ApiResponse<Void>> batchStatus(@RequestBody Map<String,Object> body) {
        List<Long> ids = (List<Long>) body.get("ids");
        int status = ((Number) body.get("status")).intValue();
        // 兼容 Long/Integer
        List<Long> longIds = ids.stream().map(v -> v instanceof Number ? ((Number) v).longValue() : Long.parseLong(String.valueOf(v))).toList();
        service.batchUpdateStatus(longIds, status);
        return ResponseEntity.ok(ApiResponse.ok());
    }
}
