package domain

import "time"

type Menu struct {
	ID         uint      `json:"id"`
	ParentID   uint      `json:"parent_id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Component  string    `json:"component"`
	Icon       string    `json:"icon"`
	Sort       int       `json:"sort"`
	Type       int       `json:"type"`
	Permission string    `json:"permission"`
	Status     int       `json:"status"`
	// Version 表示该菜单由哪个版本新增
	Version   string    `json:"version"`
	Children  []Menu    `json:"children,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
