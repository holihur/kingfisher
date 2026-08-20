package domain

import "time"

type Task struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	OwnerID      uint      `json:"owner_id"`
	DepartmentID uint      `json:"department_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
