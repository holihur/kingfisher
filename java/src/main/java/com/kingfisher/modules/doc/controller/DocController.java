package com.kingfisher.modules.doc.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.PageData;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.doc.domain.DocDirectory;
import com.kingfisher.modules.doc.domain.DocVersion;
import com.kingfisher.modules.doc.domain.Document;
import com.kingfisher.modules.doc.service.DocService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.util.List;
import java.util.Map;

@RestController
public class DocController {

    private final DocService docService;
    public DocController(DocService docService) { this.docService = docService; }

    private Long getUserId(HttpServletRequest req) {
        Long uid = (Long) req.getAttribute("user_id");
        if (uid != null) return uid;
        Object c = req.getAttribute("claims");
        if (c instanceof io.jsonwebtoken.Claims cl) return cl.get("user_id", Long.class);
        return 0L;
    }
    private boolean isAdmin(HttpServletRequest req) {
        List<String> roles = (List<String>) req.getAttribute("roles");
        if (roles == null) {
            Object c = req.getAttribute("claims");
            if (c instanceof io.jsonwebtoken.Claims cl) roles = cl.get("roles", List.class);
        }
        return roles != null && roles.contains("admin");
    }

    @GetMapping("/api/v1/public/docs/tree")
    public ResponseEntity<ApiResponse<List<Map<String,Object>>>> getPublicTree() {
        List<Map<String,Object>> tree = docService.tree(true, 0L);
        return ResponseEntity.ok(ApiResponse.ok(tree));
    }

