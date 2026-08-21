package com.kingfisher.modules.dict.service;

import com.kingfisher.common.query.Query;
import com.kingfisher.modules.dict.domain.DictEntry;
import com.kingfisher.modules.dict.domain.DictType;
import com.kingfisher.modules.dict.mapper.DictEntryMapper;
import com.kingfisher.modules.dict.mapper.DictTypeMapper;
import org.springframework.stereotype.Service;
import java.util.List;

@Service
public class DictEntryService {
    private final DictEntryMapper entryMapper;
    private final DictTypeMapper typeMapper;
    public DictEntryService(DictEntryMapper entryMapper, DictTypeMapper typeMapper) {
        this.entryMapper = entryMapper; this.typeMapper = typeMapper;
    }

    public List<DictEntry> listByTypeId(Long typeId, Query q) {
        return entryMapper.findByTypeId(typeId, q.getKeyword(), q.getFilters(), q.sortExpr(), q.offset(), q.getPageSize());
    }
    public long countByTypeId(Long typeId, Query q) { return entryMapper.countByTypeId(typeId, q.getKeyword(), q.getFilters()); }

    public List<DictEntry> listPublicByCode(String code) {
        DictType t = typeMapper.findByCode(code);
        if (t == null) throw new IllegalArgumentException("类型不存在");
        if (t.getIsPublic() == null || !t.getIsPublic()) throw new IllegalArgumentException("类型未公开");
        return entryMapper.findPublicByCode(code);
    }

    public DictEntry create(Long typeId, String label, String value, int sort, int status, String remark, String version) {
        DictEntry e = new DictEntry();
        e.setTypeId(typeId); e.setLabel(label); e.setValue(value); e.setSort(sort); e.setStatus(status); e.setRemark(remark); e.setVersion(version);
        entryMapper.insert(e);
        return e;
    }

    public void update(Long entryId, Long typeId, String label, String value, int sort, int status, String remark, String version) {
        DictEntry e = new DictEntry();
        e.setId(entryId); e.setTypeId(typeId); e.setLabel(label); e.setValue(value); e.setSort(sort); e.setStatus(status); e.setRemark(remark); e.setVersion(version);
        entryMapper.update(e);
    }

    public void delete(Long id) { entryMapper.deleteById(id); }
    public void batchDelete(List<Long> ids) { entryMapper.deleteBatch(ids); }
    public void batchUpdateStatus(List<Long> ids, int status) { entryMapper.updateStatusBatch(ids, status); }
}
