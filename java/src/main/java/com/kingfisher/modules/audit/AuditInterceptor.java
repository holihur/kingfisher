package com.kingfisher.modules.audit;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kingfisher.modules.audit.domain.AuditLog;
import com.kingfisher.modules.audit.service.AuditService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;
import org.springframework.web.util.ContentCachingRequestWrapper;

import java.util.HashMap;
import java.util.Map;

@Component
public class AuditInterceptor implements HandlerInterceptor {

    private final AuditService auditService;
    private final ObjectMapper objectMapper;

    public AuditInterceptor(AuditService auditService, ObjectMapper objectMapper) {
        this.auditService = auditService;
        this.objectMapper = objectMapper;
    }

    private static final Map<String, String> METHOD_ACTION = Map.of(
            "POST", "create",
            "PUT", "update",
            "PATCH", "update",
            "DELETE", "delete"
    );

    private static final Map<String, String> RESOURCE_LABELS = Map.ofEntries(
            Map.entry("users", "用户"),
            Map.entry("roles", "角色"),
            Map.entry("permissions", "权限"),
            Map.entry("menus", "菜单"),
            Map.entry("configs", "系统配置"),
            Map.entry("config-groups", "配置分组"),
            Map.entry("dict-types", "字典类型"),
            Map.entry("dict-entries", "字典条目"),
            Map.entry("messages", "站内信"),
            Map.entry("templates", "消息模板"),
            Map.entry("scheduled-tasks", "周期任务"),
            Map.entry("audit-logs", "审计日志"),
            Map.entry("public", "公开配置"),
            Map.entry("docs", "文档"),
            Map.entry("departments", "部门"),
            Map.entry("tasks", "业务任务"),
            Map.entry("agent", "Agent")
    );

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        request.setAttribute("_auditStart", System.currentTimeMillis());
        return true;
    }

    @Override
    public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) {
        String method = request.getMethod();
        if ("GET".equals(method) || "HEAD".equals(method) || "OPTIONS".equals(method)) return;
        Object uidObj = request.getAttribute("user_id");
        if (uidObj == null) return;
        Long userId;
        try {
            userId = uidObj instanceof Long ? (Long) uidObj : Long.valueOf(String.valueOf(uidObj));
        } catch (Exception e) {
            return;
        }
        String username = (String) request.getAttribute("username");
        if (username == null) username = "";

        int status = response.getStatus();
        String result = (status >= 200 && status < 300) ? "success" : "failure";
        String message = "";
        if (!"success".equals(result)) {
            if (status == 403) message = "权限不足";
            else if (status == 404) message = "资源不存在";
            else if (status == 401) message = "未认证";
            else if (status == 400) message = "参数错误";
            else message = "HTTP " + status;
            if (ex != null && ex.getMessage() != null) message = ex.getMessage();
        }

        String action = METHOD_ACTION.getOrDefault(method, method.toLowerCase());
        String[] res = extractResource(request);
        String resource = res[0];
        Long resourceId = null;
        try {
            if (res[1] != null) resourceId = Long.valueOf(res[1]);
        } catch (Exception ignored) {}

        String detail = buildDetail(request);
        Object diff = request.getAttribute("audit_diff");
        if (diff instanceof String s && !s.isBlank()) detail = s;

        long latency = 0;
        Object startObj = request.getAttribute("_auditStart");
        if (startObj instanceof Long l) latency = System.currentTimeMillis() - l;

        AuditLog log = new AuditLog();
        log.setUserId(userId);
        log.setUsername(username);
        log.setAction(action);
        log.setResource(resource);
        log.setResourceId(resourceId);
        log.setDetail(detail);
        log.setResult(result);
        log.setLatency(latency);
        log.setMessage(message);
        log.setIp(getClientIp(request));
        log.setUserAgent(request.getHeader("User-Agent"));
        try {
            auditService.log(log);
        } catch (Exception ignored) {
        }
    }

    private String[] extractResource(HttpServletRequest request) {
        String path = request.getRequestURI();
        String[] parts = path.split("/");
        java.util.List<String> clean = new java.util.ArrayList<>();
        for (String p : parts) {
            if (p.isEmpty() || "api".equals(p) || "v1".equals(p)) continue;
            clean.add(p);
        }
        String resourceKey = "unknown";
        for (int i = clean.size() - 1; i >= 0; i--) {
            String seg = clean.get(i);
            if (!seg.startsWith(":")) {
                try {
                    Long.parseLong(seg);
                    continue;
                } catch (NumberFormatException ignored) {}
                resourceKey = seg;
                break;
            }
        }
        String resource = RESOURCE_LABELS.getOrDefault(resourceKey, resourceKey);
        String idStr = null;
        @SuppressWarnings("unchecked")
        Map<String, String> vars = (Map<String, String>) request.getAttribute("org.springframework.web.servlet.HandlerMapping.uriTemplateVariables");
        if (vars != null) {
            if (vars.containsKey("entryId")) idStr = vars.get("entryId");
            else if (vars.containsKey("id")) idStr = vars.get("id");
            else if (!vars.isEmpty()) idStr = vars.values().iterator().next();
        }
        if (idStr == null) {
            for (int i = clean.size() - 1; i >= 0; i--) {
                try {
                    Long.parseLong(clean.get(i));
                    idStr = clean.get(i);
                    break;
                } catch (NumberFormatException ignored) {}
            }
        }
        return new String[]{resource, idStr};
    }

    private String buildDetail(HttpServletRequest request) {
        try {
            Map<String, Object> map = new HashMap<>();
            Map<String, String[]> params = request.getParameterMap();
            for (Map.Entry<String, String[]> e : params.entrySet()) {
                String k = e.getKey();
                if (isSensitive(k)) continue;
                String[] v = e.getValue();
                map.put(k, v.length == 1 ? v[0] : String.join(",", v));
            }
            if (request instanceof ContentCachingRequestWrapper wrapper) {
                byte[] body = wrapper.getContentAsByteArray();
                if (body.length > 0 && body.length < 8192) {
                    String s = new String(body, wrapper.getCharacterEncoding());
                    try {
                        @SuppressWarnings("unchecked")
                        Map<String, Object> bodyMap = objectMapper.readValue(s, Map.class);
                        for (String k : new String[]{"password","new_password","old_password","refresh_token","access_token","secret"}) {
                            bodyMap.remove(k);
                        }
                        map.putAll(bodyMap);
                    } catch (Exception ignored) {
                    }
                }
            }
            if (map.isEmpty()) return "";
            String json = objectMapper.writeValueAsString(map);
            if (json.length() > 8192) return json.substring(0, 8192);
            return json;
        } catch (Exception e) {
            return "";
        }
    }

    private boolean isSensitive(String key) {
        String k = key.toLowerCase();
        return k.contains("password") || k.contains("secret") || k.contains("token");
    }

    private String getClientIp(HttpServletRequest request) {
        String ip = request.getHeader("X-Forwarded-For");
        if (ip != null && !ip.isBlank()) return ip.split(",")[0].trim();
        ip = request.getHeader("X-Real-IP");
        if (ip != null && !ip.isBlank()) return ip;
        return request.getRemoteAddr();
    }
}
