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
class DocApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired ObjectMapper objectMapper;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() { return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken(); }
    private String nopermToken() { return jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken(); }

    @Test
    void tree_happy_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/docs/tree").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void tree_bad_noperm_should403() throws Exception {
        mockMvc.perform(get("/api/v1/docs/tree").header("Authorization", "Bearer " + nopermToken()))
                .andExpect(status().isForbidden());
    }

    @Test
    void list_happy_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/docs?dirId=1").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk());
    }

    @Test
    void createDoc_happy_shouldSucceed() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("dir_id", 1, "title", "test" + System.nanoTime(), "content", "hello"));
        mockMvc.perform(post("/api/v1/docs").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void createDoc_bad_missingTitle_should500() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("dir_id", 1, "content", "no title"));
        mockMvc.perform(post("/api/v1/docs").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().is5xxServerError());
    }

    @Test
    void publicTree_happy_noToken_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/public/docs/tree"))
                .andExpect(status().isOk());
    }
}
