package app

import (
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"gorm.io/gorm"

	"kingfisher/extends/system/domain"
)

// VersionInfo 后端构建版本信息（由 main 注入）
type VersionInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

// SystemService 系统信息服务
type SystemService struct {
	db  *gorm.DB
	rdb *redis.Client
	ver VersionInfo
}

func NewSystemService(db *gorm.DB, rdb *redis.Client, ver VersionInfo) *SystemService {
	return &SystemService{db: db, rdb: rdb, ver: ver}
}

// GetInfo 收集系统信息
func (s *SystemService) GetInfo(ctx context.Context) (*domain.SystemInfo, error) {
	info := &domain.SystemInfo{
		BackendVersion: s.ver.Version,
		BackendCommit:  s.ver.Commit,
		BuildTime:      s.ver.BuildTime,
		GoVersion:      runtime.Version(),
		Arch:           runtime.GOARCH,
	}

	// OS / host / uptime
	if hi, err := host.Info(); err == nil {
		info.OS = hi.Platform + " " + hi.PlatformVersion
		info.OSVersion = hi.KernelVersion
		info.Hostname = hi.Hostname
		info.Uptime = hi.Uptime
	}

	// CPU
	info.CPUCores = runtime.NumCPU()
	if ci, err := cpu.Info(); err == nil && len(ci) > 0 {
		info.CPUVendor = ci[0].VendorID
		info.CPUModel = ci[0].ModelName
	}
	// CPU 使用率（首次调用为 0，预热一次取第二次）
	_, _ = cpu.Percent(0, false)
	if pct, err := cpu.Percent(100*time.Millisecond, false); err == nil && len(pct) > 0 {
		info.CPUPercent = pct[0]
	}

	// 内存
	if mv, err := mem.VirtualMemory(); err == nil {
		info.MemTotal = mv.Total
		info.MemUsed = mv.Used
		info.MemUsedPct = mv.UsedPercent
		info.MemAvailable = mv.Available
	}

	// 磁盘（根分区 + 常见挂载点）
	for _, p := range []string{"/", "/home", "/data"} {
		if du, err := disk.Usage(p); err == nil {
			info.Disk = append(info.Disk, domain.DiskInfo{
				Path:        du.Path,
				FsType:      du.Fstype,
				Total:       du.Total,
				Used:        du.Used,
				UsedPercent: du.UsedPercent,
			})
		}
	}

	// 网络/带宽：两次采样计算实时速率
	if n0, err := net.IOCounters(false); err == nil && len(n0) > 0 {
		info.NetBytesRecv = n0[0].BytesRecv
		info.NetBytesSent = n0[0].BytesSent
		time.Sleep(500 * time.Millisecond)
		if n1, err := net.IOCounters(false); err == nil && len(n1) > 0 {
			dt := 0.5 // seconds
			info.NetRecvRate = float64(n1[0].BytesRecv-n0[0].BytesRecv) / dt
			info.NetSentRate = float64(n1[0].BytesSent-n0[0].BytesSent) / dt
		}
	}

	// Redis 版本
	info.RedisVersion = s.redisVersion(ctx)
	// 数据库版本 + driver
	info.DBDriver = s.dbDriver()
	info.DBVersion = s.dbVersion(ctx)

	return info, nil
}

func (s *SystemService) redisVersion(ctx context.Context) string {
	if s.rdb == nil {
		return ""
	}
	res, err := s.rdb.Info(ctx, "server").Result()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(res, "\r\n") {
		if strings.HasPrefix(line, "redis_version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "redis_version:"))
		}
	}
	return ""
}

func (s *SystemService) dbDriver() string {
	if s.db == nil {
		return ""
	}
	return s.db.Name()
}

func (s *SystemService) dbVersion(ctx context.Context) string {
	if s.db == nil {
		return ""
	}
	var v string
	switch s.db.Name() {
	case "sqlite":
		_ = s.db.Raw("SELECT sqlite_version()").Scan(&v).Error
	case "postgres":
		_ = s.db.Raw("SELECT version()").Scan(&v).Error
	default: // mysql
		_ = s.db.Raw("SELECT VERSION()").Scan(&v).Error
	}
	return v
}

// LoadAverage 系统负载（额外提供）
func (s *SystemService) LoadAverage() (float64, float64, float64) {
	if a, err := load.Avg(); err == nil {
		return a.Load1, a.Load5, a.Load15
	}
	return 0, 0, 0
}
