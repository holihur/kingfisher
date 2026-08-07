package adapter

import "time"

type configPO struct {
	ID            uint
	Key           string
	Value         string
	Remark        string
	IsPublic      bool
	Version       string
	Render        string
	RenderOptions string
	GroupID       uint
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (configPO) TableName() string { return "system_configs" }

type configGroupPO struct {
	ID        uint
	Name      string
	Sort      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (configGroupPO) TableName() string { return "config_groups" }
