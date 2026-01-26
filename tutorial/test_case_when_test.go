package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== CASE WHEN with Aggregate Tests ====================

// TestCaseWhenWithAggregates tests CASE WHEN with aggregate functions
func TestCaseWhenWithAggregates(t *testing.T) {
	o := OrderSchema
	setupTable(t, o.ModelType())
	db := getDB()

	orders := []Order{
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 50.00, Status: "completed"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 150.00, Status: "completed"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 350.00, Status: "pending"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 750.00, Status: "shipped"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 1500.00, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("Conditional counting with CASE WHEN", func(t *testing.T) {
		// MySQL: SELECT
		//          SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed_count,
		//          SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending_count,
		//          SUM(CASE WHEN status = 'shipped' THEN 1 ELSE 0 END) AS shipped_count
		//        FROM orders
		completedSum := gsql.Int(
			gsql.Cases.Int().
				When(o.Status.Eq("completed"), gsql.IntVal(1)).
				Else(gsql.IntVal(0)),
		).Sum().As("completed_count")

		pendingSum := gsql.Int(
			gsql.Cases.Int().
				When(o.Status.Eq("pending"), gsql.IntVal(1)).
				Else(gsql.IntVal(0)),
		).Sum().As("pending_count")

		shippedSum := gsql.Int(
			gsql.Cases.Int().
				When(o.Status.Eq("shipped"), gsql.IntVal(1)).
				Else(gsql.IntVal(0)),
		).Sum().As("shipped_count")

		type Result struct {
			CompletedCount int64 `gorm:"column:completed_count"`
			PendingCount   int64 `gorm:"column:pending_count"`
			ShippedCount   int64 `gorm:"column:shipped_count"`
		}

		var result Result
		err := gsql.Select(completedSum, pendingSum, shippedSum).
			From(&o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.CompletedCount != 3 {
			t.Errorf("Expected 3 completed, got %d", result.CompletedCount)
		}
		if result.PendingCount != 1 {
			t.Errorf("Expected 1 pending, got %d", result.PendingCount)
		}
		if result.ShippedCount != 1 {
			t.Errorf("Expected 1 shipped, got %d", result.ShippedCount)
		}
	})

	t.Run("Conditional sum with CASE WHEN", func(t *testing.T) {
		// MySQL: SELECT
		//          SUM(CASE WHEN total_price < 100 THEN total_price ELSE 0 END) AS small_total,
		//          SUM(CASE WHEN total_price >= 100 AND total_price < 500 THEN total_price ELSE 0 END) AS medium_total,
		//          SUM(CASE WHEN total_price >= 500 THEN total_price ELSE 0 END) AS large_total
		//        FROM orders
		smallTotal := gsql.Float(
			gsql.Cases.Float().
				When(o.TotalPrice.Lt(100), o.TotalPrice.Expr()).
				Else(gsql.FloatVal(0.0)),
		).Sum().As("small_total")

		mediumTotal := gsql.Float(
			gsql.Cases.Float().
				When(gsql.And(o.TotalPrice.Gte(100), o.TotalPrice.Lt(500)), o.TotalPrice.Expr()).
				Else(gsql.FloatVal(0.0)),
		).Sum().As("medium_total")

		largeTotal := gsql.Float(
			gsql.Cases.Float().
				When(o.TotalPrice.Gte(500), o.TotalPrice.Expr()).
				Else(gsql.FloatVal(0.0)),
		).Sum().As("large_total")

		type Result struct {
			SmallTotal  float64 `gorm:"column:small_total"`
			MediumTotal float64 `gorm:"column:medium_total"`
			LargeTotal  float64 `gorm:"column:large_total"`
		}

		var result Result
		err := gsql.Select(smallTotal, mediumTotal, largeTotal).
			From(&o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// small: 50, medium: 150+350=500, large: 750+1500=2250
		if result.SmallTotal != 50 {
			t.Errorf("Expected small_total 50, got %f", result.SmallTotal)
		}
		if result.MediumTotal != 500 {
			t.Errorf("Expected medium_total 500, got %f", result.MediumTotal)
		}
		if result.LargeTotal != 2250 {
			t.Errorf("Expected large_total 2250, got %f", result.LargeTotal)
		}
	})
}
