package com.kingfisher.modules.config;

import com.kingfisher.modules.config.service.ConfigService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class ConfigServiceTest {

    @Autowired ConfigService service;

    @Test
    void getByKey_happy_shouldReturnSeed() {
        var cfg = service.getByKey("site_name");
        assertNotNull(cfg);
        assertNotNull(cfg.getValue());
    }

    @Test
    void getByKey_bad_notFound_shouldReturnNull() {
        assertNull(service.getByKey("missing_key_" + System.nanoTime()));
    }

    @Test
    void set_happy_shouldUpsertAndReadBack() {
        String key = "test_key_" + System.nanoTime();
        service.set(key, "val", false, "1.0.0", "text", null, 1L);
        var cfg = service.getByKey(key);
        assertNotNull(cfg);
        assertEquals("val", cfg.getValue());
    }

    @Test
    void delete_happy_shouldRemove() {
        String key = "del_key_" + System.nanoTime();
        service.set(key, "v", false, "1.0.0", "text", null, 1L);
        service.delete(key);
        assertNull(service.getByKey(key));
    }

    @Test
    void getAllPublic_happy_shouldReturnSeed() {
        var list = service.getAllPublic();
        assertNotNull(list);
        assertTrue(list.stream().anyMatch(c -> "site_name".equals(c.getKey())));
    }

    @Test
    void getPublicByKey_bad_privateShouldReturnNull() {
        assertNull(service.getPublicByKey("missing_private_" + System.nanoTime()));
    }
}
