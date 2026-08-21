package com.kingfisher.modules.user.domain;

import lombok.Data;

/**
 * 部门轻量视图，对应 departments 表。
 */
@Data
public class Department {
    private Long id;
    private String name;
}
