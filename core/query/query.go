// Package query implements a generic list query DSL used by list endpoints.
//
// URL parameters:
//
//	q=keyword          — 关键词，对声明为 Searchable 的字符串字段做 LIKE 模糊匹配
//	filter=JSON        — 结构化过滤，如 {"is_public":true,"key":{"contains":"site"},"group_id":{"in":[1,2]}}
//	page / page_size   — 分页（page_size 1~100，默认 20）
//	sort=-created_at   — 排序（"-" 前缀表示倒序，字段须在白名单内）
//
// filter 支持的操作符：eq(默认)/ne/contains/in/gt/gte/lt/lte。
// 字段必须事先在 Defs 中声明，未知字段或操作符返回错误（由 handler 转 400）。
package query

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FieldType 字段类型，用于把 filter 里的 JSON 值强转为可比较的类型。
type FieldType int

const (
	TypeString FieldType = iota
	TypeBool
	TypeInt
	TypeUint
	TypeTime
)

// Field 声明一个可参与查询的字段。
type Field struct {
	Name       string    // DB 列名
	Type       FieldType // 值类型
	Searchable bool      // 是否参与 q= 关键词模糊搜索（字符串字段）
	Filterable bool      // 是否允许出现在 filter= 中
}

// Defs 字段定义表：key = URL/JSON 字段名，value = Field。
type Defs map[string]Field

// Operator 过滤操作符。
type Operator string

const (
	OpEq       Operator = "eq"       // 等于
	OpNe       Operator = "ne"       // 不等于
	OpContains Operator = "contains" // LIKE %v%
	OpIn       Operator = "in"       // IN (...)
	OpGt       Operator = "gt"
	OpGte      Operator = "gte"
	OpLt       Operator = "lt"
	OpLte      Operator = "lte"
)

var validOps = map[Operator]bool{
	OpEq: true, OpNe: true, OpContains: true, OpIn: true,
	OpGt: true, OpGte: true, OpLt: true, OpLte: true,
}

// Condition 一条过滤条件。
type Condition struct {
	Field string
	Op    Operator
	Value any
}

// Query 一次列表查询的全部条件。
type Query struct {
	Keyword  string
	Filters  []Condition
	Page     int
	PageSize int
	Sort     string // 如 "-created_at"；空则默认 id DESC
	defs     Defs
}

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Parse 从 gin query 解析查询条件。defs 为字段白名单。
func Parse(c *gin.Context, defs Defs) (*Query, error) {
	q := &Query{defs: defs}
	q.Keyword = strings.TrimSpace(c.Query("q"))
	if kw := strings.TrimSpace(c.Query("keyword")); q.Keyword == "" && kw != "" {
		q.Keyword = kw // 向后兼容的别名
	}
	q.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	if q.Page < 1 {
		q.Page = 1
	}
	q.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))
	if q.PageSize < 1 || q.PageSize > MaxPageSize {
		q.PageSize = DefaultPageSize
	}
	q.Sort = strings.TrimSpace(c.Query("sort"))

	if raw := c.Query("filter"); raw != "" {
		var rawFilters map[string]json.RawMessage
		if err := json.Unmarshal([]byte(raw), &rawFilters); err != nil {
			return nil, fmt.Errorf("filter 不是合法 JSON: %w", err)
		}
		for field, rawVal := range rawFilters {
			f, ok := defs[field]
			if !ok {
				return nil, fmt.Errorf("filter: 未知字段 %q", field)
			}
			if !f.Filterable {
				return nil, fmt.Errorf("filter: 字段 %q 不可过滤", field)
			}
			// 探测操作符包装：{"contains":"x"} / {"in":[1,2]}
			var probe map[string]json.RawMessage
			if json.Unmarshal(rawVal, &probe) == nil && len(probe) == 1 {
				opStr, valRaw := "", json.RawMessage(nil)
				for k, v := range probe {
					opStr, valRaw = k, v
				}
				op := Operator(opStr)
				if !validOps[op] {
					return nil, fmt.Errorf("filter: 字段 %q 不支持操作符 %q", field, opStr)
				}
				val, err := coerce(valRaw, f.Type, op)
				if err != nil {
					return nil, fmt.Errorf("filter: 字段 %q 值不合法: %w", field, err)
				}
				q.Filters = append(q.Filters, Condition{Field: field, Op: op, Value: val})
				continue
			}
			// 裸值 = 等值
			val, err := coerce(rawVal, f.Type, OpEq)
			if err != nil {
				return nil, fmt.Errorf("filter: 字段 %q 值不合法: %w", field, err)
			}
			q.Filters = append(q.Filters, Condition{Field: field, Op: OpEq, Value: val})
		}
	}
	return q, nil
}

