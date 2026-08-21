package com.kingfisher.modules.dict.mapper;

import com.kingfisher.modules.dict.domain.DictType;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import java.util.List;

@Mapper
public interface DictTypeMapper {
    List<DictType> findAll(@Param("keyword") String keyword,
                           @Param("filters") List<com.kingfisher.common.query.Condition> filters,
                           @Param("sort") String sort,
                           @Param("offset") int offset,
                           @Param("limit") int limit);
    long countAll(@Param("keyword") String keyword, @Param("filters") List<com.kingfisher.common.query.Condition> filters);
    DictType findById(@Param("id") Long id);
    DictType findByCode(@Param("code") String code);
    void insert(DictType dictType);
    void update(DictType dictType);
    void deleteById(@Param("id") Long id);
    void deleteBatch(@Param("ids") List<Long> ids);
    void updateStatusBatch(@Param("ids") List<Long> ids, @Param("status") int status);
    long countEntriesByTypeId(@Param("typeId") Long typeId);
}
