package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== String Function Tests ====================

// TestStringFunctions_Extended tests more string functions
func TestStringFunctions_Extended(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data with various string patterns
	products := []Product{
		{Name: "iPhone 15 Pro", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook Air M3", Category: "Electronics", Price: 1299.99, Stock: 50},
		{Name: "AirPods Max", Category: "Electronics", Price: 549.99, Stock: 200},
		{Name: "T-Shirt Blue", Category: "Clothing", Price: 29.99, Stock: 500},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("Substring function", func(t *testing.T) {
		// MySQL: SELECT SUBSTRING(products.name, 1, 6) AS prefix FROM products
		type Result struct {
			Name   string `gorm:"column:name"`
			Prefix string `gorm:"column:prefix"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Name.Substring(1, 6).As("prefix"),
		).From(&p).
			Where(p.Name.HasPrefix("iPhone")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Prefix != "iPhone" {
			t.Errorf("Expected prefix 'iPhone', got '%s'", results[0].Prefix)
		}
	})

	t.Run("Left and Right functions", func(t *testing.T) {
		// MySQL: SELECT LEFT(products.name, 3) AS left_part, RIGHT(products.name, 3) AS right_part
		type Result struct {
			Name      string `gorm:"column:name"`
			LeftPart  string `gorm:"column:left_part"`
			RightPart string `gorm:"column:right_part"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Name.Left(3).As("left_part"),
			p.Name.Right(3).As("right_part"),
		).From(&p).
			Where(p.Name.Eq("AirPods Max")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].LeftPart != "Air" {
			t.Errorf("Expected left part 'Air', got '%s'", results[0].LeftPart)
		}
		if results[0].RightPart != "Max" {
			t.Errorf("Expected right part 'Max', got '%s'", results[0].RightPart)
		}
	})

	t.Run("Locate function", func(t *testing.T) {
		// MySQL: SELECT products.name, LOCATE(' ', products.name) AS space_pos FROM products
		type Result struct {
			Name     string `gorm:"column:name"`
			SpacePos int    `gorm:"column:space_pos"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Name.Locate(" ").As("space_pos"),
		).From(&p).
			Where(p.Name.HasPrefix("iPhone")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].SpacePos != 7 {
			t.Errorf("Expected space position 7, got %d", results[0].SpacePos)
		}
	})

	t.Run("Reverse function", func(t *testing.T) {
		// MySQL: SELECT products.category, REVERSE(products.category) AS reversed FROM products
		type Result struct {
			Category string `gorm:"column:category"`
			Reversed string `gorm:"column:reversed"`
		}

		var results []Result
		err := gsql.Select(
			p.Category,
			p.Category.Reverse().As("reversed"),
		).From(&p).
			Where(p.Category.Eq("Clothing")).
			Limit(1).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Reversed != "gnihtolC" {
			t.Errorf("Expected reversed 'gnihtolC', got '%s'", results[0].Reversed)
		}
	})

	t.Run("Repeat function", func(t *testing.T) {
		// MySQL: SELECT REPEAT('*', 5) AS stars FROM products LIMIT 1
		type Result struct {
			Stars string `gorm:"column:stars"`
		}

		var results []Result
		err := gsql.Select(
			gsql.StringVal("*").Repeat(5).As("stars"),
		).From(&p).
			Limit(1).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if results[0].Stars != "*****" {
			t.Errorf("Expected stars '*****', got '%s'", results[0].Stars)
		}
	})

	t.Run("LPad and RPad functions", func(t *testing.T) {
		// MySQL: SELECT LPAD(products.stock, 6, '0') AS padded_stock FROM products
		type Result struct {
			ID          uint64 `gorm:"column:id"`
			PaddedStock string `gorm:"column:padded_stock"`
		}

		var results []Result
		err := gsql.Select(
			p.ID,
			p.Stock.CastChar().LPad(6, "0").As("padded_stock"),
		).From(&p).
			Where(p.Stock.Eq(100)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].PaddedStock != "000100" {
			t.Errorf("Expected padded stock '000100', got '%s'", results[0].PaddedStock)
		}
	})

	t.Run("ConcatWS function", func(t *testing.T) {
		// MySQL: SELECT CONCAT_WS(' - ', products.name, products.category) AS full_name FROM products
		type Result struct {
			FullName string `gorm:"column:full_name"`
		}

		var results []Result
		err := gsql.Select(
			p.Name.ConcatWS(" - ", p.Category).As("full_name"),
		).From(&p).
			Where(p.Name.HasPrefix("T-Shirt")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].FullName != "T-Shirt Blue - Clothing" {
			t.Errorf("Expected 'T-Shirt Blue - Clothing', got '%s'", results[0].FullName)
		}
	})

	t.Run("CharLength function", func(t *testing.T) {
		// MySQL: SELECT products.name, CHAR_LENGTH(products.name) AS name_len FROM products
		type Result struct {
			Name    string `gorm:"column:name"`
			NameLen int    `gorm:"column:name_len"`
		}

		var results []Result
		err := gsql.Select(
			p.Name,
			p.Name.CharLength().As("name_len"),
		).From(&p).
			Where(p.Name.Eq("AirPods Max")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].NameLen != 11 {
			t.Errorf("Expected name length 11, got %d", results[0].NameLen)
		}
	})
}
