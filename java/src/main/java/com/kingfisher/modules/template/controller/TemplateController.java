package com.kingfisher.modules.template.controller;

import com.kingfisher.common.ApiResponse;
import com.kingfisher.common.ErrorCode;
import com.kingfisher.common.PageData;
import com.kingfisher.common.RequirePerm;
import com.kingfisher.modules.template.domain.Template;
import com.kingfisher.modules.template.service.TemplateService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * 模版控制器，与 Go extends/template/transport 路由 1:1 对齐。
 */
@RestController
@RequestMapping("/api/v1/templates")
public class TemplateController {

    private final TemplateService templateService;

    public TemplateController(TemplateService templateService) {
        this.templateService = templateService;
    }

    @RequirePerm("template:list")
    @GetMapping
    public ApiResponse<PageData<List<Template>>> list(HttpServletRequest request) {
        return ApiResponse.ok(templateService.list(request));
    }

    @RequirePerm("template:list")
    @GetMapping("/{id}")
    public ApiResponse<Template> getById(@PathVariable Long id) {
        Template t = templateService.getById(id);
        if (t == null) return ApiResponse.error(ErrorCode.ERR_NOT_FOUND);
        return ApiResponse.ok(t);
    }

    @RequirePerm("template:create")
    @PostMapping
    public ApiResponse<Template> create(@RequestBody TemplateReq req) {
        try {
            Template t = templateService.create(req.getName(), req.getCode(), req.getTemplateType(),
                req.getTitle(), req.getContent(), req.getStatus() != null ? req.getStatus() : 1,
                req.getRemark(), req.getVersion());
            return ApiResponse.ok(t);
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage());
        }
    }

    @RequirePerm("template:update")
    @PutMapping("/{id}")
    public ApiResponse<Void> update(@PathVariable Long id, @RequestBody TemplateReq req) {
        try {
            templateService.update(id, req.getName(), req.getCode(), req.getTemplateType(),
                req.getTitle(), req.getContent(), req.getStatus() != null ? req.getStatus() : 1,
                req.getRemark(), req.getVersion());
            return ApiResponse.ok();
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.ERR_INVALID_PARAM, e.getMessage());
        }
    }

    @RequirePerm("template:delete")
    @DeleteMapping("/{id}")
    public ApiResponse<Void> delete(@PathVariable Long id) {
        templateService.delete(id);
        return ApiResponse.ok();
    }

    @RequirePerm("template:delete")
    @PostMapping("/batch-delete")
    public ApiResponse<Void> batchDelete(@RequestBody BatchIDsRequest req) {
        templateService.batchDelete(req.getIds());
        return ApiResponse.ok();
    }

    @RequirePerm("template:update")
    @PostMapping("/batch-status")
    public ApiResponse<Void> batchUpdateStatus(@RequestBody BatchStatusRequest req) {
        templateService.batchUpdateStatus(req.getIds(), req.getStatus());
        return ApiResponse.ok();
    }

    public static class TemplateReq {
        private String name;
        private String code;
        private String templateType;
        private String title;
        private String content;
        private Integer status;
        private String remark;
        private String version;
        public String getName() { return name; }
        public String getCode() { return code; }
        public String getTemplateType() { return templateType; }
        public String getTitle() { return title; }
        public String getContent() { return content; }
        public Integer getStatus() { return status; }
        public String getRemark() { return remark; }
        public String getVersion() { return version; }
    }

    public static class BatchIDsRequest {
        private List<Long> ids;
        public List<Long> getIds() { return ids; }
    }

    public static class BatchStatusRequest {
        private List<Long> ids;
        private Integer status;
        public List<Long> getIds() { return ids; }
        public Integer getStatus() { return status; }
    }
}
