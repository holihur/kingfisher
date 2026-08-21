package com.kingfisher.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

/**
 * JWT 配置，与 Go core/config.JWTConfig 对齐。
 */
@Data
@Component
@ConfigurationProperties(prefix = "kingfisher.jwt")
public class JwtProperties {

    /** 签名密钥，默认 change-me-in-production，可通过环境变量 JWT_SECRET 覆盖 */
    private String secret = "change-me-in-production";
    private String issuer = "kingfisher";
    /** 支持 2h / 168h 这种字符串，解析为 Duration */
    private String accessTtl = "2h";
    private String refreshTtl = "168h";
}
