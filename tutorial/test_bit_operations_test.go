package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Bit Operations Tests ====================

// TestBitOperations tests bit operations on integers
func TestBitOperations(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "Product A", Category: "Test", Price: 100, Stock: 15}, // binary: 1111
		{Name: "Product B", Category: "Test", Price: 100, Stock: 8},  // binary: 1000
		{Name: "Product C", Category: "Test", Price: 100, Stock: 3},  // binary: 0011
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("BitAnd operation", func(t *testing.T) {
		// MySQL: SELECT products.stock, products.stock & 7 AS masked FROM products
		type Result struct {
			Stock  int `gorm:"column:stock"`
			Masked int `gorm:"column:masked"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock,
			p.Stock.BitAnd(7).As("masked"), // mask with 0111
		).From(&p).
			OrderBy(p.Stock.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// 3 & 7 = 3, 8 & 7 = 0, 15 & 7 = 7
		found := false
		for _, r := range results {
			if r.Stock == 8 && r.Masked == 0 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected stock=8 with masked=0")
		}
	})

	t.Run("BitOr operation", func(t *testing.T) {
		// MySQL: SELECT products.stock | 16 AS with_flag FROM products
		type Result struct {
			Stock    int `gorm:"column:stock"`
			WithFlag int `gorm:"column:with_flag"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock,
			p.Stock.BitOr(16).As("with_flag"),
		).From(&p).
			Where(p.Stock.Eq(3)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 3 | 16 = 19
		if results[0].WithFlag != 19 {
			t.Errorf("Expected with_flag 19, got %d", results[0].WithFlag)
		}
	})

	t.Run("LeftShift and RightShift", func(t *testing.T) {
		// MySQL: SELECT products.stock << 2 AS shifted_left, products.stock >> 1 AS shifted_right
		type Result struct {
			Stock        int `gorm:"column:stock"`
			ShiftedLeft  int `gorm:"column:shifted_left"`
			ShiftedRight int `gorm:"column:shifted_right"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock,
			p.Stock.LeftShift(2).As("shifted_left"),
			p.Stock.RightShift(1).As("shifted_right"),
		).From(&p).
			Where(p.Stock.Eq(8)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 8 << 2 = 32, 8 >> 1 = 4
		if results[0].ShiftedLeft != 32 {
			t.Errorf("Expected shifted_left 32, got %d", results[0].ShiftedLeft)
		}
		if results[0].ShiftedRight != 4 {
			t.Errorf("Expected shifted_right 4, got %d", results[0].ShiftedRight)
		}
	})

	t.Run("IntDiv operation", func(t *testing.T) {
		// MySQL: SELECT products.stock DIV 4 AS div_result FROM products
		type Result struct {
			Stock     int `gorm:"column:stock"`
			DivResult int `gorm:"column:div_result"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock,
			p.Stock.IntDiv(4).As("div_result"),
		).From(&p).
			Where(p.Stock.Eq(15)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 15 DIV 4 = 3
		if results[0].DivResult != 3 {
			t.Errorf("Expected div_result 3, got %d", results[0].DivResult)
		}
	})
}
