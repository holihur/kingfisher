package com.kingfisher.modules.message.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.PageData;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.common.query.Defs;
import com.kingfisher.common.query.Field;
import com.kingfisher.common.query.FieldType;
import com.kingfisher.common.query.Query;
import com.kingfisher.modules.message.domain.Message;
import com.kingfisher.modules.message.service.MessageService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
public class MessageController {

    private final MessageService messageService;
    public MessageController(MessageService messageService) { this.messageService = messageService; }

    private Long getUserId(HttpServletRequest req) {
        Long uid = (Long) req.getAttribute("user_id");
        if (uid != null) return uid;
        Object c = req.getAttribute("claims");
        if (c instanceof io.jsonwebtoken.Claims cl) return cl.get("user_id", Long.class);
        return 0L;
    }

    private static final Defs MSG_DEFS = Defs.of(
            "title", Field.searchableString("title"),
            "is_read", Field.filterable(FieldType.BOOL, "is_read"),
            "status", Field.filterable(FieldType.STRING, "status"),
            "created_at", Field.filterable(FieldType.TIME, "created_at")
    );

    @GetMapping("/api/v1/me/messages")
    public ResponseEntity<ApiResponse<PageData<List<Message>>>> listInbox(HttpServletRequest req) {
        Long uid = getUserId(req);
        Query q = Query.parse(req, MSG_DEFS);
        List<Message> items = messageService.listByRecipient(uid, q);
        long total = messageService.countByRecipient(uid, q);
        return ResponseEntity.ok(ApiResponse.ok(new PageData<>(items, total, q.getPage(), q.getPageSize())));
    }

    @GetMapping("/api/v1/me/messages/unread-count")
    public ResponseEntity<ApiResponse<Map<String,Object>>> unreadCount(HttpServletRequest req) {
        long cnt = messageService.unreadCount(getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok(Map.of("unread_count", cnt)));
    }

    @GetMapping("/api/v1/me/messages/{id}")
    public ResponseEntity<ApiResponse<Message>> getById(HttpServletRequest req, @PathVariable Long id) {
        Message m = messageService.getById(id, getUserId(req));
        if (m == null) return ResponseEntity.status(404).body(ApiResponse.error(10005));
        return ResponseEntity.ok(ApiResponse.ok(m));
    }

    @PutMapping("/api/v1/me/messages/{id}/read")
    public ResponseEntity<ApiResponse<Void>> markRead(HttpServletRequest req, @PathVariable Long id) {
        messageService.markRead(id, getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @PostMapping("/api/v1/me/messages/batch-delete")
    public ResponseEntity<ApiResponse<Void>> batchDelete(HttpServletRequest req, @RequestBody Map<String, List<Long>> body) {
        List<Long> ids = body.get("ids");
        messageService.deleteBatch(ids, getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("message:list")
    @GetMapping("/api/v1/messages")
    public ResponseEntity<ApiResponse<PageData<List<Message>>>> listSent(HttpServletRequest req) {
        Long uid = getUserId(req);
        Query q = Query.parse(req, MSG_DEFS);
        List<Message> items = messageService.listSent(uid, q);
        long total = messageService.countSent(uid, q);
        return ResponseEntity.ok(ApiResponse.ok(new PageData<>(items, total, q.getPage(), q.getPageSize())));
    }

    @RequirePerm("message:create")
    @PostMapping("/api/v1/messages")
    public ResponseEntity<ApiResponse<Map<String,Object>>> send(HttpServletRequest req, @RequestBody Map<String,Object> body) {
        Long uid = getUserId(req);
        String title = (String) body.get("title");
        String content = (String) body.getOrDefault("content", "");
        Object recips = body.get("recipient_ids");
        if (recips == null) recips = body.get("recipient_id");
        List<Long> ids;
        if (recips instanceof List) {
            ids = ((List<?>) recips).stream().map(v -> Long.valueOf(String.valueOf(v))).toList();
        } else if (recips != null) {
            ids = List.of(Long.valueOf(String.valueOf(recips)));
        } else ids = List.of();
        if (ids.isEmpty()) return ResponseEntity.badRequest().body(ApiResponse.error(10001, "please provide recipient_ids"));
        int n = messageService.sendBatch(uid, "admin", ids, title, content);
        return ResponseEntity.ok(ApiResponse.ok(Map.of("enqueued", true, "recipients", n)));
    }

    @RequirePerm("message:list")
    @GetMapping("/api/v1/messages/batch/{batchId}")
    public ResponseEntity<ApiResponse<List<Message>>> listBatch(HttpServletRequest req, @PathVariable Long batchId) {
        List<Message> msgs = messageService.listBatchMessages(batchId, getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok(msgs));
    }

    @RequirePerm("message:update")
    @PutMapping("/api/v1/messages/batch/{batchId}/revoke")
    public ResponseEntity<ApiResponse<Void>> revokeBatch(HttpServletRequest req, @PathVariable Long batchId) {
        messageService.revokeBatch(batchId, getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok());
    }

    @RequirePerm("message:update")
    @PutMapping("/api/v1/messages/{id}/revoke")
    public ResponseEntity<ApiResponse<Void>> revoke(HttpServletRequest req, @PathVariable Long id) {
        messageService.revoke(id, getUserId(req));
        return ResponseEntity.ok(ApiResponse.ok());
    }
}
