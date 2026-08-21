package com.kingfisher.modules.system.domain;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.util.List;

/**
 * 系统信息领域实体，与 Go extends/system/domain.SystemInfo 1:1 对齐。
 */
@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class SystemInfo {
    @JsonProperty("backend_version")
    private String backendVersion;
    @JsonProperty("backend_commit")
    private String backendCommit;
    @JsonProperty("build_time")
    private String buildTime;
    @JsonProperty("java_version")
    private String javaVersion;
    @JsonProperty("go_version")
    private String goVersion;
    private String os;
    @JsonProperty("os_version")
    private String osVersion;
    private String hostname;
    private String arch;
    @JsonProperty("redis_version")
    private String redisVersion;
    @JsonProperty("db_version")
    private String dbVersion;
    @JsonProperty("db_driver")
    private String dbDriver;
    @JsonProperty("cpu_cores")
    private int cpuCores;
    @JsonProperty("cpu_percent")
    private double cpuPercent;
    @JsonProperty("cpu_vendor")
    private String cpuVendor;
    @JsonProperty("cpu_model")
    private String cpuModel;
    @JsonProperty("mem_total")
    private long memTotal;
    @JsonProperty("mem_used")
    private long memUsed;
    @JsonProperty("mem_used_percent")
    private double memUsedPercent;
    @JsonProperty("mem_available")
    private long memAvailable;
    private long uptime;
    private List<DiskInfo> disk;
    @JsonProperty("net_bytes_recv")
    private long netBytesRecv;
    @JsonProperty("net_bytes_sent")
    private long netBytesSent;
    @JsonProperty("net_recv_rate")
    private double netRecvRate;
    @JsonProperty("net_sent_rate")
    private double netSentRate;

    @Data
    public static class DiskInfo {
        private String path;
        @JsonProperty("fs_type")
        private String fsType;
        private long total;
        private long used;
        @JsonProperty("used_percent")
        private double usedPercent;
    }
}
