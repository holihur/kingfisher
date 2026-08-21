package com.kingfisher.modules.rbac.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.rbac.domain.Permission;
import com.kingfisher.modules.rbac.service.PermissionService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/v1/permissions")
public class PermissionController {
    private final PermissionService service;
    public PermissionController(PermissionService service) { this.service = service; }

    @RequirePerm("role:list")
    @GetMapping
    public ResponseEntity<ApiResponse<List<Permission>>> list() {
        List<Permission> perms = service.list();
        return ResponseEntity.ok(ApiResponse.ok(perms));
    }
}
