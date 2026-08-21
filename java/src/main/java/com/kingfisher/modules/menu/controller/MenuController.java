package com.kingfisher.modules.menu.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.menu.domain.Menu;
import com.kingfisher.modules.menu.dto.BatchIDsRequest;
import com.kingfisher.modules.menu.dto.BatchStatusRequest;
import com.kingfisher.modules.menu.dto.CreateMenuRequest;
import com.kingfisher.modules.menu.dto.UpdateMenuRequest;
import com.kingfisher.modules.menu.service.MenuService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.*;

import java.util.*;

/**
 * 菜单控制器，与 Go extends/menu/transport 路由 1:1 对齐。
 * 所有接口均在受保护路由下（JwtAuthFilter 已校验 token）。
 */
@RestController
@RequestMapping("/api/v1/menus")
public class MenuController {

    private final MenuService menuService;

    public MenuController(MenuService menuService) {
        this.menuService = menuService;
    }

    /**
     * GET /api/v1/menus/tree — 完整菜单树（需 menu:list 权限）
     */
    @RequirePerm("menu:list")
    @GetMapping("/tree")
    public ApiResponse<List<Menu>> getTree() {
        return ApiResponse.ok(menuService.getTree());
    }

    /**
     * GET /api/v1/menus/my — 当前用户角色过滤的菜单树（无需额外权限）
     */
    @GetMapping("/my")
    public ApiResponse<List<Menu>> getMyTree(HttpServletRequest request) {
        List<Long> roleIds = extractRoleIds(request);
        return ApiResponse.ok(menuService.getTreeForRole(roleIds));
    }

    /**
     * GET /api/v1/menus/{id} — 按 ID 查询菜单（需 menu:list 权限）
     */
    @RequirePerm("menu:list")
    @GetMapping("/{id}")
    public ApiResponse<Menu> getById(@PathVariable Long id) {
        Menu menu = menuService.getById(id);
        if (menu == null) {
            return ApiResponse.error(ErrorCode.ERR_NOT_FOUND);
        }
        return ApiResponse.ok(menu);
    }

    /**
     * POST /api/v1/menus — 创建菜单（需 menu:create 权限）
     */
    @RequirePerm("menu:create")
    @PostMapping
    public ApiResponse<Menu> create(@RequestBody CreateMenuRequest req) {
        Menu menu = new Menu();
        menu.setParentId(req.getParentId());
        menu.setName(req.getName());
        menu.setPath(req.getPath());
        menu.setComponent(req.getComponent());
        menu.setIcon(req.getIcon());
        menu.setSort(req.getSort());
        menu.setType(req.getType());
        menu.setPermission(req.getPermission());
        menu.setStatus(req.getStatus());
        menu.setVersion(req.getVersion());
        try {
            menuService.create(menu);
            return ApiResponse.ok(menu);
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage());
        }
    }

    /**
     * PUT /api/v1/menus/{id} — 更新菜单（需 menu:update 权限）
     * 白名单字段，防止 mass assignment，与 Go updateMenuReq 对齐。
     */
    @RequirePerm("menu:update")
    @PutMapping("/{id}")
    public ApiResponse<Void> update(@PathVariable Long id, @RequestBody UpdateMenuRequest req) {
        Map<String, Object> updates = new HashMap<>();
        if (req.getName() != null) updates.put("name", req.getName());
        if (req.getIcon() != null) updates.put("icon", req.getIcon());
        if (req.getPath() != null) updates.put("path", req.getPath());
        if (req.getComponent() != null) updates.put("component", req.getComponent());
        if (req.getSort() != null) updates.put("sort", req.getSort());
        if (req.getParentId() != null) updates.put("parent_id", req.getParentId());
        if (req.getStatus() != null) updates.put("status", req.getStatus());
        if (req.getVersion() != null) updates.put("version", req.getVersion());
        if (updates.isEmpty()) {
            return ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "无更新字段");
        }
        menuService.update(id, updates);
        return ApiResponse.ok();
    }

    /**
     * DELETE /api/v1/menus/{id} — 删除菜单（需 menu:delete 权限，含子节点检查）
     */
    @RequirePerm("menu:delete")
    @DeleteMapping("/{id}")
    public ApiResponse<Void> delete(@PathVariable Long id) {
        try {
            menuService.delete(id);
            return ApiResponse.ok();
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage());
        }
    }

    /**
     * POST /api/v1/menus/batch-delete — 批量删除（需 menu:delete 权限）
     */
    @RequirePerm("menu:delete")
    @PostMapping("/batch-delete")
    public ApiResponse<Void> batchDelete(@Valid @RequestBody BatchIDsRequest req) {
        try {
            menuService.batchDelete(req.getIds());
            return ApiResponse.ok();
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage());
        }
    }

    /**
     * POST /api/v1/menus/batch-status — 批量更新状态（需 menu:update 权限）
     */
    @RequirePerm("menu:update")
    @PostMapping("/batch-status")
    public ApiResponse<Void> batchUpdateStatus(@Valid @RequestBody BatchStatusRequest req) {
        menuService.batchUpdateStatus(req.getIds(), req.getStatus());
        return ApiResponse.ok();
    }

    // ========== 辅助方法 ==========

    /**
     * 从请求属性中提取当前用户的角色 ID 列表（由 JwtAuthFilter 注入）。
     * role_ids 在 JWT claims 中为 List<Integer>，需转为 List<Long>。
     */
    @SuppressWarnings("unchecked")
    private List<Long> extractRoleIds(HttpServletRequest request) {
        Object attr = request.getAttribute("role_ids");
        if (attr instanceof List<?> list) {
            List<Long> result = new ArrayList<>();
            for (Object item : list) {
                if (item instanceof Number n) {
                    result.add(n.longValue());
                }
            }
            return result;
        }
        return List.of();
    }
}
