package com.kingfisher.common;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.config.RateLimitProperties;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.time.Duration;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Component
@Order(Ordered.HIGHEST_PRECEDENCE + 5)
public class RateLimitFilter extends OncePerRequestFilter {

    private final RateLimitProperties props;
    private final StringRedisTemplate redisTemplate;
    private final ObjectMapper objectMapper;
    private final ConcurrentHashMap<String, AtomicInteger> memoryCounter = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, Long> memoryExpire = new ConcurrentHashMap<>();

    public RateLimitFilter(RateLimitProperties props, org.springframework.beans.factory.ObjectProvider<StringRedisTemplate> redisProvider, ObjectMapper objectMapper) {
        this.props = props;
        this.redisTemplate = redisProvider.getIfAvailable();
        this.objectMapper = objectMapper;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        if (!props.isEnabled()) {
            chain.doFilter(request, response);
            return;
        }
        String path = request.getRequestURI();
        if (path.startsWith("/api/v1/auth/health") || path.startsWith("/assets") || path.startsWith("/uploads")) {
            chain.doFilter(request, response);
            return;
        }
        String key = "ratelimit:" + getClientIp(request) + ":" + path;
        int limit = props.getRequestsPerMinute();
        if (path.contains("/auth/login") || path.contains("/auth/refresh")) limit = props.getLoginPerMinute() > 0 ? props.getLoginPerMinute() : limit;
        if (isLimited(key, limit)) {
            ApiResponse<Void> body = ApiResponse.error(ErrorCode.ERR_TOO_MANY_REQUEST);
            response.setStatus(ErrorCode.httpStatus(ErrorCode.ERR_TOO_MANY_REQUEST));
            response.setContentType("application/json;charset=UTF-8");
            response.getWriter().write(objectMapper.writeValueAsString(body));
            return;
        }
        chain.doFilter(request, response);
    }

    private boolean isLimited(String key, int limit) {
        if (redisTemplate != null) {
            try {
                Long count = redisTemplate.opsForValue().increment(key);
                if (count != null && count == 1) redisTemplate.expire(key, Duration.ofMinutes(1));
                return count != null && count > limit;
            } catch (Exception ignored) {}
        }
        long now = System.currentTimeMillis();
        Long expire = memoryExpire.get(key);
        if (expire != null && now > expire) {
            memoryCounter.remove(key);
            memoryExpire.remove(key);
        }
        AtomicInteger counter = memoryCounter.computeIfAbsent(key, k -> new AtomicInteger(0));
        int c = counter.incrementAndGet();
        if (c == 1) memoryExpire.put(key, now + 60000);
        return c > limit;
    }

    private String getClientIp(HttpServletRequest request) {
        String ip = request.getHeader("X-Forwarded-For");
        if (ip != null && !ip.isBlank()) return ip.split(",")[0].trim();
        ip = request.getHeader("X-Real-IP");
        if (ip != null && !ip.isBlank()) return ip;
        return request.getRemoteAddr();
    }
}
