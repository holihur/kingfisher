package domain
import "time"
type Role struct {
    ID uint `json:"id"`; Name string `json:"name"`; Code string `json:"code"`
    Description string `json:"description"`; Status int `json:"status"`; Level int `json:"level"`
    CreatedAt time.Time `json:"created_at"`; UpdatedAt time.Time `json:"updated_at"`
}
type Permission struct {
    ID uint `json:"id"`; Name string `json:"name"`; Code string `json:"code"`
    Resource string `json:"resource"`; Action string `json:"action"`; CreatedAt time.Time `json:"created_at"`
}
