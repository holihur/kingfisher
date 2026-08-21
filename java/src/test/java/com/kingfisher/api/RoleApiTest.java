package com.kingfisher.api;

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
class RoleApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() {
        return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken();
    }
    private String viewerToken() {
        return jwtProvider.generateToken(3L, List.of(4L), List.of("viewer"), "viewer", 1).accessToken();
    }

    @Test
    void list_happy_admin() throws Exception {
        mockMvc.perform(get("/api/v1/roles").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0))
                .andExpect(jsonPath("$.data.items").isArray());
    }

    @Test
    void list_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/roles"))
                .andExpect(status().isUnauthorized())
                .andExpect(jsonPath("$.code").value(10003));
    }

    @Test
    void list_bad_viewerForbidden_should403() throws Exception {
        String noPermToken = jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken();
        mockMvc.perform(get("/api/v1/roles").header("Authorization", "Bearer " + noPermToken))
                .andExpect(status().isForbidden())
                .andExpect(jsonPath("$.code").value(10004));
    }

    @Test
    void list_happy_viewer_shouldAlsoSucceed_becauseSeedGivesRoleList() throws Exception {
        mockMvc.perform(get("/api/v1/roles").header("Authorization", "Bearer " + viewerToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void create_happy_shouldSucceed() throws Exception {
        String body = "{\"name\":\"testRole\",\"code\":\"test_\"}" + System.nanoTime();
        mockMvc.perform(post("/api/v1/roles").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void create_bad_viewer_should403() throws Exception {
        String body = "{\"name\":\"x\",\"code\":\"x\"}";
        mockMvc.perform(post("/api/v1/roles").header("Authorization", "Bearer " + viewerToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isForbidden());
    }

    @Test
    void getById_happy_should200() throws Exception {
        mockMvc.perform(get("/api/v1/roles/1").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void getById_bad_notFound_should404() throws Exception {
        mockMvc.perform(get("/api/v1/roles/99999").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isNotFound());
    }
}
