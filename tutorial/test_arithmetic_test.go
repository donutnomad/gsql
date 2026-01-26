package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Arithmetic Operations Tests ====================

// TestArithmeticOperations tests arithmetic operations on fields
func TestArithmeticOperations(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "Product A", Category: "Test", Price: 100, Stock: 50},
		{Name: "Product B", Category: "Test", Price: 200, Stock: 30},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("Add and Sub operations", func(t *testing.T) {
		type Result struct {
			Name       string  `gorm:"column:name"`
			PricePlus  float64 `gorm:"column:price_plus"`
			PriceMinus float64 `gorm:"column:price_minus"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Price.Add(10).As("price_plus"),
			p.Price.Sub(10).As("price_minus"),
		).From(&p).
			Where(p.Price.Eq(100)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].PricePlus != 110 {
			t.Errorf("Expected price_plus 110, got %f", results[0].PricePlus)
		}
		if results[0].PriceMinus != 90 {
			t.Errorf("Expected price_minus 90, got %f", results[0].PriceMinus)
		}
	})

	t.Run("Mul and Div operations", func(t *testing.T) {
		type Result struct {
			Name       string  `gorm:"column:name"`
			PriceMul   float64 `gorm:"column:price_mul"`
			PriceDiv   float64 `gorm:"column:price_div"`
			TotalValue float64 `gorm:"column:total_value"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Price.Mul(2).As("price_mul"),
			p.Price.Div(4).As("price_div"),
			p.Price.Mul(p.Stock).As("total_value"),
		).From(&p).
			Where(p.Price.Eq(100)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].PriceMul != 200 {
			t.Errorf("Expected price_mul 200, got %f", results[0].PriceMul)
		}
		if results[0].PriceDiv != 25 {
			t.Errorf("Expected price_div 25, got %f", results[0].PriceDiv)
		}
		if results[0].TotalValue != 5000 {
			t.Errorf("Expected total_value 5000, got %f", results[0].TotalValue)
		}
	})

	t.Run("Mod and Neg operations", func(t *testing.T) {
		type Result struct {
			Stock  int `gorm:"column:stock"`
			ModVal int `gorm:"column:mod_val"`
			NegVal int `gorm:"column:neg_val"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock,
			p.Stock.Mod(7).As("mod_val"),
			p.Stock.Neg().As("neg_val"),
		).From(&p).
			Where(p.Stock.Eq(50)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// 50 MOD 7 = 1
		if results[0].ModVal != 1 {
			t.Errorf("Expected mod_val 1, got %d", results[0].ModVal)
		}
		if results[0].NegVal != -50 {
			t.Errorf("Expected neg_val -50, got %d", results[0].NegVal)
		}
	})
}