    @GetMapping("/api/v1/public/docs/{id}")
    public ResponseEntity<ApiResponse<Document>> getPublicDoc(@PathVariable Long id) {
        Document doc = docService.getPublicDoc(id);
        if (doc == null) return ResponseEntity.status(404).body(ApiResponse.error(ErrorCode.ERR_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(doc));
    }

    @RequirePerm("doc:list")
    @GetMapping("/api/v1/docs/tree")
    public ResponseEntity<ApiResponse<List<Map<String,Object>>>> getTree(HttpServletRequest req) {
        List<Map<String,Object>> tree = docService.tree(isAdmin(req), getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok(tree));
    }

    @RequirePerm("doc:create")
    @PostMapping("/api/v1/docs/dirs")
    public ResponseEntity<ApiResponse<DocDirectory>> createDir(@RequestBody DocDirectory dir) {
        DocDirectory d = docService.createDir(dir);
        return ResponseEntity.ok(ApiResponse.ok(d));
    }

    @RequirePerm("doc:update")
    @PutMapping("/api/v1/docs/dirs/{id}")
    public ResponseEntity<ApiResponse<Void>> updateDir(@PathVariable Long id, @RequestBody Map<String,Object> body) {
        try { docService.updateDir(id, body); return ResponseEntity.ok(ApiResponse.ok()); }
        catch (IllegalArgumentException e) { return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage())); }
    }

    @RequirePerm("doc:delete")
    @DeleteMapping("/api/v1/docs/dirs/{id}")
    public ResponseEntity<ApiResponse<Void>> deleteDir(@PathVariable Long id) {
        try { docService.deleteDir(id); return ResponseEntity.ok(ApiResponse.ok()); }
        catch (IllegalArgumentException e) { return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage())); }
    }

    @RequirePerm("doc:list")
    @GetMapping("/api/v1/docs")
    public ResponseEntity<ApiResponse<PageData<List<Document>>>> listDocs(HttpServletRequest req,
            @RequestParam Long dirId,
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int pageSize,
            @RequestParam(required = false) String q,
            @RequestParam(required = false) String filter,
            @RequestParam(required = false) String sort) {
        List<Document> items = docService.listDocs(dirId, q, null, sort, (page-1)*pageSize, pageSize, isAdmin(req), getUserId(req));
        long total = docService.countDocs(dirId, q, null, isAdmin(req), getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok(new PageData<>(items, total, page, pageSize)));
    }

    @RequirePerm("doc:create")
    @PostMapping("/api/v1/docs")
    public ResponseEntity<ApiResponse<Document>> createDoc(HttpServletRequest req, @RequestBody Document doc) {
        doc.setOwnerId(getUserId(req));
        Document d = docService.createDoc(doc);
        return ResponseEntity.ok(ApiResponse.ok(d));
    }

    @RequirePerm("doc:update")
    @PostMapping("/api/v1/docs/upload")
    public ResponseEntity<ApiResponse<Map<String,String>>> uploadImage(@RequestParam("file") MultipartFile file) {
        try {
            String ext = "";
            String fn = file.getOriginalFilename();
            if (fn != null && fn.contains(".")) ext = fn.substring(fn.lastIndexOf("."));
            String name = System.nanoTime() + ext;
            File dir = new File("uploads/docs");
            if (!dir.exists()) dir.mkdirs();
            File dest = new File(dir, name);
            file.transferTo(dest);
            return ResponseEntity.ok(ApiResponse.ok(Map.of("url", "/uploads/docs/" + name)));
        } catch (Exception e) { return ResponseEntity.internalServerError().body(ApiResponse.error(ErrorCode.ERR_INTERNAL)); }
    }

    @RequirePerm("doc:list")
    @GetMapping("/api/v1/docs/{id}")
    public ResponseEntity<ApiResponse<Document>> getDoc(HttpServletRequest req, @PathVariable Long id) {
        Document doc = docService.getDoc(id, isAdmin(req), getUserId(req));
        if (doc == null) return ResponseEntity.status(404).body(ApiResponse.error(ErrorCode.ERR_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(doc));
    }

    @RequirePerm("doc:update")
    @PutMapping("/api/v1/docs/{id}")
    public ResponseEntity<ApiResponse<Void>> updateDoc(HttpServletRequest req, @PathVariable Long id, @RequestBody Map<String,Object> body) {
        try { docService.updateDoc(id, body, getUserId(req)); return ResponseEntity.ok(ApiResponse.ok()); }
        catch (IllegalArgumentException e) { return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage())); }
    }

    @RequirePerm("doc:update")
    @PutMapping("/api/v1/docs/{id}/publish")
    public ResponseEntity<ApiResponse<Void>> publish(@PathVariable Long id) { docService.publish(id); return ResponseEntity.ok(ApiResponse.ok()); }

    @RequirePerm("doc:update")
    @PutMapping("/api/v1/docs/{id}/unpublish")
    public ResponseEntity<ApiResponse<Void>> unpublish(@PathVariable Long id) { docService.unpublish(id); return ResponseEntity.ok(ApiResponse.ok()); }

    @RequirePerm("doc:list")
    @GetMapping("/api/v1/docs/{id}/versions")
    public ResponseEntity<ApiResponse<List<DocVersion>>> listVersions(@PathVariable Long id) {
        return ResponseEntity.ok(ApiResponse.ok(docService.listVersions(id)));
    }

    @RequirePerm("doc:list")
    @GetMapping("/api/v1/docs/{id}/versions/{no}")
    public ResponseEntity<ApiResponse<DocVersion>> getVersion(@PathVariable Long id, @PathVariable int no) {
        DocVersion v = docService.getVersion(id, no);
        if (v == null) return ResponseEntity.status(404).body(ApiResponse.error(ErrorCode.ERR_NOT_FOUND));
        return ResponseEntity.ok(ApiResponse.ok(v));
    }

    @RequirePerm("doc:update")
    @PostMapping("/api/v1/docs/{id}/restore")
    public ResponseEntity<ApiResponse<Void>> restore(HttpServletRequest req, @PathVariable Long id, @RequestBody Map<String,Object> body) {
        int versionNo = body.get("version_no") != null ? ((Number)body.get("version_no")).intValue() : ((Number)body.get("versionNo")).intValue();
        try { docService.restore(id, versionNo, getUserId(req)); return ResponseEntity.ok(ApiResponse.ok()); }
        catch (IllegalArgumentException e) { return ResponseEntity.badRequest().body(ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage())); }
    }

    @RequirePerm("doc:delete")
    @DeleteMapping("/api/v1/docs/{id}")
    public ResponseEntity<ApiResponse<Void>> deleteDoc(@PathVariable Long id) { docService.deleteDoc(id); return ResponseEntity.ok(ApiResponse.ok()); }
}
