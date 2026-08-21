package com.kingfisher.modules.task.service;

import com.kingfisher.common.query.Query;
import com.kingfisher.modules.task.domain.ScheduledTask;
import com.kingfisher.modules.task.mapper.ScheduledTaskMapper;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

@Service
public class ScheduledTaskService {

    private final ScheduledTaskMapper mapper;
    public ScheduledTaskService(ScheduledTaskMapper mapper) { this.mapper = mapper; }

    public List<ScheduledTask> list(Query q) {
        return mapper.findAll(q.getKeyword(), q.searchableColumns(), q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
    }
    public long count(Query q) { return mapper.countAll(q.getKeyword(), q.searchableColumns(), q.getFilters()); }
    public ScheduledTask getById(Long id) { return mapper.findById(id); }

    public ScheduledTask create(String name, String taskType, String cronSpec, String payload, int enabled, String remark) {
        ScheduledTask t = new ScheduledTask();
        t.setName(name); t.setTaskType(taskType); t.setCronSpec(cronSpec); t.setPayload(payload); t.setEnabled(enabled); t.setRemark(remark);
        mapper.insert(t);
        return t;
    }

    public void update(Long id, String name, String taskType, String cronSpec, String payload, int enabled, String remark) {
        Map<String,Object> updates = Map.of("name", name, "task_type", taskType, "cron_spec", cronSpec, "payload", payload != null ? payload : "", "enabled", enabled, "remark", remark != null ? remark : "");
        // MyBatis map with string keys, convert to expected
        java.util.HashMap<String,Object> m = new java.util.HashMap<>(updates);
        mapper.update(id, m);
    }

    public void delete(Long id) { mapper.deleteById(id); }
    public void batchDelete(List<Long> ids) { mapper.deleteBatch(ids); }
    public void batchUpdateStatus(List<Long> ids, int enabled) { mapper.updateEnabledBatch(ids, enabled); }
}
