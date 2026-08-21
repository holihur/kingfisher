package com.kingfisher.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "kingfisher.smtp")
public class SmtpProperties {
    private boolean enabled = false;
    private String host = "";
    private int port = 587;
    private String username = "";
    private String password = "";
    private String from = "";
    private String fromName = "Kingfisher";
}
