package domain

import "time"

// Template 模版（消息/通知/通用，通过 TemplateType 区分）
type Template struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`          // 模板名称
	Code         string    `json:"code"`          // 模板编码（唯一）
	TemplateType string    `json:"template_type"` // general | message | email | sms
	Title        string    `json:"title"`         // 标题模板（支持 {{占位符}}）
	Content      string    `json:"content"`       // 内容模板（支持 {{占位符}}）
	Status       int       `json:"status"`        // 1=启用, 0=禁用
	Remark       string    `json:"remark"`
	Version      string    `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}