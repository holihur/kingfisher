package com.kingfisher.core.queue;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class RocketMQQueueConsumer {

    private static final Logger log = LoggerFactory.getLogger(RocketMQQueueConsumer.class);
    private final QueueProperties props;
    private final Map<String, QueueHandler> handlers = new ConcurrentHashMap<>();

    public RocketMQQueueConsumer(QueueProperties props) {
        this.props = props;
    }

    public void registerHandler(String taskType, QueueHandler handler) {
        handlers.put(taskType, handler);
        log.info("RocketMQ 消费者注册处理器 type={}", taskType);
    }

    @PostConstruct
    public void init() {
        if (!"rocketmq".equalsIgnoreCase(props.getType())) {
            log.info("RocketMQ 消费者未启用（当前类型={}）", props.getType());
            return;
        }
        log.info("RocketMQ 消费者已启动（桩实现） topic={} group={} — 未引入 rocketmq-client 时仅记录注册，生产环境请引入客户端并订阅", props.getRocketmq().getTopic(), props.getRocketmq().getConsumerGroup());
    }

    @PreDestroy
    public void destroy() {
    }
}
