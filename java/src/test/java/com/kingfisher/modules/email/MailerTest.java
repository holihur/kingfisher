package com.kingfisher.modules.email;

import com.kingfisher.config.SmtpProperties;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class MailerTest {

    @Test
    void disabledMailer_shouldNotThrow() {
        SmtpProperties props = new SmtpProperties();
        props.setEnabled(false);
        props.setHost("smtp.example.com");
        Mailer mailer = new Mailer(props);
        assertFalse(mailer.isEnabled());
        assertDoesNotThrow(() -> mailer.send("user@example.com", "subject", "<p>body</p>"));
    }

    @Test
    void enabledWithoutHost_shouldBeDisabled() {
        SmtpProperties props = new SmtpProperties();
        props.setEnabled(true);
        props.setHost("");
        Mailer mailer = new Mailer(props);
        assertFalse(mailer.isEnabled());
        assertDoesNotThrow(() -> mailer.send("a@b.com", "s", "b"));
    }

    @Test
    void enabledWithHost_shouldBeEnabled() {
        SmtpProperties props = new SmtpProperties();
        props.setEnabled(true);
        props.setHost("smtp.example.com");
        props.setPort(587);
        props.setUsername("u");
        props.setPassword("p");
        Mailer mailer = new Mailer(props);
        assertTrue(mailer.isEnabled());
    }
}
