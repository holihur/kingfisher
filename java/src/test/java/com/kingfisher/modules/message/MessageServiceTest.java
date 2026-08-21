package com.kingfisher.modules.message;

import com.kingfisher.modules.message.service.MessageService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class MessageServiceTest {

    @Autowired MessageService service;

    @Test
    void sendBatch_happy_shouldReturnCount() {
        int n = service.sendBatch(1L, "admin", List.of(2L, 3L), "title", "content");
        assertEquals(2, n);
    }

    @Test
    void sendBatch_bad_emptyRecipients_shouldReturnZero() {
        int n = service.sendBatch(1L, "admin", List.of(), "title", "content");
        assertEquals(0, n);
    }

    @Test
    void listByRecipient_happy_shouldReturnSeedOrEmpty() {
        var req = new org.springframework.mock.web.MockHttpServletRequest();
        var q = com.kingfisher.common.query.Query.parse(req, com.kingfisher.common.query.Defs.of("title", com.kingfisher.common.query.Field.searchableString("title")));
        var result = service.listByRecipient(1L, q);
        assertNotNull(result);
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L, 1L));
    }

    @Test
    void unreadCount_happy_shouldReturn() {
        long cnt = service.unreadCount(1L);
        assertTrue(cnt >= 0);
    }
}
