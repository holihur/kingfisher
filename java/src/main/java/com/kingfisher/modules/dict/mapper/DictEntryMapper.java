package com.kingfisher.modules.dict.mapper;

import com.kingfisher.modules.dict.domain.DictEntry;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import java.util.List;

@Mapper
public interface DictEntryMapper {
    List<DictEntry> findByTypeId(@Param("typeId") Long typeId,
                                 @Param("keyword") String keyword,
                                 @Param("filters") List<com.kingfisher.common.query.Condition> filters,
                                 @Param("sort") String sort,
                                 @Param("offset") int offset,
                                 @Param("limit") int limit);
    long countByTypeId(@Param("typeId") Long typeId, @Param("keyword") String keyword, @Param("filters") List<com.kingfisher.common.query.Condition> filters);
    List<DictEntry> findPublicByCode(@Param("code") String code);
    DictEntry findById(@Param("id") Long id);
    void insert(DictEntry entry);
    void update(DictEntry entry);
    void deleteById(@Param("id") Long id);
    void deleteBatch(@Param("ids") List<Long> ids);
    void updateStatusBatch(@Param("ids") List<Long> ids, @Param("status") int status);
}
