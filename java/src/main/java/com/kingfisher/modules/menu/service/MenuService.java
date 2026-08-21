package com.kingfisher.modules.menu.service;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.modules.menu.domain.Menu;
import com.kingfisher.modules.menu.mapper.MenuMapper;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 菜单服务，与 Go extends/menu/app.MenuService 1:1 对齐。
 * 缓存策略：优先 Redis，不可用时回退内存（与 Go core/cache 行为一致）。
 */
@Service
public class MenuService {

    private static final String CACHE_KEY_TREE = "menu:tree";
    private static final Duration CACHE_TTL = Duration.ofMinutes(10);

    private final MenuMapper menuMapper;
    private final StringRedisTemplate redisTemplate;
    private final ObjectMapper objectMapper;

    // 内存缓存回退
    private String memoryTreeCache;
    private long memoryTreeCacheExpireAt;

    public MenuService(MenuMapper menuMapper,
                       ObjectProvider<StringRedisTemplate> redisProvider,
                       ObjectMapper objectMapper) {
        this.menuMapper = menuMapper;
        this.redisTemplate = redisProvider.getIfAvailable();
        this.objectMapper = objectMapper;
    }

    /**
     * 获取完整菜单树（带缓存）
     */
    public List<Menu> getTree() {
        // 尝试从缓存读取
        String cached = getFromCache(CACHE_KEY_TREE);
        if (cached != null) {
            try {
                return objectMapper.readValue(cached, new TypeReference<>() {});
            } catch (Exception ignored) {
            }
        }
        List<Menu> menus = menuMapper.findAll();
        List<Menu> tree = buildTree(menus, 0L);
        // 写入缓存
        try {
            String json = objectMapper.writeValueAsString(tree);
            putToCache(CACHE_KEY_TREE, json, CACHE_TTL);
        } catch (Exception ignored) {
        }
        return tree;
    }

    /**
     * 按当前用户角色获取菜单树（多角色取并集，不缓存）
     */
    public List<Menu> getTreeForRole(List<Long> roleIds) {
        if (roleIds == null || roleIds.isEmpty()) {
            return List.of();
        }
        List<Menu> menus = menuMapper.findByRoleIds(roleIds);
        return buildTree(menus, 0L);
    }

    /**
     * 按 ID 查询单个菜单
     */
    public Menu getById(Long id) {
        return menuMapper.findById(id);
    }

    /**
     * 创建菜单（校验 path 唯一性）
     */
    public Menu create(Menu menu) {
        if (menu.getPath() != null && !menu.getPath().isEmpty()) {
            List<Menu> all = menuMapper.findAll();
            for (Menu existing : all) {
                if (menu.getPath().equals(existing.getPath())) {
                    throw new IllegalArgumentException("menu path already exists");
                }
            }
        }
        if (menu.getParentId() == null) menu.setParentId(0L);
        if (menu.getSort() == null) menu.setSort(0);
        if (menu.getType() == null) menu.setType(2);
        if (menu.getStatus() == null) menu.setStatus(1);
        menuMapper.insert(menu);
        invalidateTreeCache();
        return menu;
    }

    /**
     * 部分更新菜单
     */
    public void update(Long id, Map<String, Object> updates) {
        menuMapper.update(id, updates);
        invalidateTreeCache();
    }

    /**
     * 删除单个菜单（含子节点检查）
     */
    public void delete(Long id) {
        if (menuMapper.countChildren(id) > 0) {
            throw new IllegalArgumentException("menu has children");
        }
        menuMapper.deleteById(id);
        invalidateTreeCache();
    }

    /**
     * 批量删除（任一菜单含子节点则整批拒绝）
     */
    public void batchDelete(List<Long> ids) {
        for (Long id : ids) {
            if (menuMapper.countChildren(id) > 0) {
                throw new IllegalArgumentException("menu has children");
            }
        }
        menuMapper.deleteBatch(ids);
        invalidateTreeCache();
    }

    /**
     * 批量更新状态
     */
    public void batchUpdateStatus(List<Long> ids, int status) {
        menuMapper.updateStatusBatch(ids, status);
        invalidateTreeCache();
    }

    // ========== 内部方法 ==========

    /**
     * 递归建树，与 Go buildTree 逻辑一致
     */
    private List<Menu> buildTree(List<Menu> menus, Long parentId) {
        List<Menu> result = new ArrayList<>();
        for (Menu m : menus) {
            if (parentId.equals(m.getParentId())) {
                m.setChildren(buildTree(menus, m.getId()));
                result.add(m);
            }
        }
        return result;
    }

    private void invalidateTreeCache() {
        deleteFromCache(CACHE_KEY_TREE);
    }

    // ========== 缓存操作（Redis 优先，内存回退） ==========

    private String getFromCache(String key) {
        if (redisTemplate != null) {
            try {
                return redisTemplate.opsForValue().get(key);
            } catch (Exception ignored) {
            }
        }
        if (memoryTreeCache != null && System.currentTimeMillis() < memoryTreeCacheExpireAt) {
            return memoryTreeCache;
        }
        return null;
    }

    private void putToCache(String key, String value, Duration ttl) {
        if (redisTemplate != null) {
            try {
                redisTemplate.opsForValue().set(key, value, ttl);
                return;
            } catch (Exception ignored) {
            }
        }
        memoryTreeCache = value;
        memoryTreeCacheExpireAt = System.currentTimeMillis() + ttl.toMillis();
    }

    private void deleteFromCache(String key) {
        if (redisTemplate != null) {
            try {
                redisTemplate.delete(key);
            } catch (Exception ignored) {
            }
        }
        memoryTreeCache = null;
        memoryTreeCacheExpireAt = 0;
    }
}
