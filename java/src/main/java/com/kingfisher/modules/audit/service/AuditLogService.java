package com.kingfisher.modules.audit.service;

import com.kingfisher.common.PageData;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.audit.domain.AuditLog;
import com.kingfisher.modules.audit.mapper.AuditLogMapper;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 审计日志服务，与 Go extends/audit/app.AuditService 1:1 对齐。
 * 简化版：直接同步写入，Go 版本的 buffer/worker 批量写暂不复刻。
 */
@Service
public class AuditLogService {

    private final AuditLogMapper auditLogMapper;

    public AuditLogService(AuditLogMapper auditLogMapper) {
        this.auditLogMapper = auditLogMapper;
    }

    /**
     * 记录一条审计日志（同步写入）
     */
    public void log(AuditLog auditLog) {
        auditLogMapper.insert(auditLog);
    }

    /**
     * 批量记录审计日志
     */
    public void logBatch(List<AuditLog> logs) {
        if (logs == null || logs.isEmpty()) return;
        auditLogMapper.insertBatch(logs);
    }

    /**
     * 分页查询审计日志（支持 keyword + filter + sort）
     */
    public PageData<List<AuditLog>> list(Query query) {
        List<String> searchableColumns = query.searchableColumns();
        String sort = query.sortExpr();
        long total = auditLogMapper.countAll(query.getKeyword(), searchableColumns, query.getFilters());
        if (total == 0) {
            return new PageData<>(List.of(), 0, query.getPage(), query.getPageSize());
        }
        List<AuditLog> items = auditLogMapper.findAll(
                query.getKeyword(), searchableColumns, query.getFilters(),
                sort, query.offset(), query.getPageSize());
        return new PageData<>(items, total, query.getPage(), query.getPageSize());
    }
}
