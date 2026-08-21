package com.kingfisher.modules.email;

import com.kingfisher.config.SmtpProperties;
import jakarta.mail.internet.MimeMessage;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.mail.javamail.JavaMailSenderImpl;
import org.springframework.mail.javamail.MimeMessageHelper;
import org.springframework.stereotype.Component;

import java.util.Properties;

@Component
public class Mailer {

    private static final Logger log = LoggerFactory.getLogger(Mailer.class);
    private final SmtpProperties props;

    public Mailer(SmtpProperties props) {
        this.props = props;
    }

    public boolean isEnabled() {
        return props.isEnabled() && props.getHost() != null && !props.getHost().isBlank();
    }

    public void send(String to, String subject, String body) {
        if (!isEnabled()) {
            log.info("[邮件-模拟] to={} subject={} bodyLen={}", to, subject, body != null ? body.length() : 0);
            return;
        }
        try {
            JavaMailSenderImpl sender = new JavaMailSenderImpl();
            sender.setHost(props.getHost());
            sender.setPort(props.getPort());
            sender.setUsername(props.getUsername());
            sender.setPassword(props.getPassword());
            Properties p = sender.getJavaMailProperties();
            p.put("mail.transport.protocol", "smtp");
            p.put("mail.smtp.auth", "true");
            p.put("mail.smtp.starttls.enable", "true");
            p.put("mail.smtp.starttls.required", "true");
            p.put("mail.smtp.ssl.enable", String.valueOf(props.getPort() == 465));
            p.put("mail.smtp.connectiontimeout", "10000");
            p.put("mail.smtp.timeout", "10000");

            MimeMessage message = sender.createMimeMessage();
            MimeMessageHelper helper = new MimeMessageHelper(message, false, "UTF-8");
            String from = props.getFrom() != null && !props.getFrom().isBlank() ? props.getFrom() : props.getUsername();
            if (props.getFromName() != null && !props.getFromName().isBlank()) {
                helper.setFrom(from, props.getFromName());
            } else {
                helper.setFrom(from);
            }
            helper.setTo(to);
            helper.setSubject(subject);
            helper.setText(body, true);
            sender.send(message);
            log.info("邮件已发送 to={} subject={}", to, subject);
        } catch (Exception e) {
            log.error("邮件发送失败 to={}", to, e);
            throw new RuntimeException("邮件发送失败: " + e.getMessage(), e);
        }
    }
}
