package com.kingfisher.modules.email;

import com.kingfisher.modules.email.task.EmailTask;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class EmailTaskTest {

    @Test
    void jsonRoundTrip() {
        EmailTask task = new EmailTask("a@b.com", "hello", "<p>body</p>");
        String json = task.toJson();
        assertTrue(json.contains("a@b.com"));
        EmailTask parsed = EmailTask.fromJson(json);
        assertEquals(task.getTo(), parsed.getTo());
        assertEquals(task.getSubject(), parsed.getSubject());
        assertEquals(task.getBody(), parsed.getBody());
    }

    @Test
    void fromBytes() {
        EmailTask task = new EmailTask("x@y.com", "s", "b");
        byte[] bytes = task.toJson().getBytes();
        EmailTask parsed = EmailTask.fromBytes(bytes);
        assertEquals("x@y.com", parsed.getTo());
    }

    @Test
    void typeConstant() {
        assertEquals("email:send", EmailTask.TYPE_SEND_EMAIL);
    }
}
