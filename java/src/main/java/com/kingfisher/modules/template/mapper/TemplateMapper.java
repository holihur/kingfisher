package com.kingfisher.modules.template.mapper;

import com.kingfisher.common.query.Condition;
import com.kingfisher.modules.template.domain.Template;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;
import java.util.Map;

/**
 * 模版 Mapper，映射 SQLite templates 表。
 */
@Mapper
public interface TemplateMapper {

    List<Template> findAll(@Param("keyword") String keyword,
                           @Param("searchableColumns") List<String> searchableColumns,
                           @Param("filters") List<Condition> filters,
                           @Param("sort") String sort,
                           @Param("offset") int offset,
                           @Param("limit") int limit);

    long countAll(@Param("keyword") String keyword,
                  @Param("searchableColumns") List<String> searchableColumns,
                  @Param("filters") List<Condition> filters);

    Template findById(@Param("id") Long id);

    Template findByCode(@Param("code") String code);

    void insert(Template template);

    void update(@Param("id") Long id, @Param("updates") Map<String, Object> updates);

    void deleteById(@Param("id") Long id);

    void deleteBatch(@Param("ids") List<Long> ids);

    void updateStatusBatch(@Param("ids") List<Long> ids, @Param("status") int status);
}
