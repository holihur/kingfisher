package com.kingfisher.modules.menu;

import com.kingfisher.modules.menu.service.MenuService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class MenuServiceTest {

    @Autowired MenuService service;

    @Test
    void getTree_happy_shouldReturnSeed() {
        var tree = service.getTree();
        assertNotNull(tree);
        assertTrue(tree.size() >= 1);
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L));
    }

    @Test
    void create_happy_shouldSucceed() {
        var m = new com.kingfisher.modules.menu.domain.Menu();
        m.setName("TestMenu" + System.nanoTime());
        m.setParentId(0L);
        m.setPath("/test" + System.nanoTime());
        m.setType(1);
        var result = service.create(m);
        assertNotNull(result);
        assertNotNull(result.getId());
    }

    @Test
    void delete_bad_hasChildren_shouldThrow() {
        var parent = new com.kingfisher.modules.menu.domain.Menu();
        parent.setName("ParentDel_" + System.nanoTime());
        parent.setParentId(0L);
        parent.setPath("/parentDel" + System.nanoTime());
        parent.setType(1);
        var createdParent = service.create(parent);
        var child = new com.kingfisher.modules.menu.domain.Menu();
        child.setName("ChildDel_" + System.nanoTime());
        child.setParentId(createdParent.getId());
        child.setPath("/childDel" + System.nanoTime());
        child.setType(2);
        service.create(child);
        assertThrows(IllegalArgumentException.class, () -> service.delete(createdParent.getId()));
    }

    @Test
    void getById_happy_shouldReturnSeed() {
        assertNotNull(service.getById(1L));
    }
}
