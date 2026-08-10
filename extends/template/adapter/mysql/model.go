package adapter

import "time"

type templatePO struct {
	ID           uint
	Name         string
	Code         string
	TemplateType string
	Title        string
	Content      string
	Status       int
	Remark       string
	Version      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (templatePO) TableName() string { return "templates" }
