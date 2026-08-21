package com.kingfisher.modules.config.mapper;

import com.kingfisher.modules.config.domain.ConfigGroup;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface ConfigGroupMapper {

    List<ConfigGroup> findAll();

    ConfigGroup findById(@Param("id") Long id);

    void insert(ConfigGroup group);

    void update(@Param("id") Long id, @Param("name") String name, @Param("sort") Integer sort);

    void deleteById(@Param("id") Long id);

    void resetGroupConfigs(@Param("groupId") Long groupId);
}
