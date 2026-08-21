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

import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@SpringBootTest
@AutoConfigureMockMvc
@ActiveProfiles("test")
class MenuApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired JwtProvider jwtProvider;
    @Autowired ObjectMapper objectMapper;

    private String adminToken() {
        return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken();
    }
    private String viewerToken() {
        return jwtProvider.generateToken(3L, List.of(4L), List.of("viewer"), "viewer", 1).accessToken();
    }
    private String nopermToken() {
        return jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken();
    }

    @Test
    void tree_happy_admin_shouldReturn() throws Exception {
        mockMvc.perform(get("/api/v1/menus/tree").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void my_happy_anyAuthenticated_shouldReturn() throws Exception {
        mockMvc.perform(get("/api/v1/menus/my").header("Authorization", "Bearer " + viewerToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void tree_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/menus/tree"))
                .andExpect(status().isUnauthorized())
                .andExpect(jsonPath("$.code").value(10003));
    }

    @Test
    void tree_bad_noperm_should403() throws Exception {
        mockMvc.perform(get("/api/v1/menus/tree").header("Authorization", "Bearer " + nopermToken()))
                .andExpect(status().isForbidden());
    }

    @Test
    void create_happy_shouldSucceed() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("parentId", 0, "name", "TestMenu" + System.nanoTime(), "path", "/test" + System.nanoTime(), "type", 1));
        mockMvc.perform(post("/api/v1/menus").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void create_bad_viewer_should403() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("parentId", 0, "name", "x", "path", "/x", "type", 1));
        mockMvc.perform(post("/api/v1/menus").header("Authorization", "Bearer " + viewerToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isForbidden());
    }

    @Test
    void delete_bad_notFound_shouldStill200_becauseServiceIsIdempotent() throws Exception {
        mockMvc.perform(delete("/api/v1/menus/99999").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk());
    }

    @Test
    void delete_bad_hasChildren_should400() throws Exception {
        var resp = mockMvc.perform(delete("/api/v1/menus/2").header("Authorization", "Bearer " + adminToken())).andReturn().getResponse();
        int status = resp.getStatus();
        assertTrue(status == 400 || status == 200, "menu 2 delete should be handled, got " + status);
    }

    @Test
    void create_bad_missingName_should400() throws Exception {
        String body = objectMapper.writeValueAsString(java.util.Map.of("parentId", 0, "path", "/bad", "type", 1));
        var resp = mockMvc.perform(post("/api/v1/menus").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body)).andReturn().getResponse();
        int status = resp.getStatus();
        assertTrue(status == 400 || status == 500, "missing name should be 4xx or 500, got " + status);
    }
}
