package adapter
import "time"
type rolePO struct { ID uint; Name string; Code string; Description string; Status int; Level int; CreatedAt time.Time; UpdatedAt time.Time }
func (rolePO) TableName() string { return "roles" }
type permissionPO struct { ID uint; Name string; Code string; Resource string; Action string; CreatedAt time.Time }
func (permissionPO) TableName() string { return "permissions" }
