package dataaccess_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kingfisher/core/dataaccess"
)

type scopeRow struct {
	ID           uint
	OwnerID      uint
	DepartmentID uint
	Title        string
}

func (scopeRow) TableName() string { return "scope_rows" }

func newScopeDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/scope.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&scopeRow{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	rows := []scopeRow{
		{ID: 1, OwnerID: 10, DepartmentID: 20, Title: "self"},
		{ID: 2, OwnerID: 11, DepartmentID: 20, Title: "department"},
		{ID: 3, OwnerID: 12, DepartmentID: 30, Title: "subtree"},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed rows: %v", err)
	}
	return db
}

func TestScopesFilterByKind(t *testing.T) {
	db := newScopeDB(t)
	tests := []struct {
		name  string
		scope dataaccess.Scope
		want  int64
	}{
		{name: "all", scope: dataaccess.All(), want: 3},
		{name: "self", scope: dataaccess.Self("owner_id", 10), want: 1},
		{name: "department", scope: dataaccess.Department("department_id", []uint{20}), want: 2},
		{name: "subtree", scope: dataaccess.Subtree("department_id", []uint{20, 30}), want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int64
			if err := dataaccess.Apply(db.Model(&scopeRow{}), tt.scope).Count(&got).Error; err != nil {
				t.Fatalf("count: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d rows, want %d", got, tt.want)
			}
		})
	}
}

func TestApplyRejectsUnsafeColumn(t *testing.T) {
	db := newScopeDB(t)
	var count int64
	err := dataaccess.Apply(db.Model(&scopeRow{}), dataaccess.Self("owner_id = 1 OR 1", 10)).Count(&count).Error
	if err == nil {
		t.Fatal("unsafe column should produce an error")
	}
}
