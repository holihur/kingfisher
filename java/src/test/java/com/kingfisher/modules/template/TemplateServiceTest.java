package com.kingfisher.modules.template;

import com.kingfisher.modules.template.service.TemplateService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class TemplateServiceTest {

    @Autowired TemplateService service;

    @Test
    void list_happy_shouldReturnSeed() {
        var req = new org.springframework.mock.web.MockHttpServletRequest();
        var result = service.list(req);
        assertNotNull(result);
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L));
    }

    @Test
    void create_happy_shouldSucceed() {
        String code = "test_" + System.nanoTime();
        var t = service.create("name", code, "general", "title", "content", 1, "", "1.0.0");
        assertNotNull(t);
        assertNotNull(t.getId());
    }

    @Test
    void create_bad_duplicateCode_shouldThrow() {
        String code = "dup_" + System.nanoTime();
        service.create("n", code, "general", "t", "c", 1, "", "1.0.0");
        assertThrows(IllegalArgumentException.class, () -> service.create("n2", code, "general", "t2", "c2", 1, "", "1.0.0"));
    }

    @Test
    void getById_happy_shouldReturnSeed() {
        String code = "get_" + System.nanoTime();
        var created = service.create("n", code, "general", "t", "c", 1, "", "1.0.0");
        assertNotNull(service.getById(created.getId()));
    }
}
