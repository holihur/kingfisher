package com.kingfisher.modules.dict.service;

import com.kingfisher.common.query.Query;
import com.kingfisher.modules.dict.domain.DictType;
import com.kingfisher.modules.dict.mapper.DictTypeMapper;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.util.List;

@Service
public class DictTypeService {
    private final DictTypeMapper mapper;
    public DictTypeService(DictTypeMapper mapper) { this.mapper = mapper; }

    public List<DictType> list(Query q) {
        return mapper.findAll(q.getKeyword(), q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
    }
    public long count(Query q) { return mapper.countAll(q.getKeyword(), q.getFilters()); }
    public DictType getById(Long id) { return mapper.findById(id); }

    public DictType create(String code, String name, boolean isPublic, int status, String remark, String version) {
        if (mapper.findByCode(code) != null) throw new IllegalArgumentException("编码已存在");
        DictType t = new DictType();
        t.setCode(code); t.setName(name); t.setIsPublic(isPublic); t.setStatus(status); t.setRemark(remark); t.setVersion(version);
        mapper.insert(t);
        return t;
    }

    public void update(Long id, String code, String name, boolean isPublic, int status, String remark, String version) {
        DictType existing = mapper.findByCode(code);
        if (existing != null && !existing.getId().equals(id)) throw new IllegalArgumentException("编码已存在");
        DictType t = new DictType();
        t.setId(id); t.setCode(code); t.setName(name); t.setIsPublic(isPublic); t.setStatus(status); t.setRemark(remark); t.setVersion(version);
        mapper.update(t);
    }

    public void delete(Long id) {
        if (mapper.countEntriesByTypeId(id) > 0) throw new IllegalArgumentException("存在条目不可删除");
        mapper.deleteById(id);
    }

    @Transactional
    public void batchDelete(List<Long> ids) {
        for (Long id : ids) if (mapper.countEntriesByTypeId(id) > 0) throw new IllegalArgumentException("存在条目不可删除");
        mapper.deleteBatch(ids);
    }

    public void batchUpdateStatus(List<Long> ids, int status) { mapper.updateStatusBatch(ids, status); }
}
