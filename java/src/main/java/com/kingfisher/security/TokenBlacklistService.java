package com.kingfisher.security;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Token 黑名单（撤销）存储。
 * 优先使用 Redis，若不可用则回退到内存（适用于单机开发/测试）。
 * key: blacklist:token:{jti}
 */
@Service
public class TokenBlacklistService {

    private final StringRedisTemplate redisTemplate; // 可能为 null（可选依赖）
    private final Map<String, Long> memoryStore = new ConcurrentHashMap<>();

    public TokenBlacklistService(org.springframework.beans.factory.ObjectProvider<StringRedisTemplate> provider) {
        this.redisTemplate = provider.getIfAvailable();
    }

    public void blacklist(String jti, Duration ttl) {
        if (ttl.isNegative() || ttl.isZero()) return;
        if (redisTemplate != null) {
            try {
                redisTemplate.opsForValue().set("blacklist:token:" + jti, "1", ttl);
                return;
            } catch (Exception ignored) {
                // 回退内存
            }
        }
        long expireAt = System.currentTimeMillis() + ttl.toMillis();
        memoryStore.put(jti, expireAt);
    }

    public boolean isBlacklisted(String jti) {
        if (redisTemplate != null) {
            try {
                Boolean exists = redisTemplate.hasKey("blacklist:token:" + jti);
                if (Boolean.TRUE.equals(exists)) return true;
            } catch (Exception ignored) {
            }
        }
        Long expireAt = memoryStore.get(jti);
        if (expireAt == null) return false;
        if (System.currentTimeMillis() > expireAt) {
            memoryStore.remove(jti);
            return false;
        }
        return true;
    }
}
