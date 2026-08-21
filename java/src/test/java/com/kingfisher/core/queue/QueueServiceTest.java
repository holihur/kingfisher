package com.kingfisher.core.queue;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

import static org.junit.jupiter.api.Assertions.*;

class QueueServiceTest {

    private QueueProperties props;
    private MemoryQueueProducer memory;
    private RocketMQQueueProducer rocket;
    private RocketMQQueueConsumer consumer;
    private QueueService queueService;

    @BeforeEach
    void setUp() {
        props = new QueueProperties();
        props.setType("memory");
        memory = new MemoryQueueProducer();
        rocket = new RocketMQQueueProducer(props);
        consumer = new RocketMQQueueConsumer(props);
        queueService = new QueueService(props, memory, rocket, consumer);
        queueService.init();
    }

    @Test
    void defaultType_isMemory() {
        assertEquals("memory", queueService.getType());
    }

    @Test
    void memoryEnqueue_shouldDispatchToHandler() throws Exception {
        CountDownLatch latch = new CountDownLatch(1);
        AtomicReference<String> received = new AtomicReference<>();
        queueService.registerHandler("test:task", payload -> {
            received.set(payload);
            latch.countDown();
        });
        Task task = Task.of("test:task", "{\"hello\":\"world\"}");
        queueService.enqueue(task);
        assertTrue(latch.await(2, TimeUnit.SECONDS));
        assertEquals("{\"hello\":\"world\"}", received.get());
    }

    @Test
    void unknownTask_shouldNotThrow() {
        Task task = Task.of("unknown:type", "{}");
        assertDoesNotThrow(() -> queueService.enqueue(task));
    }

    @Test
    void switchToRocketMQ() {
        props.setType("rocketmq");
        queueService.init();
        assertEquals("rocketmq", queueService.getType());
        Task task = Task.of("email:send", "{\"to\":\"a@b.com\"}");
        assertDoesNotThrow(() -> queueService.enqueue(task));
    }

    @Test
    void taskFactory() {
        Task t = Task.ofJson("email:send", new com.kingfisher.modules.email.task.EmailTask("a@b.com", "s", "b"));
        assertEquals("email:send", t.getType());
        assertTrue(t.getPayload().contains("a@b.com"));
    }
}
