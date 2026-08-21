package com.kingfisher.core.queue;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executor;
import java.util.concurrent.Executors;

@Component
public class MemoryQueueProducer implements QueueProducer {

    private static final Logger log = LoggerFactory.getLogger(MemoryQueueProducer.class);
    private final Executor executor = Executors.newCachedThreadPool(r -> {
        Thread t = new Thread(r, "memory-queue-worker");
        t.setDaemon(true);
        return t;
    });
    private final Map<String, QueueHandler> handlers = new ConcurrentHashMap<>();

    public void registerHandler(String taskType, QueueHandler handler) {
        handlers.put(taskType, handler);
        log.info("MemoryQueue 注册处理器 type={}", taskType);
    }

    @Override
    public void enqueue(Task task) {
        log.info("MemoryQueue 入队 type={} payloadLen={}", task.getType(), task.getPayload() != null ? task.getPayload().length() : 0);
        QueueHandler h = handlers.get(task.getType());
        if (h == null) {
            log.warn("MemoryQueue 无处理器 type={}, 丢弃", task.getType());
            return;
        }
        executor.execute(() -> {
            try {
                h.handle(task.getPayload());
            } catch (Exception e) {
                log.error("MemoryQueue 处理失败 type={}", task.getType(), e);
            }
        });
    }

    @Override
    public String getType() {
        return "memory";
    }
}
