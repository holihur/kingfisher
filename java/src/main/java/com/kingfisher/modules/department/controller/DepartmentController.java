package com.kingfisher.modules.department.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.department.domain.Department;
import com.kingfisher.modules.department.service.DepartmentService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/departments")
public class DepartmentController {

    private final DepartmentService service;
    public DepartmentController(DepartmentService service) { this.service = service; }

    @RequirePerm("department:list")
    @GetMapping("/tree")
    public ResponseEntity<ApiResponse<List<Department>>> tree() {
        List<Department> tree = service.tree();
        return ResponseEntity.ok(ApiResponse.ok(tree));
    }

    @RequirePerm("department:list")
    @GetMapping
    public ResponseEntity<ApiResponse<List<Department>>> list() {
        List<Department> list = service.list();
        return ResponseEntity.ok(ApiResponse.ok(list));
    }

    @RequirePerm("department:list")
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<Department>> getById(@PathVariable Long id) {
        Department d = service.getById(id);
        if (d == null) return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_DEPT_NOT_FOUND)).body(ApiResponse.error(ErrorCode.ERR_DEPT_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(d));
    }

    @RequirePerm("department:create")
    @PostMapping
    public ResponseEntity<ApiResponse<Department>> create(@RequestBody Department dept) {
        try {
            Department d = service.create(dept);
            return ResponseEntity.ok(ApiResponse.ok(d));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("department:update")
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> update(@PathVariable Long id, @RequestBody Department dept) {
        try {
            service.update(id, dept);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("department:update")
    @PutMapping("/{id}/roles")
    public ResponseEntity<ApiResponse<Void>> assignRoles(@PathVariable Long id, @RequestBody Map<String, List<Long>> body) {
        List<Long> roleIds = body.get("role_ids");
        if (roleIds == null) roleIds = body.get("roleIds");
        service.assignRoles(id, roleIds);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("department:delete")
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable Long id) {
        try {
            service.delete(id);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            int code = e.getMessage().contains("子部门") ? ErrorCode.ERR_DEPT_HAS_CHILDREN : ErrorCode.ERR_INVALID_PARAM;
            return ResponseEntity.status(ErrorCode.httpStatus(code)).body(ApiResponse.error(code, e.getMessage()));
        }
    }
}
