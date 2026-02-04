package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== Complex UNION ALL with Subquery Tests ====================
// This test covers a complex real-world scenario similar to:
// - UNION ALL of two SELECT statements
// - Subquery as derived table
// - Multiple IF conditions
// - LEFT JOIN with multiple tables
// - NOT EXISTS subquery
// - Dynamic scopes for filtering

// TestComplexUnionAllWithSubquery tests a complex query pattern combining:
// 1. UNION ALL of two different queries
// 2. IF conditional expressions
// 3. LEFT JOIN with multiple tables
// 4. NOT EXISTS subquery
// 5. Using the union result as a derived table with filtering
func TestComplexUnionAllWithSubquery(t *testing.T) {
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
		{Name: "Charlie", Email: "charlie@test.com", Phone: "333-3333"}, // No orders
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 500.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 150.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 300.00, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("UNION ALL with CASE expressions and derived table", func(t *testing.T) {
		// Result type for the combined query
		type CombinedResult struct {
			ID         uint64  `gorm:"column:id"`
			Name       string  `gorm:"column:name"`
			Email      string  `gorm:"column:email"`
			TotalPrice float64 `gorm:"column:total_price"`
			Status     string  `gorm:"column:status"`
		}

		// Query 1: Customers with orders (using CASE to show conditional value)
		// Similar to: CASE WHEN status = 'completed' THEN total_price ELSE 0 END AS total_price
		totalPriceExpr := gsql.Cases.Float().
			When(o.Status.Eq("completed"), o.TotalPrice.Expr()).
			Else(gsql.FloatVal(0.0))

		query1 := gsql.Select(
			c.ID,
			c.Name,
			c.Email,
			totalPriceExpr.As("total_price"),
			o.Status,
		).From(&c).
			Join(gsql.LeftJoin(&o).On(c.ID.EqF(o.CustomerID)))

		// Query 2: Customers without orders (default values)
		notExistsSubquery := gsql.Select(gsql.Lit(1).As("_")).
			From(&o).
			Where(o.CustomerID.EqF(c.ID))

		query2 := gsql.Select(
			c.ID,
			c.Name,
			c.Email,
			gsql.Lit(0).As("total_price"),
			gsql.Lit("no_order").As("status"),
		).From(&c).
			Where(gsql.NotExists(notExistsSubquery))

		// Combine with UNION ALL
		unionAll := gsql.UnionAll(query1, query2)

		// Create derived table
		derivedTable := gsql.DefineTable[any, CombinedResult]("combined", CombinedResult{}, unionAll)

		// Query from derived table with filtering
		nameCol := gsql.StringFieldOf[string]("combined", "name")
		statusCol := gsql.IntFieldOf[string]("combined", "status")

		var results []CombinedResult
		err := gsql.Select(
			gsql.Field("id"),
			gsql.Field("name"),
			gsql.Field("email"),
			gsql.Field("total_price"),
			gsql.Field("status"),
		).From(&derivedTable).
			Where(
				gsql.Or(
					statusCol.Eq("completed"),
					statusCol.Eq("no_order"),
				),
			).
			OrderBy(nameCol.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// Verify results
		t.Logf("Got %d results", len(results))
		for _, r := range results {
			t.Logf("ID: %d, Name: %s, Email: %s, TotalPrice: %.2f, Status: %s",
				r.ID, r.Name, r.Email, r.TotalPrice, r.Status)
		}

		// Should have results for customers with completed orders or no orders
		if len(results) == 0 {
			t.Error("Expected some results but got none")
		}
	})

	t.Run("Complex CASE expression with multiple conditions", func(t *testing.T) {
		type PriceTier struct {
			ID       uint64 `gorm:"column:id"`
			Name     string `gorm:"column:name"`
			Tier     string `gorm:"column:tier"`
			Discount string `gorm:"column:discount"`
		}

		// Use Case-When for more complex conditions
		// Similar to: CASE WHEN status = 'completed' AND total_price > 400 THEN 'VIP' ...
		tierExpr := gsql.Cases.String().
			When(gsql.And(o.Status.Eq("completed"), o.TotalPrice.Gt(400)), gsql.StringVal("VIP")).
			When(o.TotalPrice.Gt(200), gsql.StringVal("Premium")).
			Else(gsql.StringVal("Standard"))

		discountExpr := gsql.Cases.String().
			When(o.TotalPrice.Gt(400), gsql.StringVal("30%")).
			When(o.TotalPrice.Gt(200), gsql.StringVal("15%")).
			Else(gsql.StringVal("5%"))

		var results []PriceTier
		err := gsql.Select(
			c.ID,
			c.Name,
			tierExpr.As("tier"),
			discountExpr.As("discount"),
		).From(&c).
			Join(gsql.InnerJoin(&o).On(c.ID.EqF(o.CustomerID))).
			OrderBy(c.Name.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Got %d results", len(results))
		for _, r := range results {
			t.Logf("ID: %d, Name: %s, Tier: %s, Discount: %s", r.ID, r.Name, r.Tier, r.Discount)
		}

		// Alice should have VIP tier (500 > 400 and completed)
		foundAliceVIP := false
		for _, r := range results {
			if r.Name == "Alice" && r.Tier == "VIP" && r.Discount == "30%" {
				foundAliceVIP = true
				break
			}
		}
		if !foundAliceVIP {
			t.Error("Expected Alice with VIP tier and 30% discount")
		}
	})

	t.Run("Multiple LEFT JOINs with aggregation", func(t *testing.T) {
		type CustomerSummary struct {
			ID           uint64  `gorm:"column:id"`
			Name         string  `gorm:"column:name"`
			OrderCount   int64   `gorm:"column:order_count"`
			TotalSpent   float64 `gorm:"column:total_spent"`
			AvgOrderSize float64 `gorm:"column:avg_order_size"`
		}

		var results []CustomerSummary
		err := gsql.Select(
			c.ID,
			c.Name,
			gsql.COUNT(o.ID).As("order_count"),
			o.TotalPrice.Sum().IfNull(0).As("total_spent"),
			o.TotalPrice.Avg().IfNull(0).As("avg_order_size"),
		).From(&c).
			Join(gsql.LeftJoin(&o).On(c.ID.EqF(o.CustomerID))).
			GroupBy(c.ID, c.Name).
			OrderBy(c.Name.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 customers, got %d", len(results))
		}

		for _, r := range results {
			t.Logf("ID: %d, Name: %s, Orders: %d, Total: %.2f, Avg: %.2f",
				r.ID, r.Name, r.OrderCount, r.TotalSpent, r.AvgOrderSize)
		}

		// Verify Alice has 2 orders
		for _, r := range results {
			if r.Name == "Alice" {
				if r.OrderCount != 2 {
					t.Errorf("Expected Alice to have 2 orders, got %d", r.OrderCount)
				}
				if r.TotalSpent != 650.00 {
					t.Errorf("Expected Alice total spent 650.00, got %.2f", r.TotalSpent)
				}
			}
			if r.Name == "Charlie" {
				if r.OrderCount != 0 {
					t.Errorf("Expected Charlie to have 0 orders, got %d", r.OrderCount)
				}
			}
		}
	})

	t.Run("Subquery in WHERE with IN clause", func(t *testing.T) {
		// Find customers who have high-value orders (> 200)
		// Similar to: WHERE customer_id IN (SELECT customer_id FROM orders WHERE total_price > 200)
		highValueCustomers := gsql.Select(o.CustomerID).
			From(&o).
			Where(o.TotalPrice.Gt(200))

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(c.ID.InSubquery(highValueCustomers.ToExpr())).
			OrderBy(c.Name.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 customers with high-value orders, got %d", len(results))
		}

		for _, r := range results {
			t.Logf("Customer: %s", r.Name)
		}
	})
}

// TestComplexUnionWithDynamicFiltering tests UNION with dynamic scopes-like filtering
func TestComplexUnionWithDynamicFiltering(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema

	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// Insert test data
	customers := []Customer{
		{Name: "Alice Smith", Email: "alice@test.com", Phone: "111-1111"},
		{Name: "Bob Jones", Email: "bob@test.com", Phone: "222-2222"},
		{Name: "Alice Johnson", Email: "alicej@test.com", Phone: "333-3333"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 200.00, Status: "pending"},
		{CustomerID: customers[2].ID, OrderDate: time.Now(), TotalPrice: 300.00, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("Dynamic filtering like scopes", func(t *testing.T) {
		type CombinedResult struct {
			ID         uint64  `gorm:"column:id"`
			Name       string  `gorm:"column:name"`
			TotalPrice float64 `gorm:"column:total_price"`
			Status     string  `gorm:"column:status"`
		}

		// Build base query
		baseQuery := gsql.Select(
			c.ID,
			c.Name,
			o.TotalPrice,
			o.Status,
		).From(&c).
			Join(gsql.LeftJoin(&o).On(c.ID.EqF(o.CustomerID)))

		derivedTable := gsql.DefineTable[any, CombinedResult]("combined", CombinedResult{}, baseQuery)

		nameCol := gsql.StringFieldOf[string]("combined", "name")
		statusCol := gsql.IntFieldOf[string]("combined", "status")
		priceCol := gsql.IntFieldOf[float64]("combined", "total_price")

		// Simulate dynamic filtering (like scopes)
		filterName := "Alice"       // Could be empty in real scenario
		filterStatus := "completed" // Could be empty
		minPrice := 50.0            // Could be 0

		query := gsql.Select(
			gsql.Field("id"),
			gsql.Field("name"),
			gsql.Field("total_price"),
			gsql.Field("status"),
		).From(&derivedTable)

		// Apply filters conditionally (similar to scopes.Option pattern)
		var conditions []gsql.Expression
		if filterName != "" {
			conditions = append(conditions, nameCol.Contains(filterName))
		}
		if filterStatus != "" {
			conditions = append(conditions, statusCol.Eq(filterStatus))
		}
		if minPrice > 0 {
			conditions = append(conditions, priceCol.Gte(minPrice))
		}

		if len(conditions) > 0 {
			query = query.Where(conditions...)
		}

		query = query.OrderBy(nameCol.Asc())

		var results []CombinedResult
		err := query.Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		t.Logf("Got %d results", len(results))
		for _, r := range results {
			t.Logf("ID: %d, Name: %s, TotalPrice: %.2f, Status: %s",
				r.ID, r.Name, r.TotalPrice, r.Status)
		}

		// Should find both Alice records with completed status and price >= 50
		if len(results) != 2 {
			t.Errorf("Expected 2 results (both Alice with completed orders), got %d", len(results))
		}
	})
}
