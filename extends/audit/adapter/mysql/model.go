package adapter

import "time"

type auditPO struct {
	ID         uint
	UserID     uint
	Username   string
	Action     string
	Resource   string
	ResourceID uint
	Detail     string
	IP         string
	UserAgent  string
	CreatedAt  time.Time
}

func (auditPO) TableName() string { return "audit_logs" }
