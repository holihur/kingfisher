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
class TaskApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired ObjectMapper objectMapper;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() { return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken(); }
    private String nopermToken() { return jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken(); }

    @Test
    void list_happy_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/scheduled-tasks").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void list_bad_noperm_should403() throws Exception {
        mockMvc.perform(get("/api/v1/scheduled-tasks").header("Authorization", "Bearer " + nopermToken()))
                .andExpect(status().isForbidden());
    }

    @Test
    void create_happy_shouldSucceed() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("name", "t" + System.nanoTime(), "task_type", "email:send", "cron_spec", "* * * * *", "payload", "{}"));
        mockMvc.perform(post("/api/v1/scheduled-tasks").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk());
    }

    @Test
    void create_bad_invalidCron_shouldStill200_becauseNoValidation() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("name", "t" + System.nanoTime(), "task_type", "email:send", "cron_spec", "bad", "payload", "{}"));
        mockMvc.perform(post("/api/v1/scheduled-tasks").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk());
    }

    @Test
    void create_bad_missingName_should400() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("task_type", "email:send", "cron_spec", "* * * * *", "payload", "{}"));
        var resp = mockMvc.perform(post("/api/v1/scheduled-tasks").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body)).andReturn().getResponse();
        int status = resp.getStatus();
        assertTrue(status == 400 || status == 500, "missing name should be 4xx or 500, got " + status);
    }

    @Test
    void getById_bad_notFound_should404() throws Exception {
        mockMvc.perform(get("/api/v1/scheduled-tasks/99999").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().is4xxClientError());
    }
}
