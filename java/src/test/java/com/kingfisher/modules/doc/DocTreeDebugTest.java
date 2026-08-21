package com.kingfisher.modules.doc;

import com.kingfisher.modules.doc.service.DocService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;

@SpringBootTest
@ActiveProfiles("test")
class DocTreeDebugTest {

    @Autowired DocService docService;

    @Test
    void printTree() {
        var treeAdmin = docService.tree(true, 1L);
        System.out.println("Admin tree: " + treeAdmin);
        var treeViewer = docService.tree(false, 3L);
        System.out.println("Viewer tree: " + treeViewer);
        var docs = docService.listDocs(1L, null, null, null, 0, 10, true, 1L);
        System.out.println("Docs in dir 1: " + docs);
    }
}
