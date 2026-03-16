package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// TestStarWithAlias tests SELECT with multiple Star() and field aliases
// This tests the scenario after removing duplicate field name conflict check
func TestStarWithAlias(t *testing.T) {
	o := OrderSchema
	oi := OrderItemSchema
	setupTable(t, o.ModelType())
	setupTable(t, oi.ModelType())
	db := getDB()

	// Insert test data
	orders := []Order{
		{CustomerID: 1, OrderDate: time.Now(), TotalPrice: 100, Status: "completed"},
		{CustomerID: 2, OrderDate: time.Now(), TotalPrice: 200, Status: "pending"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	orderItems := []OrderItem{
		{OrderID: orders[0].ID, ProductID: 1, Quantity: 2, UnitPrice: 50},
		{OrderID: orders[0].ID, ProductID: 2, Quantity: 1, UnitPrice: 100},
		{OrderID: orders[1].ID, ProductID: 1, Quantity: 3, UnitPrice: 50},
	}
	if err := db.Create(&orderItems).Error; err != nil {
		t.Fatalf("Failed to create order items: %v", err)
	}

	t.Run("Star with field alias - should work after removing conflict check", func(t *testing.T) {
		// This pattern was blocked by field name conflict check before
		// SELECT orders.*, order_items.*, order_items.id AS tid
		// FROM orders
		// JOIN order_items ON orders.id = order_items.order_id
		//
		// Expected: The aliased field 'tid' should override the original 'id' from order_items.*
		type Result struct {
			Order
			OrderItem
			TID uint64 `gorm:"column:tid"` // Aliased field
		}

		var results []Result
		builder := gsql.Select(
			o.Star(),
			oi.Star(),
			oi.ID.As("tid"),
		).From(&o).
			Join(gsql.InnerJoin(&oi).On(o.ID.EqF(oi.OrderID))).
			OrderBy(o.ID.Asc(), oi.ID.Asc())

		// Print SQL for verification
		sql := builder.ToSQL()
		t.Logf("Generated SQL: %s", sql)

		err := builder.Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// Verify data integrity
		for i, r := range results {
			// TID should equal OrderItem.ID
			if r.TID != r.OrderItem.ID {
				t.Errorf("Result[%d]: TID=%d should equal OrderItem.ID=%d", i, r.TID, r.OrderItem.ID)
			}

			// Order data should be correct
			if r.Order.ID != orders[r.OrderItem.OrderID-orders[0].ID].ID {
				t.Errorf("Result[%d]: Order data mismatch", i)
			}
		}
	})

	t.Run("Multiple Star with multiple aliases", func(t *testing.T) {
		// SELECT orders.*, order_items.*,
		//        order_items.id AS item_id,
		//        orders.id AS order_id
		// FROM orders
		// JOIN order_items ON orders.id = order_items.order_id
		type Result struct {
			Order
			OrderItem
			ItemID  uint64 `gorm:"column:item_id"`
			OrderID uint64 `gorm:"column:order_id"`
		}

		var results []Result
		err := gsql.Select(
			o.Star(),
			oi.Star(),
			oi.ID.As("item_id"),
			o.ID.As("order_id"),
		).From(&o).
			Join(gsql.InnerJoin(&oi).On(o.ID.EqF(oi.OrderID))).
			OrderBy(o.ID.Asc(), oi.ID.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// Verify aliases
		for i, r := range results {
			if r.ItemID != r.OrderItem.ID {
				t.Errorf("Result[%d]: ItemID=%d should equal OrderItem.ID=%d", i, r.ItemID, r.OrderItem.ID)
			}
			if r.OrderID != r.Order.ID {
				t.Errorf("Result[%d]: OrderID=%d should equal Order.ID=%d", i, r.OrderID, r.Order.ID)
			}
		}
	})

	t.Run("Star with computed field alias", func(t *testing.T) {
		// SELECT orders.*, order_items.*,
		//        order_items.quantity * order_items.unit_price AS line_total
		// FROM orders
		// JOIN order_items ON orders.id = order_items.order_id
		type Result struct {
			Order
			OrderItem
			LineTotal float64 `gorm:"column:line_total"`
		}

		var results []Result
		err := gsql.Select(
			o.Star(),
			oi.Star(),
			oi.Quantity.Mul(oi.UnitPrice).As("line_total"),
		).From(&o).
			Join(gsql.InnerJoin(&oi).On(o.ID.EqF(oi.OrderID))).
			OrderBy(o.ID.Asc(), oi.ID.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// Verify computed values
		for i, r := range results {
			expected := float64(r.Quantity) * r.UnitPrice
			if r.LineTotal != expected {
				t.Errorf("Result[%d]: LineTotal=%f, expected %f", i, r.LineTotal, expected)
			}
		}
	})

	t.Run("Three tables with Star and aliases", func(t *testing.T) {
		// Test with three tables to ensure it scales
		c := CustomerSchema
		setupTable(t, c.ModelType())

		customers := []Customer{
			{Name: "Alice", Email: "alice@test.com"},
			{Name: "Bob", Email: "bob@test.com"},
		}
		if err := db.Create(&customers).Error; err != nil {
			t.Fatalf("Failed to create customers: %v", err)
		}

		// Update orders to use real customer IDs
		db.Model(&Order{}).Where("id = ?", orders[0].ID).Update("customer_id", customers[0].ID)
		db.Model(&Order{}).Where("id = ?", orders[1].ID).Update("customer_id", customers[1].ID)

		// SELECT customers.*, orders.*, order_items.*,
		//        customers.id AS cid, orders.id AS oid, order_items.id AS iid
		// FROM customers
		// JOIN orders ON customers.id = orders.customer_id
		// JOIN order_items ON orders.id = order_items.order_id
		type Result struct {
			Customer
			Order
			OrderItem
			CID uint64 `gorm:"column:cid"`
			OID uint64 `gorm:"column:oid"`
			IID uint64 `gorm:"column:iid"`
		}

		var results []Result
		err := gsql.Select(
			c.Star(),
			o.Star(),
			oi.Star(),
			c.ID.As("cid"),
			o.ID.As("oid"),
			oi.ID.As("iid"),
		).From(&c).
			Join(
				gsql.InnerJoin(&o).On(c.ID.EqF(o.CustomerID)),
				gsql.InnerJoin(&oi).On(o.ID.EqF(oi.OrderID)),
			).
			OrderBy(c.ID.Asc(), o.ID.Asc(), oi.ID.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// Verify all aliases match original IDs
		for i, r := range results {
			if r.CID != r.Customer.ID {
				t.Errorf("Result[%d]: CID mismatch", i)
			}
			if r.OID != r.Order.ID {
				t.Errorf("Result[%d]: OID mismatch", i)
			}
			if r.IID != r.OrderItem.ID {
				t.Errorf("Result[%d]: IID mismatch", i)
			}
		}
	})
}
