package com.kingfisher.modules.email;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.core.queue.QueueService;
import com.kingfisher.core.queue.Task;
import com.kingfisher.modules.email.task.EmailTask;
import com.kingfisher.modules.email.worker.EmailWorker;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class EmailProducer {

    private static final Logger log = LoggerFactory.getLogger(EmailProducer.class);
    private final QueueService queueService;
    private final EmailWorker emailWorker;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public EmailProducer(QueueService queueService, EmailWorker emailWorker) {
        this.queueService = queueService;
        this.emailWorker = emailWorker;
    }

    public void enqueueEmail(String to, String subject, String body) {
        enqueueEmail(new EmailTask(to, subject, body));
    }

    public void enqueueEmail(EmailTask payload) {
        try {
            String json = objectMapper.writeValueAsString(payload);
            Task task = Task.of(EmailTask.TYPE_SEND_EMAIL, json);
            log.info("邮件入队 queue={} to={} subject={}", queueService.getType(), payload.getTo(), payload.getSubject());
            queueService.enqueue(task);
        } catch (Exception e) {
            log.error("邮件入队失败 to={}", payload.getTo(), e);
            throw new RuntimeException(e);
        }
    }

    public void enqueueEmailSync(String to, String subject, String body) {
        emailWorker.handleSendEmail(new EmailTask(to, subject, body));
    }
}
