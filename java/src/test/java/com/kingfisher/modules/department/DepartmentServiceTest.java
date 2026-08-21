package com.kingfisher.modules.department;

import com.kingfisher.modules.department.service.DepartmentService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class DepartmentServiceTest {

    @Autowired DepartmentService service;

    @Test
    void tree_happy_shouldBuildSeed() {
        var d = new com.kingfisher.modules.department.domain.Department();
        d.setName("TreeTest_" + System.nanoTime());
        d.setParentId(0L);
        service.create(d);
        var tree = service.tree();
        assertNotNull(tree);
        assertTrue(tree.size() >= 1);
    }

    @Test
    void getById_bad_notFound_shouldReturnNull() {
        assertNull(service.getById(99999L));
    }

    @Test
    void create_happy_shouldSucceed() {
        var d = new com.kingfisher.modules.department.domain.Department();
        d.setName("Dept_" + System.nanoTime());
        d.setParentId(0L);
        var result = service.create(d);
        assertNotNull(result);
        assertNotNull(result.getId());
    }

    @Test
    void create_bad_parentNotFound_shouldThrow() {
        var d = new com.kingfisher.modules.department.domain.Department();
        d.setName("X");
        d.setParentId(99999L);
        assertThrows(IllegalArgumentException.class, () -> service.create(d));
    }

    @Test
    void delete_bad_hasChildren_shouldThrow() {
        var parent = new com.kingfisher.modules.department.domain.Department();
        parent.setName("Parent_" + System.nanoTime());
        parent.setParentId(0L);
        var createdParent = service.create(parent);
        var child = new com.kingfisher.modules.department.domain.Department();
        child.setName("Child_" + System.nanoTime());
        child.setParentId(createdParent.getId());
        service.create(child);
        Long pid = createdParent.getId();
        assertThrows(IllegalArgumentException.class, () -> service.delete(pid));
    }

    @Test
    void getById_happy_shouldReturnSeed() {
        var d = new com.kingfisher.modules.department.domain.Department();
        d.setName("GetTest_" + System.nanoTime());
        d.setParentId(0L);
        var created = service.create(d);
        assertNotNull(service.getById(created.getId()));
    }
}
