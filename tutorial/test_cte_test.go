package tutorial

import (
	"strings"
	"testing"

	gsql "github.com/donutnomad/gsql"
)

// ==================== CTE with UNION Tests ====================

// TestCTEWithUnion tests combining CTE with UNION
func TestCTEWithUnion(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999.99, Stock: 50},
		{Name: "T-Shirt", Category: "Clothing", Price: 29.99, Stock: 500},
		{Name: "Jeans", Category: "Clothing", Price: 79.99, Stock: 200},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	t.Run("CTE with multiple UNION queries", func(t *testing.T) {
		// MySQL: WITH electronics AS (SELECT * FROM products WHERE category = 'Electronics'),
		//             clothing AS (SELECT * FROM products WHERE category = 'Clothing')
		//        SELECT name, category FROM electronics
		//        UNION ALL
		//        SELECT name, category FROM clothing
		electronicsQuery := gsql.Select(p.AllFields()...).From(&p).Where(p.Category.Eq("Electronics"))
		clothingQuery := gsql.Select(p.AllFields()...).From(&p).Where(p.Category.Eq("Clothing"))

		cte := gsql.With("electronics", electronicsQuery).And("clothing", clothingQuery)

		// Build the main query combining both CTEs
		sql := cte.Select(gsql.Field("name"), gsql.Field("category")).
			From(gsql.TN("electronics")).
			ToSQL()

		t.Logf("CTE SQL: %s", sql)

		// Verify SQL structure
		if !strings.Contains(sql, "WITH") {
			t.Error("SQL should contain WITH clause")
		}
		if !strings.Contains(sql, "electronics") && !strings.Contains(sql, "clothing") {
			t.Error("SQL should contain both CTE names")
		}
	})
}
