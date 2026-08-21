package com.kingfisher.modules.config.service;

import com.kingfisher.modules.config.domain.ConfigGroup;
import com.kingfisher.modules.config.mapper.ConfigGroupMapper;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class ConfigGroupService {

    private final ConfigGroupMapper configGroupMapper;

    public ConfigGroupService(ConfigGroupMapper configGroupMapper) {
        this.configGroupMapper = configGroupMapper;
    }

    public List<ConfigGroup> list() {
        return configGroupMapper.findAll();
    }

    public ConfigGroup create(String name, Integer sort) {
        ConfigGroup group = new ConfigGroup();
        group.setName(name);
        group.setSort(sort != null ? sort : 0);
        configGroupMapper.insert(group);
        return group;
    }

    public void update(Long id, String name, Integer sort) {
        configGroupMapper.update(id, name, sort);
    }

    public void delete(Long id) {
        configGroupMapper.resetGroupConfigs(id);
        configGroupMapper.deleteById(id);
    }
}
