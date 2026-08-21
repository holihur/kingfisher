package com.kingfisher.core.queue;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;

@Component
public class RocketMQQueueProducer implements QueueProducer {

    private static final Logger log = LoggerFactory.getLogger(RocketMQQueueProducer.class);
    private final QueueProperties props;
    private volatile boolean started = false;

    public RocketMQQueueProducer(QueueProperties props) {
        this.props = props;
    }

    @PostConstruct
    public void init() {
        if (!"rocketmq".equalsIgnoreCase(props.getType())) {
            log.info("RocketMQQueue 未启用（当前队列类型={}），跳过初始化", props.getType());
            return;
        }
        try {
            started = true;
            log.info("RocketMQQueue 已启动（桩实现） nameServer={} topic={} — 未引入 rocketmq-client 依赖时仅记录日志，生产环境请引入 org.apache.rocketmq:rocketmq-client 并替换为真实发送", props.getRocketmq().getNameServer(), props.getRocketmq().getTopic());
        } catch (Exception e) {
            log.error("RocketMQQueue 启动失败，将降级为日志模式", e);
        }
    }

    @Override
    public void enqueue(Task task) {
        if (!"rocketmq".equalsIgnoreCase(props.getType())) {
            log.warn("RocketMQQueue 当前未启用，忽略入队 type={}", task.getType());
            return;
        }
        if (!started) {
            log.warn("RocketMQQueue 未就绪（桩模式），仅记录日志 type={}", task.getType());
            return;
        }
        log.info("RocketMQ 入队（桩） type={} payloadLen={} — 需引入真实 RocketMQ 客户端后替换为 producer.send", task.getType(), task.getPayload() != null ? task.getPayload().length() : 0);
    }

    @Override
    public void close() {
        started = false;
    }

    @PreDestroy
    public void destroy() {
        close();
    }

    @Override
    public String getType() {
        return "rocketmq";
    }
}
