package com.kingfisher.modules.rbac.mapper;

import com.kingfisher.modules.rbac.domain.Permission;
import com.kingfisher.modules.rbac.domain.Role;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import java.util.List;
import java.util.Map;

@Mapper
public interface RoleMapper {
    List<Role> findAll(@Param("keyword") String keyword,
                       @Param("filters") List<com.kingfisher.common.query.Condition> filters,
                       @Param("sort") String sort,
                       @Param("offset") int offset,
                       @Param("limit") int limit);
    long countAll(@Param("keyword") String keyword, @Param("filters") List<com.kingfisher.common.query.Condition> filters);
    Role findById(@Param("id") Long id);
    Role findByCode(@Param("code") String code);
    void insert(Role role);
    void update(@Param("id") Long id, @Param("name") String name, @Param("description") String desc, @Param("status") Integer status);
    void deleteById(@Param("id") Long id);
    void deleteBatch(@Param("ids") List<Long> ids);
    void updateStatusBatch(@Param("ids") List<Long> ids, @Param("status") int status);
    void deleteRolePermissions(@Param("roleId") Long roleId);
    void insertRolePermission(@Param("roleId") Long roleId, @Param("permId") Long permId);
    List<Permission> findPermissionsByRoleId(@Param("roleId") Long roleId);
    List<Long> findMenuIdsByRoleId(@Param("roleId") Long roleId);
    void deleteRoleMenus(@Param("roleId") Long roleId);
    void insertRoleMenu(@Param("roleId") Long roleId, @Param("menuId") Long menuId);
    List<String> findPermissionCodesByUserId(@Param("userId") Long userId);

    List<Map<String, Object>> findDataScopes(@Param("roleIds") List<Long> roleIds, @Param("resource") String resource);
    void deleteDataScope(@Param("roleId") Long roleId, @Param("resource") String resource);
    void insertDataScope(@Param("roleId") Long roleId, @Param("resource") String resource, @Param("scopeType") String scopeType);
    List<Long> findDepartmentIdsByUserId(@Param("userId") Long userId);
    List<Map<String, Long>> findAllDepartments();
}
