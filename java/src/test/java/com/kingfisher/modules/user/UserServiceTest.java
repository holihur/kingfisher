package com.kingfisher.modules.user;

import com.kingfisher.modules.user.service.UserService;
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
class UserServiceTest {

    @Autowired UserService service;

    @Test
    void getById_happy_shouldReturnSeed() {
        var u = service.getById(1L);
        assertNotNull(u);
        assertEquals("admin", u.getUsername());
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L));
    }

    @Test
    void createUser_happy_shouldSucceed() {
        String username = "test_" + System.nanoTime();
        var created = service.createUser(username, "Abcd1234", "a@b.com", List.of(4L), List.of());
        assertNotNull(created);
        assertNotNull(created.getId());
        assertEquals(username, created.getUsername());
    }

    @Test
    void createUser_bad_duplicate_shouldThrow() {
        assertThrows(Exception.class, () -> service.createUser("admin", "Abcd1234", null, null, null));
    }

    @Test
    void changePassword_bad_wrongOld_shouldThrow() {
        assertThrows(Exception.class, () -> service.changePassword(1L, "wrongOldPass", "NewPass123"));
    }

    @Test
    void changePassword_happy_shouldSucceed() {
        assertDoesNotThrow(() -> service.changePassword(2L, "Abcd1234", "NewPass1234"));
    }
}
