package com.kingfisher.modules.config.mapper;

import com.kingfisher.modules.config.domain.SystemConfig;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;
import java.util.Map;

@Mapper
public interface ConfigMapper {

    List<SystemConfig> list(@Param("keyword") String keyword,
                            @Param("searchable") List<String> searchable,
                            @Param("filters") List<Map<String, Object>> filters,
                            @Param("sort") String sort,
                            @Param("offset") int offset,
                            @Param("pageSize") int pageSize);

    long count(@Param("keyword") String keyword,
               @Param("searchable") List<String> searchable,
               @Param("filters") List<Map<String, Object>> filters);

    List<SystemConfig> findAllPublic();

    SystemConfig findByKey(@Param("key") String key);

    SystemConfig findPublicByKey(@Param("key") String key);

    void upsert(@Param("key") String key,
                @Param("value") String value,
                @Param("isPublic") boolean isPublic,
                @Param("version") String version,
                @Param("render") String render,
                @Param("renderOptions") String renderOptions,
                @Param("groupId") Long groupId);

    void deleteByKey(@Param("key") String key);

    void deleteBatch(@Param("keys") List<String> keys);
}
