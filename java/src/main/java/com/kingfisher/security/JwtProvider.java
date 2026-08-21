package com.kingfisher.security;

import com.kingfisher.config.JwtProperties;
import io.jsonwebtoken.*;
import io.jsonwebtoken.security.Keys;
import org.springframework.stereotype.Component;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.time.Instant;
import java.util.Date;
import java.util.List;
import java.util.UUID;

/**
 * JWT 签发与校验，与 Go core/jwt 对齐。
 * Claims: user_id, role_ids, roles, username, jti, type, sv, exp/iat/iss
 */
@Component
public class JwtProvider {

    private final JwtProperties props;
    private final SecretKey key;

    public JwtProvider(JwtProperties props) {
        this.props = props;
        // HS256 要求 key 长度 >= 32 字节，Go 的默认 secret 较短，这里按字节填充/截断
        byte[] secretBytes = props.getSecret().getBytes(StandardCharsets.UTF_8);
        if (secretBytes.length < 32) {
            byte[] padded = new byte[32];
            System.arraycopy(secretBytes, 0, padded, 0, secretBytes.length);
            // 剩余用 0 填充，jjwt 仍可工作；生产环境请配置 32+ 字符的 secret
            secretBytes = padded;
        }
        this.key = Keys.hmacShaKeyFor(secretBytes);
    }

    public record TokenPair(String accessToken, String refreshToken, String jti) {}

    public TokenPair generateToken(Long userId, List<Long> roleIds, List<String> roleCodes,
                                   String username, int sessionVersion) {
        String jti = UUID.randomUUID().toString();
        Instant now = Instant.now();
        Date accessExp = Date.from(now.plus(parseDuration(props.getAccessTtl())));
        Date refreshExp = Date.from(now.plus(parseDuration(props.getRefreshTtl())));

        String access = Jwts.builder()
                .issuer(props.getIssuer())
                .issuedAt(Date.from(now))
                .expiration(accessExp)
                .claim("user_id", userId)
                .claim("role_ids", roleIds)
                .claim("roles", roleCodes)
                .claim("username", username)
                .claim("jti", jti)
                .claim("type", "access")
                .claim("sv", sessionVersion)
                .signWith(key)
                .compact();

        String refresh = Jwts.builder()
                .issuer(props.getIssuer())
                .issuedAt(Date.from(now))
                .expiration(refreshExp)
                .claim("user_id", userId)
                .claim("role_ids", roleIds)
                .claim("roles", roleCodes)
                .claim("username", username)
                .claim("jti", jti)
                .claim("type", "refresh")
                .claim("sv", sessionVersion)
                .signWith(key)
                .compact();

        return new TokenPair(access, refresh, jti);
    }

    /**
     * 刷新：用 refresh token 签发新的 access token（与 Go RefreshToken 对齐，换新 jti）。
     */
    public String refreshAccessToken(String refreshToken) {
        Claims claims = parseAndValidate(refreshToken, "refresh");
        Instant now = Instant.now();
        Date exp = Date.from(now.plus(parseDuration(props.getAccessTtl())));
        String newJti = UUID.randomUUID().toString();
        return Jwts.builder()
                .issuer(props.getIssuer())
                .issuedAt(Date.from(now))
                .expiration(exp)
                .claim("user_id", claims.get("user_id", Long.class))
                .claim("role_ids", claims.get("role_ids", List.class))
                .claim("roles", claims.get("roles", List.class))
                .claim("username", claims.get("username", String.class))
                .claim("jti", newJti)
                .claim("type", "access")
                .claim("sv", claims.get("sv", Integer.class))
                .signWith(key)
                .compact();
    }

    public Claims parseAccessToken(String token) {
        return parseAndValidate(token, "access");
    }

    public Claims parseRefreshToken(String token) {
        return parseAndValidate(token, "refresh");
    }

    /**
     * 供黑名单/撤销时解析 jti 与过期时间（不校验 type）。
     */
    public Claims parseCore(String token) {
        try {
            return Jwts.parser()
                    .verifyWith(key)
                    .requireIssuer(props.getIssuer())
                    .build()
                    .parseSignedClaims(token)
                    .getPayload();
        } catch (ExpiredJwtException e) {
            throw new JwtExpiredException("jwt: token expired", e);
        } catch (JwtException e) {
            throw new JwtInvalidException("parse token: " + e.getMessage(), e);
        }
    }

    private Claims parseAndValidate(String token, String expectedType) {
        Claims claims;
        try {
            claims = Jwts.parser()
                    .verifyWith(key)
                    .requireIssuer(props.getIssuer())
                    .build()
                    .parseSignedClaims(token)
                    .getPayload();
        } catch (ExpiredJwtException e) {
            throw new JwtExpiredException("jwt: token expired", e);
        } catch (JwtException e) {
            throw new JwtInvalidException("parse token: " + e.getMessage(), e);
        }
        String type = claims.get("type", String.class);
        if (!expectedType.equals(type)) {
            throw new JwtInvalidException("not an " + expectedType + " token");
        }
        return claims;
    }

    public static Duration parseDuration(String raw) {
        if (raw == null || raw.isBlank()) return Duration.ofHours(2);
        raw = raw.trim();
        // 支持 Go 风格：2h, 168h, 30m 等，Java Duration.parse 需要 ISO8601，先简单处理
        try {
            if (raw.endsWith("h")) {
                long h = Long.parseLong(raw.substring(0, raw.length() - 1));
                return Duration.ofHours(h);
            } else if (raw.endsWith("m")) {
                long m = Long.parseLong(raw.substring(0, raw.length() - 1));
                return Duration.ofMinutes(m);
            } else if (raw.endsWith("s")) {
                long s = Long.parseLong(raw.substring(0, raw.length() - 1));
                return Duration.ofSeconds(s);
            }
            return Duration.parse(raw);
        } catch (Exception e) {
            return Duration.ofHours(2);
        }
    }

    public static class JwtExpiredException extends RuntimeException {
        public JwtExpiredException(String msg, Throwable cause) { super(msg, cause); }
    }

    public static class JwtInvalidException extends RuntimeException {
        public JwtInvalidException(String msg) { super(msg); }
        public JwtInvalidException(String msg, Throwable cause) { super(msg, cause); }
    }
}
