package com.kingfisher.api;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.security.JwtProvider;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.web.servlet.MockMvc;

import java.util.List;
import java.util.Map;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@SpringBootTest
@AutoConfigureMockMvc
@ActiveProfiles("test")
class UserApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired ObjectMapper objectMapper;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() {
        return jwtProvider.generateToken(1L, java.util.List.of(1L), java.util.List.of("admin"), "admin", 1).accessToken();
    }

    private String viewerToken() {
        return jwtProvider.generateToken(3L, java.util.List.of(4L), java.util.List.of("viewer"), "viewer", 1).accessToken();
    }

    @Test
    void me_happy_shouldReturnUser() throws Exception {
        mockMvc.perform(get("/api/v1/users/me").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0))
                .andExpect(jsonPath("$.data.username").value("admin"));
    }

    @Test
    void me_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/users/me"))
                .andExpect(status().isUnauthorized())
                .andExpect(jsonPath("$.code").value(10003));
    }

    @Test
    void me_bad_expiredToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/users/me").header("Authorization", "Bearer invalid.token.here"))
                .andExpect(status().isUnauthorized());
    }

    @Test
    void list_happy_admin_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/users").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void list_bad_viewerWithoutPerm_should403() throws Exception {
        String noPermToken = jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken();
        mockMvc.perform(get("/api/v1/users").header("Authorization", "Bearer " + noPermToken))
                .andExpect(status().isForbidden())
                .andExpect(jsonPath("$.code").value(10004));
    }

    @Test
    void list_happy_viewer_shouldSucceed_becauseViewerHasUserList() throws Exception {
        mockMvc.perform(get("/api/v1/users").header("Authorization", "Bearer " + viewerToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void list_bad_invalidQuery_should400() throws Exception {
        mockMvc.perform(get("/api/v1/users?filter=" + java.net.URLEncoder.encode("{\"status\":\"not_a_number\"}", "UTF-8")).header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isBadRequest());
    }

    @Test
    void create_happy_shouldSucceed() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("username", "test_" + System.nanoTime(), "password", "Abcd1234", "email", "t@test.com", "role_ids", java.util.List.of(4)));
        mockMvc.perform(post("/api/v1/users").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void create_bad_duplicate_should400() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("username", "admin", "password", "Abcd1234"));
        mockMvc.perform(post("/api/v1/users").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().is4xxClientError());
    }

    @Test
    void create_bad_missingPassword_should400() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("username", "no_pwd_user"));
        mockMvc.perform(post("/api/v1/users").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isBadRequest());
    }

    @Test
    void getById_happy_shouldReturn() throws Exception {
        mockMvc.perform(get("/api/v1/users/1").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void getById_bad_notFound_should404() throws Exception {
        mockMvc.perform(get("/api/v1/users/99999").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isNotFound());
    }

    @Test
    void loginLogs_happy_shouldReturnPage() throws Exception {
        mockMvc.perform(get("/api/v1/users/me/login-logs").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }
}
