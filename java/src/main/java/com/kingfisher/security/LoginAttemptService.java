package com.kingfisher.security;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * 登录失败计数，防暴力破解。
 * 与 Go 的 login_fail:{username} 逻辑对齐：15 分钟窗口，>5 次拒绝。
 */
@Service
public class LoginAttemptService {

    private final StringRedisTemplate redisTemplate;
    // 内存回退：username -> {count, expireAt}
    private final Map<String, Entry> memory = new ConcurrentHashMap<>();

    private static class Entry {
        AtomicInteger count = new AtomicInteger(0);
        long expireAt;
    }

    public LoginAttemptService(org.springframework.beans.factory.ObjectProvider<StringRedisTemplate> provider) {
        this.redisTemplate = provider.getIfAvailable();
    }

    /**
     * 记录一次失败，返回当前计数（已含本次）。
     */
    public long recordFail(String username) {
        if (redisTemplate != null) {
            try {
                String key = "login_fail:" + username;
                Long cnt = redisTemplate.opsForValue().increment(key);
                if (cnt != null && cnt == 1) {
                    redisTemplate.expire(key, Duration.ofMinutes(15));
                }
                return cnt == null ? 1 : cnt;
            } catch (Exception ignored) {
            }
        }
        Entry e = memory.compute(username, (k, v) -> {
            long now = System.currentTimeMillis();
            if (v == null || now > v.expireAt) {
                Entry ne = new Entry();
                ne.count.set(1);
                ne.expireAt = now + Duration.ofMinutes(15).toMillis();
                return ne;
            } else {
                v.count.incrementAndGet();
                return v;
            }
        });
        return e.count.get();
    }

    public void clear(String username) {
        if (redisTemplate != null) {
            try {
                redisTemplate.delete("login_fail:" + username);
            } catch (Exception ignored) {
            }
        }
        memory.remove(username);
    }

    public boolean isBlocked(String username) {
        // 仅在 recordFail 时判断，此处保留接口
        return false;
    }
}
