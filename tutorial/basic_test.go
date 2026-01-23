package tutorial

import (
	"database/sql"
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/internal/fields"
)

// ==================== CRUD Tests ====================

// TestBasic_Select tests WHERE/ORDER/LIMIT/OFFSET operations
func TestBasic_Select(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999.99, Stock: 50},
		{Name: "iPad", Category: "Electronics", Price: 799.99, Stock: 75},
		{Name: "AirPods", Category: "Electronics", Price: 199.99, Stock: 200},
		{Name: "T-Shirt", Category: "Clothing", Price: 29.99, Stock: 500},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test 1: WHERE + ORDER BY DESC
	t.Run("WHERE + ORDER BY DESC", func(t *testing.T) {
		var result []Product
		// MySQL: SELECT products.* FROM products
		//        WHERE products.category = 'Electronics'
		//        ORDER BY products.price DESC
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			Order(p.Price, false). // DESC
			Find(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(result) != 4 {
			t.Errorf("Expected 4 products, got %d", len(result))
		}
		if result[0].Name != "MacBook" {
			t.Errorf("Expected first product to be MacBook, got %s", result[0].Name)
		}
	})

	// Test 2: WHERE + ORDER BY ASC
	t.Run("WHERE + ORDER BY ASC", func(t *testing.T) {
		var result []Product
		// MySQL: SELECT products.* FROM products
		//        WHERE products.category = 'Electronics'
		//        ORDER BY products.price ASC
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			OrderBy(p.Price.Asc()).
			Find(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result[0].Name != "AirPods" {
			t.Errorf("Expected first product to be AirPods, got %s", result[0].Name)
		}
	})

	// Test 3: LIMIT
	t.Run("LIMIT", func(t *testing.T) {
		var result []Product
		// MySQL: SELECT products.* FROM products
		//        ORDER BY products.price DESC
		//        LIMIT 2
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Order(p.Price, false).
			Limit(2).
			Find(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 products, got %d", len(result))
		}
	})

	// Test 4: OFFSET + LIMIT
	t.Run("OFFSET + LIMIT", func(t *testing.T) {
		var result []Product
		// MySQL: SELECT products.* FROM products
		//        ORDER BY products.price DESC
		//        LIMIT 2 OFFSET 1
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Order(p.Price, false).
			Offset(1).
			Limit(2).
			Find(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 products, got %d", len(result))
		}
		// After MacBook (1999.99), we should get iPhone (999.99) and iPad (799.99)
		if result[0].Name != "iPhone" {
			t.Errorf("Expected first product to be iPhone, got %s", result[0].Name)
		}
	})

	// Test 5: Multiple WHERE conditions
	t.Run("Multiple WHERE conditions", func(t *testing.T) {
		var result []Product
		// MySQL: SELECT products.* FROM products
		//        WHERE products.category = 'Electronics' AND products.price >= 500
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(
				p.Category.Eq("Electronics"),
				p.Price.Gte(500),
			).
			Find(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("Expected 3 products (iPhone, MacBook, iPad), got %d", len(result))
		}
	})

	// Test 6: IN operator
	t.Run("IN operator", func(t *testing.T) {
		var result []Product
		// MySQL: SELECT products.* FROM products
		//        WHERE products.name IN ('iPhone', 'MacBook')
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.In("iPhone", "MacBook")).
			Find(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 products, got %d", len(result))
		}
	})

	// Test 7: LIKE operator
	t.Run("LIKE operator", func(t *testing.T) {
		var result []Product
		// MySQL: SELECT products.* FROM products
		//        WHERE products.name LIKE '%Pod%'
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Contains("Pod")).
			Find(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(result) != 1 || result[0].Name != "AirPods" {
			t.Errorf("Expected AirPods, got %v", result)
		}
	})
}

