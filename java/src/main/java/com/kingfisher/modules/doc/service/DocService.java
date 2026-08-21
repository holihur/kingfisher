package com.kingfisher.modules.doc.service;

import com.kingfisher.modules.doc.domain.DocDirectory;
import com.kingfisher.modules.doc.domain.Document;
import com.kingfisher.modules.doc.domain.DocVersion;
import com.kingfisher.modules.doc.mapper.DocMapper;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Service
public class DocService {
    private final DocMapper mapper;
    public DocService(DocMapper mapper) { this.mapper = mapper; }

    public List<DocDirectory> listDirs() { return mapper.findAllDirs(); }

    public List<Map<String,Object>> tree(boolean isAdmin, Long userId) {
        List<DocDirectory> dirs = mapper.findAllDirs();
        List<Document> docs = isAdmin ? mapper.listAllPublicDocs() : mapper.listAllVisibleDocs(userId, isAdmin);
        Map<Long, List<Document>> docsByDir = docs.stream().collect(Collectors.groupingBy(Document::getDirId));
        return buildTree(0L, dirs, docsByDir);
    }

    private List<Map<String,Object>> buildTree(Long parentId, List<DocDirectory> allDirs, Map<Long, List<Document>> docsByDir) {
        List<Map<String,Object>> res = new ArrayList<>();
        for (DocDirectory d : allDirs) {
            if (Objects.equals(d.getParentId(), parentId)) {
                Map<String,Object> node = new HashMap<>();
                node.put("id", d.getId());
                node.put("name", d.getName());
                node.put("sort", d.getSort());
                node.put("status", d.getStatus());
                node.put("visibility", d.getVisibility());
                node.put("children", buildTree(d.getId(), allDirs, docsByDir));
                List<Document> ds = docsByDir.getOrDefault(d.getId(), List.of());
                if (!ds.isEmpty()) {
                    List<Map<String, Object>> docNodes = new ArrayList<>();
                    for (Document doc : ds) {
                        Map<String,Object> docNode = new HashMap<>();
                        docNode.put("id", doc.getId());
                        docNode.put("title", doc.getTitle());
                        docNode.put("status", doc.getStatus());
                        docNode.put("visibility", doc.getVisibility());
                        docNodes.add(docNode);
                    }
                    node.put("docs", docNodes);
                }
                res.add(node);
            }
        }
        return res;
    }

    public DocDirectory createDir(DocDirectory dir) { mapper.insertDir(dir); return dir; }
    public void updateDir(Long id, Map<String,Object> updates) { mapper.updateDir(id, updates); }
    public void deleteDir(Long id) {
        if (mapper.countDirChildren(id) > 0) throw new IllegalArgumentException("目录下有子目录，不可删除");
        if (mapper.countDirDocuments(id) > 0) throw new IllegalArgumentException("目录下存在文档，不可删除");
        mapper.deleteDir(id);
    }

    public List<Document> listDocs(Long dirId, String keyword, List<com.kingfisher.common.query.Condition> filters, String sort, int offset, int limit, boolean isAdmin, Long userId) {
        List<String> searchable = List.of("title", "content");
        return mapper.listDocs(dirId, keyword, searchable, filters, sort, offset, limit, userId, isAdmin);
    }
    public long countDocs(Long dirId, String keyword, List<com.kingfisher.common.query.Condition> filters, boolean isAdmin, Long userId) {
        List<String> searchable = List.of("title", "content");
        return mapper.countDocs(dirId, keyword, searchable, filters, userId, isAdmin);
    }

    public Document getDoc(Long id, boolean isAdmin, Long userId) { return mapper.getDocById(id, userId, isAdmin); }
    public Document getPublicDoc(Long id) { return mapper.getPublicDoc(id); }

    @Transactional
    public Document createDoc(Document doc) {
        mapper.insertDoc(doc);
        DocVersion v = new DocVersion();
        v.setDocId(doc.getId()); v.setVersionNo(1); v.setTitle(doc.getTitle()); v.setContent(doc.getContent()); v.setOwnerId(doc.getOwnerId());
        mapper.insertVersion(v);
        return doc;
    }

    @Transactional
    public void updateDoc(Long id, Map<String,Object> updates, Long userId) {
        Document doc = mapper.getDocById(id, userId, true);
        if (doc == null) throw new IllegalArgumentException("文档不存在");
        int nextVer = doc.getCurrentVersion() + 1;
        if (updates.containsKey("content") || updates.containsKey("title")) {
            DocVersion v = new DocVersion();
            v.setDocId(id); v.setVersionNo(nextVer); v.setTitle((String) updates.getOrDefault("title", doc.getTitle())); v.setContent((String) updates.getOrDefault("content", doc.getContent())); v.setOwnerId(userId);
            mapper.insertVersion(v);
            updates.put("current_version", nextVer);
        }
        mapper.updateDoc(id, updates);
    }

    public void publish(Long id) { mapper.setDocStatus(id, "published", java.time.LocalDateTime.now()); }
    public void unpublish(Long id) { mapper.setDocStatus(id, "draft", null); }
    public void deleteDoc(Long id) { mapper.deleteDoc(id); mapper.deleteDocVersions(id); }
    public List<DocVersion> listVersions(Long docId) { return mapper.listVersions(docId); }
    public DocVersion getVersion(Long docId, int versionNo) { return mapper.getVersion(docId, versionNo); }
    @Transactional
    public void restore(Long docId, int versionNo, Long userId) {
        DocVersion v = mapper.getVersion(docId, versionNo);
        if (v == null) throw new IllegalArgumentException("版本不存在");
        Map<String,Object> updates = Map.of("title", v.getTitle(), "content", v.getContent());
        updateDoc(docId, updates, userId);
    }
}
