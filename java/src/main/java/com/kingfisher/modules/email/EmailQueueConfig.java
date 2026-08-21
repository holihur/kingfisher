package com.kingfisher.modules.email;

import com.kingfisher.core.queue.QueueService;
import com.kingfisher.core.queue.Task;
import com.kingfisher.modules.email.task.EmailTask;
import com.kingfisher.modules.email.worker.EmailWorker;
import jakarta.annotation.PostConstruct;
import org.springframework.stereotype.Component;

@Component
public class EmailQueueConfig {

    private final QueueService queueService;
    private final EmailWorker emailWorker;

    public EmailQueueConfig(QueueService queueService, EmailWorker emailWorker) {
        this.queueService = queueService;
        this.emailWorker = emailWorker;
    }

    @PostConstruct
    public void register() {
        queueService.registerHandler(EmailTask.TYPE_SEND_EMAIL, payload -> {
            EmailTask task = EmailTask.fromJson(payload);
            emailWorker.handleSendEmail(task);
        });
    }
}