// TestBasic_Insert tests single and batch insert operations
func TestBasic_Insert(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Test 1: Single insert
	t.Run("Single insert", func(t *testing.T) {
		product := Product{
			Name:     "Watch",
			Category: "Electronics",
			Price:    399.99,
			Stock:    150,
		}
		if err := db.Create(&product).Error; err != nil {
			t.Fatalf("Failed to insert product: %v", err)
		}
		if product.ID == 0 {
			t.Error("Expected product ID to be set after insert")
		}

		// Verify insertion
		var result Product
		// MySQL: SELECT products.* FROM products WHERE products.id = ?
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.ID.Eq(product.ID)).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Name != "Watch" {
			t.Errorf("Expected Watch, got %s", result.Name)
		}
	})

	// Test 2: Batch insert
	t.Run("Batch insert", func(t *testing.T) {
		products := []Product{
			{Name: "Phone", Category: "Electronics", Price: 599.99, Stock: 80},
			{Name: "Laptop", Category: "Electronics", Price: 1299.99, Stock: 40},
			{Name: "Tablet", Category: "Electronics", Price: 449.99, Stock: 60},
		}
		if err := db.Create(&products).Error; err != nil {
			t.Fatalf("Failed to batch insert: %v", err)
		}

		// Verify insertion
		// MySQL: SELECT COUNT(*) FROM products WHERE products.name IN ('Phone', 'Laptop', 'Tablet')
		count, err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.In("Phone", "Laptop", "Tablet")).
			Count(db)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected 3 products, got %d", count)
		}
	})

	// Test 3: Insert with nullable field
	t.Run("Insert with nullable field", func(t *testing.T) {
		product := Product{
			Name:        "Special Item",
			Category:    "Special",
			Price:       99.99,
			Stock:       10,
			Description: sql.NullString{String: "This is a special item", Valid: true},
		}
		if err := db.Create(&product).Error; err != nil {
			t.Fatalf("Failed to insert: %v", err)
		}

		var result Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.ID.Eq(product.ID)).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if !result.Description.Valid || result.Description.String != "This is a special item" {
			t.Errorf("Description mismatch: %v", result.Description)
		}
	})
}

// TestBasic_Update tests conditional update operations
func TestBasic_Update(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "Product A", Category: "Cat1", Price: 100, Stock: 10},
		{Name: "Product B", Category: "Cat1", Price: 200, Stock: 20},
		{Name: "Product C", Category: "Cat2", Price: 300, Stock: 30},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test 1: Update single field with condition
	t.Run("Update single field", func(t *testing.T) {
		// MySQL: UPDATE products SET price = 150.0 WHERE products.name = 'Product A'
		result := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Eq("Product A")).
			Update(db, map[string]any{"price": 150.0})
		if result.Error != nil {
			t.Fatalf("Update failed: %v", result.Error)
		}
		if result.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
		}

		// Verify update
		var updated Product
		// MySQL: SELECT products.* FROM products WHERE products.name = 'Product A' LIMIT 1
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Eq("Product A")).
			First(db, &updated)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if updated.Price != 150.0 {
			t.Errorf("Expected price 150.0, got %f", updated.Price)
		}
	})

	// Test 2: Update multiple fields
	t.Run("Update multiple fields", func(t *testing.T) {
		// MySQL: UPDATE products SET price = 250.0, stock = 25 WHERE products.name = 'Product B'
		result := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Eq("Product B")).
			Update(db, map[string]any{
				"price": 250.0,
				"stock": 25,
			})
		if result.Error != nil {
			t.Fatalf("Update failed: %v", result.Error)
		}

		var updated Product
		err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Eq("Product B")).
			First(db, &updated)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if updated.Price != 250.0 || updated.Stock != 25 {
			t.Errorf("Update mismatch: price=%f, stock=%d", updated.Price, updated.Stock)
		}
	})

	// Test 3: Update with category condition (multiple rows)
	t.Run("Update multiple rows", func(t *testing.T) {
		result := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("Cat1")).
			Update(db, map[string]any{"category": "Category1"})
		if result.Error != nil {
			t.Fatalf("Update failed: %v", result.Error)
		}
		if result.RowsAffected != 2 {
			t.Errorf("Expected 2 rows affected, got %d", result.RowsAffected)
		}
	})
}

