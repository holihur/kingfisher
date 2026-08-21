package com.kingfisher.modules.user.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.modules.audit.domain.AuditLog;
import com.kingfisher.modules.audit.service.AuditService;
import com.kingfisher.modules.user.dto.LoginRequest;
import com.kingfisher.modules.user.dto.LoginResponse;
import com.kingfisher.modules.user.dto.RefreshRequest;
import com.kingfisher.modules.user.service.AuthService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * 认证控制器，公开接口无需鉴权。
 */
@RestController
@RequestMapping("/api/v1/auth")
public class AuthController {

    private final AuthService authService;
    private final AuditService auditService;

    public AuthController(AuthService authService, AuditService auditService) {
        this.authService = authService;
        this.auditService = auditService;
    }

    /**
     * 登录 — 成功/失败均写入 audit_logs（action=login, resource=auth），与 Go auditLogin 对齐
     */
    @PostMapping("/login")
    public ResponseEntity<ApiResponse<LoginResponse>> login(@Valid @RequestBody LoginRequest req, HttpServletRequest request) {
        long start = System.currentTimeMillis();
        try {
            AuthService.LoginResult r = authService.login(req.getUsername(), req.getPassword());
            LoginResponse resp = new LoginResponse(r.accessToken(), r.refreshToken(), r.user(), r.landingPage());
            // 登录日志：成功
            auditLogin(request, r.user().getId(), req.getUsername(), "success", start);
            return ResponseEntity.ok(ApiResponse.ok(resp));
        } catch (AuthService.LoginFailedException e) {
            int code = e.getCode();
            // 登录日志：失败（userId=0，与 Go 一致）
            auditLogin(request, 0L, req.getUsername(), "failure", start);
            ApiResponse<LoginResponse> body = ApiResponse.error(code);
            return ResponseEntity.status(ErrorCode.httpStatus(code)).body(body);
        }
    }

    private void auditLogin(HttpServletRequest request, Long userId, String username, String result, long start) {
        try {
            AuditLog log = new AuditLog();
            log.setUserId(userId);
            log.setUsername(username);
            log.setAction("login");
            log.setResource("auth");
            log.setResult(result);
            log.setLatency(System.currentTimeMillis() - start);
            log.setIp(getClientIp(request));
            log.setUserAgent(request.getHeader("User-Agent"));
            log.setDetail("{\"username\":\"" + username + "\"}");
            log.setMessage(result);
            auditService.log(log);
        } catch (Exception ignored) {
            // 审计失败不影响主流程
        }
    }

    /**
     * 刷新 access_token
     */
    @PostMapping("/refresh")
    public ResponseEntity<ApiResponse<Map<String, String>>> refresh(@Valid @RequestBody RefreshRequest req) {
        try {
            // 兼容前端两种字段命名：refresh_token / refreshToken
            String token = req.getRefreshToken();
            if (token == null || token.isBlank()) {
                return ResponseEntity.status(400).body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "refresh_token 不能为空"));
            }
            String newAccess = authService.refresh(token);
            Map<String, String> data = Map.of("access_token", newAccess);
            return ResponseEntity.ok(ApiResponse.ok(data));
        } catch (com.kingfisher.security.JwtProvider.JwtExpiredException e) {
            return ResponseEntity.status(401).body(ApiResponse.error(ErrorCode.ERR_TOKEN_EXPIRED));
        } catch (com.kingfisher.security.JwtProvider.JwtInvalidException e) {
            return ResponseEntity.status(401).body(ApiResponse.error(ErrorCode.ERR_TOKEN_INVALID));
        } catch (Exception e) {
            return ResponseEntity.status(401).body(ApiResponse.error(ErrorCode.ERR_TOKEN_INVALID));
        }
    }

    @PostMapping("/logout")
    public ResponseEntity<ApiResponse<Void>> logout(@RequestHeader(value = "Authorization", required = false) String auth,
                                                     @RequestBody(required = false) Map<String, String> body,
                                                     HttpServletRequest request) {
        String token = null;
        if (auth != null && auth.startsWith("Bearer ")) {
            token = auth.substring(7);
        } else if (body != null) {
            token = body.get("refresh_token");
            if (token == null) token = body.get("refreshToken");
            if (token == null) token = body.get("access_token");
        }
        if (token != null) {
            authService.logout(token);
        }
        try {
            Long userId = (Long) request.getAttribute("user_id");
            String username = (String) request.getAttribute("username");
            if (username == null) username = "unknown";
            AuditLog log = new AuditLog();
            log.setUserId(userId != null ? userId : 0L);
            log.setUsername(username);
            log.setAction("logout");
            log.setResource("auth");
            log.setResult("success");
            log.setIp(getClientIp(request));
            log.setUserAgent(request.getHeader("User-Agent"));
            log.setLatency(0L);
            auditService.log(log);
        } catch (Exception ignored) {
        }
        return ResponseEntity.ok(ApiResponse.ok());
    }

    private String getClientIp(HttpServletRequest request) {
        String ip = request.getHeader("X-Forwarded-For");
        if (ip != null && !ip.isBlank()) return ip.split(",")[0].trim();
        ip = request.getHeader("X-Real-IP");
        if (ip != null && !ip.isBlank()) return ip;
        return request.getRemoteAddr();
    }

    @PostMapping("/forgot-password")
    public ResponseEntity<ApiResponse<Void>> forgotPassword(@RequestBody Map<String, String> body) {
        String email = body.get("email");
        if (email == null || email.isBlank()) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "email 不能为空"));
        try {
            authService.forgotPassword(email);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.internalServerError().body(ApiResponse.error(ErrorCode.ERR_INTERNAL));
        }
    }

    @PostMapping("/reset-password")
    public ResponseEntity<ApiResponse<Void>> resetPassword(@RequestBody Map<String, String> body) {
        String token = body.get("token");
        String newPassword = body.get("new_password");
        if (newPassword == null) newPassword = body.get("newPassword");
        if (token == null || token.isBlank() || newPassword == null || newPassword.isBlank()) return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, "token 和 new_password 不能为空"));
        try {
            authService.resetPassword(token, newPassword);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.internalServerError().body(ApiResponse.error(ErrorCode.ERR_INTERNAL));
        }
    }

    /**
     * 健康检查
     */
    @GetMapping("/health")
    public ApiResponse<String> health() {
        return ApiResponse.ok("ok");
    }
}
