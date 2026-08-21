package com.kingfisher.modules.dict.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.dict.domain.DictEntry;
import com.kingfisher.modules.dict.service.DictEntryService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
public class DictEntryController {

    private final DictEntryService service;
    private static final Defs DEFS = Defs.of(
            "label", new Field("label", FieldType.STRING, true, true),
            "value", new Field("value", FieldType.STRING, true, true),
            "remark", new Field("remark", FieldType.STRING, true, false),
            "status", new Field("status", FieldType.INT, false, true),
            "sort", new Field("sort", FieldType.INT, false, true),
            "created_at", new Field("created_at", FieldType.TIME, false, true)
    );

    public DictEntryController(DictEntryService service) { this.service = service; }

    @RequirePerm("dict:list")
    @GetMapping("/api/v1/dict-types/{id}/entries")
    public ResponseEntity<ApiResponse<Map<String,Object>>> listByType(@PathVariable Long id, HttpServletRequest request) {
        try {
            Query q = Query.parse(request, DEFS);
            List<DictEntry> items = service.listByTypeId(id, q);
            long total = service.countByTypeId(id, q);
            Map<String,Object> data = Map.of("items", items, "total", total, "page", q.getPage(), "page_size", q.getPageSize());
            return ResponseEntity.ok(ApiResponse.ok(data));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @GetMapping("/api/v1/public/dicts/{code}/entries")
    public ResponseEntity<ApiResponse<List<DictEntry>>> getPublic(@PathVariable String code) {
        try {
            List<DictEntry> list = service.listPublicByCode(code);
            return ResponseEntity.ok(ApiResponse.ok(list));
        } catch (IllegalArgumentException e) {
            int codeErr = e.getMessage().contains("未公开") ? ErrorCode.ERR_DICT_TYPE_NOT_PUBLIC : ErrorCode.ERR_DICT_TYPE_NOT_FOUND;
            return ResponseEntity.status(ErrorCode.httpStatus(codeErr)).body(ApiResponse.error(codeErr, e.getMessage()));
        }
    }

    @RequirePerm("dict:create")
    @PostMapping("/api/v1/dict-types/{id}/entries")
    public ResponseEntity<ApiResponse<DictEntry>> create(@PathVariable Long id, @RequestBody Map<String,Object> body) {
        String label = (String) body.get("label");
        String value = (String) body.get("value");
        int sort = body.get("sort") != null ? ((Number) body.get("sort")).intValue() : 0;
        int status = body.get("status") != null ? ((Number) body.get("status")).intValue() : 1;
        String remark = (String) body.getOrDefault("remark", "");
        String version = (String) body.getOrDefault("version", "");
        DictEntry e = service.create(id, label, value, sort, status, remark, version);
        return ResponseEntity.ok(ApiResponse.ok(e));
    }

    @RequirePerm("dict:update")
    @PutMapping("/api/v1/dict-types/{id}/entries/{entryId}")
    public ResponseEntity<ApiResponse<Void>> update(@PathVariable Long id, @PathVariable Long entryId, @RequestBody Map<String,Object> body) {
        String label = (String) body.get("label");
        String value = (String) body.get("value");
        int sort = body.get("sort") != null ? ((Number) body.get("sort")).intValue() : 0;
        int status = body.get("status") != null ? ((Number) body.get("status")).intValue() : 1;
        String remark = (String) body.getOrDefault("remark", "");
        String version = (String) body.getOrDefault("version", "");
        service.update(entryId, id, label, value, sort, status, remark, version);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("dict:delete")
    @DeleteMapping("/api/v1/dict-types/{id}/entries/{entryId}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable Long id, @PathVariable Long entryId) {
        service.delete(entryId);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("dict:delete")
    @PostMapping("/api/v1/dict-types/{id}/entries/batch-delete")
    public ResponseEntity<ApiResponse<Void>> batchDelete(@PathVariable Long id, @RequestBody Map<String,List<Long>> body) {
        service.batchDelete(body.get("ids"));
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("dict:update")
    @PostMapping("/api/v1/dict-types/{id}/entries/batch-status")
    public ResponseEntity<ApiResponse<Void>> batchStatus(@PathVariable Long id, @RequestBody Map<String,Object> body) {
        List<Long> ids = (List<Long>) body.get("ids");
        int status = ((Number) body.get("status")).intValue();
        List<Long> longIds = ids.stream().map(v -> v instanceof Number ? ((Number)v).longValue() : Long.parseLong(String.valueOf(v))).toList();
        service.batchUpdateStatus(longIds, status);
        return ResponseEntity.ok(ApiResponse.ok());
    }
}
