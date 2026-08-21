package com.kingfisher.common.query;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class Condition {
    private String field;
    private Operator op;
    private Object value;
}
