package com.kingfisher.api;

import com.kingfisher.security.JwtProvider;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import java.util.List;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@SpringBootTest
@AutoConfigureMockMvc
@ActiveProfiles("test")
class AuditApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() { return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken(); }
    private String nopermToken() { return jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken(); }

    @Test
    void list_happy_admin_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/audit-logs").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void list_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/audit-logs"))
                .andExpect(status().isUnauthorized());
    }

    @Test
    void list_bad_noperm_should403() throws Exception {
        mockMvc.perform(get("/api/v1/audit-logs").header("Authorization", "Bearer " + nopermToken()))
                .andExpect(status().isForbidden());
    }

    @Test
    void list_withFilter_happy_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/audit-logs").param("filter", "{\"action\":\"login\"}").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk());
    }
}
