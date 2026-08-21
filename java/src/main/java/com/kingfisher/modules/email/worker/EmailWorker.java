package com.kingfisher.modules.email.worker;

import com.kingfisher.modules.email.Mailer;
import com.kingfisher.modules.email.task.EmailTask;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class EmailWorker {

    private static final Logger log = LoggerFactory.getLogger(EmailWorker.class);
    private final Mailer mailer;

    public EmailWorker(Mailer mailer) {
        this.mailer = mailer;
    }

    public String getName() {
        return "email";
    }

    public String getTaskType() {
        return EmailTask.TYPE_SEND_EMAIL;
    }

    public void handleSendEmail(String jsonPayload) {
        EmailTask p = EmailTask.fromJson(jsonPayload);
        handleSendEmail(p);
    }

    public void handleSendEmail(byte[] payload) {
        EmailTask p = EmailTask.fromBytes(payload);
        handleSendEmail(p);
    }

    public void handleSendEmail(EmailTask p) {
        if (p.getTo() == null || p.getTo().isBlank()) {
            throw new IllegalArgumentException("invalid email payload (empty to)");
        }
        try {
            mailer.send(p.getTo(), p.getSubject(), p.getBody());
        } catch (Exception e) {
            log.error("邮件发送失败 to={}", p.getTo(), e);
            throw e;
        }
    }
}
