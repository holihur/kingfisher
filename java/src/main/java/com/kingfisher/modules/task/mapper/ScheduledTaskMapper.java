package com.kingfisher.modules.task.mapper;

import com.kingfisher.common.query.Condition;
import com.kingfisher.modules.task.domain.ScheduledTask;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;
import java.util.Map;

/**
 * 周期任务 Mapper，映射 SQLite scheduled_tasks 表。
 */
@Mapper
public interface ScheduledTaskMapper {

    List<ScheduledTask> findAll(@Param("keyword") String keyword,
                                @Param("searchableColumns") List<String> searchableColumns,
                                @Param("filters") List<Condition> filters,
                                @Param("sort") String sort,
                                @Param("offset") int offset,
                                @Param("limit") int limit);

    long countAll(@Param("keyword") String keyword,
                  @Param("searchableColumns") List<String> searchableColumns,
                  @Param("filters") List<Condition> filters);

    ScheduledTask findById(@Param("id") Long id);

    void insert(ScheduledTask task);

    void update(@Param("id") Long id, @Param("updates") Map<String, Object> updates);

    void deleteById(@Param("id") Long id);

    void deleteBatch(@Param("ids") List<Long> ids);

    void updateEnabledBatch(@Param("ids") List<Long> ids, @Param("enabled") int enabled);
}
