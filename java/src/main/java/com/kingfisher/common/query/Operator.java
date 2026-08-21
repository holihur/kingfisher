package com.kingfisher.common.query;

public enum Operator {
    EQ("eq"), NE("ne"), CONTAINS("contains"), IN("in"), GT("gt"), GTE("gte"), LT("lt"), LTE("lte");
    private final String value;
    Operator(String v){ this.value=v; }
    public static Operator from(String s){
        for(Operator o: values()) if(o.value.equals(s)) return o;
        return null;
    }
}
