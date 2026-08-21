package com.kingfisher.modules.dict;

import com.kingfisher.modules.dict.domain.DictType;
import com.kingfisher.modules.dict.service.DictTypeService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class DictServiceTest {

    @Autowired DictTypeService service;

    @Test
    void create_happy_shouldSucceed() {
        String code = "test_" + System.nanoTime();
        DictType created = service.create(code, "测试", false, 1, "", "1.0.0");
        assertNotNull(created);
        assertNotNull(created.getId());
        assertEquals(code, created.getCode());
    }

    @Test
    void create_bad_duplicateCode_shouldThrow() {
        assertThrows(IllegalArgumentException.class, () -> service.create("gender", "性别", false, 1, "", "1.0.0"));
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L));
    }

    @Test
    void list_happy_shouldReturn() {
        var list = service.list(new com.kingfisher.common.query.Query());
        assertNotNull(list);
    }

    @Test
    void getById_happy_shouldReturnSeed() {
        assertNotNull(service.getById(1L));
        assertEquals("gender", service.getById(1L).getCode());
    }
}
