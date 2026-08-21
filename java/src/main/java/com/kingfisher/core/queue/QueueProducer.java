package com.kingfisher.core.queue;

public interface QueueProducer {
    void enqueue(Task task);
    default void close() {}
    String getType();
}
