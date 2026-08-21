package com.kingfisher.modules.audit.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.PageData;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.audit.domain.AuditLog;
import com.kingfisher.modules.audit.service.AuditService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

/**
 * 审计日志控制器，与 Go extends/audit/transport 路由 1:1 对齐。
 */
@RestController
@RequestMapping("/api/v1/audit-logs")
public class AuditController {

    private final AuditService auditService;

    public AuditController(AuditService auditService) {
        this.auditService = auditService;
    }

    /**
     * GET /api/v1/audit-logs — 审计日志列表（需 audit:list 权限）
     */
    @RequirePerm("audit:list")
    @GetMapping
    public ApiResponse<PageData<List<AuditLog>>> list(HttpServletRequest request) {
        return ApiResponse.ok(auditService.list(request));
    }
}
