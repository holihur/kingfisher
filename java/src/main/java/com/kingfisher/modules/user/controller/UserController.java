package com.kingfisher.modules.user.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.PageData;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.rbac.service.RoleService;
import com.kingfisher.modules.user.domain.User;
import com.kingfisher.modules.user.dto.*;
import com.kingfisher.modules.user.mapper.UserMapper;
import com.kingfisher.modules.user.service.UserService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;

@RestController
@RequestMapping("/api/v1/users")
public class UserController {

    private static final Defs USER_DEFS = Defs.of(
            "username", Field.searchableString("username"),
            "email", Field.searchableString("email"),
            "nickname", Field.searchableString("nickname"),
            "status", Field.filterable(FieldType.INT, "status"),
            "role_id", Field.filterable(FieldType.UINT, "role_id"),
            "department_id", Field.filterable(FieldType.UINT, "department_id"),
            "created_at", Field.filterable(FieldType.TIME, "created_at"),
            "updated_at", Field.filterable(FieldType.TIME, "updated_at")
    );

    private final UserMapper userMapper;
    private final UserService userService;
    private final RoleService roleService;
    private final com.kingfisher.modules.audit.service.AuditService auditService;

    public UserController(UserMapper userMapper, UserService userService, RoleService roleService,
                          com.kingfisher.modules.audit.service.AuditService auditService) {
        this.userMapper = userMapper;
        this.userService = userService;
        this.roleService = roleService;
        this.auditService = auditService;
    }

    // ==================== /me 端点（无需 RBAC） ====================

    @GetMapping("/me")
    public ApiResponse<User> me(HttpServletRequest request) {
        Long userId = getUserId(request);
        User user = userService.getById(userId);
        if (user != null) user.setPassword(null);
        return ApiResponse.ok(user);
    }

    @PutMapping("/me")
    public ApiResponse<User> updateMe(HttpServletRequest request, @Valid @RequestBody UpdateMeRequest req) {
        Long userId = getUserId(request);
        userService.updateProfile(userId, req.getEmail(), req.getNickname(), req.getAvatar());
        User user = userService.getById(userId);
        if (user != null) user.setPassword(null);
        return ApiResponse.ok(user);
    }

    @GetMapping("/me/permissions")
    public ApiResponse<List<String>> myPermissions(HttpServletRequest request) {
        Long userId = getUserId(request);
        // 与 Go 的 RoleService.GetUserPermissions 1:1：直接角色 ∪ 部门继承角色 → role_permissions → permissions.code
        List<String> perms = roleService.getUserPermissions(userId);
        if (perms == null) return ApiResponse.ok(List.of());
        return ApiResponse.ok(perms);
    }

    @GetMapping("/me/login-logs")
    public ApiResponse<PageData<List<com.kingfisher.modules.audit.domain.AuditLog>>> myLoginLogs(HttpServletRequest request) {
        Long userId = getUserId(request);
        return ApiResponse.ok(auditService.listLoginLogs(userId, request));
    }

