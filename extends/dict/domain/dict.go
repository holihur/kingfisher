package domain

import "time"

// DictType 字典类型
type DictType struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`      // 唯一编码，如 "gender"
	Name      string    `json:"name"`      // 显示名称，如 "性别"
	IsPublic  bool      `json:"is_public"` // 是否公开（公开项可通过公共 API 读取）
	Status    int       `json:"status"`    // 1=启用, 0=禁用
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DictEntry 字典条目
type DictEntry struct {
	ID        uint      `json:"id"`
	TypeID    uint      `json:"type_id"` // 关联 dict_types.id
	Label     string    `json:"label"`   // 显示标签，如 "男"
	Value     string    `json:"value"`   // 实际值，如 "male"
	Sort      int       `json:"sort"`    // 排序
	Status    int       `json:"status"`  // 1=启用, 0=禁用
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
