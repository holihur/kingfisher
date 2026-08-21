package com.kingfisher.modules.config.service;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.config.domain.ConfigGroup;
import com.kingfisher.modules.config.domain.SystemConfig;
import com.kingfisher.modules.config.mapper.ConfigGroupMapper;
import com.kingfisher.modules.config.mapper.ConfigMapper;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.util.List;
import java.util.Map;

@Service
public class ConfigService {

    private static final String CACHE_PUBLIC_KEY = "config:public";
    private static final Duration CACHE_TTL = Duration.ofMinutes(5);

    private final ConfigMapper configMapper;
    private final ConfigGroupMapper configGroupMapper;
    private final StringRedisTemplate redisTemplate;
    private final ObjectMapper objectMapper;

    private String memoryPublicCache;
    private long memoryPublicCacheExpireAt;

    public ConfigService(ConfigMapper configMapper,
                         ConfigGroupMapper configGroupMapper,
                         ObjectProvider<StringRedisTemplate> redisProvider,
                         ObjectMapper objectMapper) {
        this.configMapper = configMapper;
        this.configGroupMapper = configGroupMapper;
        this.redisTemplate = redisProvider.getIfAvailable();
        this.objectMapper = objectMapper;
    }

    public List<SystemConfig> list(Query query) {
        List<String> searchable = query.searchableColumns();
        List<Map<String, Object>> filters = query.getFilters().stream()
                .map(f -> Map.<String, Object>of("field", f.getField(), "op", f.getOp().name().toLowerCase(), "value", f.getValue()))
                .toList();
        return configMapper.list(query.getKeyword(), searchable, filters, query.sortExpr(), query.offset(), query.getPageSize());
    }

    public long count(Query query) {
        List<String> searchable = query.searchableColumns();
        List<Map<String, Object>> filters = query.getFilters().stream()
                .map(f -> Map.<String, Object>of("field", f.getField(), "op", f.getOp().name().toLowerCase(), "value", f.getValue()))
                .toList();
        return configMapper.count(query.getKeyword(), searchable, filters);
    }

    public List<SystemConfig> getAllPublic() {
        String cached = getFromCache(CACHE_PUBLIC_KEY);
        if (cached != null) {
            try {
                return objectMapper.readValue(cached, new TypeReference<>() {});
            } catch (Exception ignored) {
            }
        }
        List<SystemConfig> configs = configMapper.findAllPublic();
        try {
            putToCache(CACHE_PUBLIC_KEY, objectMapper.writeValueAsString(configs), CACHE_TTL);
        } catch (Exception ignored) {
        }
        return configs;
    }

    public SystemConfig getByKey(String key) {
        return configMapper.findByKey(key);
    }

    public SystemConfig getPublicByKey(String key) {
        return configMapper.findPublicByKey(key);
    }

    public void set(String key, String value, boolean isPublic, String version, String render, String renderOptions, Long groupId) {
        configMapper.upsert(key, value, isPublic, version, render, renderOptions, groupId);
        deleteFromCache(CACHE_PUBLIC_KEY);
        deleteFromCache("config:" + key);
    }

    public void delete(String key) {
        configMapper.deleteByKey(key);
        deleteFromCache(CACHE_PUBLIC_KEY);
    }

    public void batchDelete(List<String> keys) {
        configMapper.deleteBatch(keys);
        deleteFromCache(CACHE_PUBLIC_KEY);
    }

    // ========== ConfigGroup ==========

    public List<ConfigGroup> listGroups() {
        return configGroupMapper.findAll();
    }

    public ConfigGroup createGroup(String name, int sort) {
        ConfigGroup group = new ConfigGroup();
        group.setName(name);
        group.setSort(sort);
        configGroupMapper.insert(group);
        return group;
    }

    public void updateGroup(Long id, String name, int sort) {
        configGroupMapper.update(id, name, sort);
    }

    @Transactional
    public void deleteGroup(Long id) {
        configGroupMapper.resetGroupConfigs(id);
        configGroupMapper.deleteById(id);
    }

    // ========== 缓存操作 ==========

    private String getFromCache(String key) {
        if (redisTemplate != null) {
            try {
                return redisTemplate.opsForValue().get(key);
            } catch (Exception ignored) {
            }
        }
        if (memoryPublicCache != null && System.currentTimeMillis() < memoryPublicCacheExpireAt && CACHE_PUBLIC_KEY.equals(key)) {
            return memoryPublicCache;
        }
        return null;
    }

    private void putToCache(String key, String value, Duration ttl) {
        if (redisTemplate != null) {
            try {
                redisTemplate.opsForValue().set(key, value, ttl);
                return;
            } catch (Exception ignored) {
            }
        }
        if (CACHE_PUBLIC_KEY.equals(key)) {
            memoryPublicCache = value;
            memoryPublicCacheExpireAt = System.currentTimeMillis() + ttl.toMillis();
        }
    }

    private void deleteFromCache(String key) {
        if (redisTemplate != null) {
            try {
                redisTemplate.delete(key);
            } catch (Exception ignored) {
            }
        }
        if (CACHE_PUBLIC_KEY.equals(key)) {
            memoryPublicCache = null;
            memoryPublicCacheExpireAt = 0;
        }
    }
}
