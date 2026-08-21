package com.kingfisher.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * CORS 配置
 */
@Data
@Component
@ConfigurationProperties(prefix = "kingfisher.cors")
public class CorsProperties {
    private List<String> allowedOrigins = List.of("http://localhost:5173");
    private List<String> allowedMethods = List.of("GET", "POST", "PUT", "DELETE", "OPTIONS");
    private List<String> allowedHeaders = List.of("Authorization", "Content-Type", "X-Request-ID");
    private Boolean allowCredentials = true;
}
