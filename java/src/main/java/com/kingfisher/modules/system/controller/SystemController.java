package com.kingfisher.modules.system.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.system.domain.SystemInfo;
import com.kingfisher.modules.system.service.SystemService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 系统信息控制器，与 Go extends/system/transport 路由 1:1 对齐。
 */
@RestController
@RequestMapping("/api/v1/system")
public class SystemController {

    private final SystemService systemService;

    public SystemController(SystemService systemService) {
        this.systemService = systemService;
    }

    /**
     * GET /api/v1/system/info — 系统信息（需 system:list 权限）
     */
    @RequirePerm("system:list")
    @GetMapping("/info")
    public ApiResponse<SystemInfo> getInfo() {
        return ApiResponse.ok(systemService.getInfo());
    }
}
