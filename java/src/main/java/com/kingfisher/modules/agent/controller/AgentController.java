package com.kingfisher.modules.agent.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.agent.domain.Conversation;
import com.kingfisher.modules.agent.service.AgentService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/agent")
public class AgentController {

    private final AgentService agentService;
    public AgentController(AgentService agentService) { this.agentService = agentService; }

    private Long getUserId(HttpServletRequest request) {
        Long uid = (Long) request.getAttribute("user_id");
        if (uid != null) return uid;
        Object claims = request.getAttribute("claims");
        if (claims instanceof io.jsonwebtoken.Claims c) return c.get("user_id", Long.class);
        return null;
    }

    @RequirePerm("agent:list")
    @GetMapping("/conversations")
    public ResponseEntity<ApiResponse<List<Conversation>>> listConversations(HttpServletRequest request) {
        try {
            Long userId = getUserId(request);
            List<Conversation> list = agentService.listConversations(userId);
            return ResponseEntity.ok(ApiResponse.ok(list));
        } catch (AgentService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode())).body(ApiResponse.error(e.getCode(), e.getMessage()));
        }
    }

    @RequirePerm("agent:list")
    @PostMapping("/conversations")
    public ResponseEntity<ApiResponse<Conversation>> createConversation(HttpServletRequest request, @RequestBody Map<String,String> body) {
        try {
            Long userId = getUserId(request);
            String title = body.getOrDefault("title", "新会话");
            Conversation c = agentService.createConversation(userId, title);
            return ResponseEntity.ok(ApiResponse.ok(c));
        } catch (AgentService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode())).body(ApiResponse.error(e.getCode(), e.getMessage()));
        }
    }

    @RequirePerm("agent:list")
    @GetMapping("/conversations/{id}")
    public ResponseEntity<ApiResponse<Conversation>> getConversation(HttpServletRequest request, @PathVariable Long id) {
        try {
            Long userId = getUserId(request);
            Conversation c = agentService.getConversation(id, userId);
            return ResponseEntity.ok(ApiResponse.ok(c));
        } catch (AgentService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode())).body(ApiResponse.error(e.getCode(), e.getMessage()));
        }
    }

    @RequirePerm("agent:list")
    @GetMapping("/conversations/{id}/messages")
    public ResponseEntity<ApiResponse<List<com.kingfisher.modules.agent.domain.AgentMessage>>> listMessages(HttpServletRequest request, @PathVariable Long id) {
        try {
            Long userId = getUserId(request);
            var messages = agentService.listMessages(id, userId);
            return ResponseEntity.ok(ApiResponse.ok(messages));
        } catch (AgentService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode())).body(ApiResponse.error(e.getCode(), e.getMessage()));
        }
    }

    @RequirePerm("agent:list")
    @DeleteMapping("/conversations/{id}")
    public ResponseEntity<ApiResponse<Void>> deleteConversation(HttpServletRequest request, @PathVariable Long id) {
        try {
            Long userId = getUserId(request);
            agentService.deleteConversation(id, userId);
            return ResponseEntity.ok(ApiResponse.ok());
        } catch (AgentService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode())).body(ApiResponse.error(e.getCode(), e.getMessage()));
        }
    }

    @RequirePerm("agent:list")
    @PostMapping("/chat")
    public ResponseEntity<ApiResponse<Map<String,Object>>> chat(HttpServletRequest request, @RequestBody Map<String,Object> body) {
        try {
            Long userId = getUserId(request);
            Long conversationId = body.get("conversation_id") != null ? Long.valueOf(String.valueOf(body.get("conversation_id"))) : null;
            if (conversationId == null && body.get("conversationId") != null) conversationId = Long.valueOf(String.valueOf(body.get("conversationId")));
            String content = (String) body.get("content");
            if (content == null) content = (String) body.get("message");
            if (conversationId == null) {
                Conversation c = agentService.createConversation(userId, content != null && content.length() > 20 ? content.substring(0,20) : "新会话");
                conversationId = c.getId();
            }
            String token = request.getHeader("Authorization");
            if (token != null && token.startsWith("Bearer ")) token = token.substring(7);
            Map<String,Object> reply = agentService.chat(conversationId, userId, content, token);
            return ResponseEntity.ok(ApiResponse.ok(reply));
        } catch (AgentService.BizException e) {
            return ResponseEntity.status(ErrorCode.httpStatus(e.getCode())).body(ApiResponse.error(e.getCode(), e.getMessage()));
        } catch (Exception e) {
            return ResponseEntity.internalServerError().body(ApiResponse.error(ErrorCode.ERR_INTERNAL, e.getMessage()));
        }
    }
}
