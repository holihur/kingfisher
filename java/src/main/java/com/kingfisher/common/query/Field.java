package com.kingfisher.common.query;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class Field {
    private String name;
    private FieldType type;
    private boolean searchable;
    private boolean filterable;

    public static Field searchableString(String name) {
        return new Field(name, FieldType.STRING, true, true);
    }
    public static Field filterable(FieldType type, String name) {
        return new Field(name, type, false, true);
    }
}
