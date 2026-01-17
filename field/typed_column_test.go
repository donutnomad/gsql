package field

import (
	"testing"
)

// mockTable implements TableName() for testing
type mockTable struct {
	tableName string
}

func (m mockTable) TableName() string { return m.tableName }

func TestIntColumn_From(t *testing.T) {
	sub := mockTable{tableName: "sub"}
	cnt := IntColumn("cnt").From(sub)

	if cnt.Base.tableName != "sub" {
		t.Errorf("Expected tableName 'sub', got '%s'", cnt.Base.tableName)
	}
	if cnt.Base.columnName != "cnt" {
		t.Errorf("Expected columnName 'cnt', got '%s'", cnt.Base.columnName)
	}

	// Verify comparison works
	condition := cnt.Gt(10)
	if condition == nil {
		t.Error("Gt should return non-nil expression")
	}
}

func TestFloatColumn_From(t *testing.T) {
	sub := mockTable{tableName: "derived"}
	total := FloatColumn("total").From(sub)

	if total.Base.tableName != "derived" {
		t.Errorf("Expected tableName 'derived', got '%s'", total.Base.tableName)
	}
	if total.Base.columnName != "total" {
		t.Errorf("Expected columnName 'total', got '%s'", total.Base.columnName)
	}

	// Verify comparison works
	condition := total.Gte(100.5)
	if condition == nil {
		t.Error("Gte should return non-nil expression")
	}
}

func TestStringColumn_From(t *testing.T) {
	sub := mockTable{tableName: "names"}
	name := StringColumn("name").From(sub)

	if name.Base.tableName != "names" {
		t.Errorf("Expected tableName 'names', got '%s'", name.Base.tableName)
	}

	// Verify pattern matching works
	condition := name.Like("%test%")
	if condition == nil {
		t.Error("Like should return non-nil expression")
	}
}

func TestBoolColumn_From(t *testing.T) {
	sub := mockTable{tableName: "flags"}
	active := BoolColumn("active").From(sub)

	if active.Base.tableName != "flags" {
		t.Errorf("Expected tableName 'flags', got '%s'", active.Base.tableName)
	}
	if active.Base.columnName != "active" {
		t.Errorf("Expected columnName 'active', got '%s'", active.Base.columnName)
	}

	// Verify comparison works
	condition := active.Eq(true)
	if condition == nil {
		t.Error("Eq should return non-nil expression")
	}
}

func TestTimeColumn_From(t *testing.T) {
	sub := mockTable{tableName: "events"}
	createdAt := TimeColumn("created_at").From(sub)

	if createdAt.Base.tableName != "events" {
		t.Errorf("Expected tableName 'events', got '%s'", createdAt.Base.tableName)
	}
	if createdAt.Base.columnName != "created_at" {
		t.Errorf("Expected columnName 'created_at', got '%s'", createdAt.Base.columnName)
	}
}

func TestGenericColumn_From(t *testing.T) {
	sub := mockTable{tableName: "custom"}
	custom := Column[uint64]("custom_id").From(sub)

	if custom.Base.tableName != "custom" {
		t.Errorf("Expected tableName 'custom', got '%s'", custom.Base.tableName)
	}
	if custom.Base.columnName != "custom_id" {
		t.Errorf("Expected columnName 'custom_id', got '%s'", custom.Base.columnName)
	}

	// Verify comparison works with generic type
	condition := custom.Gt(100)
	if condition == nil {
		t.Error("Gt should return non-nil expression")
	}
}
