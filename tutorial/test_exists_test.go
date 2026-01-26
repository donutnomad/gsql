package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Exists / Not Exists Tests ====================

// TestExistsNotExists tests EXISTS and NOT EXISTS subqueries
func TestExistsNotExists(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// Insert test data
	customers := []Customer{
		{Name: "Alice", Email: "alice@test.com", Phone: "111-1111"},
		{Name: "Bob", Email: "bob@test.com", Phone: "222-2222"},
		{Name: "NoOrders", Email: "noorders@test.com", Phone: "999-9999"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 200.00, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("EXISTS - find customers with orders", func(t *testing.T) {
		// Build EXISTS subquery
		// MySQL: SELECT customers.* FROM customers
		//        WHERE EXISTS (SELECT 1 FROM orders WHERE orders.customer_id = customers.id)
		subquery := gsql.Select(gsql.Lit(1).As("_")).
			From(&o).
			Where(o.CustomerID.EqF(c.ID))

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(gsql.Exists(subquery)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 customers with orders, got %d", len(results))
		}
	})

	t.Run("NOT EXISTS - find customers without orders", func(t *testing.T) {
		// MySQL: SELECT customers.* FROM customers
		//        WHERE NOT EXISTS (SELECT 1 FROM orders WHERE orders.customer_id = customers.id)
		subquery := gsql.Select(gsql.Lit(1).As("_")).
			From(&o).
			Where(o.CustomerID.EqF(c.ID))

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(gsql.NotExists(subquery)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 customer without orders, got %d", len(results))
		}
		if results[0].Name != "NoOrders" {
			t.Errorf("Expected NoOrders, got %s", results[0].Name)
		}
	})

	t.Run("EXISTS with condition - customers with high-value orders", func(t *testing.T) {
		// MySQL: SELECT customers.* FROM customers
		//        WHERE EXISTS (SELECT 1 FROM orders WHERE orders.customer_id = customers.id AND orders.total_price > 150)
		subquery := gsql.Select(gsql.Lit(1).As("_")).
			From(&o).
			Where(
				o.CustomerID.EqF(c.ID),
				o.TotalPrice.Gt(150),
			)

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(gsql.Exists(subquery)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 customer with high-value order, got %d", len(results))
		}
		if results[0].Name != "Bob" {
			t.Errorf("Expected Bob, got %s", results[0].Name)
		}
	})
}
