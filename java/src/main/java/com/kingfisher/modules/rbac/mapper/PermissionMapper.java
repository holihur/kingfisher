package com.kingfisher.modules.rbac.mapper;

import com.kingfisher.modules.rbac.domain.Permission;
import org.apache.ibatis.annotations.Mapper;
import java.util.List;

@Mapper
public interface PermissionMapper {
    List<Permission> findAll();
}
