package com.kingfisher.modules.rbac.service;

import com.kingfisher.modules.rbac.domain.Permission;
import com.kingfisher.modules.rbac.mapper.PermissionMapper;
import org.springframework.stereotype.Service;
import java.util.List;

@Service
public class PermissionService {
    private final PermissionMapper mapper;
    public PermissionService(PermissionMapper mapper) { this.mapper = mapper; }
    public List<Permission> list() { return mapper.findAll(); }
}
