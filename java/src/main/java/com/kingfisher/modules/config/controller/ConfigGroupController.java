package com.kingfisher.modules.config.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.config.domain.ConfigGroup;
import com.kingfisher.modules.config.dto.ConfigGroupRequest;
import com.kingfisher.modules.config.service.ConfigService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/v1/config-groups")
public class ConfigGroupController {

    private final ConfigService configService;
    public ConfigGroupController(ConfigService configService) { this.configService = configService; }

    @RequirePerm("config:list")
    @GetMapping
    public ResponseEntity<ApiResponse<List<ConfigGroup>>> list() {
        List<ConfigGroup> groups = configService.listGroups();
        return ResponseEntity.ok(ApiResponse.ok(groups));
    }

    @RequirePerm("config:update")
    @PostMapping
    public ResponseEntity<ApiResponse<ConfigGroup>> create(@RequestBody ConfigGroupRequest req) {
        if (req.getName() == null || req.getName().isBlank()) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "name 不能为空"));
        ConfigGroup g = configService.createGroup(req.getName(), req.getSort() != null ? req.getSort() : 0);
        return ResponseEntity.ok(ApiResponse.ok(g));
    }

    @RequirePerm("config:update")
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> update(@PathVariable Long id, @RequestBody ConfigGroupRequest req) {
        configService.updateGroup(id, req.getName(), req.getSort() != null ? req.getSort() : 0);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("config:update")
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable Long id) {
        configService.deleteGroup(id);
        return ResponseEntity.ok(ApiResponse.ok());
    }
}
