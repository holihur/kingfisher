package com.kingfisher.modules.menu.mapper;

import com.kingfisher.modules.menu.domain.Menu;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;
import java.util.Map;

/**
 * 菜单 Mapper，基于 MyBatis，映射 SQLite menus / role_menus。
 * 与 Go extends/menu/adapter/mysql/repo.go 方法 1:1 对齐。
 */
@Mapper
public interface MenuMapper {

    /** 查询全部菜单（按 sort ASC），用于建树 */
    List<Menu> findAll();

    /** 按 ID 查询单个菜单 */
    Menu findById(@Param("id") Long id);

    /** 按 parent_id 查询子菜单 */
    List<Menu> findByParentId(@Param("parentId") Long parentId);

    /** 按角色 ID 列表查询菜单（JOIN role_menus），用于当前用户菜单过滤 */
    List<Menu> findByRoleIds(@Param("roleIds") List<Long> roleIds);

    /** 新增菜单 */
    void insert(Menu menu);

    /** 部分更新菜单（动态 SQL） */
    void update(@Param("id") Long id, @Param("updates") Map<String, Object> updates);

    /** 删除单个菜单 */
    void deleteById(@Param("id") Long id);

    /** 批量删除菜单 */
    void deleteBatch(@Param("ids") List<Long> ids);

    /** 批量更新菜单状态 */
    void updateStatusBatch(@Param("ids") List<Long> ids, @Param("status") int status);

    /** 检查指定菜单是否有子节点 */
    int countChildren(@Param("parentId") Long parentId);
}
