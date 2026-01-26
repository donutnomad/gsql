package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Complex Subquery Tests ====================

// TestComplexSubqueries tests various subquery patterns
func TestComplexSubqueries(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	p := ProductSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	customers := []Customer{
		{Name: "Alice", Email: "alice@test.com", Phone: "111-1111"},
		{Name: "Bob", Email: "bob@test.com", Phone: "222-2222"},
		{Name: "Charlie", Email: "charlie@test.com", Phone: "333-3333"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 500.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 300.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999.99, Stock: 50},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("Scalar subquery in SELECT using LEFT JOIN", func(t *testing.T) {
		// Instead of correlated subquery in SELECT, use LEFT JOIN with aggregation
		// MySQL: SELECT customers.name, COUNT(orders.id) AS order_count
		//        FROM customers
		//        LEFT JOIN orders ON customers.id = orders.customer_id
		//        GROUP BY customers.id, customers.name
		type Result struct {
			Name       string `gorm:"column:name"`
			OrderCount int64  `gorm:"column:order_count"`
		}

		var results []Result
		err := gsql.Select(
			c.Name,
			gsql.COUNT(o.ID).As("order_count"),
		).From(&c).
			Join(gsql.LeftJoin(&o).On(c.ID.EqF(o.CustomerID))).
			GroupBy(c.ID, c.Name).
			OrderBy(c.Name.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		for _, r := range results {
			switch r.Name {
			case "Alice":
				if r.OrderCount != 2 {
					t.Errorf("Expected Alice to have 2 orders, got %d", r.OrderCount)
				}
			case "Bob":
				if r.OrderCount != 1 {
					t.Errorf("Expected Bob to have 1 order, got %d", r.OrderCount)
				}
			case "Charlie":
				if r.OrderCount != 0 {
					t.Errorf("Expected Charlie to have 0 orders, got %d", r.OrderCount)
				}
			}
		}
	})

	t.Run("Subquery with comparison to aggregate", func(t *testing.T) {
		// Find orders above average price
		// MySQL: SELECT * FROM orders WHERE total_price > (SELECT AVG(total_price) FROM orders)
		avgSubquery := gsql.Select(o.TotalPrice.Avg().As("avg_price")).From(&o)

		var results []Order
		err := gsql.Select(o.AllFields()...).
			From(&o).
			Where(o.TotalPrice.GtF(avgSubquery.ToExpr())).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Average is 300, so only 500 order should match
		if len(results) != 1 {
			t.Errorf("Expected 1 result above average, got %d", len(results))
		}
		if results[0].TotalPrice != 500 {
			t.Errorf("Expected order with price 500, got %f", results[0].TotalPrice)
		}
	})

	t.Run("Subquery with MAX", func(t *testing.T) {
		// Find the customer with the highest single order
		// MySQL: SELECT * FROM customers WHERE id = (SELECT customer_id FROM orders ORDER BY total_price DESC LIMIT 1)
		maxOrderSubquery := gsql.Select(o.CustomerID).
			From(&o).
			OrderBy(o.TotalPrice.Desc()).
			Limit(1)

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(c.ID.EqF(maxOrderSubquery.ToExpr())).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if len(results) > 0 && results[0].Name != "Alice" {
			t.Errorf("Expected Alice (highest order 500), got %s", results[0].Name)
		}
	})

	t.Run("Multiple subqueries in WHERE", func(t *testing.T) {
		// Find customers who:
		// 1. Have at least one order
		// 2. Their orders are above the average
		// MySQL: SELECT * FROM customers
		//        WHERE id IN (SELECT customer_id FROM orders WHERE total_price > (SELECT AVG(total_price) FROM orders))
		avgSubquery := gsql.Select(o.TotalPrice.Avg().As("avg_price")).From(&o)
		customerSubquery := gsql.Select(o.CustomerID).
			From(&o).
			Where(o.TotalPrice.GtF(avgSubquery.ToExpr()))

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(c.ID.InSubquery(customerSubquery.ToExpr())).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 customer with above-average order, got %d", len(results))
		}
	})
}
