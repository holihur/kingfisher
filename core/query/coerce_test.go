package query

import (
	"encoding/json"
	"testing"
	"time"
)

// TestCoerceBool 覆盖布尔类型强转：JSON 布尔、字符串 "true"/"1"/"false"/"0"、非法值。
func TestCoerceBool(t *testing.T) {
	cases := []struct {
		raw  string
		want bool
	}{
		{`true`, true},
		{`false`, false},
		{`"true"`, true},
		{`"1"`, true},
		{`"false"`, false},
		{`"0"`, false},
	}
	for _, c := range cases {
		got, err := coerce(json.RawMessage(c.raw), TypeBool, OpEq)
		if err != nil {
			t.Fatalf("coerce %s: %v", c.raw, err)
		}
		if got != c.want {
			t.Errorf("coerce %s: want %v, got %v", c.raw, c.want, got)
		}
	}
	// 非法布尔
	if _, err := coerce(json.RawMessage(`"maybe"`), TypeBool, OpEq); err == nil {
		t.Error("invalid bool string should error")
	}
	if _, err := coerce(json.RawMessage(`123`), TypeBool, OpEq); err == nil {
		t.Error("numeric bool should error")
	}
}

// TestCoerceInt 覆盖整数强转：JSON 数字、字符串数字、非法值。
func TestCoerceInt(t *testing.T) {
	v, err := coerce(json.RawMessage(`42`), TypeInt, OpEq)
	if err != nil || v != 42 {
		t.Fatalf("int: %v %v", v, err)
	}
	v, err = coerce(json.RawMessage(`"42"`), TypeInt, OpEq)
	if err != nil || v != 42 {
		t.Fatalf("int string: %v %v", v, err)
	}
	if _, err := coerce(json.RawMessage(`"abc"`), TypeInt, OpEq); err == nil {
		t.Error("non-numeric int string should error")
	}
	if _, err := coerce(json.RawMessage(`1.5`), TypeInt, OpEq); err == nil {
		t.Error("float int should error")
	}
}

// TestCoerceUint 覆盖无符号整数强转。
func TestCoerceUint(t *testing.T) {
	v, err := coerce(json.RawMessage(`7`), TypeUint, OpEq)
	if err != nil || v != uint64(7) {
		t.Fatalf("uint: %v %v", v, err)
	}
	v, err = coerce(json.RawMessage(`"7"`), TypeUint, OpEq)
	if err != nil || v != uint64(7) {
		t.Fatalf("uint string: %v %v", v, err)
	}
	if _, err := coerce(json.RawMessage(`"-1"`), TypeUint, OpEq); err == nil {
		t.Error("negative uint should error")
	}
	if _, err := coerce(json.RawMessage(`"abc"`), TypeUint, OpEq); err == nil {
		t.Error("non-numeric uint string should error")
	}
}

// TestCoerceTime 覆盖 RFC3339 时间解析。
func TestCoerceTime(t *testing.T) {
	v, err := coerce(json.RawMessage(`"2026-01-02T15:04:05Z"`), TypeTime, OpEq)
	if err != nil {
		t.Fatal(err)
	}
	if tm, ok := v.(time.Time); !ok || tm.Year() != 2026 {
		t.Errorf("unexpected time: %v", v)
	}
	if _, err := coerce(json.RawMessage(`"not-a-time"`), TypeTime, OpEq); err == nil {
		t.Error("invalid time should error")
	}
	if _, err := coerce(json.RawMessage(`123`), TypeTime, OpEq); err == nil {
		t.Error("non-string time should error")
	}
}

// TestCoerceString 覆盖字符串与类型不匹配。
func TestCoerceString(t *testing.T) {
	v, err := coerce(json.RawMessage(`"hello"`), TypeString, OpEq)
	if err != nil || v != "hello" {
		t.Fatalf("string: %v %v", v, err)
	}
	if _, err := coerce(json.RawMessage(`123`), TypeString, OpEq); err == nil {
		t.Error("numeric string should error")
	}
}

// TestCoerceIn 覆盖 IN 操作符：合法数组、非数组、元素类型错误。
func TestCoerceIn(t *testing.T) {
	v, err := coerce(json.RawMessage(`[1,2,3]`), TypeInt, OpIn)
	if err != nil {
		t.Fatal(err)
	}
	arr, ok := v.([]any)
	if !ok || len(arr) != 3 {
		t.Fatalf("in result: %v", v)
	}
	if _, err := coerce(json.RawMessage(`{"a":1}`), TypeInt, OpIn); err == nil {
		t.Error("non-array in value should error")
	}
	if _, err := coerce(json.RawMessage(`[1,"x"]`), TypeInt, OpIn); err == nil {
		t.Error("array with bad element should error")
	}
}

// TestCoerceEqNil 覆盖 JSON null 值：字符串类型下解析为 ""（不报错）。
func TestCoerceEqNil(t *testing.T) {
	v, err := coerce(json.RawMessage(`null`), TypeString, OpEq)
	if err != nil {
		t.Fatal(err)
	}
	if v != "" {
		t.Errorf("null string coerces to empty; got %#v", v)
	}
}

// TestNeAndRangeOps 覆盖 ne / gt / lt / lte 操作符在 Parse 中的生效路径。
func TestNeAndRangeOps(t *testing.T) {
	db := setupDB(t)
	// ne
	q, err := Parse(newCtx(`filter={"category":{"ne":"site"}}`), testDefs)
	if err != nil {
		t.Fatal(err)
	}
	var items []item
	if total, err := q.Find(db.Model(&item{}), &items); err != nil || total != 2 {
		t.Fatalf("ne filter: err=%v total=%d", err, total)
	}
	// lt
	q, _ = Parse(newCtx(`filter={"sort":{"lt":3}}`), testDefs)
	items = nil
	if total, _ := q.Find(db.Model(&item{}), &items); total != 2 {
		t.Errorf("lt filter: total=%d", total)
	}
	// lte
	q, _ = Parse(newCtx(`filter={"sort":{"lte":2}}`), testDefs)
	items = nil
	if total, _ := q.Find(db.Model(&item{}), &items); total != 2 {
		t.Errorf("lte filter: total=%d", total)
	}
	// gt
	q, _ = Parse(newCtx(`filter={"sort":{"gt":3}}`), testDefs)
	items = nil
	if total, _ := q.Find(db.Model(&item{}), &items); total != 1 {
		t.Errorf("gt filter: total=%d", total)
	}
}
