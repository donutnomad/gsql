package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== IFF (Typed IF) Tests ====================

// TestIFF tests the IFF function for typed conditional expressions
func TestIFF(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "InStock", Category: "Test", Price: 100, Stock: 50},
		{Name: "OutOfStock", Category: "Test", Price: 200, Stock: 0},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("IFF with string result", func(t *testing.T) {
		type Result struct {
			Name   string `gorm:"column:name"`
			Status string `gorm:"column:status"`
		}

		// Use IF for typed IF expression
		statusExpr := gsql.IF(
			p.Stock.Gt(0),
			p.Name.Upper(), // In stock - use uppercase name
			p.Name.Lower(), // Out of stock - use lowercase name
		).As("status")

		var results []Result
		err := gsql.Select(p.Name, statusExpr).
			From(&p).
			OrderBy(p.Name.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		for _, r := range results {
			if r.Name == "InStock" && r.Status != "INSTOCK" {
				t.Errorf("Expected INSTOCK for in-stock product, got '%s'", r.Status)
			}
			if r.Name == "OutOfStock" && r.Status != "outofstock" {
				t.Errorf("Expected outofstock for out-of-stock product, got '%s'", r.Status)
			}
		}
	})
}
