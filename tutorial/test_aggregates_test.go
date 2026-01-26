package tutorial

import (
	"strings"
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== COUNT_IF Tests ====================

// TestCountIF tests the COUNT_IF function
func TestCountIF(t *testing.T) {
	o := OrderSchema
	setupTable(t, o.ModelType())
	db := getDB()

	orders := []Order{
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 50, Status: "completed"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 150, Status: "completed"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 250, Status: "pending"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 350, Status: "shipped"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 450, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("COUNT_IF with single condition", func(t *testing.T) {
		// MySQL: SELECT COUNT(IF(status = 'completed', 1, NULL)) AS completed_count
		type Result struct {
			CompletedCount int64 `gorm:"column:completed_count"`
		}

		var result Result
		err := gsql.Select(
			gsql.COUNT_IF(o.Status.Eq("completed")).As("completed_count"),
		).From(&o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.CompletedCount != 3 {
			t.Errorf("Expected 3 completed orders, got %d", result.CompletedCount)
		}
	})

	t.Run("COUNT_IF with price condition", func(t *testing.T) {
		// MySQL: SELECT COUNT(IF(total_price > 200, 1, NULL)) AS high_value_count
		type Result struct {
			HighValueCount int64 `gorm:"column:high_value_count"`
		}

		var result Result
		err := gsql.Select(
			gsql.COUNT_IF(o.TotalPrice.Gt(200)).As("high_value_count"),
		).From(&o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.HighValueCount != 3 {
			t.Errorf("Expected 3 high value orders, got %d", result.HighValueCount)
		}
	})
}

// ==================== GROUP_CONCAT Tests ====================

// TestGroupConcat tests the GROUP_CONCAT function
func TestGroupConcat(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999, Stock: 50},
		{Name: "AirPods", Category: "Electronics", Price: 199, Stock: 200},
		{Name: "T-Shirt", Category: "Clothing", Price: 29, Stock: 500},
		{Name: "Jeans", Category: "Clothing", Price: 79, Stock: 300},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("GROUP_CONCAT with default separator", func(t *testing.T) {
		type Result struct {
			Category string `gorm:"column:category"`
			Names    string `gorm:"column:names"`
		}

		var results []Result
		err := gsql.Select(
			p.Category,
			gsql.GROUP_CONCAT(p.Name).As("names"),
		).From(&p).
			GroupBy(p.Category).
			OrderBy(p.Category.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 categories, got %d", len(results))
		}
		for _, r := range results {
			if r.Category == "Clothing" {
				if !strings.Contains(r.Names, "T-Shirt") || !strings.Contains(r.Names, "Jeans") {
					t.Errorf("Expected T-Shirt and Jeans in names, got '%s'", r.Names)
				}
			}
		}
	})

	t.Run("GROUP_CONCAT with custom separator", func(t *testing.T) {
		type Result struct {
			Category string `gorm:"column:category"`
			Names    string `gorm:"column:names"`
		}

		var results []Result
		err := gsql.Select(
			p.Category,
			gsql.GROUP_CONCAT(p.Name, " | ").As("names"),
		).From(&p).
			GroupBy(p.Category).
			Where(p.Category.Eq("Clothing")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if !strings.Contains(results[0].Names, " | ") {
			t.Errorf("Expected ' | ' separator in names, got '%s'", results[0].Names)
		}
	})
}

// ==================== COUNT_DISTINCT Tests ====================

// TestCountDistinct tests the COUNT_DISTINCT function
func TestCountDistinct(t *testing.T) {
	o := OrderSchema
	setupTable(t, o.ModelType())
	db := getDB()

	orders := []Order{
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 100, Status: "completed"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 200, Status: "pending"},
		{CustomerID: 2, OrderDate: time.Now(), TotalPrice: 150, Status: "completed"},
		{CustomerID: 2, OrderDate: time.Now(), TotalPrice: 250, Status: "completed"},
		{CustomerID: 3, OrderDate: time.Now(), TotalPrice: 50, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("COUNT_DISTINCT customers", func(t *testing.T) {
		type Result struct {
			UniqueCustomers int64 `gorm:"column:unique_customers"`
		}

		var result Result
		err := gsql.Select(
			gsql.COUNT_DISTINCT(o.CustomerID).As("unique_customers"),
		).From(&o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.UniqueCustomers != 3 {
			t.Errorf("Expected 3 unique customers, got %d", result.UniqueCustomers)
		}
	})

	t.Run("COUNT_DISTINCT statuses", func(t *testing.T) {
		type Result struct {
			UniqueStatuses int64 `gorm:"column:unique_statuses"`
		}

		var result Result
		err := gsql.Select(
			gsql.COUNT_DISTINCT(o.Status).As("unique_statuses"),
		).From(&o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.UniqueStatuses != 2 {
			t.Errorf("Expected 2 unique statuses, got %d", result.UniqueStatuses)
		}
	})

	t.Run("COUNT_DISTINCT with field method", func(t *testing.T) {
		type Result struct {
			UniqueCustomers int64 `gorm:"column:unique_customers"`
		}

		var result Result
		err := gsql.Select(
			o.CustomerID.CountDistinct().As("unique_customers"),
		).From(&o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.UniqueCustomers != 3 {
			t.Errorf("Expected 3 unique customers, got %d", result.UniqueCustomers)
		}
	})
}
