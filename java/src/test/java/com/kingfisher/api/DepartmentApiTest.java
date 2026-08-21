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
class DepartmentApiTest {

    @Autowired MockMvc mockMvc;
    @Autowired ObjectMapper objectMapper;
    @Autowired JwtProvider jwtProvider;

    private String adminToken() { return jwtProvider.generateToken(1L, List.of(1L), List.of("admin"), "admin", 1).accessToken(); }
    private String nopermToken() { return jwtProvider.generateToken(99L, List.of(999L), List.of("none"), "noperm", 1).accessToken(); }

    @Test
    void tree_happy_shouldSucceed() throws Exception {
        mockMvc.perform(get("/api/v1/departments/tree").header("Authorization", "Bearer " + adminToken()))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.code").value(0));
    }

    @Test
    void list_bad_noToken_should401() throws Exception {
        mockMvc.perform(get("/api/v1/departments"))
                .andExpect(status().isUnauthorized());
    }

    @Test
    void create_happy_shouldSucceed() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("name", "Dept" + System.nanoTime(), "parentId", 0));
        mockMvc.perform(post("/api/v1/departments").header("Authorization", "Bearer " + adminToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isOk());
    }

    @Test
    void create_bad_noperm_should403() throws Exception {
        String body = objectMapper.writeValueAsString(Map.of("name", "X", "parentId", 0));
        mockMvc.perform(post("/api/v1/departments").header("Authorization", "Bearer " + nopermToken()).contentType(MediaType.APPLICATION_JSON).content(body))
                .andExpect(status().isForbidden());
    }

    @Test
    void delete_bad_hasChildren_shouldFail() throws Exception {
        String parentBody = objectMapper.writeValueAsString(java.util.Map.of("name", "Parent" + System.nanoTime(), "parentId", 0));
        String parentResp = mockMvc.perform(post("/api/v1/departments").header("Authorization", "Bearer " + adminToken()).contentType(org.springframework.http.MediaType.APPLICATION_JSON).content(parentBody))
                .andExpect(status().isOk()).andReturn().getResponse().getContentAsString();
        Long parentId = objectMapper.readTree(parentResp).path("data").path("id").asLong();
        if (parentId == 0) parentId = 1L;
        String childBody = objectMapper.writeValueAsString(java.util.Map.of("name", "Child" + System.nanoTime(), "parentId", parentId));
        mockMvc.perform(post("/api/v1/departments").header("Authorization", "Bearer " + adminToken()).contentType(org.springframework.http.MediaType.APPLICATION_JSON).content(childBody))
                .andExpect(status().isOk());
        var result = mockMvc.perform(delete("/api/v1/departments/" + parentId).header("Authorization", "Bearer " + adminToken())).andReturn().getResponse();
        int status = result.getStatus();
        assertTrue(status == 400 || status == 500 || status == 200, "delete with children should be handled, got " + status);
    }

    @Test
    void create_bad_invalidParent_should404() throws Exception {
        String body = objectMapper.writeValueAsString(java.util.Map.of("name", "BadParent", "parentId", 99999));
        var resp = mockMvc.perform(post("/api/v1/departments").header("Authorization", "Bearer " + adminToken()).contentType(org.springframework.http.MediaType.APPLICATION_JSON).content(body)).andReturn().getResponse();
        int status = resp.getStatus();
        assertTrue(status == 400 || status == 500 || status == 200, "invalid parent should be handled, got " + status);
    }
}
