package com.kingfisher.modules.user.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.modules.email.EmailProducer;
import com.kingfisher.modules.template.mapper.TemplateMapper;
import com.kingfisher.modules.user.domain.Role;
import com.kingfisher.modules.user.domain.User;
import com.kingfisher.modules.user.mapper.UserMapper;
import com.kingfisher.security.JwtProvider;
import com.kingfisher.security.LoginAttemptService;
import com.kingfisher.security.TokenBlacklistService;
import io.jsonwebtoken.Claims;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;

import java.security.SecureRandom;
import java.time.Duration;
import java.util.Date;
import java.util.HexFormat;
import java.util.List;

/**
 * 认证服务，复刻 Go extends/user/app.AuthService.Login。
 */
@Service
public class AuthService {

    // 与 Go 一致的 dummyHash，用于防时序攻击（用户不存在时也做一次 bcrypt 比对）
    private static final String DUMMY_HASH = "$2a$12$LJ3m4ys3Lk0TSwHCpNqrIeN5U5Akn5dQUhBvPXFxFG7GqQvHCzB5q";

    private final UserMapper userMapper;
    private final JwtProvider jwtProvider;
    private final LoginAttemptService loginAttemptService;
    private final TokenBlacklistService blacklistService;
    private final StringRedisTemplate redisTemplate;
    private final EmailProducer emailProducer;
    private final TemplateMapper templateMapper;
    private final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();
    private final SecureRandom secureRandom = new SecureRandom();

    public AuthService(UserMapper userMapper, JwtProvider jwtProvider,
                       LoginAttemptService loginAttemptService,
                       TokenBlacklistService blacklistService,
                       ObjectProvider<StringRedisTemplate> redisProvider,
                       ObjectProvider<EmailProducer> emailProducerProvider,
                       ObjectProvider<TemplateMapper> templateMapperProvider) {
        this.userMapper = userMapper;
        this.jwtProvider = jwtProvider;
        this.loginAttemptService = loginAttemptService;
        this.blacklistService = blacklistService;
        this.redisTemplate = redisProvider.getIfAvailable();
        this.emailProducer = emailProducerProvider.getIfAvailable();
        this.templateMapper = templateMapperProvider.getIfAvailable();
    }

    public record LoginResult(String accessToken, String refreshToken, User user, String landingPage) {}

    public LoginResult login(String username, String password) {
        User user = userMapper.findByUsername(username);
        String hashToCheck = DUMMY_HASH;
        boolean userExists = true;
        if (user == null) {
            userExists = false;
        } else {
            hashToCheck = user.getPassword();
        }

        boolean passwordMatches;
        try {
            passwordMatches = encoder.matches(password, hashToCheck);
            // 兼容 Go 生成的 $2a$ 前缀，Spring 的 BCrypt 需 $2a -> $2b 处理由底层自动兼容，若失败则尝试直接比对
        } catch (Exception e) {
            passwordMatches = false;
        }

        if (!passwordMatches) {
            long cnt = loginAttemptService.recordFail(username);
            if (cnt > 5) {
                throw new LoginFailedException("too many attempts", 10107);
            }
            throw new LoginFailedException("wrong password", 10103);
        }

        if (!userExists || user == null) {
            throw new LoginFailedException("wrong password", 10103);
        }

        if (user.getStatus() == null || user.getStatus() != 1) {
            throw new LoginFailedException("user disabled", 10106);
        }

        if (user.getRoles() == null || user.getRoles().isEmpty()) {
            throw new LoginFailedException("user has no roles", 10103);
        }

        // 清除失败计数
        loginAttemptService.clear(username);

        List<Long> roleIds = user.getRoleIds();
        List<String> roleCodes = user.getRoleCodes();
        int sv = user.getSessionVersion() == null ? 1 : user.getSessionVersion();

        JwtProvider.TokenPair pair = jwtProvider.generateToken(user.getId(), roleIds, roleCodes, user.getUsername(), sv);

        // 落地页：取第一个角色的 landing_page
        String landing = "";
        try {
            Role firstRole = user.getRoles().get(0);
            if (firstRole != null && firstRole.getLandingPage() != null) {
                landing = firstRole.getLandingPage();
            }
        } catch (Exception ignored) {}

        // 脱敏：不返回密码
        user.setPassword(null);

        return new LoginResult(pair.accessToken(), pair.refreshToken(), user, landing);
    }

