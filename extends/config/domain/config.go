package domain

import "time"

type SystemConfig struct {
	ID        uint      `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Remark    string    `json:"remark"`
	// IsPublic 是否公开：公开项可在未登录状态下通过 /api/v1/public/configs 读取
	IsPublic  bool      `json:"is_public"`
	// Version 表示该配置由哪个版本新增
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
