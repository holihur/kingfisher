package com.kingfisher.modules.rbac.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.rbac.domain.Permission;
import com.kingfisher.modules.rbac.domain.Role;
import com.kingfisher.modules.rbac.service.RoleService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/roles")
public class RoleController {

    private final RoleService roleService;

    private static final Defs ROLE_DEFS = Defs.of(
            "name", new Field("name", FieldType.STRING, true, true),
            "code", new Field("code", FieldType.STRING, true, true),
            "description", new Field("description", FieldType.STRING, true, false),
            "status", new Field("status", FieldType.INT, false, true),
            "level", new Field("level", FieldType.INT, false, true),
            "created_at", new Field("created_at", FieldType.TIME, false, true)
    );

    public RoleController(RoleService roleService) { this.roleService = roleService; }

    @RequirePerm("role:list")
    @GetMapping
    public ResponseEntity<ApiResponse<Map<String, Object>>> list(HttpServletRequest request) {
        try {
            Query q = Query.parse(request, ROLE_DEFS);
            List<Role> items = roleService.list(q);
            long total = roleService.count(q);
            Map<String, Object> data = Map.of("items", items, "total", total, "page", q.getPage(), "page_size", q.getPageSize());
            return ResponseEntity.ok(ApiResponse.ok(data));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_INVALID_PARAM))
                    .body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("role:list")
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<Role>> getById(@PathVariable Long id) {
        Role r = roleService.getById(id);
        if (r == null) return ResponseEntity.status(404).body(ApiResponse.error(ErrorCode.ERR_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(r));
    }

    @RequirePerm("role:create")
    @PostMapping
    public ResponseEntity<ApiResponse<Role>> create(@RequestBody Role role) {
        try {
            roleService.create(role);
            return ResponseEntity.ok(ApiResponse.ok(role));
        } catch (Exception e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("role:update")
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> update(@PathVariable Long id, @RequestBody Map<String, Object> body) {
        try {
            String name = (String) body.get("name");
            String desc = (String) body.get("description");
            Integer status = body.get("status") != null ? ((Number) body.get("status")).intValue() : null;
            roleService.update(id, name, desc, status);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("role:delete")
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable Long id) {
        try {
            roleService.delete(id);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("role:delete")
    @PostMapping("/batch-delete")
    public ResponseEntity<ApiResponse<Void>> batchDelete(@RequestBody Map<String, List<Long>> body) {
        try {
            List<Long> ids = body.get("ids");
            roleService.batchDelete(ids);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("role:update")
    @PostMapping("/batch-status")
    public ResponseEntity<ApiResponse<Void>> batchStatus(@RequestBody Map<String, Object> body) {
        try {
            List<Long> ids = (List<Long>) body.get("ids");
            // 兼容 Integer/Long
            Object s = body.get("status");
            int status = s instanceof Number ? ((Number) s).intValue() : Integer.parseInt(String.valueOf(s));
            // 转换 ids 类型
            List<Long> longIds = ids.stream().map(v -> v instanceof Number ? ((Number) v).longValue() : Long.parseLong(String.valueOf(v))).toList();
            roleService.batchUpdateStatus(longIds, status);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("role:list")
    @GetMapping("/{id}/permissions")
    public ResponseEntity<ApiResponse<List<Permission>>> getPermissions(@PathVariable Long id) {
        List<Permission> perms = roleService.getRolePermissions(id);
        return ResponseEntity.ok(ApiResponse.ok(perms));
    }

    @RequirePerm("role:update")
    @PutMapping("/{id}/permissions")
    public ResponseEntity<ApiResponse<Void>> assignPerms(@PathVariable Long id, @RequestBody Map<String, List<Long>> body) {
        List<Long> permIds = body.get("permission_ids");
        if (permIds == null) permIds = body.get("permissionIds");
        roleService.assignPermissions(id, permIds);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("role:list")
    @GetMapping("/{id}/menus")
    public ResponseEntity<ApiResponse<List<Long>>> getMenus(@PathVariable Long id) {
        List<Long> mids = roleService.getRoleMenus(id);
        return ResponseEntity.ok(ApiResponse.ok(mids));
    }

    @RequirePerm("role:update")
    @PutMapping("/{id}/menus")
    public ResponseEntity<ApiResponse<Void>> assignMenus(@PathVariable Long id, @RequestBody Map<String, List<Long>> body) {
        List<Long> menuIds = body.get("menu_ids");
        if (menuIds == null) menuIds = body.get("menuIds");
        roleService.assignMenus(id, menuIds);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("role:list")
    @GetMapping("/{id}/data-scope")
    public ResponseEntity<ApiResponse<Map<String, Object>>> getDataScope(@PathVariable Long id, @RequestParam(required = false) String resource) {
        String res = resource != null ? resource : "worktask";
        String scope = roleService.getDataScope(id, res);
        return ResponseEntity.ok(ApiResponse.ok(Map.of("resource", res, "scope_type", scope != null ? scope : "")));
    }

    @RequirePerm("role:update")
    @PutMapping("/{id}/data-scope")
    public ResponseEntity<ApiResponse<Void>> setDataScope(@PathVariable Long id, @RequestBody Map<String, String> body) {
        String resource = body.getOrDefault("resource", "worktask");
        String scopeType = body.get("scope_type");
        if (scopeType == null) scopeType = body.get("scopeType");
        if (scopeType == null || scopeType.isBlank()) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "scope_type 不能为空"));
        if (!List.of("all", "self", "department", "department_subtree").contains(scopeType)) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "非法 scope_type"));
        }
        roleService.setDataScope(id, resource, scopeType);
        return ResponseEntity.ok(ApiResponse.ok());
    }
}
