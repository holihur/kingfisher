package com.kingfisher.core.queue;

import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class QueueService implements QueueProducer {

    private static final Logger log = LoggerFactory.getLogger(QueueService.class);
    private final QueueProperties props;
    private final MemoryQueueProducer memoryProducer;
    private final RocketMQQueueProducer rocketMQProducer;
    private final RocketMQQueueConsumer rocketMQConsumer;
    private QueueProducer active;

    public QueueService(QueueProperties props, MemoryQueueProducer memoryProducer, RocketMQQueueProducer rocketMQProducer, RocketMQQueueConsumer rocketMQConsumer) {
        this.props = props;
        this.memoryProducer = memoryProducer;
        this.rocketMQProducer = rocketMQProducer;
        this.rocketMQConsumer = rocketMQConsumer;
    }

    @PostConstruct
    public void init() {
        String t = props.getType() != null ? props.getType().toLowerCase() : "memory";
        if ("rocketmq".equals(t)) {
            active = rocketMQProducer;
            log.info("队列已选择 RocketMQ 模式");
        } else {
            active = memoryProducer;
            log.info("队列已选择 Memory 模式");
        }
    }

    @Override
    public void enqueue(Task task) {
        if (active == null) init();
        active.enqueue(task);
    }

    public void enqueue(String type, String payload) {
        enqueue(Task.of(type, payload));
    }

    public void enqueueJson(String type, Object payloadObj) {
        enqueue(Task.ofJson(type, payloadObj));
    }

    @Override
    public String getType() {
        return active != null ? active.getType() : props.getType();
    }

    public void registerHandler(String taskType, QueueHandler handler) {
        memoryProducer.registerHandler(taskType, handler);
        rocketMQConsumer.registerHandler(taskType, handler);
        log.info("队列处理器已注册 type={} -> {}", taskType, handler.getClass().getSimpleName());
    }

    public List<String> supportedTypes() {
        return List.of("memory", "rocketmq");
    }
}
