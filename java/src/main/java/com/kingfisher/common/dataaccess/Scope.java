package com.kingfisher.common.dataaccess;

import lombok.Data;

import java.util.List;

@Data
public class Scope {
    private String type;
    private String field;
    private Object value;
    private List<Long> ids;

    public static Scope all() {
        Scope s = new Scope();
        s.type = "all";
        return s;
    }

    public static Scope self(String field, Long userId) {
        Scope s = new Scope();
        s.type = "self";
        s.field = field;
        s.value = userId;
        return s;
    }

    public static Scope department(String field, List<Long> deptIds) {
        Scope s = new Scope();
        s.type = "department";
        s.field = field;
        s.ids = deptIds;
        return s;
    }

    public static Scope subtree(String field, List<Long> deptIds) {
        Scope s = new Scope();
        s.type = "subtree";
        s.field = field;
        s.ids = deptIds;
        return s;
    }

    public boolean isAll() { return "all".equals(type); }
    public boolean isSelf() { return "self".equals(type); }
    public boolean isDepartment() { return "department".equals(type); }
    public boolean isSubtree() { return "subtree".equals(type); }
}
