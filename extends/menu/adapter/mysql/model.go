package adapter

import "time"

type menuPO struct {
	ID         uint
	ParentID   uint
	Name       string
	Path       string
	Component  string
	Icon       string
	Sort       int
	Type       int
	Permission string
	Status     int
	Version    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (menuPO) TableName() string { return "menus" }
