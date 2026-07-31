package adapter

import "time"

type configPO struct {
	ID        uint
	Key       string
	Value     string
	Remark    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (configPO) TableName() string { return "system_configs" }
