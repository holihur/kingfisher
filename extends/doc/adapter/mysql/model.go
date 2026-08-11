package adapter

import "time"

// docDirectoryPO 目录表（GORM PO；SQLite AutoMigrate 使用，生产走 migrations SQL）
type docDirectoryPO struct {
	ID        uint
	ParentID  uint
	Name      string
	Sort      int
	Status    int
	Version   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (docDirectoryPO) TableName() string { return "doc_directories" }

// docDirRolePO 目录-角色授权（复合主键；无授权记录 = 默认拒绝，仅 admin 可见）
type docDirRolePO struct {
	DirID  uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (docDirRolePO) TableName() string { return "doc_dir_roles" }

// documentPO 文档表
type documentPO struct {
	ID             uint
	DirID          uint
	Title          string
	Content        string
	OwnerID        uint
	Visibility     string
	Status         string
	CurrentVersion int
	Sort           int
	PublishedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (documentPO) TableName() string { return "documents" }

// docVersionPO 版本历史表（append-only；UNIQUE(doc_id, version_no)）
type docVersionPO struct {
	ID        uint
	DocID     uint
	VersionNo int
	Title     string
	Content   string
	OwnerID   uint
	Note      string
	CreatedAt time.Time
}

func (docVersionPO) TableName() string { return "doc_versions" }
