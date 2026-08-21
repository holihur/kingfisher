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
class TemplateApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired ObjectMapper objectMapper;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() { return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken(); }
    private String nopermToken() { return jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken(); }

    @Test
    void list_happy_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/templates").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void list_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/templates"))
                .andExpect(status().isUnauthorized());
    }

    @Test
    void list_bad_noperm_should403() throws Exception {
        mockMvc.perform(get("/api/v1/templates").header("Authorization", "Bearer " + nopermToken()))
                .andExpect(status().isForbidden());
    }

    @Test
    void create_happy_shouldSucceed() throws Exception {
        String code = "test_" + System.nanoTime();
        String body = objectMapper.writeValueAsString(Map.of("name", "n", "code", code, "templateType", "general", "title", "t", "content", "c"));
        mockMvc.perform(post("/api/v1/templates").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk());
    }

    @Test
    void create_bad_duplicate_shouldFail() throws Exception {
        String code = "dup_" + System.nanoTime();
        String body1 = objectMapper.writeValueAsString(Map.of("name", "n", "code", code, "templateType", "general", "title", "t", "content", "c"));
        mockMvc.perform(post("/api/v1/templates").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body1))
                .andExpect(status().isOk());
        String body2 = objectMapper.writeValueAsString(Map.of("name", "n2", "code", code, "templateType", "general", "title", "t2", "content", "c2"));
        mockMvc.perform(post("/api/v1/templates").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body2))
                .andExpect(status().is4xxClientError());
    }

    @Test
    void create_bad_missingCode_should400() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("name", "n", "templateType", "general", "title", "t", "content", "c"));
        mockMvc.perform(post("/api/v1/templates").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().is4xxClientError());
    }
}
