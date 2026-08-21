package com.kingfisher.common.query;

import java.util.HashMap;

public class Defs extends HashMap<String, Field> {
    public static Defs of(Object... kvs){
        Defs d = new Defs();
        for(int i=0;i<kvs.length;i+=2) d.put((String)kvs[i], (Field)kvs[i+1]);
        return d;
    }
}
