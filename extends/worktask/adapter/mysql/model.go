package adapter

import "time"

type taskPO struct {
	ID           uint `gorm:"primaryKey"`
	Title        string
	Description  string
	OwnerID      uint   `gorm:"index;not null"`
	DepartmentID uint   `gorm:"index;not null"`
	Status       string `gorm:"index;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (taskPO) TableName() string { return "tasks" }
