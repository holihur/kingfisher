package com.kingfisher.modules.system.service;

import com.kingfisher.modules.system.domain.SystemInfo;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Service;

import java.lang.management.ManagementFactory;
import java.lang.management.MemoryUsage;
import java.net.InetAddress;

/**
 * 系统信息服务，与 Go extends/system/app.SystemService 对齐。
 * 简化版：使用 JVM 信息替代 gopsutil。
 */
@Service
public class SystemService {

    private final JdbcTemplate jdbcTemplate;

    @Value("${kingfisher.build.version:dev}")
    private String buildVersion;

    @Value("${kingfisher.build.commit:unknown}")
    private String buildCommit;

    @Value("${kingfisher.build.time:unknown}")
    private String buildTime;

    public SystemService(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public SystemInfo getInfo() {
        SystemInfo info = new SystemInfo();

        info.setBackendVersion(buildVersion);
        info.setBackendCommit(buildCommit);
        info.setBuildTime(buildTime);
        String jv = System.getProperty("java.version");
        info.setJavaVersion(jv);
        info.setGoVersion(jv);

        info.setOs(System.getProperty("os.name"));
        info.setOsVersion(System.getProperty("os.version"));
        info.setArch(System.getProperty("os.arch"));
        try {
            info.setHostname(InetAddress.getLocalHost().getHostName());
        } catch (Exception e) {
            info.setHostname("unknown");
        }

        info.setCpuCores(Runtime.getRuntime().availableProcessors());
        info.setCpuPercent(0.0);
        info.setCpuVendor("");
        info.setCpuModel("");

        MemoryUsage heap = ManagementFactory.getMemoryMXBean().getHeapMemoryUsage();
        long totalMem = heap.getMax();
        if (totalMem == -1) totalMem = heap.getCommitted();
        long usedMem = heap.getUsed();
        info.setMemTotal(totalMem > 0 ? totalMem : 0);
        info.setMemUsed(usedMem);
        info.setMemUsedPercent(totalMem > 0 ? (double) usedMem / totalMem * 100 : 0);
        info.setMemAvailable(totalMem > usedMem ? totalMem - usedMem : 0);
        info.setUptime(ManagementFactory.getRuntimeMXBean().getUptime() / 1000);

        info.setDbDriver("sqlite");
        try {
            String version = jdbcTemplate.queryForObject("SELECT sqlite_version()", String.class);
            info.setDbVersion(version);
        } catch (Exception e) {
            info.setDbVersion("unknown");
        }
        info.setRedisVersion("");
        info.setDisk(java.util.List.of());
        info.setNetBytesRecv(0);
        info.setNetBytesSent(0);
        info.setNetRecvRate(0);
        info.setNetSentRate(0);

        return info;
    }
}
