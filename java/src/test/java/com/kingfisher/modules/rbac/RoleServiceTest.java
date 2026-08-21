package com.kingfisher.modules.rbac;

import com.kingfisher.modules.rbac.service.RoleService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class RoleServiceTest {

    @Autowired RoleService service;

    @Test
    void getById_happy_shouldReturnSeed() {
        var r = service.getById(1L);
        assertNotNull(r);
        assertEquals("admin", r.getCode());
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L));
    }

    @Test
    void create_happy_shouldSucceed() {
        var r = new com.kingfisher.modules.rbac.domain.Role();
        r.setName("test_" + System.nanoTime());
        r.setCode("test_" + System.nanoTime());
        r.setLevel(1);
        assertDoesNotThrow(() -> service.create(r));
        assertNotNull(r.getId());
    }

    @Test
    void delete_bad_admin_shouldThrow() {
        assertThrows(IllegalArgumentException.class, () -> service.delete(1L));
    }

    @Test
    void getUserPermissions_happy_shouldReturnSeed() {
        var perms = service.getUserPermissions(1L);
        assertNotNull(perms);
        assertTrue(perms.contains("user:list"));
    }

    @Test
    void getUserPermissions_bad_unknownUser_shouldReturnEmpty() {
        var perms = service.getUserPermissions(99999L);
        assertTrue(perms == null || perms.isEmpty());
    }
}