// TestBasic_Delete tests conditional delete operations
func TestBasic_Delete(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "Delete A", Category: "ToDelete", Price: 100, Stock: 10},
		{Name: "Delete B", Category: "ToDelete", Price: 200, Stock: 20},
		{Name: "Keep C", Category: "ToKeep", Price: 300, Stock: 30},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test 1: Delete single row
	t.Run("Delete single row", func(t *testing.T) {
		// MySQL: DELETE FROM products WHERE products.name = 'Delete A'
		result := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Eq("Delete A")).
			Delete(db, &Product{})
		if result.Error != nil {
			t.Fatalf("Delete failed: %v", result.Error)
		}
		if result.RowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
		}

		// Verify deletion
		// MySQL: SELECT EXISTS(SELECT 1 FROM products WHERE products.name = 'Delete A')
		exist, err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Eq("Delete A")).
			Exist(db)
		if err != nil {
			t.Fatalf("Exist check failed: %v", err)
		}
		if exist {
			t.Error("Expected product to be deleted")
		}
	})

	// Test 2: Delete by category (multiple rows)
	t.Run("Delete multiple rows", func(t *testing.T) {
		// MySQL: DELETE FROM products WHERE products.category = 'ToDelete'
		result := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Category.Eq("ToDelete")).
			Delete(db, &Product{})
		if result.Error != nil {
			t.Fatalf("Delete failed: %v", result.Error)
		}
		if result.RowsAffected != 1 { // Only Delete B remains
			t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
		}

		// Verify Keep C still exists
		// MySQL: SELECT EXISTS(SELECT 1 FROM products WHERE products.name = 'Keep C')
		exist, err := gsql.Select(p.AllFields()...).
			From(&p).
			Where(p.Name.Eq("Keep C")).
			Exist(db)
		if err != nil {
			t.Fatalf("Exist check failed: %v", err)
		}
		if !exist {
			t.Error("Keep C should not be deleted")
		}
	})
}

// ==================== Function Tests ====================

