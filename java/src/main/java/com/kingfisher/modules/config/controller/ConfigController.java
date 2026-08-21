package com.kingfisher.modules.config.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.config.domain.SystemConfig;
import com.kingfisher.modules.config.dto.BatchKeysRequest;
import com.kingfisher.modules.config.dto.SetConfigRequest;
import com.kingfisher.modules.config.service.ConfigService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import jakarta.servlet.http.HttpServletRequest;
import java.io.File;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1")
public class ConfigController {

    private final ConfigService configService;

    private static final Defs CONFIG_DEFS = Defs.of(
            "key", new Field("key", FieldType.STRING, true, true),
            "value", new Field("value", FieldType.STRING, true, false),
            "remark", new Field("remark", FieldType.STRING, true, false),
            "is_public", new Field("is_public", FieldType.BOOL, false, true),
            "version", new Field("version", FieldType.STRING, false, true),
            "render", new Field("render", FieldType.STRING, false, true),
            "group_id", new Field("group_id", FieldType.UINT, false, true),
            "created_at", new Field("created_at", FieldType.TIME, false, true),
            "updated_at", new Field("updated_at", FieldType.TIME, false, true)
    );

    public ConfigController(ConfigService configService) { this.configService = configService; }

    @GetMapping("/public/configs")
    public ResponseEntity<ApiResponse<List<SystemConfig>>> getPublicAll() {
        List<SystemConfig> list = configService.getAllPublic();
        return ResponseEntity.ok(ApiResponse.ok(list));
    }

    @GetMapping("/public/configs/{key}")
    public ResponseEntity<ApiResponse<SystemConfig>> getPublic(@PathVariable String key) {
        SystemConfig c = configService.getPublicByKey(key);
        if (c == null) return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_CONFIG_NOT_FOUND)).body(ApiResponse.error(ErrorCode.ERR_CONFIG_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(c));
    }

    @RequirePerm("config:list")
    @GetMapping("/configs")
    public ResponseEntity<ApiResponse<Map<String, Object>>> getAll(HttpServletRequest request) {
        try {
            Query q = Query.parse(request, CONFIG_DEFS);
            List<SystemConfig> items = configService.list(q);
            long total = configService.count(q);
            Map<String, Object> data = Map.of("items", items, "total", total, "page", q.getPage(), "page_size", q.getPageSize());
            return ResponseEntity.ok(ApiResponse.ok(data));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_INVALID_PARAM)).body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("config:list")
    @GetMapping("/configs/{key}")
    public ResponseEntity<ApiResponse<SystemConfig>> get(@PathVariable String key) {
        SystemConfig c = configService.getByKey(key);
        if (c == null) return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_CONFIG_NOT_FOUND)).body(ApiResponse.error(ErrorCode.ERR_CONFIG_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(c));
    }

    @RequirePerm("config:update")
    @PutMapping("/configs/{key}")
    public ResponseEntity<ApiResponse<Void>> set(@PathVariable String key, @RequestBody SetConfigRequest req) {
        if (req.getValue() == null) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "value 不能为空"));
        boolean isPublic = req.getIsPublic() != null ? req.getIsPublic() : false;
        // 若未传 is_public，尝试保留原值
        if (req.getIsPublic() == null) {
            SystemConfig existing = configService.getByKey(key);
            if (existing != null) isPublic = existing.getIsPublic() != null && existing.getIsPublic();
        }
        configService.set(key, req.getValue(), isPublic, req.getVersion(), req.getRender(), req.getRenderOptions(), req.getGroupId());
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("config:update")
    @DeleteMapping("/configs/{key}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable String key) {
        configService.delete(key);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("config:update")
    @PostMapping("/configs/batch-delete")
    public ResponseEntity<ApiResponse<Void>> batchDelete(@RequestBody BatchKeysRequest req) {
        configService.batchDelete(req.getKeys());
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("config:update")
    @PostMapping("/configs/upload-image")
    public ResponseEntity<ApiResponse<Map<String, String>>> uploadImage(@RequestParam("file") MultipartFile file) {
        if (file.isEmpty()) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "请选择文件"));
        String filename = file.getOriginalFilename() != null ? file.getOriginalFilename().toLowerCase() : "";
        if (!(filename.endsWith(".png") || filename.endsWith(".jpg") || filename.endsWith(".jpeg") || filename.endsWith(".gif") || filename.endsWith(".webp"))) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "不支持的文件类型，仅支持 png/jpg/jpeg/gif/webp"));
        }
        if (file.getSize() > 2 * 1024 * 1024) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "文件大小不能超过 2MB"));
        try {
            String ext = filename.substring(filename.lastIndexOf('.'));
            String newName = System.nanoTime() + ext;
            File dir = new File("uploads/configs");
            if (!dir.exists()) dir.mkdirs();
            File dest = new File(dir, newName);
            file.transferTo(dest);
            Map<String, String> data = Map.of("url", "/uploads/configs/" + newName);
            return ResponseEntity.ok(ApiResponse.ok(data));
        } catch (Exception e) {
            return ResponseEntity.internalServerError().body(ApiResponse.error(ErrorCode.ERR_INTERNAL));
        }
    }
}
