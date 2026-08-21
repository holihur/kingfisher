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
class MessageApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired ObjectMapper objectMapper;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() { return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken(); }

    @Test
    void inbox_happy_shouldReturn() throws Exception {
        mockMvc.perform(get("/api/v1/me/messages").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void inbox_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/me/messages"))
                .andExpect(status().isUnauthorized());
    }

    @Test
    void send_happy_admin_shouldSucceed() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("recipient_ids", List.of(3), "title", "hello", "content", "world"));
        mockMvc.perform(post("/api/v1/messages").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void send_bad_missingRecipients_should400() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("title", "no recipients"));
        mockMvc.perform(post("/api/v1/messages").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isBadRequest());
    }

    @Test
    void listSent_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/messages"))
                .andExpect(status().isUnauthorized());
    }
}