    @PostMapping("/me/avatar")
    public ResponseEntity<ApiResponse<Map<String, String>>> uploadAvatar(
            HttpServletRequest request, @RequestParam("file") MultipartFile file) {
        Long userId = getUserId(request);
        String ext = getExtension(file.getOriginalFilename());
        if (!List.of(".png", ".jpg", ".jpeg", ".gif", ".webp").contains(ext)) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "不支持的文件类型，仅支持 png/jpg/jpeg/gif/webp"));
        }
        if (file.getSize() > 2 * 1024 * 1024) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "文件大小不能超过 2MB"));
        }
        String detected = file.getContentType();
        if (detected == null || !detected.startsWith("image/")) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "不支持的文件内容，仅支持图片文件"));
        }

        String filename = userId + "_" + System.nanoTime() + ext;
        Path uploadDir = Paths.get("uploads", "avatars");
        try {
            Files.createDirectories(uploadDir);
            Path savePath = uploadDir.resolve(filename);
            file.transferTo(savePath.toFile());
        } catch (IOException e) {
            return ResponseEntity.internalServerError().body(ApiResponse.error(ErrorCode.ERR_INTERNAL));
        }
        String avatarUrl = "/uploads/avatars/" + filename;
        userService.updateProfile(userId, "", "", avatarUrl);
        return ResponseEntity.ok(ApiResponse.ok(Map.of("url", avatarUrl)));
    }

    @PutMapping("/me/password")
    public ResponseEntity<ApiResponse<Void>> changePassword(HttpServletRequest request,
                                                            @Valid @RequestBody ChangePasswordRequest req) {
        Long userId = getUserId(request);
        try {
            userService.changePassword(userId, req.getOldPassword(), req.getNewPassword());
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (UserService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode()))
                    .body(ApiResponse.error(e.getCode()));
        }
    }

    // ==================== CRUD 端点（需 RBAC 权限） ====================

    @RequirePerm("user:create")
    @PostMapping
    public ResponseEntity<ApiResponse<User>> create(@Valid @RequestBody CreateUserRequest req) {
        try {
            User user = userService.createUser(req.getUsername(), req.getPassword(),
                    req.getEmail(), req.getRoleIds(), req.getDeptIds());
            return ResponseEntity.ok(ApiResponse.ok(user));
        } catch (UserService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode()))
                    .body(ApiResponse.error(e.getCode()));
        }
    }

    @RequirePerm("user:list")
    @GetMapping
    public ResponseEntity<ApiResponse<PageData<List<User>>>> list(HttpServletRequest request) {
        try {
            Query query = Query.parse(request, USER_DEFS);
            UserService.PageResult result = userService.listUsers(query);
            return ResponseEntity.ok(ApiResponse.page(result.items(), result.total(), result.page(), result.pageSize()));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        }
    }

    @RequirePerm("user:list")
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<User>> getById(@PathVariable Long id) {
        User user = userService.getById(id);
        if (user == null) {
            return ResponseEntity.status(404).body(ApiResponse.error(ErrorCode.ERR_NOT_FOUND));
        }
        return ResponseEntity.ok(ApiResponse.ok(user));
    }

    @RequirePerm("user:update")
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> update(HttpServletRequest request, @PathVariable Long id,
                                                    @Valid @RequestBody UpdateUserRequest req) {
        Long currentUserId = getUserId(request);
        if (id.equals(currentUserId)) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "不能修改自己"));
        }
        Map<String, Object> updates = new HashMap<>();
        if (req.getEmail() != null) updates.put("email", req.getEmail());
        if (req.getStatus() != null) updates.put("status", req.getStatus());
        if (req.getRoleIds() != null) {
            if (req.getRoleIds().isEmpty() && (req.getDeptIds() == null || req.getDeptIds().isEmpty())) {
                return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "至少需要分配一个角色或部门"));
            }
            updates.put("role_ids", req.getRoleIds());
        }
        if (req.getDeptIds() != null) updates.put("dept_ids", req.getDeptIds());
        if (updates.isEmpty()) return ResponseEntity.ok(ApiResponse.ok());
        userService.updateUser(id, updates);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("user:delete")
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(HttpServletRequest request, @PathVariable Long id) {
        Long currentUserId = getUserId(request);
        if (id.equals(currentUserId)) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "不能删除自己"));
        }
        userService.deleteUser(id);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("user:delete")
    @PostMapping("/batch-delete")
    public ResponseEntity<ApiResponse<Void>> batchDelete(@Valid @RequestBody BatchIdsRequest req) {
        userService.batchDelete(req.getIds());
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("user:update")
    @PostMapping("/batch-status")
    public ResponseEntity<ApiResponse<Void>> batchUpdateStatus(@Valid @RequestBody BatchStatusRequest req) {
        userService.batchUpdateStatus(req.getIds(), req.getStatus());
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("user:update")
    @DeleteMapping("/{id}/sessions")
    public ResponseEntity<ApiResponse<Void>> revokeSessions(@PathVariable Long id) {
        userService.revokeSessions(id);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    // ==================== Helpers ====================

    private Long getUserId(HttpServletRequest request) {
        Long userId = (Long) request.getAttribute("user_id");
        if (userId != null) return userId;
        Object claims = request.getAttribute("claims");
        if (claims instanceof io.jsonwebtoken.Claims c) {
            return c.get("user_id", Long.class);
        }
        return null;
    }

    private String getExtension(String filename) {
        if (filename == null) return "";
        int dot = filename.lastIndexOf('.');
        return dot >= 0 ? filename.substring(dot).toLowerCase() : "";
    }
}
