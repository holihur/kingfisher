package com.kingfisher.modules.agent;

import com.kingfisher.modules.agent.mapper.AgentMapper;
import com.kingfisher.modules.agent.service.AgentService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

class AgentServiceTest {

    private AgentMapper mapper;
    private AgentService service;

    @BeforeEach
    void setUp() {
        mapper = mock(AgentMapper.class);
        service = new AgentService(mapper, mock(com.kingfisher.modules.agent.service.LlmClient.class), mock(com.kingfisher.modules.config.service.ConfigService.class));
        try {
            var f = AgentService.class.getDeclaredField("enabled");
            f.setAccessible(true);
            f.set(service, true);
        } catch (Exception ignored) {}
    }

    @Test
    void listConversations_happy_shouldReturn() {
        when(mapper.listConversations(1L)).thenReturn(java.util.List.of());
        var list = service.listConversations(1L);
        assertNotNull(list);
    }

    @Test
    void getConversation_bad_notFound_shouldThrow() {
        when(mapper.findConversationById(999L, 1L)).thenReturn(null);
        assertThrows(Exception.class, () -> service.getConversation(999L, 1L));
    }

    @Test
    void createConversation_happy_shouldSucceed() {
        var c = service.createConversation(1L, "test");
        assertNotNull(c);
    }
}
