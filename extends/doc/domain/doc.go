// Package doc implements document management logic.
package domain

import "time"

// 文档可见性
const (
	VisibilityShared  = "shared"  // 共享：目录可见者均可读
	VisibilityPrivate = "private" // 私有：仅作者（+admin）可读/改
)

// 文档生命周期状态
const (
	DocStatusDraft     = "draft"     // 草稿：仅作者/admin 可见
	DocStatusPublished = "published" // 已发布：按可见性规则公开
)

// DocDirectory 文档目录（树形）
type DocDirectory struct {
	ID           uint           `json:"id"`
	ParentID     uint           `json:"parent_id"`
	Name         string         `json:"name"`
	Sort         int            `json:"sort"`
	Status       int            `json:"status"`
	Version      string         `json:"version"`
	GrantedRoles []uint         `json:"granted_roles,omitempty"` // 可见角色 id 白名单（空 = 默认拒绝：仅 admin 可见）
	Docs         []DocTreeItem  `json:"docs,omitempty"`          // 该目录下当前用户可见的文档（叶子节点）
	Children     []DocDirectory `json:"children,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// DocTreeItem 目录树中的文档叶子节点（轻量，不含正文）
type DocTreeItem struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Visibility string `json:"visibility"`
}

// Document 文档
type Document struct {
	ID             uint       `json:"id"`
	DirID          uint       `json:"dir_id"`
	Title          string     `json:"title"`
	Content        string     `json:"content"` // Quill 输出的 HTML（当前最新内容）
	OwnerID        uint       `json:"owner_id"`
	OwnerName      string     `json:"owner_name,omitempty"`
	Visibility     string     `json:"visibility"`      // shared | private
	Status         string     `json:"status"`          // draft | published
	CurrentVersion int        `json:"current_version"` // 当前指向 doc_versions 的最新 version_no
	Sort           int        `json:"sort"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DocVersion 文档版本（append-only 历史）
type DocVersion struct {
	ID        uint      `json:"id"`
	DocID     uint      `json:"doc_id"`
	VersionNo int       `json:"version_no"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	OwnerID   uint      `json:"owner_id"`
	OwnerName string    `json:"owner_name,omitempty"`
	Note      string    `json:"note"` // 变更说明
	CreatedAt time.Time `json:"created_at"`
}
