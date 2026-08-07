package adapter

import "time"

type dictTypePO struct {
	ID        uint
	Code      string
	Name      string
	IsPublic  bool
	Status    int
	Remark    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (dictTypePO) TableName() string { return "dict_types" }

type dictEntryPO struct {
	ID        uint
	TypeID    uint
	Label     string
	Value     string
	Sort      int
	Status    int
	Remark    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (dictEntryPO) TableName() string { return "dict_entries" }
