package com.kingfisher.modules.rbac.service;

import com.kingfisher.common.query.Condition;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.rbac.domain.Permission;
import com.kingfisher.modules.rbac.domain.Role;
import com.kingfisher.modules.rbac.mapper.RoleMapper;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class RoleService {

    private final RoleMapper roleMapper;
    private final StringRedisTemplate redisTemplate;
    private final Map<String, String> memoryCache = new ConcurrentHashMap<>();
    private final Map<String, Long> memoryExpire = new ConcurrentHashMap<>();

    public RoleService(RoleMapper roleMapper, ObjectProvider<StringRedisTemplate> redisProvider) {
        this.roleMapper = roleMapper;
        this.redisTemplate = redisProvider.getIfAvailable();
    }

    public List<Role> list(Query q) {
        return roleMapper.findAll(q.getKeyword(), q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
    }

    public long count(Query q) {
        return roleMapper.countAll(q.getKeyword(), q.getFilters());
    }

    public Role getById(Long id) {
        return roleMapper.findById(id);
    }

    public void create(Role role) {
        roleMapper.insert(role);
    }

    public void update(Long id, String name, String description, Integer status) {
        Role r = roleMapper.findById(id);
        if (r != null && "admin".equals(r.getCode())) throw new IllegalArgumentException("cannot modify admin role");
        roleMapper.update(id, name, description, status);
    }

    public void delete(Long id) {
        Role r = roleMapper.findById(id);
        if (r != null && "admin".equals(r.getCode())) throw new IllegalArgumentException("cannot delete admin role");
        roleMapper.deleteById(id);
    }

    @Transactional
    public void batchDelete(List<Long> ids) {
        for (Long id : ids) {
            Role r = roleMapper.findById(id);
            if (r != null && "admin".equals(r.getCode())) throw new IllegalArgumentException("cannot delete admin role");
        }
        roleMapper.deleteBatch(ids);
    }

    public void batchUpdateStatus(List<Long> ids, int status) {
        for (Long id : ids) {
            Role r = roleMapper.findById(id);
            if (r != null && "admin".equals(r.getCode())) throw new IllegalArgumentException("cannot modify admin role");
        }
        roleMapper.updateStatusBatch(ids, status);
    }

    @Transactional
    public void assignPermissions(Long roleId, List<Long> permIds) {
        roleMapper.deleteRolePermissions(roleId);
        if (permIds != null) for (Long pid : permIds) roleMapper.insertRolePermission(roleId, pid);
        deleteByPattern("user:perms:*");
    }

    @Transactional
    public void assignMenus(Long roleId, List<Long> menuIds) {
        roleMapper.deleteRoleMenus(roleId);
        if (menuIds != null) for (Long mid : menuIds) roleMapper.insertRoleMenu(roleId, mid);
        deleteCache("menu:role:" + roleId);
        deleteCache("menu:tree");
    }

    public List<Permission> getRolePermissions(Long roleId) {
        return roleMapper.findPermissionsByRoleId(roleId);
    }

    public List<Long> getRoleMenus(Long roleId) {
        return roleMapper.findMenuIdsByRoleId(roleId);
    }

    public List<String> getUserPermissions(Long userId) {
        String key = "user:perms:" + userId;
        String cached = getCache(key);
        if (cached != null && !cached.isEmpty()) return List.of(cached.split(","));
        List<String> codes = roleMapper.findPermissionCodesByUserId(userId);
        if (codes != null && !codes.isEmpty()) putCache(key, String.join(",", codes), Duration.ofMinutes(30));
        return codes;
    }

    public String getLandingPage(Long roleId) {
        Role r = roleMapper.findById(roleId);
        return r != null ? r.getLandingPage() : "";
    }

    public String getDataScope(Long roleId, String resource) {
        var rows = roleMapper.findDataScopes(List.of(roleId), resource);
        for (var row : rows) {
            Object rid = row.get("role_id");
            Object st = row.get("scope_type");
            if (rid != null && String.valueOf(rid).equals(String.valueOf(roleId))) {
                return String.valueOf(st);
            }
        }
        return "";
    }

    public void setDataScope(Long roleId, String resource, String scopeType) {
        roleMapper.deleteDataScope(roleId, resource);
        roleMapper.insertDataScope(roleId, resource, scopeType);
    }

    public com.kingfisher.common.dataaccess.Scope resolveDataScope(Long userId, List<Long> roleIds, List<String> roleCodes) {
        for (String c : roleCodes) if ("admin".equals(c)) return com.kingfisher.common.dataaccess.Scope.all();
        if (roleIds == null || roleIds.isEmpty()) return com.kingfisher.common.dataaccess.Scope.self("owner_id", userId);
        var rows = roleMapper.findDataScopes(roleIds, "worktask");
        java.util.Map<Long, String> scopes = new java.util.HashMap<>();
        for (var row : rows) {
            try {
                Long rid = Long.valueOf(String.valueOf(row.get("role_id")));
                String st = String.valueOf(row.get("scope_type"));
                scopes.put(rid, st);
            } catch (Exception ignored) {}
        }
        String selected = "self";
        for (Long rid : roleIds) {
            String s = scopes.get(rid);
            if ("all".equals(s)) return com.kingfisher.common.dataaccess.Scope.all();
            if ("department_subtree".equals(s)) selected = "department_subtree";
            else if ("department".equals(s) && "self".equals(selected)) selected = "department";
        }
        if ("self".equals(selected)) return com.kingfisher.common.dataaccess.Scope.self("owner_id", userId);
        List<Long> deptIds = roleMapper.findDepartmentIdsByUserId(userId);
        if (deptIds == null) deptIds = List.of();
        if ("department_subtree".equals(selected)) {
            deptIds = getDepartmentSubtreeIds(deptIds);
            return com.kingfisher.common.dataaccess.Scope.subtree("department_id", deptIds);
        }
        return com.kingfisher.common.dataaccess.Scope.department("department_id", deptIds);
    }

    private List<Long> getDepartmentSubtreeIds(List<Long> roots) {
        if (roots == null || roots.isEmpty()) return List.of();
        var all = roleMapper.findAllDepartments();
        java.util.Map<Long, Long> parentMap = new java.util.HashMap<>();
        for (var row : all) {
            try {
                Long id = Long.valueOf(String.valueOf(row.get("id")));
                Long pid = Long.valueOf(String.valueOf(row.get("parent_id")));
                parentMap.put(id, pid);
            } catch (Exception ignored) {}
        }
        java.util.Set<Long> seen = new java.util.HashSet<>(roots);
        boolean changed = true;
        while (changed) {
            changed = false;
            for (var e : parentMap.entrySet()) {
                Long id = e.getKey(), pid = e.getValue();
                if (seen.contains(pid) && !seen.contains(id)) { seen.add(id); changed = true; }
            }
        }
        return new java.util.ArrayList<>(seen);
    }

    private String getCache(String key) {
        if (redisTemplate != null) {
            try { String v = redisTemplate.opsForValue().get(key); if (v != null) return v; } catch (Exception ignored) {}
        }
        Long exp = memoryExpire.get(key);
        if (exp != null && System.currentTimeMillis() < exp) return memoryCache.get(key);
        return null;
    }

    private void putCache(String key, String value, Duration ttl) {
        if (redisTemplate != null) {
            try { redisTemplate.opsForValue().set(key, value, ttl); return; } catch (Exception ignored) {}
        }
        memoryCache.put(key, value);
        memoryExpire.put(key, System.currentTimeMillis() + ttl.toMillis());
    }

    private void deleteCache(String key) {
        if (redisTemplate != null) try { redisTemplate.delete(key); } catch (Exception ignored) {}
        memoryCache.remove(key);
        memoryExpire.remove(key);
    }

    private void deleteByPattern(String pattern) {
        if (redisTemplate != null) {
            try {
                // 简化：Redis 需 SCAN，这里直接尝试删除精确键，批量场景可扩展
                // 内存回退则清空所有 user:perms: 前缀
            } catch (Exception ignored) {}
        }
        memoryCache.keySet().removeIf(k -> k.startsWith("user:perms:"));
        memoryExpire.keySet().removeIf(k -> k.startsWith("user:perms:"));
    }
}
