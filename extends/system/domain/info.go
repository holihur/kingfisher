package domain

// SystemInfo 系统信息（服务监控展示用）
type SystemInfo struct {
	// 后端版本
	BackendVersion string `json:"backend_version"`
	BackendCommit  string `json:"backend_commit"`
	BuildTime      string `json:"build_time"`
	GoVersion      string `json:"go_version"`
	// 运行环境
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	Hostname  string `json:"hostname"`
	Arch      string `json:"arch"`
	// 依赖版本
	RedisVersion string `json:"redis_version"`
	DBVersion    string `json:"db_version"`
	DBDriver     string `json:"db_driver"`
	// CPU
	CPUCores   int     `json:"cpu_cores"`
	CPUPercent float64 `json:"cpu_percent"`
	CPUVendor  string  `json:"cpu_vendor"`
	CPUModel   string  `json:"cpu_model"`
	// 内存
	MemTotal     uint64  `json:"mem_total"` // bytes
	MemUsed      uint64  `json:"mem_used"`  // bytes
	MemUsedPct   float64 `json:"mem_used_percent"`
	MemAvailable uint64  `json:"mem_available"` // bytes
	// 磁盘
	Disk []DiskInfo `json:"disk"`
	// 网络/带宽
	NetBytesRecv uint64  `json:"net_bytes_recv"` // 累计接收字节
	NetBytesSent uint64  `json:"net_bytes_sent"` // 累计发送字节
	NetRecvRate  float64 `json:"net_recv_rate"`  // 实时接收速率 bytes/s
	NetSentRate  float64 `json:"net_sent_rate"`  // 实时发送速率 bytes/s
	// 运行时长
	Uptime uint64 `json:"uptime"` // 秒
}

// DiskInfo 磁盘使用情况
type DiskInfo struct {
	Path        string  `json:"path"`
	FsType      string  `json:"fs_type"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}
