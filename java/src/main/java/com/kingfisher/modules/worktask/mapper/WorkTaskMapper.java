package com.kingfisher.modules.worktask.mapper;

import com.kingfisher.common.query.Condition;
import com.kingfisher.modules.worktask.domain.WorkTask;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;
import java.util.Map;

/**
 * 工作任务 Mapper，映射 SQLite tasks 表。
 */
@Mapper
public interface WorkTaskMapper {

    List<WorkTask> findAll(@Param("keyword") String keyword,
                           @Param("searchableColumns") List<String> searchableColumns,
                           @Param("filters") List<Condition> filters,
                           @Param("sort") String sort,
                           @Param("offset") int offset,
                           @Param("limit") int limit,
                           @Param("scope") com.kingfisher.common.dataaccess.Scope scope);

    long countAll(@Param("keyword") String keyword,
                  @Param("searchableColumns") List<String> searchableColumns,
                  @Param("filters") List<Condition> filters,
                  @Param("scope") com.kingfisher.common.dataaccess.Scope scope);

    WorkTask findById(@Param("id") Long id, @Param("scope") com.kingfisher.common.dataaccess.Scope scope);

    void insert(WorkTask task);

    void update(@Param("id") Long id, @Param("updates") Map<String, Object> updates);

    void deleteById(@Param("id") Long id, @Param("scope") com.kingfisher.common.dataaccess.Scope scope);

    // 兼容旧 isAdmin 接口（委托给 Scope）
    default List<WorkTask> findAll(String k, List<String> sc, List<Condition> f, String s, int o, int l, Long userId, boolean isAdmin) {
        return findAll(k, sc, f, s, o, l, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId));
    }
    default long countAll(String k, List<String> sc, List<Condition> f, Long userId, boolean isAdmin) {
        return countAll(k, sc, f, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId));
    }
    default WorkTask findById(Long id, Long userId, boolean isAdmin) {
        return findById(id, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId));
    }
    default void deleteById(Long id, Long userId, boolean isAdmin) {
        deleteById(id, isAdmin ? com.kingfisher.common.dataaccess.Scope.all() : com.kingfisher.common.dataaccess.Scope.self("owner_id", userId));
    }
}