// coerce 把 JSON 值按字段类型转成可比较的 Go 值。op=in 时值应为数组。
func coerce(raw json.RawMessage, t FieldType, op Operator) (any, error) {
	if op == OpIn {
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, fmt.Errorf("in 需要数组值")
		}
		out := make([]any, 0, len(arr))
		for _, e := range arr {
			v, err := coerce(e, t, OpEq)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	}
	switch t {
	case TypeBool:
		// 兼容 JSON 布尔 true/false 与字符串 "true"/"false"（前端 select 提交字符串）
		var s string
		if json.Unmarshal(raw, &s) == nil {
			switch s {
			case "true", "1":
				return true, nil
			case "false", "0":
				return false, nil
			default:
				return nil, fmt.Errorf("布尔值需为 true/false")
			}
		}
		var b bool
		if err := json.Unmarshal(raw, &b); err != nil {
			return nil, err
		}
		return b, nil
	case TypeInt:
		var i int
		if err := json.Unmarshal(raw, &i); err != nil {
			// 兼容字符串数字（如 "1"）
			var s string
			if err2 := json.Unmarshal(raw, &s); err2 != nil {
				return nil, err
			}
			n, err3 := strconv.Atoi(s)
			if err3 != nil {
				return nil, fmt.Errorf("整数格式不正确")
			}
			return n, nil
		}
		return i, nil
	case TypeUint:
		var u uint64
		if err := json.Unmarshal(raw, &u); err != nil {
			var s string
			if err2 := json.Unmarshal(raw, &s); err2 != nil {
				return nil, err
			}
			n, err3 := strconv.ParseUint(s, 10, 64)
			if err3 != nil {
				return nil, fmt.Errorf("无符号整数格式不正确")
			}
			return n, nil
		}
		return u, nil
	case TypeTime:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		tm, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, fmt.Errorf("时间需为 RFC3339 格式")
		}
		return tm, nil
	default: // TypeString
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return s, nil
	}
}

// Find 对 base 应用过滤/排序/分页并写入 items，返回 total。
// base 应为带 Model 的 *gorm.DB 链，如 r.db.WithContext(ctx).Model(&userPO{})；
// 调用方可先附加固定条件（如 dict-entries 的 type_id = ?）。
func (q *Query) Find(base *gorm.DB, items any) (int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = DefaultPageSize
	}
	scopes := q.scopes()

	countDB := base.Session(&gorm.Session{})
	for _, s := range scopes {
		countDB = countDB.Scopes(s)
	}
	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return 0, err
	}

	listDB := base
	for _, s := range scopes {
		listDB = listDB.Scopes(s)
	}
	err := listDB.Order(q.SortExpr()).
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(items).Error
	return total, err
}

// SortExpr 返回经过白名单校验的 ORDER BY 表达式。
func (q *Query) SortExpr() string {
	col := q.Sort
	desc := false
	if strings.HasPrefix(col, "-") {
		desc = true
		col = strings.TrimPrefix(col, "-")
	}
	if f, ok := q.defs[col]; ok {
		if desc {
			return f.Name + " DESC"
		}
		return f.Name + " ASC"
	}
	return "id DESC"
}

func (q *Query) scopes() []func(*gorm.DB) *gorm.DB {
	var scopes []func(*gorm.DB) *gorm.DB

	if q.Keyword != "" {
		kw := q.Keyword
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			var cols []string
			for _, f := range q.defs {
				if f.Searchable && f.Type == TypeString {
					cols = append(cols, f.Name)
				}
			}
			if len(cols) == 0 {
				return db
			}
			like := "%" + escapeLike(kw) + "%"
			conds := make([]string, 0, len(cols))
			args := make([]any, 0, len(cols))
			for _, c := range cols {
				conds = append(conds, c+" LIKE ?")
				args = append(args, like)
			}
			return db.Where(strings.Join(conds, " OR "), args...)
		})
	}

	for _, fc := range q.Filters {
		f := q.defs[fc.Field]
		c := fc
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			col := f.Name
			switch c.Op {
			case OpEq:
				if c.Value == nil {
					return db.Where(col + " IS NULL")
				}
				return db.Where(col+" = ?", c.Value)
			case OpNe:
				return db.Where(col+" <> ?", c.Value)
			case OpContains:
				return db.Where(col+" LIKE ?", "%"+escapeLike(fmt.Sprint(c.Value))+"%")
			case OpIn:
				if arr, ok := c.Value.([]any); ok {
					return db.Where(col+" IN ?", arr)
				}
				return db.Where(col+" IN ?", []any{c.Value})
			case OpGt:
				return db.Where(col+" > ?", c.Value)
			case OpGte:
				return db.Where(col+" >= ?", c.Value)
			case OpLt:
				return db.Where(col+" < ?", c.Value)
			case OpLte:
				return db.Where(col+" <= ?", c.Value)
			default:
				return db
			}
		})
	}
	return scopes
}

// escapeLike 转义 LIKE 通配符，避免用户输入被当作通配。
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
