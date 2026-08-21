package com.kingfisher.modules.worktask.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.PageData;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.worktask.domain.WorkTask;
import com.kingfisher.modules.worktask.service.WorktaskService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/tasks")
public class WorktaskController {

    private final WorktaskService worktaskService;
    private final com.kingfisher.modules.rbac.service.RoleService roleService;
    public WorktaskController(WorktaskService worktaskService, com.kingfisher.modules.rbac.service.RoleService roleService) { this.worktaskService = worktaskService; this.roleService = roleService; }

    private Long getUserId(HttpServletRequest req) {
        Long uid = (Long) req.getAttribute("user_id");
        if (uid != null) return uid;
        Object c = req.getAttribute("claims");
        if (c instanceof io.jsonwebtoken.Claims cl) return cl.get("user_id", Long.class);
        return 0L;
    }
    private boolean isAdmin(HttpServletRequest req) {
        List<String> roles = (List<String>) req.getAttribute("roles");
        if (roles == null) {
            Object c = req.getAttribute("claims");
            if (c instanceof io.jsonwebtoken.Claims cl) roles = cl.get("roles", List.class);
        }
        return roles != null && roles.contains("admin");
    }

    @SuppressWarnings("unchecked")
    private List<Long> getRoleIds(HttpServletRequest req) {
        Object v = req.getAttribute("role_ids");
        if (v instanceof List<?> list) {
            List<Long> r = new java.util.ArrayList<>();
            for (Object o : list) if (o instanceof Number n) r.add(n.longValue());
            return r;
        }
        Object c = req.getAttribute("claims");
        if (c instanceof io.jsonwebtoken.Claims cl) {
            Object ids = cl.get("role_ids");
            if (ids instanceof List<?> list) {
                List<Long> r = new java.util.ArrayList<>();
                for (Object o : list) if (o instanceof Number n) r.add(n.longValue());
                return r;
            }
        }
        return List.of();
    }

    @SuppressWarnings("unchecked")
    private List<String> getRoleCodes(HttpServletRequest req) {
        Object v = req.getAttribute("roles");
        if (v instanceof List<?> list) {
            List<String> r = new java.util.ArrayList<>();
            for (Object o : list) r.add(String.valueOf(o));
            return r;
        }
        Object c = req.getAttribute("claims");
        if (c instanceof io.jsonwebtoken.Claims cl) {
            Object roles = cl.get("roles");
            if (roles instanceof List<?> list) {
                List<String> r = new java.util.ArrayList<>();
                for (Object o : list) r.add(String.valueOf(o));
                return r;
            }
        }
        return List.of();
    }

    private com.kingfisher.common.dataaccess.Scope resolveScope(HttpServletRequest req) {
        Long uid = getUserId(req);
        List<Long> roleIds = getRoleIds(req);
        List<String> roleCodes = getRoleCodes(req);
        try {
            return roleService.resolveDataScope(uid, roleIds, roleCodes);
        } catch (Exception e) {
            return com.kingfisher.common.dataaccess.Scope.self("owner_id", uid);
        }
    }

    private static final Defs TASK_DEFS = Defs.of(
            "title", Field.searchableString("title"),
            "status", Field.filterable(FieldType.STRING, "status"),
            "created_at", Field.filterable(FieldType.TIME, "created_at")
    );

    @RequirePerm("worktask:list")
    @GetMapping
    public ResponseEntity<ApiResponse<PageData<List<WorkTask>>>> list(HttpServletRequest request) {
        Query q = Query.parse(request, TASK_DEFS);
        var scope = resolveScope(request);
        List<WorkTask> items = worktaskService.list(q, scope);
        long total = worktaskService.count(q, scope);
        return ResponseEntity.ok(ApiResponse.ok(new PageData<>(items, total, q.getPage(), q.getPageSize())));
    }

    @RequirePerm("worktask:list")
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<WorkTask>> getById(HttpServletRequest request, @PathVariable Long id) {
        WorkTask t = worktaskService.getById(id, resolveScope(request));
        if (t == null) return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_NOT_FOUND)).body(ApiResponse.error(ErrorCode.ERR_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(t));
    }

    @RequirePerm("worktask:create")
    @PostMapping
    public ResponseEntity<ApiResponse<WorkTask>> create(HttpServletRequest request, @RequestBody Map<String,Object> body) {
        String title = (String) body.get("title");
        if (title == null || title.isBlank()) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "title 不能为空"));
        String description = (String) body.getOrDefault("description", "");
        Long deptId = body.get("department_id") != null ? Long.valueOf(String.valueOf(body.get("department_id"))) : (body.get("departmentId") != null ? Long.valueOf(String.valueOf(body.get("departmentId"))) : 0L);
        WorkTask t = worktaskService.create(title, description, getUserId(request), deptId, "pending");
        return ResponseEntity.ok(ApiResponse.ok(t));
    }

    @RequirePerm("worktask:update")
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> update(@PathVariable Long id, @RequestBody Map<String,Object> body) {
        worktaskService.update(id, body);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("worktask:delete")
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(HttpServletRequest request, @PathVariable Long id) {
        worktaskService.delete(id, resolveScope(request));
        return ResponseEntity.ok(ApiResponse.ok());
    }
}
