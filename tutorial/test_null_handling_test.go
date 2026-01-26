package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== NULL Handling Tests ====================

// TestNullHandling tests IFNULL, COALESCE, NULLIF functions
func TestNullHandling(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "Product A", Category: "Test", Price: 100, Stock: 50},
		{Name: "Product B", Category: "Test", Price: 200, Stock: 0},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Update one product to have NULL description using gsql API
	gsql.Select(p.AllFields()...).
		From(&p).
		Where(p.Name.Eq("Product A")).
		Update(db, map[string]any{"description": nil})

	gsql.Select(p.AllFields()...).
		From(&p).
		Where(p.Name.Eq("Product B")).
		Update(db, map[string]any{"description": "Has description"})

	t.Run("IfNull function", func(t *testing.T) {
		// MySQL: SELECT products.name, COALESCE(products.description, 'No description') AS desc
		type Result struct {
			Name string `gorm:"column:name"`
			Desc string `gorm:"column:desc"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Description.IfNull("No description").As("desc"),
		).From(&p).
			OrderBy(p.Name.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		for _, r := range results {
			if r.Name == "Product A" && r.Desc != "No description" {
				t.Errorf("Expected 'No description' for Product A, got '%s'", r.Desc)
			}
			if r.Name == "Product B" && r.Desc != "Has description" {
				t.Errorf("Expected 'Has description' for Product B, got '%s'", r.Desc)
			}
		}
	})

	t.Run("NullIf function", func(t *testing.T) {
		// MySQL: SELECT products.name, NULLIF(products.stock, 0) AS stock_or_null
		type Result struct {
			Name        string `gorm:"column:name"`
			StockOrNull *int   `gorm:"column:stock_or_null"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Stock.NullIf(0).As("stock_or_null"),
		).From(&p).
			OrderBy(p.Name.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		for _, r := range results {
			if r.Name == "Product A" && (r.StockOrNull == nil || *r.StockOrNull != 50) {
				t.Errorf("Expected stock 50 for Product A")
			}
			if r.Name == "Product B" && r.StockOrNull != nil {
				t.Errorf("Expected NULL stock for Product B (stock was 0)")
			}
		}
	})
}

// ==================== IS NULL / IS NOT NULL Tests ====================

// TestIsNullIsNotNull tests NULL checking conditions
func TestIsNullIsNotNull(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "Product A", Category: "Test", Price: 100, Stock: 50},
		{Name: "Product B", Category: "Test", Price: 200, Stock: 100},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Set one product's description to NULL explicitly using gsql API
	gsql.Select(p.AllFields()...).
		From(&p).
		Where(p.Name.Eq("Product A")).
		Update(db, map[string]any{"description": nil})

	gsql.Select(p.AllFields()...).
		From(&p).
		Where(p.Name.Eq("Product B")).
		Update(db, map[string]any{"description": "Has desc"})

	t.Run("IsNull condition", func(t *testing.T) {
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Description.IsNull()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 product with NULL description, got %d", len(results))
		}
		if results[0].Name != "Product A" {
			t.Errorf("Expected Product A, got %s", results[0].Name)
		}
	})

	t.Run("IsNotNull condition", func(t *testing.T) {
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Description.IsNotNull()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 product with non-NULL description, got %d", len(results))
		}
		if results[0].Name != "Product B" {
			t.Errorf("Expected Product B, got %s", results[0].Name)
		}
	})
}
