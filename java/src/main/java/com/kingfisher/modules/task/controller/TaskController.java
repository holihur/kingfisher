package com.kingfisher.modules.task.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.PageData;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.task.domain.ScheduledTask;
import com.kingfisher.modules.task.service.ScheduledTaskService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/scheduled-tasks")
public class TaskController {

    private final ScheduledTaskService taskService;
    public TaskController(ScheduledTaskService taskService) { this.taskService = taskService; }

    private static final Defs TASK_DEFS = Defs.of(
            "name", Field.searchableString("name"),
            "task_type", Field.filterable(FieldType.STRING, "task_type"),
            "cron_spec", Field.searchableString("cron_spec"),
            "enabled", Field.filterable(FieldType.INT, "enabled"),
            "created_at", Field.filterable(FieldType.TIME, "created_at")
    );

    @RequirePerm("task:list")
    @GetMapping
    public ResponseEntity<ApiResponse<PageData<List<ScheduledTask>>>> list(HttpServletRequest request) {
        Query q = Query.parse(request, TASK_DEFS);
        List<ScheduledTask> items = taskService.list(q);
        long total = taskService.count(q);
        return ResponseEntity.ok(ApiResponse.ok(new PageData<>(items, total, q.getPage(), q.getPageSize())));
    }

    @RequirePerm("task:create")
    @PostMapping
    public ResponseEntity<ApiResponse<ScheduledTask>> create(@RequestBody Map<String,Object> body) {
        String name = (String) body.get("name");
        String taskType = (String) body.get("task_type");
        if (taskType == null) taskType = (String) body.get("taskType");
        String cronSpec = (String) body.get("cron_spec");
        if (cronSpec == null) cronSpec = (String) body.get("cronSpec");
        String payload = body.get("payload") != null ? String.valueOf(body.get("payload")) : "";
        int enabled = body.get("enabled") != null ? ((Number) body.get("enabled")).intValue() : 1;
        String remark = (String) body.getOrDefault("remark", "");
        ScheduledTask t = taskService.create(name, taskType, cronSpec, payload, enabled, remark);
        return ResponseEntity.ok(ApiResponse.ok(t));
    }

    @RequirePerm("task:list")
    @GetMapping("/types")
    public ResponseEntity<ApiResponse<List<Map<String, String>>>> types() {
        List<Map<String, String>> types = List.of(
                Map.of("type", "email:send", "label", "邮件发送", "payload_example", "{\"to\":\"user@example.com\",\"subject\":\"标题\",\"body\":\"正文\"}"),
                Map.of("type", "nop:run", "label", "空操作", "payload_example", "{\"note\":\"test\"}")
        );
        return ResponseEntity.ok(ApiResponse.ok(types));
    }

    @RequirePerm("task:list")
    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<ScheduledTask>> getById(@PathVariable Long id) {
        ScheduledTask t = taskService.getById(id);
        if (t == null) return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_TASK_NOT_FOUND)).body(ApiResponse.error(ErrorCode.ERR_TASK_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(t));
    }

    @RequirePerm("task:update")
    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> update(@PathVariable Long id, @RequestBody Map<String,Object> body) {
        String name = (String) body.get("name");
        String taskType = (String) body.get("task_type");
        if (taskType == null) taskType = (String) body.get("taskType");
        String cronSpec = (String) body.get("cron_spec");
        if (cronSpec == null) cronSpec = (String) body.get("cronSpec");
        String payload = body.get("payload") != null ? String.valueOf(body.get("payload")) : "";
        int enabled = body.get("enabled") != null ? ((Number) body.get("enabled")).intValue() : 1;
        String remark = (String) body.getOrDefault("remark", "");
        taskService.update(id, name, taskType, cronSpec, payload, enabled, remark);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("task:delete")
    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> delete(@PathVariable Long id) {
        taskService.delete(id);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("task:delete")
    @PostMapping("/batch-delete")
    public ResponseEntity<ApiResponse<Void>> batchDelete(@RequestBody Map<String,List<Long>> body) {
        taskService.batchDelete(body.get("ids"));
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("task:update")
    @PostMapping("/batch-status")
    public ResponseEntity<ApiResponse<Void>> batchStatus(@RequestBody Map<String,Object> body) {
        List<Long> ids = (List<Long>) body.get("ids");
        List<Long> longIds = ids.stream().map(v -> v instanceof Number ? ((Number)v).longValue() : Long.parseLong(String.valueOf(v))).toList();
        int enabled = ((Number) body.get("enabled")).intValue();
        if (body.get("status") != null) enabled = ((Number) body.get("status")).intValue();
        taskService.batchUpdateStatus(longIds, enabled);
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("task:update")
    @PostMapping("/{id}/trigger")
    public ResponseEntity<ApiResponse<Void>> trigger(@PathVariable Long id) {
        ScheduledTask t = taskService.getById(id);
        if (t == null) return ResponseEntity.status(ErrorCode.httpStatus(ErrorCode.ERR_TASK_NOT_FOUND)).body(ApiResponse.error(ErrorCode.ERR_TASK_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("task:update")
    @PostMapping("/{id}/run")
    public ResponseEntity<ApiResponse<Void>> run(@PathVariable Long id) {
        return trigger(id);
    }
}
