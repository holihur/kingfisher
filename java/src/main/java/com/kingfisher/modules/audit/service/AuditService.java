package com.kingfisher.modules.audit.service;

import com.kingfisher.common.PageData;
import com.kingfisher.common.query.*;
import com.kingfisher.modules.audit.domain.AuditLog;
import com.kingfisher.modules.audit.mapper.AuditLogMapper;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 审计日志服务，与 Go extends/audit/app.AuditService 对齐。
 * 简化版：直接同步写入，不做异步 buffer。
 */
@Service
public class AuditService {

    private static final Defs AUDIT_DEFS = Defs.of(
        "username", Field.searchableString("username"),
        "ip", Field.searchableString("ip"),
        "user_agent", Field.searchableString("user_agent"),
        "user_id", Field.filterable(FieldType.UINT, "user_id"),
        "resource", Field.filterable(FieldType.STRING, "resource"),
        "action", Field.filterable(FieldType.STRING, "action"),
        "resource_id", Field.filterable(FieldType.UINT, "resource_id"),
        "result", Field.filterable(FieldType.STRING, "result"),
        "created_at", Field.filterable(FieldType.TIME, "created_at")
    );

    private final AuditLogMapper auditLogMapper;

    public AuditService(AuditLogMapper auditLogMapper) {
        this.auditLogMapper = auditLogMapper;
    }

    /** 分页查询审计日志 */
    public PageData<List<AuditLog>> list(HttpServletRequest request) {
        Query q = Query.parse(request, AUDIT_DEFS);
        List<String> searchable = q.searchableColumns();
        List<AuditLog> items = auditLogMapper.findAll(
            q.getKeyword(), searchable, q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
        long total = auditLogMapper.countAll(q.getKeyword(), searchable, q.getFilters());
        return new PageData<>(items, total, q.getPage(), q.getPageSize());
    }

    public PageData<List<AuditLog>> listByUserId(Long userId, HttpServletRequest request) {
        Query q = Query.parse(request, AUDIT_DEFS);
        q.getFilters().add(new Condition("user_id", Operator.EQ, userId));
        List<String> searchable = q.searchableColumns();
        List<AuditLog> items = auditLogMapper.findAll(
            q.getKeyword(), searchable, q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
        long total = auditLogMapper.countAll(q.getKeyword(), searchable, q.getFilters());
        return new PageData<>(items, total, q.getPage(), q.getPageSize());
    }

    public PageData<List<AuditLog>> listLoginLogs(Long userId, HttpServletRequest request) {
        Query q = Query.parse(request, AUDIT_DEFS);
        q.getFilters().add(new Condition("user_id", Operator.EQ, userId));
        q.getFilters().add(new Condition("action", Operator.EQ, "login"));
        q.getFilters().add(new Condition("resource", Operator.EQ, "auth"));
        List<String> searchable = q.searchableColumns();
        List<AuditLog> items = auditLogMapper.findAll(
            q.getKeyword(), searchable, q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
        long total = auditLogMapper.countAll(q.getKeyword(), searchable, q.getFilters());
        return new PageData<>(items, total, q.getPage(), q.getPageSize());
    }

    /** 记录一条审计日志（简化版，同步写入） */
    public void log(AuditLog log) {
        auditLogMapper.insert(log);
    }

    /** 批量记录审计日志 */
    public void logBatch(List<AuditLog> logs) {
        if (logs != null && !logs.isEmpty()) {
            auditLogMapper.insertBatch(logs);
        }
    }
}
