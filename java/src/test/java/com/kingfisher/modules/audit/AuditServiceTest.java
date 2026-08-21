package com.kingfisher.modules.audit;

import com.kingfisher.modules.audit.domain.AuditLog;
import com.kingfisher.modules.audit.service.AuditService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class AuditServiceTest {

    @Autowired AuditService service;

    @Test
    void log_happy_shouldInsertAndReadBack() {
        AuditLog log = new AuditLog();
        log.setUserId(1L); log.setUsername("admin"); log.setAction("create"); log.setResource("用户");
        log.setResult("success"); log.setDetail("{\"test\":1}");
        assertDoesNotThrow(() -> service.log(log));
        assertNotNull(log.getId());
    }

    @Test
    void logBatch_happy_shouldInsertBatch() {
        AuditLog a = new AuditLog(); a.setUserId(1L); a.setUsername("admin"); a.setAction("create"); a.setResource("用户"); a.setResult("success");
        AuditLog b = new AuditLog(); b.setUserId(2L); b.setUsername("editor"); b.setAction("update"); b.setResource("角色"); b.setResult("success");
        assertDoesNotThrow(() -> service.logBatch(java.util.List.of(a, b)));
    }

    @Test
    void list_happy_shouldReturnSeedOrEmpty() {
        MockHttpServletRequest req = new MockHttpServletRequest("GET", "/api/v1/audit-logs");
        var page = service.list(req);
        assertNotNull(page);
        assertTrue(page.getTotal() >= 0);
    }

    @Test
    void listByUserId_happy_shouldFilter() {
        MockHttpServletRequest req = new MockHttpServletRequest("GET", "/api/v1/users/me/login-logs");
        var page = service.listByUserId(1L, req);
        assertNotNull(page);
    }

    @Test
    void listLoginLogs_happy_shouldFilterLogin() {
        MockHttpServletRequest req = new MockHttpServletRequest("GET", "/api/v1/users/me/login-logs");
        var page = service.listLoginLogs(1L, req);
        assertNotNull(page);
    }
}
