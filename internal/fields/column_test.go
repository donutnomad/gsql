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
	cnt := IntColumn[int]("cnt").From(sub)

	if cnt.TableName() != "sub" {
		t.Errorf("Expected tableName 'sub', got '%s'", cnt.TableName())
	}
	if cnt.ColumnName() != "cnt" {
		t.Errorf("Expected columnName 'cnt', got '%s'", cnt.ColumnName())
	}

	// Verify comparison works
	_ = cnt.Gt(10)
}

func TestFloatColumn_From(t *testing.T) {
	sub := mockTable{tableName: "derived"}
	total := FloatColumn[any]("total").From(sub)

	if total.TableName() != "derived" {
		t.Errorf("Expected tableName 'derived', got '%s'", total.TableName())
	}
	if total.ColumnName() != "total" {
		t.Errorf("Expected columnName 'total', got '%s'", total.ColumnName())
	}

	// Verify comparison works
	_ = total.Gte(100.5)
}

func TestStringColumn_From(t *testing.T) {
	sub := mockTable{tableName: "names"}
	name := StringColumn[string]("name").From(sub)

	if name.TableName() != "names" {
		t.Errorf("Expected tableName 'names', got '%s'", name.TableName())
	}

	// Verify pattern matching works
	_ = name.Like("%test%")
}

func TestBoolColumn_From(t *testing.T) {
	sub := mockTable{tableName: "flags"}
	active := BoolColumn("active").From(sub)

	if active.TableName() != "flags" {
		t.Errorf("Expected tableName 'flags', got '%s'", active.TableName())
	}
	if active.ColumnName() != "active" {
		t.Errorf("Expected columnName 'active', got '%s'", active.ColumnName())
	}

	// Verify comparison works
	_ = active.Eq(true)
}

func TestTimeColumn_From(t *testing.T) {
	sub := mockTable{tableName: "events"}
	createdAt := TimeColumn[any]("created_at").From(sub)

	if createdAt.TableName() != "events" {
		t.Errorf("Expected tableName 'events', got '%s'", createdAt.TableName())
	}
	if createdAt.ColumnName() != "created_at" {
		t.Errorf("Expected columnName 'created_at', got '%s'", createdAt.ColumnName())
	}
}

func TestGenericColumn_From(t *testing.T) {
	sub := mockTable{tableName: "custom"}
	custom := Column[uint64]("custom_id").From(sub)

	if custom.TableName() != "custom" {
		t.Errorf("Expected tableName 'custom', got '%s'", custom.TableName())
	}
	if custom.ColumnName() != "custom_id" {
		t.Errorf("Expected columnName 'custom_id', got '%s'", custom.ColumnName())
	}

	// Verify comparison works with generic type
	_ = custom.Eq(100)
}
