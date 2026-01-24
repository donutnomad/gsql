package fields

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

	if cnt.Base.TableName() != "sub" {
		t.Errorf("Expected tableName 'sub', got '%s'", cnt.Base.TableName())
	}
	if cnt.Base.ColumnName() != "cnt" {
		t.Errorf("Expected columnName 'cnt', got '%s'", cnt.Base.ColumnName())
	}

	// Verify comparison works
	_ = cnt.Gt(10)
}

func TestFloatColumn_From(t *testing.T) {
	sub := mockTable{tableName: "derived"}
	total := FloatColumn("total").From(sub)

	if total.Base.TableName() != "derived" {
		t.Errorf("Expected tableName 'derived', got '%s'", total.Base.TableName())
	}
	if total.Base.ColumnName() != "total" {
		t.Errorf("Expected columnName 'total', got '%s'", total.Base.ColumnName())
	}

	// Verify comparison works
	_ = total.Gte(100.5)
}

func TestStringColumn_From(t *testing.T) {
	sub := mockTable{tableName: "names"}
	name := StringColumn("name").From(sub)

	if name.Base.TableName() != "names" {
		t.Errorf("Expected tableName 'names', got '%s'", name.Base.TableName())
	}

	// Verify pattern matching works
	_ = name.Like("%test%")
}

func TestBoolColumn_From(t *testing.T) {
	sub := mockTable{tableName: "flags"}
	active := BoolColumn("active").From(sub)

	if active.Base.TableName() != "flags" {
		t.Errorf("Expected tableName 'flags', got '%s'", active.Base.TableName())
	}
	if active.Base.ColumnName() != "active" {
		t.Errorf("Expected columnName 'active', got '%s'", active.Base.ColumnName())
	}

	// Verify comparison works
	_ = active.Eq(true)
}

func TestTimeColumn_From(t *testing.T) {
	sub := mockTable{tableName: "events"}
	createdAt := TimeColumn("created_at").From(sub)

	if createdAt.Base.TableName() != "events" {
		t.Errorf("Expected tableName 'events', got '%s'", createdAt.Base.TableName())
	}
	if createdAt.Base.ColumnName() != "created_at" {
		t.Errorf("Expected columnName 'created_at', got '%s'", createdAt.Base.ColumnName())
	}
}

func TestGenericColumn_From(t *testing.T) {
	sub := mockTable{tableName: "custom"}
	custom := Column[uint64]("custom_id").From(sub)

	if custom.Base.TableName() != "custom" {
		t.Errorf("Expected tableName 'custom', got '%s'", custom.Base.TableName())
	}
	if custom.Base.ColumnName() != "custom_id" {
		t.Errorf("Expected columnName 'custom_id', got '%s'", custom.Base.ColumnName())
	}

	// Verify comparison works with generic type
	_ = custom.Eq(100)
}
