package domain

import "time"

type SystemConfig struct {
	ID        uint      `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Remark    string    `json:"remark"`
	// IsPublic 是否公开：公开项可在未登录状态下通过 /api/v1/public/configs 读取
	IsPublic bool `json:"is_public"`
	// Version 表示该配置由哪个版本新增
	Version string `json:"version"`
	// Render 前端渲染组件：text|number|switch|select|textarea
	Render string `json:"render"`
	// RenderOptions 渲染组件配置（JSON），如 select 的选项
	RenderOptions string    `json:"render_options"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
