package com.kingfisher.modules.task;

import com.kingfisher.modules.task.service.ScheduledTaskService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class TaskServiceTest {

    @Autowired ScheduledTaskService service;

    @Test
    void list_happy_shouldReturnSeed() {
        var req = new org.springframework.mock.web.MockHttpServletRequest();
        var q = com.kingfisher.common.query.Query.parse(req, com.kingfisher.common.query.Defs.of("name", com.kingfisher.common.query.Field.searchableString("name")));
        var result = service.list(q);
        assertNotNull(result);
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L));
    }

    @Test
    void create_happy_shouldSucceed() {
        var t = service.create("test_" + System.nanoTime(), "email:send", "* * * * *", "{}", 1, "");
        assertNotNull(t);
        assertNotNull(t.getId());
    }

    @Test
    void getById_happy_shouldReturnCreated() {
        var created = service.create("get_" + System.nanoTime(), "email:send", "* * * * *", "{}", 1, "");
        assertNotNull(service.getById(created.getId()));
    }
}
