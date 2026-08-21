package com.kingfisher.modules.audit;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.modules.audit.mapper.AuditLogMapper;
import com.kingfisher.modules.audit.service.AuditService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class AuditInterceptorTest {

    @Autowired AuditService auditService;
    @Autowired AuditLogMapper auditLogMapper;
    private AuditInterceptor interceptor;

    @BeforeEach
    void setUp() {
        interceptor = new AuditInterceptor(auditService, new ObjectMapper());
    }

    @Test
    void getRequest_shouldNotLog() throws Exception {
        long before = auditLogMapper.countAll("", java.util.List.of(), java.util.List.of());
        MockHttpServletRequest req = new MockHttpServletRequest("GET", "/api/v1/users");
        req.setAttribute("user_id", 1L);
        req.setAttribute("username", "admin");
        MockHttpServletResponse resp = new MockHttpServletResponse();
        resp.setStatus(200);
        interceptor.preHandle(req, resp, new Object());
        interceptor.afterCompletion(req, resp, new Object(), null);
        long after = auditLogMapper.countAll("", java.util.List.of(), java.util.List.of());
        assertEquals(before, after);
    }

    @Test
    void unauthenticatedPost_shouldNotLog() throws Exception {
        long before = auditLogMapper.countAll("", java.util.List.of(), java.util.List.of());
        MockHttpServletRequest req = new MockHttpServletRequest("POST", "/api/v1/users");
        MockHttpServletResponse resp = new MockHttpServletResponse();
        resp.setStatus(200);
        interceptor.preHandle(req, resp, new Object());
        interceptor.afterCompletion(req, resp, new Object(), null);
        long after = auditLogMapper.countAll("", java.util.List.of(), java.util.List.of());
        assertEquals(before, after);
    }

    @Test
    void authenticatedPost_shouldLog() throws Exception {
        long before = auditLogMapper.countAll("", java.util.List.of(), java.util.List.of());
        MockHttpServletRequest req = new MockHttpServletRequest("POST", "/api/v1/users");
        req.setAttribute("user_id", 1L);
        req.setAttribute("username", "admin");
        MockHttpServletResponse resp = new MockHttpServletResponse();
        resp.setStatus(200);
        interceptor.preHandle(req, resp, new Object());
        Thread.sleep(10);
        interceptor.afterCompletion(req, resp, new Object(), null);
        long after = auditLogMapper.countAll("", java.util.List.of(), java.util.List.of());
        assertEquals(before + 1, after);
    }

    @Test
    void failureShouldLogWithFailureResult() throws Exception {
        MockHttpServletRequest req = new MockHttpServletRequest("DELETE", "/api/v1/users/1");
        req.setAttribute("user_id", 1L);
        req.setAttribute("username", "admin");
        MockHttpServletResponse resp = new MockHttpServletResponse();
        resp.setStatus(403);
        interceptor.preHandle(req, resp, new Object());
        interceptor.afterCompletion(req, resp, new Object(), null);
        var logs = auditLogMapper.findAll("", java.util.List.of(), java.util.List.of(), "id DESC", 0, 1);
        assertFalse(logs.isEmpty());
        assertEquals("failure", logs.get(0).getResult());
        assertNotNull(logs.get(0).getLatency());
    }
}
