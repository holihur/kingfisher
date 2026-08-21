package com.kingfisher.modules.user.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

/**
 * 用户领域实体，对应 users 表 + 关联 roles/departments。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class User {

    private Long id;
    private String username;
    private String nickname;
    /** bcrypt 哈希，不序列化 */
    @JsonIgnore
    private String password;
    private String email;
    private String avatar;
    private Integer status;
    @JsonIgnore
    private Integer sessionVersion;
    @JsonProperty("created_at")
    private LocalDateTime createdAt;
    @JsonProperty("updated_at")
    private LocalDateTime updatedAt;

    /** 关联角色（通过 user_roles 聚合，含部门继承角色） */
    private List<Role> roles;

    /** 直接分配的角色 ID（user_roles），供编辑表单回填 */
    @JsonProperty("direct_role_ids")
    private List<Long> directRoleIds;

    /** 用户所属部门 ID 列表（user_departments） */
    @JsonProperty("dept_ids")
    private List<Long> deptIds;

    /** 用户所属部门轻量视图 */
    private List<Department> departments;

    /** 有效角色 ID = 直接分配 ∪ 部门继承（去重） */
    @JsonProperty("role_ids")
    public List<Long> getRoleIds() {
        if (roles == null) return List.of();
        return roles.stream().map(Role::getId).distinct().toList();
    }

    @JsonIgnore
    public List<String> getRoleCodes() {
        if (roles == null) return List.of();
        return roles.stream().map(Role::getCode).distinct().toList();
    }
}
