package query

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type item struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:64"`
	Category  string `gorm:"size:32"`
	Enabled   bool   `gorm:"default:false"`
	Sort      int    `gorm:"default:0"`
	CreatedAt time.Time
}

func (item) TableName() string { return "items" }

var testDefs = Defs{
	"name":       {Name: "name", Type: TypeString, Searchable: true, Filterable: true},
	"category":   {Name: "category", Type: TypeString, Searchable: true, Filterable: true},
	"enabled":    {Name: "enabled", Type: TypeBool, Filterable: true},
	"sort":       {Name: "sort", Type: TypeInt, Filterable: true},
	"created_at": {Name: "created_at", Type: TypeTime, Filterable: true},
	"id":         {Name: "id", Type: TypeUint, Filterable: true},
}

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&item{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	rows := []item{
		{Name: "site_name", Category: "site", Enabled: true, Sort: 1, CreatedAt: now},
		{Name: "site_logo", Category: "site", Enabled: true, Sort: 2, CreatedAt: now.Add(time.Minute)},
		{Name: "max_attempts", Category: "security", Enabled: false, Sort: 3, CreatedAt: now.Add(2 * time.Minute)},
		{Name: "lockout", Category: "security", Enabled: true, Sort: 4, CreatedAt: now.Add(3 * time.Minute)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func newCtx(query string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/?"+query, nil)
	return c
}

func TestKeywordSearch(t *testing.T) {
	db := setupDB(t)
	q, err := Parse(newCtx("q=site"), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	total, err := q.Find(db.Model(&item{}), &items)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("want 2 matches for q=site, got %d", total)
	}
}

func TestEqFilter(t *testing.T) {
	db := setupDB(t)
	q, err := Parse(newCtx(`filter={"category":"security"}`), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	total, err := q.Find(db.Model(&item{}), &items)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("want 2 security, got %d", total)
	}
}

func TestBoolFilter(t *testing.T) {
	db := setupDB(t)
	q, err := Parse(newCtx(`filter={"enabled":true}`), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	total, err := q.Find(db.Model(&item{}), &items)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("want 3 enabled, got %d", total)
	}
}

func TestContainsOp(t *testing.T) {
	db := setupDB(t)
	q, err := Parse(newCtx(`filter={"name":{"contains":"site"}}`), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	total, err := q.Find(db.Model(&item{}), &items)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("want 2 with name contains site, got %d", total)
	}
}

func TestInOp(t *testing.T) {
	db := setupDB(t)
	q, err := Parse(newCtx(`filter={"sort":{"in":[1,4]}}`), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	total, err := q.Find(db.Model(&item{}), &items)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("want 2 with sort in [1,4], got %d", total)
	}
}

func TestGteOp(t *testing.T) {
	db := setupDB(t)
	q, err := Parse(newCtx(`filter={"sort":{"gte":3}}`), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	total, err := q.Find(db.Model(&item{}), &items)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("want 2 with sort>=3, got %d", total)
	}
}

func TestPaginationAndSort(t *testing.T) {
	db := setupDB(t)
	q, err := Parse(newCtx("page=2&page_size=2&sort=-name"), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	total, err := q.Find(db.Model(&item{}), &items)
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 {
		t.Errorf("want total 4, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items on page 2, got %d", len(items))
	}
	// sort=-name → 倒序 site_name, site_logo, max_attempts, lockout
	// 第 2 页应为 max_attempts, lockout
	if items[0].Name != "max_attempts" || items[1].Name != "lockout" {
		t.Errorf("unexpected page order: %v, %v", items[0].Name, items[1].Name)
	}
}

func TestUnknownFieldRejected(t *testing.T) {
	_, err := Parse(newCtx(`filter={"nope":"x"}`), testDefs)
	if err == nil {
		t.Fatal("want error for unknown filter field")
	}
}

func TestUnknownOpRejected(t *testing.T) {
	_, err := Parse(newCtx(`filter={"name":{"bogus":"x"}}`), testDefs)
	if err == nil {
		t.Fatal("want error for unknown op")
	}
}
