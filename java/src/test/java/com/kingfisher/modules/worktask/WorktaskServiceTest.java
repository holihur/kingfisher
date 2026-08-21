package com.kingfisher.modules.worktask;

import com.kingfisher.common.query.Query;
import com.kingfisher.modules.worktask.mapper.WorkTaskMapper;
import com.kingfisher.modules.worktask.service.WorktaskService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

class WorktaskServiceTest {

    private WorkTaskMapper mapper;
    private WorktaskService service;

    @BeforeEach
    void setUp() {
        mapper = mock(WorkTaskMapper.class);
        service = new WorktaskService(mapper);
    }

    @Test
    void create_happy_shouldSucceed() {
        var task = service.create("title", "desc", 1L, 1L, "pending");
        assertNotNull(task);
        assertEquals("title", task.getTitle());
        verify(mapper).insert(any());
    }

    @Test
    void create_bad_nullTitle_shouldStillCreate() {
        var task = service.create(null, "desc", 1L, 1L, "pending");
        assertNotNull(task);
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        when(mapper.findById(999L, 1L, true)).thenReturn(null);
        assertNull(service.getById(999L, 1L, true));
    }

    @Test
    void update_happy_shouldCallMapper() {
        assertDoesNotThrow(() -> service.update(1L, Map.of("title", "new")));
        verify(mapper).update(eq(1L), any());
    }
}
