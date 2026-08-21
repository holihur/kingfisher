package com.kingfisher.modules.audit.mapper;

import com.kingfisher.common.query.Condition;
import com.kingfisher.modules.audit.domain.AuditLog;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

/**
 * 审计日志 Mapper，基于 MyBatis，映射 SQLite audit_logs。
 */
@Mapper
public interface AuditLogMapper {

    /** 分页查询审计日志（支持 keyword 模糊 + filter + sort） */
    List<AuditLog> findAll(@Param("keyword") String keyword,
                           @Param("searchableColumns") List<String> searchableColumns,
                           @Param("filters") List<Condition> filters,
                           @Param("sort") String sort,
                           @Param("offset") int offset,
                           @Param("limit") int limit);

    /** 统计总数（同上条件） */
    long countAll(@Param("keyword") String keyword,
                  @Param("searchableColumns") List<String> searchableColumns,
                  @Param("filters") List<Condition> filters);

    /** 插入单条审计日志 */
    void insert(AuditLog auditLog);

    /** 批量插入审计日志 */
    void insertBatch(@Param("list") List<AuditLog> list);
}
