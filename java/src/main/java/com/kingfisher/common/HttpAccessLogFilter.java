package com.kingfisher.common;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.MDC;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;
import org.springframework.web.util.ContentCachingRequestWrapper;
import org.springframework.web.util.ContentCachingResponseWrapper;

import java.io.IOException;
import java.util.UUID;

@Component
@Order(Ordered.HIGHEST_PRECEDENCE + 10)
public class HttpAccessLogFilter extends OncePerRequestFilter {

    private static final Logger log = LoggerFactory.getLogger(HttpAccessLogFilter.class);

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        long start = System.currentTimeMillis();
        String traceId = UUID.randomUUID().toString().substring(0, 8);
        MDC.put("traceId", traceId);
        ContentCachingRequestWrapper reqWrapper = new ContentCachingRequestWrapper(request);
        ContentCachingResponseWrapper respWrapper = new ContentCachingResponseWrapper(response);
        try {
            chain.doFilter(reqWrapper, respWrapper);
        } finally {
            long latency = System.currentTimeMillis() - start;
            String method = request.getMethod();
            String uri = request.getRequestURI();
            String query = request.getQueryString();
            if (query != null) uri += "?" + query;
            int status = respWrapper.getStatus();
            String ip = getClientIp(request);
            String ua = request.getHeader("User-Agent");
            Object uid = request.getAttribute("user_id");
            String user = uid != null ? String.valueOf(uid) : "-";
            String username = (String) request.getAttribute("username");
            if (username != null) user += "/" + username;
            if (uri.startsWith("/assets") || uri.startsWith("/uploads") || uri.equals("/api/v1/auth/health")) {
                log.debug("{} {} {} {} {}ms user={} ip={}", method, uri, status, latency, user, ip);
            } else {
                log.info("{} {} {} {}ms user={} ip={} trace={} ua={}", method, uri, status, latency, user, ip, traceId, ua != null && ua.length() > 80 ? ua.substring(0, 80) : ua);
            }
            MDC.remove("traceId");
            respWrapper.copyBodyToResponse();
        }
    }

    private String getClientIp(HttpServletRequest request) {
        String ip = request.getHeader("X-Forwarded-For");
        if (ip != null && !ip.isBlank()) return ip.split(",")[0].trim();
        ip = request.getHeader("X-Real-IP");
        if (ip != null && !ip.isBlank()) return ip;
        return request.getRemoteAddr();
    }
}
