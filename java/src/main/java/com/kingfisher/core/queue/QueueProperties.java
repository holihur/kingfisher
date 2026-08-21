package com.kingfisher.core.queue;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "kingfisher.queue")
public class QueueProperties {
    private String type = "memory";
    private Rocketmq rocketmq = new Rocketmq();

    @Data
    public static class Rocketmq {
        private String nameServer = "127.0.0.1:9876";
        private String topic = "kingfisher-tasks";
        private String producerGroup = "kingfisher-producer";
        private String consumerGroup = "kingfisher-consumer";
        private int retryTimes = 3;
        private int sendTimeout = 3000;
    }
}
