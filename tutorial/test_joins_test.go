package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Comprehensive JOIN Tests ====================

// TestComplexJoins tests more complex JOIN scenarios
func TestComplexJoins(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	oi := OrderItemSchema
	p := ProductSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	setupTable(t, oi.ModelType())
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert comprehensive test data
	customers := []Customer{
		{Name: "Alice", Email: "alice@test.com", Phone: "111"},
		{Name: "Bob", Email: "bob@test.com", Phone: "222"},
		{Name: "Charlie", Email: "charlie@test.com", Phone: "333"}, // No orders
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	products := []Product{
		{Name: "Laptop", Category: "Electronics", Price: 1000, Stock: 50},
		{Name: "Mouse", Category: "Electronics", Price: 50, Stock: 200},
		{Name: "Keyboard", Category: "Electronics", Price: 80, Stock: 150},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 1050, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 80, Status: "pending"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 50, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	orderItems := []OrderItem{
		{OrderID: orders[0].ID, ProductID: products[0].ID, Quantity: 1, UnitPrice: 1000},
		{OrderID: orders[0].ID, ProductID: products[1].ID, Quantity: 1, UnitPrice: 50},
		{OrderID: orders[1].ID, ProductID: products[2].ID, Quantity: 1, UnitPrice: 80},
		{OrderID: orders[2].ID, ProductID: products[1].ID, Quantity: 1, UnitPrice: 50},
	}
	if err := db.Create(&orderItems).Error; err != nil {
		t.Fatalf("Failed to create order items: %v", err)
	}

	t.Run("Self-referential data with aggregation", func(t *testing.T) {
		// Find total spending per customer
		// MySQL: SELECT customers.name, SUM(orders.total_price) AS total_spent
		//        FROM customers
		//        LEFT JOIN orders ON customers.id = orders.customer_id
		//        GROUP BY customers.id, customers.name
		type Result struct {
			Name       string   `gorm:"column:name"`
			TotalSpent *float64 `gorm:"column:total_spent"`
		}

		var results []Result
		err := gsql.Select(
			c.Name,
			o.TotalPrice.Sum().As("total_spent"),
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
				if r.TotalSpent == nil || *r.TotalSpent != 1130 {
					t.Errorf("Expected Alice total 1130, got %v", r.TotalSpent)
				}
			case "Bob":
				if r.TotalSpent == nil || *r.TotalSpent != 50 {
					t.Errorf("Expected Bob total 50, got %v", r.TotalSpent)
				}
			case "Charlie":
				if r.TotalSpent != nil {
					t.Errorf("Expected Charlie total NULL, got %v", *r.TotalSpent)
				}
			}
		}
	})

	t.Run("JOIN with multiple conditions", func(t *testing.T) {
		// Find completed orders with items
		// MySQL: SELECT c.name, o.total_price, oi.quantity, p.name AS product_name
		//        FROM order_items oi
		//        INNER JOIN orders o ON oi.order_id = o.id AND o.status = 'completed'
		//        INNER JOIN customers c ON o.customer_id = c.id
		//        INNER JOIN products p ON oi.product_id = p.id
		type Result struct {
			CustomerName string  `gorm:"column:customer_name"`
			TotalPrice   float64 `gorm:"column:total_price"`
			Quantity     int     `gorm:"column:quantity"`
			ProductName  string  `gorm:"column:product_name"`
		}

		var results []Result
		err := gsql.Select(
			c.Name.As("customer_name"),
			o.TotalPrice,
			oi.Quantity,
			p.Name.As("product_name"),
		).From(&oi).
			Join(
				gsql.InnerJoin(&o).On(oi.OrderID.EqF(o.ID)).And(o.Status.Eq("completed")),
				gsql.InnerJoin(&c).On(o.CustomerID.EqF(c.ID)),
				gsql.InnerJoin(&p).On(oi.ProductID.EqF(p.ID)),
			).
			OrderBy(o.TotalPrice.Desc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Should have 3 items from completed orders
		if len(results) != 3 {
			t.Errorf("Expected 3 results from completed orders, got %d", len(results))
		}
	})
}
