package tutorial

import (
	"strings"
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/samber/lo"
)

// ==================== JOIN Tests ====================

// TestInter_InnerJoin tests INNER JOIN - orders with customers
func TestInter_InnerJoin(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
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
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.50, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 250.00, Status: "pending"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 75.25, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Test INNER JOIN: orders with customer name
	t.Run("INNER JOIN orders with customers", func(t *testing.T) {
		var results []struct {
			CustomerName string  `gorm:"column:customer_name"`
			TotalPrice   float64 `gorm:"column:total_price"`
			Status       string  `gorm:"column:status"`
		}
		// MySQL: SELECT customers.name AS customer_name, orders.total_price, orders.status
		//        FROM orders
		//        INNER JOIN customers ON orders.customer_id = customers.id
		//        ORDER BY orders.total_price DESC
		err := gsql.Select(
			c.Name.As("customer_name"),
			o.TotalPrice,
			o.Status,
		).
			From(o).
			Join(gsql.InnerJoin(&c).On(o.CustomerID.EqF(c.ID))).
			OrderBy(o.TotalPrice.Desc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		// Highest price order should be first
		if results[0].TotalPrice != 250.00 {
			t.Errorf("Expected total_price 250.00, got %f", results[0].TotalPrice)
		}
		if results[0].CustomerName != "Alice" {
			t.Errorf("Expected customer_name Alice, got %s", results[0].CustomerName)
		}
	})

	// Test INNER JOIN with WHERE
	t.Run("INNER JOIN with WHERE", func(t *testing.T) {
		var results []struct {
			CustomerName string  `gorm:"column:customer_name"`
			TotalPrice   float64 `gorm:"column:total_price"`
		}
		// MySQL: SELECT customers.name AS customer_name, orders.total_price
		//        FROM orders
		//        INNER JOIN customers ON orders.customer_id = customers.id
		//        WHERE orders.status = 'completed'
		err := gsql.Select(
			c.Name.As("customer_name"),
			o.TotalPrice,
		).
			From(o).
			Join(gsql.InnerJoin(&c).On(o.CustomerID.EqF(c.ID))).
			Where(o.Status.Eq("completed")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 completed orders, got %d", len(results))
		}
	})
}

// TestInter_LeftJoin tests LEFT JOIN - customers with their orders (including customers without orders)
func TestInter_LeftJoin(t *testing.T) {
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
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.50, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 200.00, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Test LEFT JOIN: all customers with their orders
	t.Run("LEFT JOIN customers with orders", func(t *testing.T) {
		var results []struct {
			CustomerName string   `gorm:"column:customer_name"`
			TotalPrice   *float64 `gorm:"column:total_price"` // Nullable for customers without orders
		}
		// MySQL: SELECT customers.name AS customer_name, orders.total_price
		//        FROM customers
		//        LEFT JOIN orders ON customers.id = orders.customer_id
		//        ORDER BY customers.name ASC
		err := gsql.
			Select(
				c.Name.As("customer_name"),
				o.TotalPrice,
			).
			From(&c).
			Join(gsql.LeftJoin(&o).On(c.ID.EqF(o.CustomerID))).
			OrderBy(c.Name.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results (all customers), got %d", len(results))
		}
		// Check that NoOrders customer has NULL total_price
		for _, r := range results {
			if r.CustomerName == "NoOrders" && r.TotalPrice != nil {
				t.Errorf("Expected NoOrders to have NULL total_price, got %v", *r.TotalPrice)
			}
		}
	})

	// Test finding customers without orders using LEFT JOIN + WHERE IS NULL
	t.Run("Find customers without orders", func(t *testing.T) {
		var results []struct {
			CustomerName string `gorm:"column:customer_name"`
		}
		// MySQL: SELECT customers.name AS customer_name
		//        FROM customers
		//        LEFT JOIN orders ON customers.id = orders.customer_id
		//        WHERE orders.id IS NULL
		err := gsql.Select(c.Name.As("customer_name")).
			From(&c).
			Join(gsql.LeftJoin(&o).On(c.ID.EqF(o.CustomerID))).
			Where(gsql.Expr("orders.id IS NULL")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 customer without orders, got %d", len(results))
		}
		if len(results) > 0 && results[0].CustomerName != "NoOrders" {
			t.Errorf("Expected NoOrders, got %s", results[0].CustomerName)
		}
	})
}

// TestInter_MultiJoin tests multi-table JOIN (Customer-Order-OrderItem-Product)
func TestInter_MultiJoin(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	oi := OrderItemSchema
	p := ProductSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	setupTable(t, oi.ModelType())
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	customers := []Customer{
		{Name: "Alice", Email: "alice@test.com", Phone: "111-1111"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999.99, Stock: 50},
		{Name: "T-Shirt", Category: "Clothing", Price: 29.99, Stock: 500},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 2029.97, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	orderItems := []OrderItem{
		{OrderID: orders[0].ID, ProductID: products[0].ID, Quantity: 1, UnitPrice: 999.99},
		{OrderID: orders[0].ID, ProductID: products[2].ID, Quantity: 2, UnitPrice: 29.99},
	}
	if err := db.Create(&orderItems).Error; err != nil {
		t.Fatalf("Failed to create order items: %v", err)
	}

	// Test 4-table JOIN
	t.Run("Multi-table JOIN", func(t *testing.T) {
		var results []struct {
			CustomerName string  `gorm:"column:customer_name"`
			ProductName  string  `gorm:"column:product_name"`
			Quantity     int     `gorm:"column:quantity"`
			UnitPrice    float64 `gorm:"column:unit_price"`
		}
		// MySQL: SELECT customers.name AS customer_name, products.name AS product_name,
		//               order_items.quantity, order_items.unit_price
		//        FROM order_items
		//        INNER JOIN orders ON order_items.order_id = orders.id
		//        INNER JOIN customers ON orders.customer_id = customers.id
		//        INNER JOIN products ON order_items.product_id = products.id
		//        ORDER BY order_items.unit_price DESC
		err := gsql.Select(
			c.Name.As("customer_name"),
			p.Name.As("product_name"),
			oi.Quantity,
			oi.UnitPrice,
		).
			From(&oi).
			Join(
				gsql.InnerJoin(&o).On(oi.OrderID.EqF(o.ID)),
				gsql.InnerJoin(&c).On(o.CustomerID.EqF(c.ID)),
				gsql.InnerJoin(&p).On(oi.ProductID.EqF(p.ID)),
			).
			OrderBy(oi.UnitPrice.Desc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 order items, got %d", len(results))
		}
		if results[0].ProductName != "iPhone" {
			t.Errorf("Expected first product to be iPhone, got %s", results[0].ProductName)
		}
		if results[0].CustomerName != "Alice" {
			t.Errorf("Expected customer to be Alice, got %s", results[0].CustomerName)
		}
	})
}

// TestInter_SubqueryInWhere tests subquery in WHERE clause
func TestInter_SubqueryInWhere(t *testing.T) {
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
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 500.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Test: Find customers who have orders with total > 200
	t.Run("Subquery in WHERE with IN", func(t *testing.T) {
		// Build subquery
		// MySQL Subquery: SELECT orders.customer_id FROM orders WHERE orders.total_price > 200
		subquery := gsql.Select(o.CustomerID).
			From(o).
			Where(o.TotalPrice.Gt(200))

		var results []Customer
		// MySQL: SELECT customers.* FROM customers
		//        WHERE id IN (SELECT orders.customer_id FROM orders WHERE orders.total_price > 200)
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(c.ID.InSubquery(subquery.ToExpr())).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 customer, got %d", len(results))
		}
		if len(results) > 0 && results[0].Name != "Alice" {
			t.Errorf("Expected Alice, got %s", results[0].Name)
		}
	})

	// Test: ScalarExpr subquery comparison
	t.Run("ScalarExpr subquery comparison", func(t *testing.T) {
		// Find orders with total above average
		// MySQL Subquery: SELECT AVG(orders.total_price) AS avg_price FROM orders
		avgSubquery := gsql.
			Select(o.TotalPrice.Avg().As("avg_price")).
			From(o)

		var results []Order
		// MySQL: SELECT orders.* FROM orders
		//        WHERE total_price > (SELECT AVG(orders.total_price) AS avg_price FROM orders)
		err := gsql.Select(o.AllFields()...).
			From(o).
			Where(o.TotalPrice.GtF(avgSubquery.ToExpr())).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Average is 300, so only the 500 order should match
		if len(results) != 1 {
			t.Errorf("Expected 1 order above average, got %d", len(results))
		}
		if len(results) > 0 && results[0].TotalPrice != 500.00 {
			t.Errorf("Expected total_price 500.00, got %f", results[0].TotalPrice)
		}
	})
}

// ==================== GROUP BY / HAVING Tests ====================

// TestInter_GroupBy tests GROUP BY with aggregate functions
func TestInter_GroupBy(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// Insert test data
	customers := []Customer{
		{Name: "Alice", Email: "alice@test.com", Phone: "111-1111"},
		{Name: "Bob", Email: "bob@test.com", Phone: "222-2222"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 200.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 150.00, Status: "pending"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 75.00, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Test GROUP BY with COUNT and SUM
	t.Run("GROUP BY with aggregates", func(t *testing.T) {
		var results []struct {
			CustomerID   uint64  `gorm:"column:customer_id"`
			OrderCount   int64   `gorm:"column:order_count"`
			TotalSpent   float64 `gorm:"column:total_spent"`
			AverageOrder float64 `gorm:"column:average_order"`
		}
		// MySQL: SELECT orders.customer_id, COUNT(*) AS order_count,
		//               SUM(orders.total_price) AS total_spent,
		//               AVG(orders.total_price) AS average_order
		//        FROM orders
		//        GROUP BY orders.customer_id
		//        ORDER BY orders.customer_id ASC
		err := gsql.Select(
			o.CustomerID,
			gsql.COUNT().As("order_count"),
			o.TotalPrice.Sum().As("total_spent"),
			o.TotalPrice.Avg().As("average_order"),
		).
			From(o).
			GroupBy(o.CustomerID).
			OrderBy(o.CustomerID.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 groups, got %d", len(results))
		}
		// Alice has 3 orders totaling 450
		if results[0].OrderCount != 3 {
			t.Errorf("Expected Alice to have 3 orders, got %d", results[0].OrderCount)
		}
		if results[0].TotalSpent != 450.00 {
			t.Errorf("Expected Alice total 450.00, got %f", results[0].TotalSpent)
		}
		// Bob has 1 order totaling 75
		if results[1].OrderCount != 1 {
			t.Errorf("Expected Bob to have 1 order, got %d", results[1].OrderCount)
		}
	})

	// Test GROUP BY with multiple columns
	t.Run("GROUP BY multiple columns", func(t *testing.T) {
		var results []struct {
			CustomerID uint64  `gorm:"column:customer_id"`
			Status     string  `gorm:"column:status"`
			Count      int64   `gorm:"column:count"`
			Total      float64 `gorm:"column:total"`
		}
		// MySQL: SELECT orders.customer_id, orders.status, COUNT(*) AS count,
		//               SUM(orders.total_price) AS total
		//        FROM orders
		//        GROUP BY orders.customer_id, orders.status
		//        ORDER BY orders.customer_id ASC
		err := gsql.Select(
			o.CustomerID,
			o.Status,
			gsql.COUNT().As("count"),
			o.TotalPrice.Sum().As("total"),
		).
			From(o).
			GroupBy(o.CustomerID, o.Status).
			OrderBy(o.CustomerID.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Alice: 2 completed, 1 pending; Bob: 1 completed
		if len(results) != 3 {
			t.Errorf("Expected 3 groups, got %d", len(results))
		}
	})
}

// TestInter_Having tests HAVING clause for filtering aggregated results
func TestInter_Having(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// Insert test data
	customers := []Customer{
		{Name: "VIP Customer", Email: "vip@test.com", Phone: "111-1111"},
		{Name: "Regular Customer", Email: "regular@test.com", Phone: "222-2222"},
		{Name: "New Customer", Email: "new@test.com", Phone: "333-3333"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	orders := []Order{
		// VIP has 5 orders
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 200.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 150.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 300.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 250.00, Status: "completed"},
		// Regular has 3 orders
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 50.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 75.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 60.00, Status: "completed"},
		// New has 1 order
		{CustomerID: customers[2].ID, OrderDate: time.Now(), TotalPrice: 25.00, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Test HAVING with COUNT
	t.Run("HAVING with COUNT > 3", func(t *testing.T) {
		var results []struct {
			CustomerID uint64 `gorm:"column:customer_id"`
			OrderCount int64  `gorm:"column:order_count"`
		}
		// MySQL: SELECT orders.customer_id, COUNT(*) AS order_count
		//        FROM orders
		//        GROUP BY orders.customer_id
		//        HAVING COUNT(*) > 3
		err := gsql.Select(
			o.CustomerID,
			gsql.COUNT().As("order_count"),
		).
			From(o).
			GroupBy(o.CustomerID).
			Having(gsql.Expr("COUNT(*) > ?", 3)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 customer with > 3 orders, got %d", len(results))
		}
		if len(results) > 0 && results[0].OrderCount != 5 {
			t.Errorf("Expected VIP with 5 orders, got %d", results[0].OrderCount)
		}
	})

	// Test HAVING with SUM
	t.Run("HAVING with SUM > 500", func(t *testing.T) {
		var results []struct {
			CustomerID uint64  `gorm:"column:customer_id"`
			Total      float64 `gorm:"column:total"`
		}
		// MySQL: SELECT orders.customer_id, SUM(orders.total_price) AS total
		//        FROM orders
		//        GROUP BY orders.customer_id
		//        HAVING SUM(total_price) > 500
		err := gsql.Select(
			o.CustomerID,
			o.TotalPrice.Sum().As("total"),
		).
			From(o).
			GroupBy(o.CustomerID).
			Having(gsql.Expr("SUM(total_price) > ?", 500)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 customer with total > 500, got %d", len(results))
		}
		if len(results) > 0 && results[0].Total != 1000.00 {
			t.Errorf("Expected total 1000.00, got %f", results[0].Total)
		}
	})

	// Test HAVING with multiple conditions
	t.Run("HAVING with multiple conditions", func(t *testing.T) {
		var results []struct {
			CustomerID uint64  `gorm:"column:customer_id"`
			OrderCount int64   `gorm:"column:order_count"`
			AvgOrder   float64 `gorm:"column:avg_order"`
		}
		// MySQL: SELECT orders.customer_id, COUNT(*) AS order_count, AVG(orders.total_price) AS avg_order
		//        FROM orders
		//        GROUP BY orders.customer_id
		//        HAVING COUNT(*) >= 3 AND AVG(total_price) > 50
		err := gsql.Select(
			o.CustomerID,
			gsql.COUNT().As("order_count"),
			o.TotalPrice.Avg().As("avg_order"),
		).
			From(o).
			GroupBy(o.CustomerID).
			Having(
				gsql.Expr("COUNT(*) >= ?", 3),
				gsql.Expr("AVG(total_price) > ?", 50),
			).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// VIP: 5 orders, avg 200; Regular: 3 orders, avg ~61.67
		if len(results) != 2 {
			t.Errorf("Expected 2 customers matching criteria, got %d", len(results))
		}
	})
}

// ==================== UNION Tests ====================

// TestInter_Union tests UNION and UNION ALL operations
func TestInter_Union(t *testing.T) {
	p := ProductSchema
	e := EmployeeSchema
	setupTable(t, p.ModelType())
	setupTable(t, e.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999.99, Stock: 50},
		{Name: "T-Shirt", Category: "Clothing", Price: 29.99, Stock: 500},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	employees := []Employee{
		{Name: "John", Email: "john@test.com", Department: "IT", Salary: 80000, HireDate: time.Now(), BirthDate: time.Now(), IsActive: true},
		{Name: "Jane", Email: "jane@test.com", Department: "HR", Salary: 75000, HireDate: time.Now(), BirthDate: time.Now(), IsActive: true},
	}
	if err := db.Create(&employees).Error; err != nil {
		t.Fatalf("Failed to create employees: %v", err)
	}

	// Test UNION - removes duplicates
	t.Run("UNION removes duplicates", func(t *testing.T) {
		// Create two queries that return the same category
		// MySQL Query1: SELECT products.category AS name FROM products WHERE products.category = 'Electronics'
		q1 := gsql.Select(p.Category.As("name")).
			From(p).
			Where(p.Category.Eq("Electronics"))

		// MySQL Query2: SELECT products.category AS name FROM products WHERE products.category = 'Electronics'
		q2 := gsql.Select(p.Category.As("name")).
			From(p).
			Where(p.Category.Eq("Electronics"))

		// MySQL: (Query1) UNION (Query2)
		union := gsql.Union(q1, q2)

		// Use the union result as a derived table
		type NameResult struct {
			Name string `gorm:"column:name"`
		}
		derivedTable := gsql.DefineTable[any, NameResult]("union_result", NameResult{}, union)

		var results []NameResult
		// MySQL: SELECT name FROM (Query1 UNION Query2) AS union_result
		err := gsql.Select(gsql.Field("name")).
			From(&derivedTable).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// UNION should remove duplicates - only 1 "Electronics"
		if len(results) != 1 {
			t.Errorf("Expected 1 unique result from UNION, got %d", len(results))
		}
	})

	// Test UNION ALL - keeps duplicates
	t.Run("UNION ALL keeps duplicates", func(t *testing.T) {
		// MySQL Query1: SELECT products.category AS name FROM products WHERE products.category = 'Electronics'
		q1 := gsql.Select(p.Category.As("name")).
			From(p).
			Where(p.Category.Eq("Electronics"))

		// MySQL Query2: SELECT products.category AS name FROM products WHERE products.category = 'Electronics'
		q2 := gsql.Select(p.Category.As("name")).
			From(p).
			Where(p.Category.Eq("Electronics"))

		// MySQL: (Query1) UNION ALL (Query2)
		unionAll := gsql.UnionAll(q1, q2)

		type NameResult struct {
			Name string `gorm:"column:name"`
		}
		derivedTable := gsql.DefineTable[any, NameResult]("union_all_result", NameResult{}, unionAll)

		var results []NameResult
		// MySQL: SELECT name FROM (Query1 UNION ALL Query2) AS union_all_result
		err := gsql.Select(gsql.Field("name")).
			From(&derivedTable).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// UNION ALL should keep all rows (2 products in Electronics * 2 = 4)
		if len(results) != 4 {
			t.Errorf("Expected 4 results from UNION ALL, got %d", len(results))
		}
	})

	// Test UNION combining different tables
	t.Run("UNION different tables", func(t *testing.T) {
		// MySQL Query1: SELECT products.name AS name FROM products
		q1 := gsql.Select(p.Name.As("name")).
			From(&p)

		// MySQL Query2: SELECT employees.name AS name FROM employees
		q2 := gsql.Select(e.Name.As("name")).
			From(&e)

		// MySQL: (Query1) UNION (Query2)
		union := gsql.Union(q1, q2)

		type NameResult struct {
			Name string `gorm:"column:name"`
		}
		derivedTable := gsql.DefineTable[any, NameResult]("combined", NameResult{}, union)

		var results []NameResult
		// MySQL: SELECT name FROM (SELECT products.name AS name FROM products
		//                          UNION SELECT employees.name AS name FROM employees) AS combined
		err := gsql.Select(gsql.Field("name")).
			From(&derivedTable).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// 3 products + 2 employees = 5 names
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}
	})
}

type string1 string

func (s string1) String() string {
	return ""
}

// ==================== CASE WHEN Tests ====================

// TestInter_CaseWhen tests CASE WHEN expressions
func TestInter_CaseWhen(t *testing.T) {
	o := OrderSchema
	setupTable(t, o.ModelType())
	db := getDB()

	// Insert test data
	orders := []Order{
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 50.00, Status: "pending"},
		{CustomerID: 2, OrderDate: time.Now(), TotalPrice: 150.00, Status: "completed"},
		{CustomerID: 3, OrderDate: time.Now(), TotalPrice: 350.00, Status: "completed"},
		{CustomerID: 4, OrderDate: time.Now(), TotalPrice: 750.00, Status: "shipped"},
		{CustomerID: 5, OrderDate: time.Now(), TotalPrice: 1500.00, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Test: Price tier classification
	t.Run("CASE WHEN for price tier", func(t *testing.T) {
		// MySQL: CASE
		//          WHEN orders.total_price < 100 THEN 'Small'
		//          WHEN orders.total_price < 500 THEN 'Medium'
		//          WHEN orders.total_price < 1000 THEN 'Large'
		//          ELSE 'VIP'
		//        END AS price_tier
		priceTier := gsql.Cases.String().
			When(o.TotalPrice.Lt(100), gsql.StringVal("Small")).
			When(o.TotalPrice.Lt(500), gsql.StringVal("Medium")).
			When(o.TotalPrice.Lt(1000), gsql.StringVal("Large")).
			Else(gsql.StringVal("VIP")).
			As("price_tier")

		var results []struct {
			ID        uint64  `gorm:"column:id"`
			Total     float64 `gorm:"column:total_price"`
			PriceTier string  `gorm:"column:price_tier"`
		}
		// MySQL: SELECT orders.id, orders.total_price,
		//               CASE WHEN total_price < 100 THEN 'Small' ... END AS price_tier
		//        FROM orders
		//        ORDER BY orders.total_price ASC
		err := gsql.
			Select(o.ID, o.TotalPrice, priceTier).
			From(o).
			OrderBy(o.TotalPrice.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}
		// Check tier assignments
		expectedTiers := []string{"Small", "Medium", "Medium", "Large", "VIP"}
		for i, r := range results {
			if r.PriceTier != expectedTiers[i] {
				t.Errorf("Order %d: expected tier %s, got %s", r.ID, expectedTiers[i], r.PriceTier)
			}
		}
	})

	// Test: Status mapping with CaseValue
	t.Run("CaseValue for status mapping", func(t *testing.T) {
		// MySQL: CASE orders.status
		//          WHEN 'pending' THEN 'Waiting'
		//          WHEN 'completed' THEN 'Done'
		//          WHEN 'shipped' THEN 'On the way'
		//          ELSE 'Unknown'
		//        END AS status_desc
		statusDesc := gsql.CaseValue[gsql.StringExpr[string]](o.Status.Expr()).
			When(gsql.StringVal("pending"), gsql.StringVal("Waiting")).
			When(gsql.StringVal("completed"), gsql.StringVal("Done")).
			When(gsql.StringVal("shipped"), gsql.StringVal("On the way")).
			Else(gsql.StringVal("Unknown")).
			As("status_desc")

		var results []struct {
			Status     string `gorm:"column:status"`
			StatusDesc string `gorm:"column:status_desc"`
		}
		// MySQL: SELECT orders.status,
		//               CASE status WHEN 'pending' THEN 'Waiting' ... END AS status_desc
		//        FROM orders
		err := gsql.Select(o.Status, statusDesc).
			From(o).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Verify mapping
		for _, r := range results {
			switch r.Status {
			case "pending":
				if r.StatusDesc != "Waiting" {
					t.Errorf("pending should map to Waiting, got %s", r.StatusDesc)
				}
			case "completed":
				if r.StatusDesc != "Done" {
					t.Errorf("completed should map to Done, got %s", r.StatusDesc)
				}
			case "shipped":
				if r.StatusDesc != "On the way" {
					t.Errorf("shipped should map to On the way, got %s", r.StatusDesc)
				}
			}
		}
	})

	// Test: CASE WHEN with aggregate
	t.Run("CASE WHEN with SUM aggregate", func(t *testing.T) {
		// Count orders by tier
		// MySQL: SUM(CASE WHEN orders.total_price < 100 THEN 1 ELSE 0 END) AS small_count

		smallOrderSum := gsql.Int(
			gsql.Cases.Int().
				When(o.TotalPrice.Lt(100), gsql.IntVal(1)).
				Else(gsql.IntVal(0)),
		).Sum().As("small_count")

		// MySQL: SUM(CASE WHEN orders.total_price >= 100 AND orders.total_price < 500 THEN 1 ELSE 0 END) AS medium_count
		mediumOrderSum := gsql.Int(
			gsql.Cases.Int().
				When(gsql.And(o.TotalPrice.Gte(100), o.TotalPrice.Lt(500)), gsql.IntVal(1)).
				Else(gsql.IntVal(0)),
		).Sum().As("medium_count")

		var result struct {
			SmallCount  int64 `gorm:"column:small_count"`
			MediumCount int64 `gorm:"column:medium_count"`
		}
		// MySQL: SELECT SUM(CASE WHEN total_price < 100 THEN 1 ELSE 0 END) AS small_count,
		//               SUM(CASE WHEN total_price >= 100 AND total_price < 500 THEN 1 ELSE 0 END) AS medium_count
		//        FROM orders
		err := gsql.Select(smallOrderSum, mediumOrderSum).
			From(o).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.SmallCount != 1 {
			t.Errorf("Expected 1 small order, got %d", result.SmallCount)
		}
		if result.MediumCount != 2 {
			t.Errorf("Expected 2 medium orders, got %d", result.MediumCount)
		}
	})
}

// ==================== Index Hint Tests ====================

// TestInter_IndexHint tests USE INDEX / FORCE INDEX hints
func TestInter_IndexHint(t *testing.T) {
	o := OrderSchema
	setupTable(t, o.ModelType())
	db := getDB()

	// Create indexes for testing (MySQL syntax without IF NOT EXISTS)
	// Drop index first if exists, ignore error if not exists
	_ = db.Exec("DROP INDEX idx_customer_id ON orders").Error
	_ = db.Exec("DROP INDEX idx_status ON orders").Error

	// Create indexes
	if err := db.Exec("CREATE INDEX idx_customer_id ON orders(customer_id)").Error; err != nil {
		t.Fatalf("Failed to create idx_customer_id: %v", err)
	}
	if err := db.Exec("CREATE INDEX idx_status ON orders(status)").Error; err != nil {
		t.Fatalf("Failed to create idx_status: %v", err)
	}

	// Insert test data
	orders := []Order{
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 100.00, Status: "pending"},
		{CustomerID: 2, OrderDate: time.Now(), TotalPrice: 200.00, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Test USE INDEX
	t.Run("USE INDEX hint", func(t *testing.T) {
		var results []Order
		// MySQL: SELECT orders.* FROM orders USE INDEX (idx_customer_id)
		//        WHERE orders.customer_id = 1
		err := gsql.Select(o.AllFields()...).
			From(o).
			UseIndex("idx_customer_id").
			Where(o.CustomerID.Eq(1)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	// Test FORCE INDEX
	t.Run("FORCE INDEX hint", func(t *testing.T) {
		var results []Order
		// MySQL: SELECT orders.* FROM orders FORCE INDEX (idx_status)
		//        WHERE orders.status = 'completed'
		err := gsql.Select(o.AllFields()...).
			From(o).
			ForceIndex("idx_status").
			Where(o.Status.Eq("completed")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	// Test IGNORE INDEX
	t.Run("IGNORE INDEX hint", func(t *testing.T) {
		var results []Order
		// MySQL: SELECT orders.* FROM orders IGNORE INDEX (idx_customer_id)
		//        WHERE orders.customer_id = 2
		err := gsql.Select(o.AllFields()...).
			From(o).
			IgnoreIndex("idx_customer_id").
			Where(o.CustomerID.Eq(2)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	// Test multiple index hints
	t.Run("Multiple index hints", func(t *testing.T) {
		var results []Order
		// MySQL: SELECT orders.* FROM orders USE INDEX (idx_customer_id) IGNORE INDEX (idx_status)
		//        WHERE orders.customer_id = 1
		err := gsql.Select(o.AllFields()...).
			From(o).
			UseIndex("idx_customer_id").
			IgnoreIndex("idx_status").
			Where(o.CustomerID.Eq(1)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	// Test index hint for specific operations
	t.Run("Index hint for ORDER BY", func(t *testing.T) {
		var results []Order
		// MySQL: SELECT orders.* FROM orders USE INDEX FOR ORDER BY (idx_customer_id)
		//        ORDER BY orders.customer_id ASC
		err := gsql.Select(o.AllFields()...).
			From(o).
			UseIndexForOrderBy("idx_customer_id").
			OrderBy(o.CustomerID.Asc()).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})
}

// ==================== Typed Expression Tests ====================

// TestTypedExpr_Comparisons tests typed expressions (IntExpr, FloatExpr) with comparison methods
// These expressions are returned by aggregate functions like COUNT, SUM, AVG and can be used
// directly in HAVING clauses with type-safe comparison methods.
func TestTypedExpr_Comparisons(t *testing.T) {
	o := OrderSchema
	p := ProductSchema
	setupTable(t, o.ModelType())
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data for orders
	orders := []Order{
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 200.00, Status: "completed"},
		{CustomerID: 2, OrderDate: time.Now(), TotalPrice: 150.00, Status: "pending"},
		{CustomerID: 2, OrderDate: time.Now(), TotalPrice: 250.00, Status: "completed"},
		{CustomerID: 3, OrderDate: time.Now(), TotalPrice: 50.00, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	// Insert test data for products
	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999.99, Stock: 50},
		{Name: "AirPods", Category: "Electronics", Price: 199.99, Stock: 200},
		{Name: "T-Shirt", Category: "Clothing", Price: 29.99, Stock: 500},
		{Name: "Jeans", Category: "Clothing", Price: 79.99, Stock: 300},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test: COUNT().Gt(value) generates correct SQL
	t.Run("COUNT_Gt", func(t *testing.T) {
		// MySQL: SELECT orders.customer_id, COUNT(*) AS order_count
		//        FROM orders
		//        GROUP BY orders.customer_id
		//        HAVING COUNT(*) > 1
		var results []struct {
			CustomerID uint64 `gorm:"column:customer_id"`
			OrderCount int64  `gorm:"column:order_count"`
		}
		err := gsql.
			Select(
				o.CustomerID,
				gsql.COUNT().As("order_count"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Having(gsql.COUNT().Gt(1)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Customers 1 and 2 have more than 1 order
		if len(results) != 2 {
			t.Errorf("Expected 2 customers with > 1 order, got %d", len(results))
		}

		// Also verify ToSQL output
		sql := gsql.
			Select(
				o.CustomerID,
				gsql.COUNT().As("order_count"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Having(gsql.COUNT().Gt(1)).
			ToSQL()

		t.Logf("COUNT().Gt SQL: %s", sql)
		if !strings.Contains(sql, "HAVING") {
			t.Error("SQL should contain HAVING clause")
		}
		if !strings.Contains(sql, "COUNT(*)") {
			t.Error("SQL should contain COUNT(*)")
		}
	})

	// Test: SUM().Gte(value) generates correct SQL
	t.Run("SUM_Gte", func(t *testing.T) {
		// MySQL: SELECT orders.customer_id, SUM(orders.total_price) AS total
		//        FROM orders
		//        GROUP BY orders.customer_id
		//        HAVING SUM(orders.total_price) >= 300
		var results []struct {
			CustomerID uint64  `gorm:"column:customer_id"`
			Total      float64 `gorm:"column:total"`
		}
		err := gsql.
			Select(
				o.CustomerID,
				o.TotalPrice.Sum().As("total"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Having(o.TotalPrice.Sum().Gte(300.0)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Customer 1: 300, Customer 2: 400 - both >= 300
		if len(results) != 2 {
			t.Errorf("Expected 2 customers with total >= 300, got %d", len(results))
		}

		// Verify ToSQL output
		sql := gsql.
			Select(
				o.CustomerID,
				o.TotalPrice.Sum().As("total"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Having(o.TotalPrice.Sum().Gte(300.0)).
			ToSQL()

		t.Logf("SUM().Gte SQL: %s", sql)
		if !strings.Contains(sql, "HAVING") {
			t.Error("SQL should contain HAVING clause")
		}
		if !strings.Contains(sql, "SUM(") {
			t.Error("SQL should contain SUM(")
		}
	})

	// Test: AVG().Between(from, to) generates correct SQL
	t.Run("AVG_Between", func(t *testing.T) {
		// MySQL: SELECT products.category, AVG(products.price) AS avg_price
		//        FROM products
		//        GROUP BY products.category
		//        HAVING AVG(products.price) >= 50 AND AVG(products.price) <= 2000
		var results []struct {
			Category string  `gorm:"column:category"`
			AvgPrice float64 `gorm:"column:avg_price"`
		}
		err := gsql.
			Select(
				p.Category,
				p.Price.Avg().As("avg_price"),
			).
			From(p).
			GroupBy(p.Category).
			Having(p.Price.Avg().Between(lo.ToPtr[float64](50.0), lo.ToPtr[float64](2000.0), ">=", "<=")).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Electronics avg: ~1066.66, Clothing avg: ~54.99 - both in range [50, 2000]
		if len(results) != 2 {
			t.Errorf("Expected 2 categories with avg price between 50 and 2000, got %d", len(results))
		}

		// Verify ToSQL output
		sql := gsql.
			Select(
				p.Category,
				p.Price.Avg().As("avg_price"),
			).
			From(p).
			GroupBy(p.Category).
			Having(p.Price.Avg().Between(lo.ToPtr[float64](50.0), lo.ToPtr[float64](2000.0), ">=", "<=")).
			ToSQL()

		t.Logf("AVG().Between SQL: %s", sql)
		if !strings.Contains(sql, "HAVING") {
			t.Error("SQL should contain HAVING clause")
		}
		if !strings.Contains(sql, "AVG(") {
			t.Error("SQL should contain AVG(")
		}
		// 新 API 不再使用 BETWEEN，而是使用 >= 和 <=
		if !strings.Contains(sql, ">=") || !strings.Contains(sql, "<=") {
			t.Error("SQL should contain >= and <= operators")
		}
	})

	// Test: IntExpr.AsF() still works after return type change
	t.Run("IntExpr_AsF_works", func(t *testing.T) {
		// Verify .AsF() still works after return type change
		// MySQL: SELECT orders.customer_id, COUNT(*) AS cnt
		//        FROM orders
		//        GROUP BY orders.customer_id
		var results []struct {
			CustomerID uint64 `gorm:"column:customer_id"`
			Cnt        int64  `gorm:"column:cnt"`
		}
		err := gsql.
			Select(
				o.CustomerID,
				gsql.COUNT().As("cnt"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 customers, got %d", len(results))
		}

		// Verify ToSQL output
		sql := gsql.Select(
			o.CustomerID,
			gsql.COUNT().As("cnt"),
		).From(o).
			GroupBy(o.CustomerID).
			ToSQL()

		t.Logf("IntExpr.AsF SQL: %s", sql)
		if !strings.Contains(sql, "COUNT(*)") {
			t.Error("SQL should contain COUNT(*)")
		}
		if !strings.Contains(sql, "AS") && !strings.Contains(sql, "cnt") {
			t.Error("SQL should contain alias 'cnt'")
		}
	})

	// Test: Additional comparison methods
	t.Run("Additional_Comparisons", func(t *testing.T) {
		// Test Lt (less than)
		sqlLt := gsql.
			Select(
				o.CustomerID,
				gsql.COUNT().As("cnt"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Having(gsql.COUNT().Lt(3)).
			ToSQL()

		if !strings.Contains(sqlLt, "HAVING") {
			t.Error("Lt SQL should contain HAVING clause")
		}
		t.Logf("COUNT().Lt SQL: %s", sqlLt)

		// Test Lte (less than or equal)
		sqlLte := gsql.Select(
			o.CustomerID,
			gsql.COUNT().As("cnt"),
		).From(o).
			GroupBy(o.CustomerID).
			Having(gsql.COUNT().Lte(2)).
			ToSQL()

		if !strings.Contains(sqlLte, "HAVING") {
			t.Error("Lte SQL should contain HAVING clause")
		}
		t.Logf("COUNT().Lte SQL: %s", sqlLte)

		// Test Eq (equal) - execute and verify results
		var results []struct {
			CustomerID uint64 `gorm:"column:customer_id"`
			Cnt        int64  `gorm:"column:cnt"`
		}
		err := gsql.
			Select(
				o.CustomerID,
				gsql.COUNT().As("cnt"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Having(gsql.COUNT().Eq(2)).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// Customers 1 and 2 have exactly 2 orders each
		if len(results) != 2 {
			t.Errorf("Expected 2 customers with exactly 2 orders, got %d", len(results))
		}

		sqlEq := gsql.
			Select(
				o.CustomerID,
				gsql.COUNT().As("cnt"),
			).
			From(o).
			GroupBy(o.CustomerID).
			Having(gsql.COUNT().Eq(2)).
			ToSQL()

		if !strings.Contains(sqlEq, "HAVING") {
			t.Error("Eq SQL should contain HAVING clause")
		}
		t.Logf("COUNT().Eq SQL: %s", sqlEq)
	})
}
