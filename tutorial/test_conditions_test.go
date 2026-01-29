package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
	"github.com/samber/lo"
)

// ==================== Between Tests ====================

// TestBetween tests BETWEEN conditions
func TestBetween(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "Cheap", Category: "Test", Price: 10, Stock: 100},
		{Name: "Medium", Category: "Test", Price: 50, Stock: 50},
		{Name: "Expensive", Category: "Test", Price: 100, Stock: 10},
		{Name: "VeryExpensive", Category: "Test", Price: 500, Stock: 5},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("Between condition", func(t *testing.T) {
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Price.Between(lo.ToPtr[float64](40), lo.ToPtr[float64](110), ">=", "<=")).
			OrderBy(p.Price.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 products between 40 and 110, got %d", len(results))
		}
	})

	t.Run("NotBetween condition", func(t *testing.T) {
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Price.NotBetween(lo.ToPtr[float64](40), lo.ToPtr[float64](110), "<", ">")).
			OrderBy(p.Price.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 products not between 40 and 110, got %d", len(results))
		}
	})
}

// ==================== Field Comparison Tests ====================

// TestFieldComparisons tests field-to-field comparisons
func TestFieldComparisons(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "LowMargin", Category: "Test", Price: 100, Stock: 150}, // stock > price
		{Name: "HighMargin", Category: "Test", Price: 200, Stock: 50}, // stock < price
		{Name: "Equal", Category: "Test", Price: 100, Stock: 100},     // stock = price
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("Field greater than field", func(t *testing.T) {
		// Find products where stock > price (numerically)
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Stock.GtF(p.Price.CastSigned())).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 product where stock > price, got %d", len(results))
		}
		if results[0].Name != "LowMargin" {
			t.Errorf("Expected LowMargin, got %s", results[0].Name)
		}
	})

	t.Run("Field equal to field", func(t *testing.T) {
		// Find products where stock = price (numerically)
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Stock.EqF(p.Price.CastSigned())).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 product where stock = price, got %d", len(results))
		}
		if results[0].Name != "Equal" {
			t.Errorf("Expected Equal, got %s", results[0].Name)
		}
	})
}

// ==================== And / Or Combination Tests ====================

// TestAndOrCombinations tests AND and OR condition combinations
func TestAndOrCombinations(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999, Stock: 50},
		{Name: "Mouse", Category: "Electronics", Price: 29, Stock: 500},
		{Name: "T-Shirt", Category: "Clothing", Price: 29, Stock: 200},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("AND combination", func(t *testing.T) {
		// Electronics AND price > 500
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(
				gsql.And(
					p.Category.Eq("Electronics"),
					p.Price.Gt(500),
				),
			).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 electronics > 500, got %d", len(results))
		}
	})

	t.Run("OR combination", func(t *testing.T) {
		// Clothing OR price < 50
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(
				gsql.Or(
					p.Category.Eq("Clothing"),
					p.Price.Lt(50),
				),
			).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// T-Shirt (Clothing) + Mouse (price 29)
		if len(results) != 2 {
			t.Errorf("Expected 2 results (Clothing OR price < 50), got %d", len(results))
		}
	})

	t.Run("Complex AND/OR", func(t *testing.T) {
		// (Electronics AND price > 500) OR (Clothing AND price < 50)
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(
				gsql.Or(
					gsql.And(p.Category.Eq("Electronics"), p.Price.Gt(500)),
					gsql.And(p.Category.Eq("Clothing"), p.Price.Lt(50)),
				),
			).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// iPhone, MacBook (Electronics > 500) + T-Shirt (Clothing < 50)
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})
}