// TestFunc_DateTime tests date/time functions: NOW, YEAR, MONTH, DAY, DATE_FORMAT, DATEDIFF, DATE_ADD
func TestFunc_DateTime(t *testing.T) {
	e := EmployeeSchema
	setupTable(t, e.ModelType())
	db := getDB()

	// Insert test data
	hireDate := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
	birthDate := time.Date(1990, 3, 20, 0, 0, 0, 0, time.UTC)
	employees := []Employee{
		{Name: "John", Email: "john@test.com", Department: "IT", Salary: 80000, HireDate: hireDate, BirthDate: birthDate, IsActive: true},
		{Name: "Jane", Email: "jane@test.com", Department: "HR", Salary: 75000, HireDate: time.Date(2019, 1, 10, 0, 0, 0, 0, time.UTC), BirthDate: time.Date(1985, 7, 5, 0, 0, 0, 0, time.UTC), IsActive: true},
	}
	if err := db.Create(&employees).Error; err != nil {
		t.Fatalf("Failed to create employees: %v", err)
	}

	// Test NOW()
	t.Run("NOW", func(t *testing.T) {
		var result struct {
			CurrentTime time.Time `gorm:"column:current_time"`
		}
		// MySQL: SELECT NOW() AS current_time FROM employees LIMIT 1
		err := gsql.Select(gsql.NOW().AsF("current_time")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// NOW() should return current time (within reasonable range)
		if time.Since(result.CurrentTime) > time.Minute {
			t.Errorf("NOW() returned unexpected time: %v", result.CurrentTime)
		}
	})

	// Test YEAR()
	t.Run("YEAR", func(t *testing.T) {
		var result struct {
			HireYear int `gorm:"column:hire_year"`
		}
		// MySQL: SELECT YEAR(employees.hire_date) AS hire_year FROM employees
		//        WHERE employees.name = 'John' LIMIT 1
		err := gsql.Select(e.HireDate.Year().As("hire_year")).
			From(&e).
			Where(e.Name.Eq("John")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.HireYear != 2020 {
			t.Errorf("Expected year 2020, got %d", result.HireYear)
		}
	})

	// Test MONTH()
	t.Run("MONTH", func(t *testing.T) {
		var result struct {
			HireMonth int `gorm:"column:hire_month"`
		}
		// MySQL: SELECT MONTH(employees.hire_date) AS hire_month FROM employees
		//        WHERE employees.name = 'John' LIMIT 1
		err := gsql.Select(e.HireDate.Month().As("hire_month")).
			From(&e).
			Where(e.Name.Eq("John")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.HireMonth != 6 {
			t.Errorf("Expected month 6, got %d", result.HireMonth)
		}
	})

	// Test DAY()
	t.Run("DAY", func(t *testing.T) {
		var result struct {
			HireDay int `gorm:"column:hire_day"`
		}
		// MySQL: SELECT DAY(employees.hire_date) AS hire_day FROM employees
		//        WHERE employees.name = 'John' LIMIT 1
		err := gsql.Select(e.HireDate.Day().As("hire_day")).
			From(&e).
			Where(e.Name.Eq("John")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.HireDay != 15 {
			t.Errorf("Expected day 15, got %d", result.HireDay)
		}
	})

	// Test DATE_FORMAT()
	t.Run("DATE_FORMAT", func(t *testing.T) {
		var result struct {
			FormattedDate string `gorm:"column:formatted_date"`
		}
		err := gsql.Select(e.HireDate.Format("%Y-%m-%d").As("formatted_date")).
			From(&e).
			Where(e.Name.Eq("John")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.FormattedDate != "2020-06-15" {
			t.Errorf("Expected 2020-06-15, got %s", result.FormattedDate)
		}
	})

	// Test DATEDIFF()
	t.Run("DATEDIFF", func(t *testing.T) {
		var result struct {
			DaysDiff int `gorm:"column:days_diff"`
		}
		// Calculate days between hire dates of John and Jane
		date1 := fields.NewDateTimeExpr[time.Time](gsql.Lit("2020-06-15"))
		date2 := fields.NewDateTimeExpr[time.Time](gsql.Lit("2019-01-10"))
		err := gsql.Select(date1.DateDiff(date2).As("days_diff")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		// 2020-06-15 - 2019-01-10 = 522 days
		if result.DaysDiff != 522 {
			t.Errorf("Expected 522 days diff, got %d", result.DaysDiff)
		}
	})

	// Test DATE_ADD()
	t.Run("DATE_ADD", func(t *testing.T) {
		var result struct {
			FutureDate time.Time `gorm:"column:future_date"`
		}
		err := gsql.Select(e.HireDate.AddInterval("30 DAY").As("future_date")).
			From(&e).
			Where(e.Name.Eq("John")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		expected := hireDate.AddDate(0, 0, 30)
		if result.FutureDate.Format("2006-01-02") != expected.Format("2006-01-02") {
			t.Errorf("Expected %v, got %v", expected, result.FutureDate)
		}
	})
}

// TestFunc_String tests string functions: CONCAT, UPPER, LOWER, SUBSTRING, LENGTH, TRIM, REPLACE
func TestFunc_String(t *testing.T) {
	e := EmployeeSchema
	setupTable(t, e.ModelType())
	db := getDB()

	// Insert test data
	employees := []Employee{
		{Name: "  John Doe  ", Email: "john@test.com", Department: "IT", Salary: 80000, HireDate: time.Now(), BirthDate: time.Now(), IsActive: true},
	}
	if err := db.Create(&employees).Error; err != nil {
		t.Fatalf("Failed to create employees: %v", err)
	}

	// Test CONCAT()
	t.Run("CONCAT", func(t *testing.T) {
		var result struct {
			FullInfo string `gorm:"column:full_info"`
		}
		err := gsql.Select(e.Name.Concat(gsql.Lit(" - "), e.Department.ToExpr()).As("full_info")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.FullInfo != "  John Doe   - IT" {
			t.Errorf("Expected '  John Doe   - IT', got '%s'", result.FullInfo)
		}
	})

	// Test UPPER()
	t.Run("UPPER", func(t *testing.T) {
		var result struct {
			UpperName string `gorm:"column:upper_name"`
		}
		err := gsql.Select(e.Department.Upper().As("upper_name")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.UpperName != "IT" {
			t.Errorf("Expected 'IT', got '%s'", result.UpperName)
		}
	})

	// Test LOWER()
	t.Run("LOWER", func(t *testing.T) {
		var result struct {
			LowerDept string `gorm:"column:lower_dept"`
		}
		err := gsql.Select(e.Department.Lower().As("lower_dept")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.LowerDept != "it" {
			t.Errorf("Expected 'it', got '%s'", result.LowerDept)
		}
	})

	// Test SUBSTRING()
	t.Run("SUBSTRING", func(t *testing.T) {
		var result struct {
			SubStr string `gorm:"column:sub_str"`
		}
		// Extract "john" from email
		err := gsql.Select(e.Email.Substring(1, 4).As("sub_str")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.SubStr != "john" {
			t.Errorf("Expected 'john', got '%s'", result.SubStr)
		}
	})

	// Test LENGTH()
	t.Run("LENGTH", func(t *testing.T) {
		var result struct {
			EmailLen int `gorm:"column:email_len"`
		}
		err := gsql.Select(e.Email.Length().As("email_len")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.EmailLen != 13 { // "john@test.com" = 13 chars
			t.Errorf("Expected 13, got %d", result.EmailLen)
		}
	})

	// Test TRIM()
	t.Run("TRIM", func(t *testing.T) {
		var result struct {
			TrimmedName string `gorm:"column:trimmed_name"`
		}
		err := gsql.Select(e.Name.Trim().As("trimmed_name")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.TrimmedName != "John Doe" {
			t.Errorf("Expected 'John Doe', got '%s'", result.TrimmedName)
		}
	})

	// Test REPLACE()
	t.Run("REPLACE", func(t *testing.T) {
		var result struct {
			ReplacedEmail string `gorm:"column:replaced_email"`
		}
		err := gsql.Select(e.Email.Replace("test.com", "example.org").As("replaced_email")).
			From(&e).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.ReplacedEmail != "john@example.org" {
			t.Errorf("Expected 'john@example.org', got '%s'", result.ReplacedEmail)
		}
	})
}

// TestFunc_Numeric tests numeric functions: ABS, CEIL, FLOOR, ROUND, MOD, POWER, SQRT
func TestFunc_Numeric(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "Test Product", Category: "Test", Price: 123.456, Stock: 17},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test ABS()
	t.Run("ABS", func(t *testing.T) {
		var result struct {
			AbsVal float64 `gorm:"column:abs_val"`
		}
		err := gsql.Select(fields.NewFloatExpr[float64](gsql.Lit(-99.5)).Abs().As("abs_val")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.AbsVal != 99.5 {
			t.Errorf("Expected 99.5, got %f", result.AbsVal)
		}
	})

	// Test CEIL()
	t.Run("CEIL", func(t *testing.T) {
		var result struct {
			CeilVal int `gorm:"column:ceil_val"`
		}
		err := gsql.Select(p.Price.Ceil().As("ceil_val")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.CeilVal != 124 {
			t.Errorf("Expected 124, got %d", result.CeilVal)
		}
	})

	// Test FLOOR()
	t.Run("FLOOR", func(t *testing.T) {
		var result struct {
			FloorVal int `gorm:"column:floor_val"`
		}
		err := gsql.Select(p.Price.Floor().As("floor_val")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.FloorVal != 123 {
			t.Errorf("Expected 123, got %d", result.FloorVal)
		}
	})

	// Test ROUND()
	t.Run("ROUND", func(t *testing.T) {
		var result struct {
			RoundVal float64 `gorm:"column:round_val"`
		}
		err := gsql.Select(p.Price.Round(2).As("round_val")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.RoundVal != 123.46 {
			t.Errorf("Expected 123.46, got %f", result.RoundVal)
		}
	})

	// Test MOD()
	t.Run("MOD", func(t *testing.T) {
		var result struct {
			ModVal int `gorm:"column:mod_val"`
		}
		err := gsql.Select(p.Stock.Mod(5).As("mod_val")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.ModVal != 2 { // 17 % 5 = 2
			t.Errorf("Expected 2, got %d", result.ModVal)
		}
	})

	// Test POWER()
	t.Run("POWER", func(t *testing.T) {
		var result struct {
			PowerVal float64 `gorm:"column:power_val"`
		}
		err := gsql.Select(fields.NewIntExpr[int](gsql.Lit(2)).Pow(3).As("power_val")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.PowerVal != 8 {
			t.Errorf("Expected 8, got %f", result.PowerVal)
		}
	})

	// Test SQRT()
	t.Run("SQRT", func(t *testing.T) {
		var result struct {
			SqrtVal float64 `gorm:"column:sqrt_val"`
		}
		err := gsql.Select(fields.NewIntExpr[int](gsql.Lit(16)).Sqrt().As("sqrt_val")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.SqrtVal != 4 {
			t.Errorf("Expected 4, got %f", result.SqrtVal)
		}
	})
}

// TestFunc_Aggregate tests aggregate functions: COUNT, SUM, AVG, MAX, MIN
func TestFunc_Aggregate(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "Product A", Category: "Electronics", Price: 100, Stock: 10},
		{Name: "Product B", Category: "Electronics", Price: 200, Stock: 20},
		{Name: "Product C", Category: "Electronics", Price: 300, Stock: 30},
		{Name: "Product D", Category: "Clothing", Price: 50, Stock: 100},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test COUNT()
	t.Run("COUNT", func(t *testing.T) {
		var result struct {
			Total int64 `gorm:"column:total"`
		}
		// MySQL: SELECT COUNT(*) AS total FROM products
		err := gsql.Select(gsql.COUNT().As("total")).
			From(&p).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Total != 4 {
			t.Errorf("Expected 4, got %d", result.Total)
		}
	})

	// Test COUNT with column
	t.Run("COUNT with column", func(t *testing.T) {
		var result struct {
			Count int64 `gorm:"column:count"`
		}
		// MySQL: SELECT COUNT(products.id) AS count FROM products
		//        WHERE products.category = 'Electronics'
		err := gsql.Select(gsql.COUNT(p.ID).As("count")).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Count != 3 {
			t.Errorf("Expected 3, got %d", result.Count)
		}
	})

	// Test SUM()
	t.Run("SUM", func(t *testing.T) {
		var result struct {
			TotalPrice float64 `gorm:"column:total_price"`
		}
		// MySQL: SELECT SUM(products.price) AS total_price FROM products
		//        WHERE products.category = 'Electronics'
		err := gsql.Select(p.Price.Sum().As("total_price")).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.TotalPrice != 600 {
			t.Errorf("Expected 600, got %f", result.TotalPrice)
		}
	})

	// Test AVG()
	t.Run("AVG", func(t *testing.T) {
		var result struct {
			AvgPrice float64 `gorm:"column:avg_price"`
		}
		// MySQL: SELECT AVG(products.price) AS avg_price FROM products
		//        WHERE products.category = 'Electronics'
		err := gsql.Select(p.Price.Avg().As("avg_price")).
			From(&p).
			Where(p.Category.Eq("Electronics")).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.AvgPrice != 200 {
			t.Errorf("Expected 200, got %f", result.AvgPrice)
		}
	})

	// Test MAX()
	t.Run("MAX", func(t *testing.T) {
		var result struct {
			MaxPrice float64 `gorm:"column:max_price"`
		}
		err := gsql.Select(p.Price.Max().As("max_price")).
			From(&p).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.MaxPrice != 300 {
			t.Errorf("Expected 300, got %f", result.MaxPrice)
		}
	})

	// Test MIN()
	t.Run("MIN", func(t *testing.T) {
		var result struct {
			MinPrice float64 `gorm:"column:min_price"`
		}
		err := gsql.Select(p.Price.Min().As("min_price")).
			From(&p).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.MinPrice != 50 {
			t.Errorf("Expected 50, got %f", result.MinPrice)
		}
	})

	// Test GROUP BY with aggregate
	t.Run("GROUP BY with aggregate", func(t *testing.T) {
		var results []struct {
			Category   string  `gorm:"column:category"`
			TotalStock int     `gorm:"column:total_stock"`
			AvgPrice   float64 `gorm:"column:avg_price"`
		}
		err := gsql.Select(
			p.Category,
			p.Stock.Sum().As("total_stock"),
			p.Price.Avg().As("avg_price"),
		).
			From(&p).
			GroupBy(p.Category).
			Order(p.Category, true).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("Expected 2 groups, got %d", len(results))
		}
		// Find Clothing result
		for _, r := range results {
			if r.Category == "Clothing" {
				if r.TotalStock != 100 {
					t.Errorf("Clothing total stock: expected 100, got %d", r.TotalStock)
				}
			}
		}
	})
}

// TestFunc_FlowControl tests flow control functions: IF, IFNULL, NULLIF
func TestFunc_FlowControl(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "In Stock", Category: "Test", Price: 100, Stock: 50},
		{Name: "Out of Stock", Category: "Test", Price: 200, Stock: 0},
		{Name: "With Desc", Category: "Test", Price: 150, Stock: 25, Description: sql.NullString{String: "Has description", Valid: true}},
		{Name: "No Desc", Category: "Test", Price: 175, Stock: 30, Description: sql.NullString{Valid: false}},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test IF()
	t.Run("IF", func(t *testing.T) {
		var results []struct {
			Name         string `gorm:"column:name"`
			Availability string `gorm:"column:availability"`
		}
		err := gsql.Select(
			p.Name,
			gsql.IF(p.Stock.Gt(0), gsql.Lit("Available"), gsql.Lit("Out of Stock")).AsF("availability"),
		).
			From(&p).
			Order(p.ID, true).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if results[0].Availability != "Available" {
			t.Errorf("Expected 'Available', got '%s'", results[0].Availability)
		}
		if results[1].Availability != "Out of Stock" {
			t.Errorf("Expected 'Out of Stock', got '%s'", results[1].Availability)
		}
	})

	// Test IFNULL()
	t.Run("IFNULL", func(t *testing.T) {
		var results []struct {
			Name string `gorm:"column:name"`
			Desc string `gorm:"column:desc"`
		}
		err := gsql.Select(
			p.Name,
			gsql.IFNULL(p.Description.ToExpr(), gsql.Lit("No description")).AsF("desc"),
		).
			From(&p).
			Where(p.Name.In("With Desc", "No Desc")).
			Order(p.ID, true).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		if results[0].Desc != "Has description" {
			t.Errorf("Expected 'Has description', got '%s'", results[0].Desc)
		}
		if results[1].Desc != "No description" {
			t.Errorf("Expected 'No description', got '%s'", results[1].Desc)
		}
	})

	// Test NULLIF()
	t.Run("NULLIF", func(t *testing.T) {
		var results []struct {
			Name      string `gorm:"column:name"`
			StockNull *int   `gorm:"column:stock_null"`
		}
		// NULLIF(stock, 0) returns NULL if stock is 0
		err := gsql.Select(
			p.Name,
			gsql.NULLIF(p.Stock.ToExpr(), gsql.Lit(0)).AsF("stock_null"),
		).
			From(&p).
			Where(p.Name.In("In Stock", "Out of Stock")).
			Order(p.ID, true).
			Find(db, &results)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if results[0].StockNull == nil || *results[0].StockNull != 50 {
			t.Errorf("Expected 50, got %v", results[0].StockNull)
		}
		if results[1].StockNull != nil {
			t.Errorf("Expected nil for zero stock, got %v", *results[1].StockNull)
		}
	})
}

// TestFunc_TypeConvert tests type conversion functions: CAST, CONVERT
func TestFunc_TypeConvert(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "Test", Category: "Test", Price: 123.456, Stock: 100},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test CAST to SIGNED
	t.Run("CAST to SIGNED", func(t *testing.T) {
		var result struct {
			IntPrice int64 `gorm:"column:int_price"`
		}

		err := gsql.Select(p.Price.CastUnsigned().As("int_price")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.IntPrice != 123 {
			t.Errorf("Expected 123, got %d", result.IntPrice)
		}
	})

	// Test CAST to CHAR
	t.Run("CAST to CHAR", func(t *testing.T) {
		var result struct {
			StrStock string `gorm:"column:str_stock"`
		}
		err := gsql.Select(p.Stock.CastChar().As("str_stock")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.StrStock != "100" {
			t.Errorf("Expected '100', got '%s'", result.StrStock)
		}
	})

	// Test CONVERT to SIGNED
	t.Run("CONVERT to SIGNED", func(t *testing.T) {
		var result struct {
			IntPrice int64 `gorm:"column:int_price"`
		}
		err := gsql.Select(p.Price.CastUnsigned().As("int_price")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.IntPrice != 123 {
			t.Errorf("Expected 123, got %d", result.IntPrice)
		}
	})

	// Test CAST with DECIMAL
	t.Run("CAST with DECIMAL", func(t *testing.T) {
		var result struct {
			DecPrice float64 `gorm:"column:dec_price"`
		}
		err := gsql.Select(p.Price.CastDecimal(10, 2).As("dec_price")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.DecPrice != 123.46 {
			t.Errorf("Expected 123.46, got %f", result.DecPrice)
		}
	})
}

// TestFunc_Arithmetic tests arithmetic operations: gsql.Add, Sub, Mul, Div, Mod
func TestFunc_Arithmetic(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	products := []Product{
		{Name: "Test", Category: "Test", Price: 100, Stock: 20},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Test Add
	t.Run("Add", func(t *testing.T) {
		var result struct {
			Total float64 `gorm:"column:total"`
		}
		err := gsql.Select(p.Price.Add(50).As("total")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Total != 150 {
			t.Errorf("Expected 150, got %f", result.Total)
		}
	})

	// Test Sub
	t.Run("Sub", func(t *testing.T) {
		var result struct {
			Diff float64 `gorm:"column:diff"`
		}
		err := gsql.Select(p.Price.Sub(30).As("diff")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Diff != 70 {
			t.Errorf("Expected 70, got %f", result.Diff)
		}
	})

	// Test Mul
	t.Run("Mul", func(t *testing.T) {
		var result struct {
			Product float64 `gorm:"column:product"`
		}
		err := gsql.Select(p.Price.Mul(p.Stock.ToExpr()).As("product")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Product != 2000 { // 100 * 20
			t.Errorf("Expected 2000, got %f", result.Product)
		}
	})

	// Test Div
	t.Run("Div", func(t *testing.T) {
		var result struct {
			Quotient float64 `gorm:"column:quotient"`
		}
		err := gsql.Select(p.Price.Div(4).As("quotient")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Quotient != 25 {
			t.Errorf("Expected 25, got %f", result.Quotient)
		}
	})

	// Test Mod
	t.Run("Mod", func(t *testing.T) {
		var result struct {
			Remainder float64 `gorm:"column:remainder"`
		}
		err := gsql.Select(p.Stock.Mod(7).As("remainder")).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Remainder != 6 { // 20 % 7 = 6
			t.Errorf("Expected 6, got %f", result.Remainder)
		}
	})

	// Test complex arithmetic expression
	t.Run("Complex arithmetic", func(t *testing.T) {
		var result struct {
			Value float64 `gorm:"column:value"`
		}
		// (price * stock) - (price / 2) = (100 * 20) - (100 / 2) = 2000 - 50 = 1950
		err := gsql.Select(
			p.Price.Mul(p.Stock.ToExpr()).Sub(p.Price.Div(2)).As("value"),
		).
			From(&p).
			Limit(1).
			First(db, &result)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if result.Value != 1950 {
			t.Errorf("Expected 1950, got %f", result.Value)
		}
	})
}
