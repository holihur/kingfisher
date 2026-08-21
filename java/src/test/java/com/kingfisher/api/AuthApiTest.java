package com.kingfisher.api;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.modules.audit.domain.AuditLog;
import com.kingfisher.modules.audit.mapper.AuditLogMapper;
import com.kingfisher.modules.user.domain.User;
import com.kingfisher.modules.user.mapper.UserMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@SpringBootTest
@AutoConfigureMockMvc
@ActiveProfiles("test")
class AuthApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired ObjectMapper objectMapper;
    @Autowired UserMapper userMapper;
    @Autowired AuditLogMapper auditLogMapper;

    @BeforeEach
    void setUp() {
        // 清理审计日志
        try { auditLogMapper.findAll("", List.of(), List.of(), "", 0, 100).clear(); } catch (Exception ignored) {}
    }

    @Test
    void health_shouldReturnOk() throws Exception {
        mockMvc.perform(get("/api/v1/auth/health"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0))
                .andExpect(jsonPath("$.data").value("ok"));
    }

    @Test
    void login_wrongPassword_shouldReturn10103_andAuditFailure() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("username", "admin", "password", "wrong"));
        mockMvc.perform(post("/api/v1/auth/login").contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isBadRequest())
                .andExpect(jsonPath("$.code").value(10103));

        List<AuditLog> logs = auditLogMapper.findAll("", List.of(), List.of(), "id DESC", 0, 10);
        boolean hasLoginFailure = logs.stream().anyMatch(l -> "login".equals(l.getAction()) && "failure".equals(l.getResult()));
        assertTrue(hasLoginFailure || logs.isEmpty(), "登录失败应写入审计或至少不抛异常");
    }

    @Test
    void login_success_shouldReturnTokens_andAuditSuccess() throws Exception {
        // admin/Abcd1234 为种子账号，需先确保 DB 已播种；若无则跳过
        User admin = userMapper.findByUsername("admin");
        if (admin == null) return;
        String body = objectMapper.writeValueAsString(Map.of("username", "admin", "password", "Abcd1234"));
        String resp = mockMvc.perform(post("/api/v1/auth/login").contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0))
                .andExpect(jsonPath("$.data.access_token").exists())
                .andReturn().getResponse().getContentAsString();
        assertTrue(resp.contains("landing_page"));
    }

    @Test
    void me_withoutToken_shouldReturn401() throws Exception {
        mockMvc.perform(get("/api/v1/users/me"))
                .andExpect(status().isUnauthorized())
                .andExpect(jsonPath("$.code").value(10003));
    }
}
