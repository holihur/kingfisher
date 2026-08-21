package com.kingfisher.common.query;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.Data;

import jakarta.servlet.http.HttpServletRequest;
import java.util.*;

/**
 * Go core/query 的 Java 复刻，支持 q/filter/page/page_size/sort。
 * 用法: Query.parse(request, defs)
 */
@Data
public class Query {
    private String keyword;
    private List<Condition> filters = new ArrayList<>();
    private int page = 1;
    private int pageSize = 20;
    private String sort;
    private Defs defs;

    public static final int DEFAULT_PAGE_SIZE = 20;
    public static final int MAX_PAGE_SIZE = 100;

    private static final ObjectMapper MAPPER = new ObjectMapper();

    public static Query parse(HttpServletRequest request, Defs defs) {
        Query q = new Query();
        q.defs = defs;
        String keyword = request.getParameter("q");
        if (keyword == null || keyword.isBlank()) keyword = request.getParameter("keyword");
        q.keyword = keyword != null ? keyword.trim() : "";
        try { q.page = Integer.parseInt(request.getParameter("page")); } catch (Exception e) { q.page = 1; }
        if (q.page < 1) q.page = 1;
        try {
            String ps = request.getParameter("page_size");
            if (ps == null) ps = request.getParameter("pageSize");
            if (ps != null) q.pageSize = Integer.parseInt(ps);
        } catch (Exception e) { q.pageSize = DEFAULT_PAGE_SIZE; }
        if (q.pageSize < 1 || q.pageSize > MAX_PAGE_SIZE) q.pageSize = DEFAULT_PAGE_SIZE;
        q.sort = request.getParameter("sort");
        if (q.sort != null) q.sort = q.sort.trim();

        String filterRaw = request.getParameter("filter");
        if (filterRaw != null && !filterRaw.isBlank()) {
            try {
                Map<String, Object> rawFilters = MAPPER.readValue(filterRaw, new TypeReference<Map<String, Object>>() {});
                for (Map.Entry<String, Object> e : rawFilters.entrySet()) {
                    String field = e.getKey();
                    Field f = defs.get(field);
                    if (f == null) throw new IllegalArgumentException("filter: 未知字段 \"" + field + "\"");
                    if (!f.isFilterable()) throw new IllegalArgumentException("filter: 字段 \"" + field + "\" 不可过滤");
                    Object rawVal = e.getValue();
                    if (rawVal instanceof Map) {
                        Map<String, Object> opMap = (Map<String, Object>) rawVal;
                        if (opMap.size() == 1) {
                            String opStr = opMap.keySet().iterator().next();
                            Object valRaw = opMap.get(opStr);
                            Operator op = Operator.from(opStr);
                            if (op == null) throw new IllegalArgumentException("filter: 字段 \"" + field + "\" 不支持操作符 \"" + opStr + "\"");
                            Object coerced = coerce(valRaw, f.getType(), op);
                            q.filters.add(new Condition(field, op, coerced));
                            continue;
                        }
                    }
                    Object coerced = coerce(rawVal, f.getType(), Operator.EQ);
                    q.filters.add(new Condition(field, Operator.EQ, coerced));
                }
            } catch (IllegalArgumentException ex) { throw ex; }
            catch (Exception ex) { throw new IllegalArgumentException("filter 不是合法 JSON: " + ex.getMessage()); }
        }
        return q;
    }

    private static Object coerce(Object raw, FieldType type, Operator op) {
        if (op == Operator.IN) {
            if (!(raw instanceof List)) throw new IllegalArgumentException("in 需要数组值");
            List<Object> out = new ArrayList<>();
            for (Object e : (List<Object>) raw) out.add(coerce(e, type, Operator.EQ));
            return out;
        }
        switch (type) {
            case BOOL:
                if (raw instanceof Boolean) return raw;
                if (raw instanceof String) {
                    String s = (String) raw;
                    if (s.equals("true") || s.equals("1")) return true;
                    if (s.equals("false") || s.equals("0")) return false;
                    throw new IllegalArgumentException("布尔值需为 true/false");
                }
                throw new IllegalArgumentException("布尔值需为 true/false");
            case INT:
                if (raw instanceof Number) return ((Number) raw).intValue();
                if (raw instanceof String) {
                    try { return Integer.parseInt((String) raw); } catch (NumberFormatException e) { throw new IllegalArgumentException("整数格式不正确"); }
                }
                throw new IllegalArgumentException("整数格式不正确");
            case UINT:
                if (raw instanceof Number) return ((Number) raw).longValue();
                if (raw instanceof String) {
                    try { return Long.parseLong((String) raw); } catch (NumberFormatException e) { throw new IllegalArgumentException("无符号整数格式不正确"); }
                }
                throw new IllegalArgumentException("无符号整数格式不正确");
            case TIME:
                if (raw instanceof String) {
                    try { return java.time.Instant.parse((String) raw); } catch (Exception e) { throw new IllegalArgumentException("时间需为 RFC3339 格式"); }
                }
                throw new IllegalArgumentException("时间需为 RFC3339 格式");
            default:
                if (raw instanceof String) return raw;
                return String.valueOf(raw);
        }
    }

    public String sortExpr() {
        String col = sort;
        boolean desc = false;
        if (col != null && col.startsWith("-")) { desc = true; col = col.substring(1); }
        Field f = col != null ? defs.get(col) : null;
        if (f != null) return f.getName() + (desc ? " DESC" : " ASC");
        return "id DESC";
    }

    public int offset() { return (page - 1) * pageSize; }

    public List<String> searchableColumns() {
        List<String> cols = new ArrayList<>();
        for (Field f : defs.values()) if (f.isSearchable() && f.getType() == FieldType.STRING) cols.add(f.getName());
        return cols;
    }
}
