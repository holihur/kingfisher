package com.kingfisher.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "kingfisher.rate-limit")
public class RateLimitProperties {
    private boolean enabled = false;
    private int requestsPerMinute = 60;
    private int loginPerMinute = 5;
}
