// Package dataaccess 提供显式的数据访问范围 GORM Scope。
package dataaccess

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

type Kind string

const (
	KindAll        Kind = "all"
	KindSelf       Kind = "self"
	KindDepartment Kind = "department"
	KindSubtree    Kind = "subtree"
)

type Scope struct {
	kind  Kind
	field string
	ids   []uint
	user  uint
}

func All() Scope { return Scope{kind: KindAll} }

func Self(field string, userID uint) Scope {
	return Scope{kind: KindSelf, field: field, user: userID}
}

func Department(field string, departmentIDs []uint) Scope {
	return Scope{kind: KindDepartment, field: field, ids: append([]uint(nil), departmentIDs...)}
}

func Subtree(field string, departmentIDs []uint) Scope {
	return Scope{kind: KindSubtree, field: field, ids: append([]uint(nil), departmentIDs...)}
}

func Apply(db *gorm.DB, scope Scope) *gorm.DB {
	if scope.kind == KindAll {
		return db
	}
	if !validColumn(scope.field) {
		_ = db.AddError(fmt.Errorf("data access: invalid column %q", scope.field))
		return db
	}
	switch scope.kind {
	case KindSelf:
		return db.Where(scope.field+" = ?", scope.user)
	case KindDepartment, KindSubtree:
		if len(scope.ids) == 0 {
			return db.Where("1 = 0")
		}
		return db.Where(scope.field+" IN ?", scope.ids)
	default:
		_ = db.AddError(fmt.Errorf("data access: unsupported scope %q", scope.kind))
		return db
	}
}

var validColumnPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func validColumn(column string) bool { return validColumnPattern.MatchString(column) }
