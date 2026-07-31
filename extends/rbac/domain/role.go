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
type Menu struct {
    ID uint `json:"id"`; ParentID uint `json:"parent_id"`; Name string `json:"name"`
    Path string `json:"path"`; Component string `json:"component"`; Icon string `json:"icon"`
    Sort int `json:"sort"`; Type int `json:"type"`; Permission string `json:"permission"`; Status int `json:"status"`
}
