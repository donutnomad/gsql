package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Type Conversion Tests ====================

// TestTypeConversions tests CAST and type conversion functions
func TestTypeConversions(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "Product 123", Category: "Test", Price: 99.567, Stock: 42},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("CastSigned and CastUnsigned", func(t *testing.T) {
		// MySQL: SELECT CAST(products.price AS SIGNED) AS price_int
		type Result struct {
			PriceInt int64 `gorm:"column:price_int"`
		}

		var results []Result
		err := gsql.Select(
			p.Price.CastSigned().As("price_int"),
		).From(&p).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].PriceInt != 100 { // 99.567 rounds to 100
			t.Errorf("Expected price_int 100, got %d", results[0].PriceInt)
		}
	})

	t.Run("CastDecimal", func(t *testing.T) {
		// MySQL: SELECT CAST(products.price AS DECIMAL(10,1)) AS price_decimal
		type Result struct {
			PriceDecimal float64 `gorm:"column:price_decimal"`
		}

		var results []Result
		err := gsql.Select(
			p.Price.CastDecimal(10, 1).As("price_decimal"),
		).From(&p).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if results[0].PriceDecimal != 99.6 {
			t.Errorf("Expected price_decimal 99.6, got %f", results[0].PriceDecimal)
		}
	})

	t.Run("Hex and Bin functions", func(t *testing.T) {
		// MySQL: SELECT HEX(products.stock) AS hex_val, BIN(products.stock) AS bin_val
		type Result struct {
			HexVal string `gorm:"column:hex_val"`
			BinVal string `gorm:"column:bin_val"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock.Hex().As("hex_val"),
			p.Stock.Bin().As("bin_val"),
		).From(&p).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// 42 in hex is 2A, in binary is 101010
		if results[0].HexVal != "2A" {
			t.Errorf("Expected hex '2A', got '%s'", results[0].HexVal)
		}
		if results[0].BinVal != "101010" {
			t.Errorf("Expected bin '101010', got '%s'", results[0].BinVal)
		}
	})
}
