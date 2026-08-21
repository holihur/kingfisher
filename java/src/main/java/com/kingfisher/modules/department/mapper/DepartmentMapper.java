package com.kingfisher.modules.department.mapper;

import com.kingfisher.modules.department.domain.Department;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import java.util.List;

@Mapper
public interface DepartmentMapper {
    List<Department> findAll();
    Department findById(@Param("id") Long id);
    List<Department> findByParentId(@Param("parentId") Long parentId);
    void insert(Department dept);
    void update(Department dept);
    void deleteById(@Param("id") Long id);
    long countChildren(@Param("parentId") Long parentId);
    List<Long> findRoleIdsByDeptId(@Param("deptId") Long deptId);
    void deleteDeptRoles(@Param("deptId") Long deptId);
    void insertDeptRole(@Param("deptId") Long deptId, @Param("roleId") Long roleId);
}
