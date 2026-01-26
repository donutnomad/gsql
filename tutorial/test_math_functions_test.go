package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Math Function Tests ====================

// TestMathFunctions_Extended tests mathematical functions
func TestMathFunctions_Extended(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "Product A", Category: "Test", Price: 100.567, Stock: 45},
		{Name: "Product B", Category: "Test", Price: 200.123, Stock: 30},
		{Name: "Product C", Category: "Test", Price: 50.999, Stock: 100},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("Log and Exp functions", func(t *testing.T) {
		// MySQL: SELECT LOG(products.price) AS log_price, EXP(1) AS e FROM products
		type Result struct {
			Price    float64 `gorm:"column:price"`
			LogPrice float64 `gorm:"column:log_price"`
			E        float64 `gorm:"column:e"`
		}

		var results []Result
		err := gsql.Select(
			p.Price,
			p.Price.Log().As("log_price"),
			gsql.FloatVal(1.0).Exp().As("e"),
		).From(&p).
			Where(p.Price.Gt(100)).
			Limit(1).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		// log(100.567) ~ 4.61
		if results[0].LogPrice < 4.5 || results[0].LogPrice > 4.7 {
			t.Errorf("Expected log_price around 4.61, got %f", results[0].LogPrice)
		}
		// e ~ 2.718
		if results[0].E < 2.7 || results[0].E > 2.8 {
			t.Errorf("Expected e around 2.718, got %f", results[0].E)
		}
	})

	t.Run("Pow and Sqrt functions", func(t *testing.T) {
		// MySQL: SELECT POW(products.stock, 2) AS squared, SQRT(products.stock) AS sqrt FROM products
		type Result struct {
			Stock   int     `gorm:"column:stock"`
			Squared float64 `gorm:"column:squared"`
			Sqrt    float64 `gorm:"column:sqrt"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock,
			p.Stock.Pow(2).As("squared"),
			p.Stock.Sqrt().As("sqrt"),
		).From(&p).
			Where(p.Stock.Eq(100)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Squared != 10000 {
			t.Errorf("Expected squared 10000, got %f", results[0].Squared)
		}
		if results[0].Sqrt != 10 {
			t.Errorf("Expected sqrt 10, got %f", results[0].Sqrt)
		}
	})

	t.Run("Sign function", func(t *testing.T) {
		// MySQL: SELECT SIGN(products.stock - 50) AS sign_val FROM products
		type Result struct {
			Stock   int   `gorm:"column:stock"`
			SignVal int64 `gorm:"column:sign_val"`
		}

		var results []Result
		err := gsql.Select(
			p.Stock,
			p.Stock.Sub(50).Sign().As("sign_val"),
		).From(&p).
			OrderBy(p.Stock.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		// stock=30 -> sign(-20)=-1, stock=45 -> sign(-5)=-1, stock=100 -> sign(50)=1
		found := false
		for _, r := range results {
			if r.Stock == 100 && r.SignVal == 1 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find stock=100 with sign=1")
		}
	})

	t.Run("Greatest and Least functions", func(t *testing.T) {
		// MySQL: SELECT GREATEST(products.price, 100) AS max_price, LEAST(products.stock, 50) AS min_stock FROM products
		type Result struct {
			Price    float64 `gorm:"column:price"`
			MaxPrice float64 `gorm:"column:max_price"`
			MinStock int     `gorm:"column:min_stock"`
		}

		var results []Result
		err := gsql.Select(
			p.Price,
			p.Price.Greatest(100.0).As("max_price"),
			p.Stock.Least(50).As("min_stock"),
		).From(&p).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		for _, r := range results {
			// max_price should always be >= 100
			if r.MaxPrice < 100 {
				t.Errorf("Expected max_price >= 100, got %f for price %f", r.MaxPrice, r.Price)
			}
		}
	})

	t.Run("Truncate function", func(t *testing.T) {
		// MySQL: SELECT TRUNCATE(products.price, 1) AS truncated FROM products
		type Result struct {
			Price     float64 `gorm:"column:price"`
			Truncated float64 `gorm:"column:truncated"`
		}

		var results []Result
		err := gsql.Select(
			p.Price,
			p.Price.Truncate(1).As("truncated"),
		).From(&p).
			Where(p.Price.Eq(100.567)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Truncated != 100.5 {
			t.Errorf("Expected truncated 100.5, got %f", results[0].Truncated)
		}
	})
}
