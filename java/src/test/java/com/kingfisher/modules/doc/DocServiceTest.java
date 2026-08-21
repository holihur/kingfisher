package com.kingfisher.modules.doc;

import com.kingfisher.modules.doc.service.DocService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class DocServiceTest {

    @Autowired DocService service;

    @Test
    void tree_happy_shouldReturnSeed() {
        var tree = service.tree(true, 1L);
        assertNotNull(tree);
    }

    @Test
    void getDoc_bad_notFound_shouldReturnNull() {
        assertNull(service.getDoc(99999L, true, 1L));
    }

    @Test
    void createDir_happy_shouldSucceed() {
        var dir = new com.kingfisher.modules.doc.domain.DocDirectory();
        dir.setName("test_" + System.nanoTime());
        dir.setParentId(0L);
        var result = service.createDir(dir);
        assertNotNull(result);
        assertNotNull(result.getId());
    }

    @Test
    void deleteDir_bad_hasChildren_shouldThrow() {
        assertThrows(IllegalArgumentException.class, () -> service.deleteDir(1L));
    }

    @Test
    void getDoc_happy_shouldReturnSeed() {
        var doc = new com.kingfisher.modules.doc.domain.Document();
        doc.setDirId(1L);
        doc.setTitle("TestDoc_" + System.nanoTime());
        doc.setContent("hello");
        doc.setOwnerId(1L);
        var created = service.createDoc(doc);
        assertNotNull(service.getDoc(created.getId(), true, 1L));
    }
}
