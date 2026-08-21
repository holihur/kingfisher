package com.kingfisher.modules.department.service;

import com.kingfisher.modules.department.domain.Department;
import com.kingfisher.modules.department.mapper.DepartmentMapper;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.util.*;

@Service
public class DepartmentService {
    private final DepartmentMapper mapper;
    public DepartmentService(DepartmentMapper mapper) { this.mapper = mapper; }

    public List<Department> list() { return mapper.findAll(); }

    public List<Department> tree() {
        List<Department> all = mapper.findAll();
        Map<Long, Department> map = new HashMap<>();
        for (Department d : all) map.put(d.getId(), d);
        List<Department> roots = new ArrayList<>();
        for (Department d : all) {
            Long pid = d.getParentId() == null ? 0L : d.getParentId();
            if (pid == 0) roots.add(d);
            else {
                Department p = map.get(pid);
                if (p != null) {
                    if (p.getChildren() == null) p.setChildren(new ArrayList<>());
                    p.getChildren().add(d);
                } else roots.add(d);
            }
        }
        return roots;
    }

    public Department getById(Long id) { return mapper.findById(id); }

    public Department create(Department d) {
        if (d.getParentId() != null && d.getParentId() != 0 && mapper.findById(d.getParentId()) == null) throw new IllegalArgumentException("父部门不存在");
        mapper.insert(d);
        return d;
    }

    public void update(Long id, Department d) {
        d.setId(id);
        // 循环引用检查
        if (d.getParentId() != null && d.getParentId().equals(id)) throw new IllegalArgumentException("不能将自己设为父部门");
        mapper.update(d);
    }

    public void delete(Long id) {
        if (mapper.countChildren(id) > 0) throw new IllegalArgumentException("存在子部门不可删除");
        mapper.deleteById(id);
    }

    public List<Long> getRoleIds(Long deptId) { return mapper.findRoleIdsByDeptId(deptId); }

    @Transactional
    public void assignRoles(Long deptId, List<Long> roleIds) {
        mapper.deleteDeptRoles(deptId);
        if (roleIds != null) for (Long rid : roleIds) mapper.insertDeptRole(deptId, rid);
    }
}
