package tutorial

import (
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== IF Expression in Update Tests ====================

// TestIF_CaseWhen_Update 测试IF(CASE WHEN)表达式在Update场景中的使用
// 模拟场景: 条件更新，只在满足条件时才更新字段值，类似于:
// UPDATE products SET price = CASE WHEN stock < ? THEN ? ELSE price END WHERE ...
func TestIF_CaseWhen_Update(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	products := []Product{
		{Name: "ProductA", Category: "Electronics", Price: 100, Stock: 5},
		{Name: "ProductB", Category: "Electronics", Price: 200, Stock: 50},
		{Name: "ProductC", Category: "Electronics", Price: 300, Stock: 3},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("Conditional update with IF expression", func(t *testing.T) {
		// 模拟: 只有当 stock < 10 时才将 price 更新为 999，否则保持原值
		// SQL: UPDATE products SET price = CASE WHEN stock < 10 THEN 999 ELSE price END
		//      WHERE category = 'Electronics'
		newPriceExpr := gsql.IF(
			p.Stock.Lt(10),
			gsql.FloatVal[float64](999),
			p.Price.Expr(),
		)

		result := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			Update(db, map[string]any{
				"price": newPriceExpr,
			})
		if result.Error != nil {
			t.Fatalf("Update failed: %v", result.Error)
		}
		// RowsAffected 只计算实际变化的行，ProductB 的 price 保持不变，所以是 2
		if result.RowsAffected != 2 {
			t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
		}

		// 验证结果
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			OrderBy(p.Name.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// ProductA: stock=5 < 10, price 应更新为 999
		if results[0].Price != 999 {
			t.Errorf("ProductA: expected price 999, got %f", results[0].Price)
		}
		// ProductB: stock=50 >= 10, price 应保持 200
		if results[1].Price != 200 {
			t.Errorf("ProductB: expected price 200, got %f", results[1].Price)
		}
		// ProductC: stock=3 < 10, price 应更新为 999
		if results[2].Price != 999 {
			t.Errorf("ProductC: expected price 999, got %f", results[2].Price)
		}
	})

	t.Run("Multiple conditional updates with IF expression", func(t *testing.T) {
		// 重置数据
		db.Exec("UPDATE products SET price = 100, stock = 5 WHERE name = 'ProductA'")
		db.Exec("UPDATE products SET price = 200, stock = 50 WHERE name = 'ProductB'")
		db.Exec("UPDATE products SET price = 300, stock = 3 WHERE name = 'ProductC'")

		// 模拟多字段条件更新，类似于:
		// UPDATE products
		// SET price = CASE WHEN stock < 10 THEN price + 50 ELSE price END,
		//     stock = CASE WHEN stock < 10 THEN stock + 100 ELSE stock END
		// WHERE category = 'Electronics'
		threshold := 10
		newPriceExpr := gsql.IF(
			p.Stock.Lt(threshold),
			p.Price.Add(50),
			p.Price.Expr(),
		)
		newStockExpr := gsql.IF(
			p.Stock.Lt(threshold),
			p.Stock.Add(100),
			p.Stock.Expr(),
		)

		result := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			Update(db, map[string]any{
				"price": newPriceExpr,
				"stock": newStockExpr,
			})
		if result.Error != nil {
			t.Fatalf("Update failed: %v", result.Error)
		}

		// 验证结果
		var results []Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			OrderBy(p.Name.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// ProductA: stock=5 < 10 → price=100+50=150, stock=5+100=105
		if results[0].Price != 150 {
			t.Errorf("ProductA: expected price 150, got %f", results[0].Price)
		}
		if results[0].Stock != 105 {
			t.Errorf("ProductA: expected stock 105, got %d", results[0].Stock)
		}
		// ProductB: stock=50 >= 10 → price=200, stock=50 (不变)
		if results[1].Price != 200 {
			t.Errorf("ProductB: expected price 200, got %f", results[1].Price)
		}
		if results[1].Stock != 50 {
			t.Errorf("ProductB: expected stock 50, got %d", results[1].Stock)
		}
		// ProductC: stock=3 < 10 → price=300+50=350, stock=3+100=103
		if results[2].Price != 350 {
			t.Errorf("ProductC: expected price 350, got %f", results[2].Price)
		}
		if results[2].Stock != 103 {
			t.Errorf("ProductC: expected stock 103, got %d", results[2].Stock)
		}
	})
}

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
