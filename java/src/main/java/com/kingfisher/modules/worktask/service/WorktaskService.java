package com.kingfisher.modules.worktask.service;

import com.kingfisher.common.query.Query;
import com.kingfisher.modules.worktask.domain.WorkTask;
import com.kingfisher.modules.worktask.mapper.WorkTaskMapper;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

@Service
public class WorktaskService {

    private final WorkTaskMapper mapper;
    public WorktaskService(WorkTaskMapper mapper) { this.mapper = mapper; }

    public List<WorkTask> list(Query q, com.kingfisher.common.dataaccess.Scope scope) {
        return mapper.findAll(q.getKeyword(), q.searchableColumns(), q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize(), scope);
    }
    public long count(Query q, com.kingfisher.common.dataaccess.Scope scope) {
        return mapper.countAll(q.getKeyword(), q.searchableColumns(), q.getFilters(), scope);
    }
    public WorkTask getById(Long id, com.kingfisher.common.dataaccess.Scope scope) { return mapper.findById(id, scope); }

    public List<WorkTask> list(Query q, Long userId, boolean isAdmin) {
        return list(q, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId));
    }
    public long count(Query q, Long userId, boolean isAdmin) {
        return count(q, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId));
    }
    public WorkTask getById(Long id, Long userId, boolean isAdmin) { return getById(id, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId)); }

    public WorkTask create(String title, String description, Long ownerId, Long departmentId, String status) {
        WorkTask t = new WorkTask();
        t.setTitle(title); t.setDescription(description); t.setOwnerId(ownerId); t.setDepartmentId(departmentId); t.setStatus(status != null ? status : "pending");
        mapper.insert(t);
        return t;
    }

    public void update(Long id, Map<String,Object> updates) { mapper.update(id, updates); }
    public void delete(Long id, com.kingfisher.common.dataaccess.Scope scope) { mapper.deleteById(id, scope); }
    public void delete(Long id, Long userId, boolean isAdmin) { delete(id, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId)); }
}
