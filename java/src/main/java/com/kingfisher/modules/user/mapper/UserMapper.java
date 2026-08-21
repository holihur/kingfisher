package com.kingfisher.modules.user.mapper;

import com.kingfisher.modules.user.domain.Department;
import com.kingfisher.modules.user.domain.Role;
import com.kingfisher.modules.user.domain.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;
import java.util.Map;

/**
 * 用户 Mapper，基于 MyBatis，映射 SQLite users/user_roles/roles/user_departments/departments。
 */
@Mapper
public interface UserMapper {

    // ==================== 查询 ====================

    /** 按用户名查询用户（含 roles 关联） */
    User findByUsername(@Param("username") String username);

    /** 按 ID 查询用户（含 roles） */
    User findById(@Param("id") Long id);

    /** 按邮箱查询用户 */
    User findByEmail(@Param("email") String email);

    /** 查询角色落地页 */
    Role findRoleById(@Param("id") Long id);

    /**
     * 分页查询用户列表（含 roles），支持关键词搜索 + 字段过滤。
     * params 包含: keyword, status, roleId, departmentId, createdAtStart, createdAtEnd, sort, offset, limit
     */
    List<User> findAll(@Param("params") Map<String, Object> params);

    /** 统计满足条件的用户总数 */
    long countAll(@Param("params") Map<String, Object> params);

    /** 查询用户直接分配的角色 ID（user_roles） */
    List<Long> findDirectRoleIds(@Param("userId") Long userId);

    /** 查询用户所属部门 ID（user_departments） */
    List<Long> findDeptIds(@Param("userId") Long userId);

    /** 查询用户所属部门详情（user_departments join departments） */
    List<Department> findDepartments(@Param("userId") Long userId);

    /** 查询用户通过部门继承的角色 ID（department_roles join user_departments） */
    List<Long> findDeptInheritedRoleIds(@Param("userId") Long userId);

    /** 批量按 ID 查询用户（含 roles），用于列表回填 */
    List<User> findByIds(@Param("ids") List<Long> ids);

    // ==================== 写入 ====================

    /** 插入用户，返回自增 ID */
    int insert(@Param("user") User user);

    /** 更新用户基础字段（白名单：nickname, email, avatar, status） */
    int update(@Param("id") Long id, @Param("fields") Map<String, Object> fields);

    /** 软删除用户（置 deleted_at） */
    int delete(@Param("id") Long id);

    /** 批量软删除 */
    int deleteBatch(@Param("ids") List<Long> ids);

    /** 批量更新状态 */
    int updateStatusBatch(@Param("ids") List<Long> ids, @Param("status") Integer status);

    /** 递增 session_version（吊销会话） */
    int incrementSessionVersion(@Param("id") Long id);

    int updatePassword(@Param("id") Long id, @Param("password") String password);

    // ==================== 关联表操作 ====================

    /** 删除用户-角色关联 */
    int deleteUserRoles(@Param("userId") Long userId);

    /** 插入用户-角色关联 */
    int insertUserRole(@Param("userId") Long userId, @Param("roleId") Long roleId);

    /** 删除用户-部门关联 */
    int deleteUserDepartments(@Param("userId") Long userId);

    /** 插入用户-部门关联 */
    int insertUserDepartment(@Param("userId") Long userId, @Param("departmentId") Long departmentId);
}
