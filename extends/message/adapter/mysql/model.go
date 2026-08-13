package adapter

import "time"

type messagePO struct {
	ID          uint
	SenderID    uint
	SenderType  string
	RecipientID uint
	BatchID     int64
	Title       string
	Content     string
	Status      string
	IsRead      bool
	ReadAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (messagePO) TableName() string { return "messages" }
