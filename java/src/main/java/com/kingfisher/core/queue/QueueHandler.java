package com.kingfisher.core.queue;

@FunctionalInterface
public interface QueueHandler {
    void handle(String payload) throws Exception;
}
