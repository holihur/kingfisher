package com.kingfisher.modules.template.service;

import com.kingfisher.common.PageData;
import com.kingfisher.common.query.*;
import com.kingfisher.modules.template.domain.Template;
import com.kingfisher.modules.template.mapper.TemplateMapper;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 模版服务，与 Go extends/template/app.TemplateService 对齐。
 */
@Service
public class TemplateService {

    private static final Defs TEMPLATE_DEFS = Defs.of(
        "name", Field.searchableString("name"),
        "code", Field.searchableString("code"),
        "remark", Field.searchableString("remark"),
        "template_type", Field.filterable(FieldType.STRING, "template_type"),
        "status", Field.filterable(FieldType.INT, "status"),
        "created_at", Field.filterable(FieldType.TIME, "created_at")
    );

    private final TemplateMapper templateMapper;

    public TemplateService(TemplateMapper templateMapper) {
        this.templateMapper = templateMapper;
    }

    public PageData<List<Template>> list(HttpServletRequest request) {
        Query q = Query.parse(request, TEMPLATE_DEFS);
        List<String> searchable = q.searchableColumns();
        List<Template> items = templateMapper.findAll(
            q.getKeyword(), searchable, q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
        long total = templateMapper.countAll(q.getKeyword(), searchable, q.getFilters());
        return new PageData<>(items, total, q.getPage(), q.getPageSize());
    }

    public Template getById(Long id) {
        return templateMapper.findById(id);
    }

    public Template create(String name, String code, String templateType, String title,
                           String content, int status, String remark, String version) {
        if (templateMapper.findByCode(code) != null) {
            throw new IllegalArgumentException("模板编码已存在");
        }
        Template t = new Template();
        t.setName(name);
        t.setCode(code);
        t.setTemplateType(templateType != null ? templateType : "general");
        t.setTitle(title);
        t.setContent(content);
        t.setStatus(status);
        t.setRemark(remark);
        t.setVersion(version);
        templateMapper.insert(t);
        return t;
    }

    public void update(Long id, String name, String code, String templateType, String title,
                       String content, int status, String remark, String version) {
        Template existing = templateMapper.findByCode(code);
        if (existing != null && !existing.getId().equals(id)) {
            throw new IllegalArgumentException("模板编码已存在");
        }
        Map<String, Object> updates = new HashMap<>();
        updates.put("name", name);
        updates.put("code", code);
        updates.put("template_type", templateType != null ? templateType : "general");
        updates.put("title", title);
        updates.put("content", content);
        updates.put("status", status);
        updates.put("remark", remark);
        updates.put("version", version);
        templateMapper.update(id, updates);
    }

    public void delete(Long id) {
        templateMapper.deleteById(id);
    }

    public void batchDelete(List<Long> ids) {
        templateMapper.deleteBatch(ids);
    }

    public void batchUpdateStatus(List<Long> ids, int status) {
        templateMapper.updateStatusBatch(ids, status);
    }
}