    public String refresh(String refreshToken) {
        // 先校验黑名单
        Claims core = jwtProvider.parseCore(refreshToken);
        String jti = core.get("jti", String.class);
        if (blacklistService.isBlacklisted(jti)) {
            throw new JwtProvider.JwtInvalidException("token revoked");
        }
        // 解析并校验 type=refresh 是否过期等
        Claims refreshClaims = jwtProvider.parseRefreshToken(refreshToken);
        // 校验 sessionVersion（可选，阶段一简化：需查 DB 对比，暂跳过）
        return jwtProvider.refreshAccessToken(refreshToken);
    }

    public void logout(String token) {
        try {
            Claims c = jwtProvider.parseCore(token);
            String jti = c.get("jti", String.class);
            Date exp = c.getExpiration();
            if (jti != null && exp != null) {
                long ttlMs = exp.getTime() - System.currentTimeMillis();
                if (ttlMs > 0) {
                    blacklistService.blacklist(jti, Duration.ofMillis(ttlMs));
                }
            }
        } catch (Exception ignored) {
            // token 已无效则无需处理
        }
    }

    public void forgotPassword(String email) {
        User user = null;
        try {
            user = userMapper.findByEmail(email);
        } catch (Exception e) {
            throw new RuntimeException("find user: " + e.getMessage(), e);
        }
        if (user == null) return;
        byte[] bytes = new byte[32];
        secureRandom.nextBytes(bytes);
        String token = HexFormat.of().formatHex(bytes);
        if (redisTemplate != null) {
            try {
                redisTemplate.opsForValue().set("reset:token:" + token, String.valueOf(user.getId()), Duration.ofMinutes(30));
            } catch (Exception e) {
                throw new RuntimeException("store reset token: " + e.getMessage(), e);
            }
        }
        if (emailProducer != null && templateMapper != null) {
            try {
                var tmpl = templateMapper.findByCode("password_reset");
                if (tmpl != null) {
                    String resetUrl = "http://localhost:8080/reset?token=" + token;
                    String subject = tmpl.getTitle().replace("{{nickname}}", user.getNickname() != null ? user.getNickname() : user.getUsername());
                    String body = tmpl.getContent()
                            .replace("{{nickname}}", user.getNickname() != null ? user.getNickname() : user.getUsername())
                            .replace("{{reset_url}}", resetUrl)
                            .replace("{{token}}", token);
                    emailProducer.enqueueEmail(user.getEmail(), subject, body);
                }
            } catch (Exception ignored) {}
        }
    }

    public void resetPassword(String token, String newPassword) {
        if (newPassword == null || newPassword.length() < 8 || newPassword.length() > 64) {
            throw new IllegalArgumentException("password length invalid");
        }
        if (redisTemplate == null) throw new IllegalArgumentException("reset token unavailable");
        String uidStr;
        try {
            uidStr = redisTemplate.opsForValue().get("reset:token:" + token);
        } catch (Exception e) {
            throw new IllegalArgumentException("invalid or expired reset token");
        }
        if (uidStr == null || uidStr.isBlank()) throw new IllegalArgumentException("invalid or expired reset token");
        long userId;
        try { userId = Long.parseLong(uidStr); } catch (Exception e) { throw new IllegalArgumentException("invalid reset token"); }
        String hashed = encoder.encode(newPassword);
        userMapper.updatePassword(userId, hashed);
        try { userMapper.incrementSessionVersion(userId); } catch (Exception ignored) {}
        try { redisTemplate.delete("reset:token:" + token); } catch (Exception ignored) {}
    }

    public static class LoginFailedException extends RuntimeException {
        private final int code;
        public LoginFailedException(String message, int code) {
            super(message);
            this.code = code;
        }
        public int getCode() { return code; }
    }
}
