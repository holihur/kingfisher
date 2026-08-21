package com.kingfisher.security;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.modules.user.mapper.UserMapper;
import io.jsonwebtoken.Claims;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

/**
 * JWT 鉴权过滤器，对受保护路由校验 access_token。
 * 行为与 Go AuthMiddleware 对齐：过期 -> 10104，未认证/无效 -> 10003，session_version 过旧 -> 10105
 */
@Component
public class JwtAuthFilter extends OncePerRequestFilter {

    private final JwtProvider jwtProvider;
    private final TokenBlacklistService blacklistService;
    private final UserMapper userMapper;
    private final ObjectMapper objectMapper;

    public JwtAuthFilter(JwtProvider jwtProvider, TokenBlacklistService blacklistService,
                         UserMapper userMapper, ObjectMapper objectMapper) {
        this.jwtProvider = jwtProvider;
        this.blacklistService = blacklistService;
        this.userMapper = userMapper;
        this.objectMapper = objectMapper;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
            throws ServletException, IOException {

        String path = request.getRequestURI();

        // 公开接口直接放行
        if (isPublic(path)) {
            filterChain.doFilter(request, response);
            return;
        }

        // 仅拦截 /api/v1 下的受保护接口
        if (!path.startsWith("/api/v1/")) {
            filterChain.doFilter(request, response);
            return;
        }

        String auth = request.getHeader("Authorization");
        if (auth == null || !auth.startsWith("Bearer ")) {
            writeError(response, ErrorCode.ERR_UNAUTHORIZED);
            return;
        }

        String token = auth.substring(7);
        try {
            Claims claims = jwtProvider.parseAccessToken(token);
            String jti = claims.get("jti", String.class);
            if (jti != null && blacklistService.isBlacklisted(jti)) {
                writeError(response, ErrorCode.ERR_TOKEN_INVALID);
                return;
            }
            // session_version 校验
            Long userId = claims.get("user_id", Long.class);
            Integer sv = claims.get("sv", Integer.class);
            if (userId != null && sv != null) {
                try {
                    var user = userMapper.findById(userId);
                    if (user != null && user.getSessionVersion() != null && sv < user.getSessionVersion()) {
                        writeError(response, ErrorCode.ERR_TOKEN_INVALID);
                        return;
                    }
                } catch (Exception ignored) {
                    // 查询失败不阻塞，放行（与 Go svp==nil 时跳过一致）
                }
            }

            // 注入请求上下文，供下游 Controller 使用
            request.setAttribute("user_id", claims.get("user_id", Long.class));
            request.setAttribute("username", claims.get("username", String.class));
            request.setAttribute("role_ids", claims.get("role_ids", java.util.List.class));
            request.setAttribute("claims", claims);

            filterChain.doFilter(request, response);
        } catch (JwtProvider.JwtExpiredException e) {
            writeError(response, ErrorCode.ERR_TOKEN_EXPIRED);
        } catch (JwtProvider.JwtInvalidException e) {
            writeError(response, ErrorCode.ERR_TOKEN_INVALID);
        } catch (Exception e) {
            writeError(response, ErrorCode.ERR_TOKEN_INVALID);
        }
    }

    private boolean isPublic(String path) {
        return path.startsWith("/api/v1/auth/login")
                || path.startsWith("/api/v1/auth/refresh")
                || path.startsWith("/api/v1/auth/register")
                || path.startsWith("/api/v1/auth/forgot-password")
                || path.startsWith("/api/v1/auth/reset-password")
                || path.startsWith("/api/v1/auth/health")
                || path.equals("/api/v1/auth/logout") // logout 允许无 token 调用
                || path.startsWith("/api/v1/public/"); // 公开文档等公开端点
    }

    private void writeError(HttpServletResponse response, int code) throws IOException {
        ApiResponse<Void> body = ApiResponse.error(code);
        response.setStatus(ErrorCode.httpStatus(code));
        response.setContentType("application/json;charset=UTF-8");
        response.getWriter().write(objectMapper.writeValueAsString(body));
    }
}
