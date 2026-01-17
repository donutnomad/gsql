package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
)

// ==================== CTE Tests ====================

// TestAdv_BasicCTE tests basic CTE (Common Table Expression)
func TestAdv_BasicCTE(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 1500, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Charlie", Amount: 2000, SaleDate: time.Now()},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// Basic CTE: select high sales records
	// Use TableName(...).Ptr() for CTE name reference
	cte := gsql.With("high_sales",
		gsql.Select(s.AllFields()...).
			From(&s).
			Where(s.Amount.Gt(1200)),
	)

	type Result struct {
		Salesperson string  `gorm:"column:salesperson"`
		Amount      float64 `gorm:"column:amount"`
	}

	// Build the CTE query
	query := cte.Select(gsql.Field("salesperson"), gsql.Field("amount")).
		From(gsql.TN("high_sales"))

	sql := query.ToSQL()
	t.Logf("CTE SQL: %s", sql)

	// Note: CTE execution via Find() has a bug in gsql where BuildClauses
	// doesn't include "CTE". Using db.Raw(ToSQL()) as a workaround.
	var results []Result
	err := db.Raw(sql).Scan(&results).Error

	if err != nil {
		t.Fatalf("Basic CTE failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 high sales records, got %d", len(results))
	}
}

// TestAdv_RecursiveCTE tests recursive CTE for hierarchical data
func TestAdv_RecursiveCTE(t *testing.T) {
	o := OrgNodeSchema
	setupTable(t, o.ModelType())
	db := getDB()

	// Insert organization tree data
	// CEO (id=1) -> VP1 (id=2), VP2 (id=3) -> Manager1 (id=4)
	nodes := []OrgNode{
		{ID: 1, Name: "CEO", ParentID: nil, Level: 0},
		{ID: 2, Name: "VP Engineering", ParentID: ptr(uint64(1)), Level: 1},
		{ID: 3, Name: "VP Sales", ParentID: ptr(uint64(1)), Level: 1},
		{ID: 4, Name: "Manager", ParentID: ptr(uint64(2)), Level: 2},
	}
	for _, node := range nodes {
		if err := db.Create(&node).Error; err != nil {
			t.Fatalf("Failed to create node: %v", err)
		}
	}

	// Recursive CTE - get organization tree
	// Use gsql.Expr for IS NULL check since field doesn't have IsNull method
	sql := gsql.WithRecursive("org_tree",
		gsql.Select(o.ID, o.Name, o.ParentID, o.Level).
			From(&o).
			Where(gsql.Expr("parent_id IS NULL")),
	).Select(gsql.Star).
		From(gsql.TN("org_tree")).
		ToSQL()

	if sql == "" {
		t.Error("Recursive CTE SQL should not be empty")
	}
	t.Logf("Recursive CTE SQL: %s", sql)
}

// TestAdv_MultipleCTE tests multiple CTEs in a single query
func TestAdv_MultipleCTE(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Bob", Amount: 2000, SaleDate: time.Now()},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// Multiple CTEs
	sql := gsql.With("north_sales",
		gsql.Select(s.AllFields()...).From(&s).Where(s.Region.Eq("North")),
	).And("south_sales",
		gsql.Select(s.AllFields()...).From(&s).Where(s.Region.Eq("South")),
	).Select(gsql.Star).
		From(gsql.TN("north_sales")).
		ToSQL()

	if sql == "" {
		t.Error("Multiple CTE SQL should not be empty")
	}
	t.Logf("Multiple CTE SQL: %s", sql)
}

// ==================== Window Function Tests ====================

// TestAdv_RowNumber tests ROW_NUMBER() window function
func TestAdv_RowNumber(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 1500, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Charlie", Amount: 2000, SaleDate: time.Now()},
		{Region: "South", Salesperson: "David", Amount: 1800, SaleDate: time.Now()},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// ROW_NUMBER() OVER (PARTITION BY region ORDER BY amount DESC)
	rn := gsql.RowNumber().
		PartitionBy(s.Region).
		OrderBy(s.Amount, true).
		AsF("row_num")

	type Result struct {
		Region      string  `gorm:"column:region"`
		Salesperson string  `gorm:"column:salesperson"`
		Amount      float64 `gorm:"column:amount"`
		RowNum      int     `gorm:"column:row_num"`
	}

	var results []Result
	err := gsql.Select(s.Region, s.Salesperson, s.Amount, rn).
		From(&s).
		Find(db, &results)

	if err != nil {
		t.Fatalf("ROW_NUMBER failed: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	// Verify first place in each region
	for _, r := range results {
		if r.RowNum == 1 {
			if r.Region == "North" && r.Salesperson != "Bob" {
				t.Errorf("North #1 should be Bob, got %s", r.Salesperson)
			}
			if r.Region == "South" && r.Salesperson != "Charlie" {
				t.Errorf("South #1 should be Charlie, got %s", r.Salesperson)
			}
		}
	}
}

// TestAdv_RankDenseRank tests RANK() and DENSE_RANK() window functions
func TestAdv_RankDenseRank(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data with same amounts
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 1000, SaleDate: time.Now()}, // Same amount
		{Region: "North", Salesperson: "Charlie", Amount: 500, SaleDate: time.Now()},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// RANK()
	rank := gsql.Rank().
		OrderBy(s.Amount, true).
		AsF("rank_num")

	// DENSE_RANK()
	denseRank := gsql.DenseRank().
		OrderBy(s.Amount, true).
		AsF("dense_rank_num")

	type Result struct {
		Salesperson  string  `gorm:"column:salesperson"`
		Amount       float64 `gorm:"column:amount"`
		RankNum      int     `gorm:"column:rank_num"`
		DenseRankNum int     `gorm:"column:dense_rank_num"`
	}

	var results []Result
	err := gsql.Select(s.Salesperson, s.Amount, rank, denseRank).
		From(&s).
		Find(db, &results)

	if err != nil {
		t.Fatalf("RANK/DENSE_RANK failed: %v", err)
	}

	t.Logf("RANK/DENSE_RANK results: %+v", results)
}

// TestAdv_LagLead tests LAG() and LEAD() window functions using raw SQL
// Note: LAG/LEAD are not yet implemented in gsql, so we use raw SQL
func TestAdv_LagLead(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data
	baseTime := time.Now()
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 100, SaleDate: baseTime},
		{Region: "North", Salesperson: "Alice", Amount: 150, SaleDate: baseTime.Add(24 * time.Hour)},
		{Region: "North", Salesperson: "Alice", Amount: 200, SaleDate: baseTime.Add(48 * time.Hour)},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	type Result struct {
		Amount     float64 `gorm:"column:amount"`
		PrevAmount float64 `gorm:"column:prev_amount"`
		NextAmount float64 `gorm:"column:next_amount"`
	}

	// Use raw SQL for LAG/LEAD since gsql doesn't implement them yet
	var results []Result
	err := db.Raw(`
		SELECT
			amount,
			COALESCE(LAG(amount, 1) OVER (ORDER BY sale_date ASC), 0) as prev_amount,
			COALESCE(LEAD(amount, 1) OVER (ORDER BY sale_date ASC), 0) as next_amount
		FROM sales_records
		WHERE salesperson = 'Alice'
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("LAG/LEAD failed: %v", err)
	}

	t.Logf("LAG/LEAD results: %+v", results)
}

// TestAdv_SumOver tests SUM() with window frame using raw SQL
// Note: SumOver is not yet implemented in gsql, so we use raw SQL
func TestAdv_SumOver(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 100, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 200, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Charlie", Amount: 300, SaleDate: time.Now()},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	type Result struct {
		Region      string  `gorm:"column:region"`
		Salesperson string  `gorm:"column:salesperson"`
		Amount      float64 `gorm:"column:amount"`
		RegionTotal float64 `gorm:"column:region_total"`
	}

	// Use raw SQL for SUM OVER since gsql doesn't implement SumOver yet
	var results []Result
	err := db.Raw(`
		SELECT
			region,
			salesperson,
			amount,
			SUM(amount) OVER (PARTITION BY region) as region_total
		FROM sales_records
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("SUM OVER failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify region totals
	for _, r := range results {
		if r.Region == "North" && r.RegionTotal != 300 {
			t.Errorf("North total should be 300, got %f", r.RegionTotal)
		}
		if r.Region == "South" && r.RegionTotal != 300 {
			t.Errorf("South total should be 300, got %f", r.RegionTotal)
		}
	}
}

// ==================== JSON Tests ====================

// TestAdv_JsonExtract tests JSON_EXTRACT function
func TestAdv_JsonExtract(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// Insert JSON data
	profiles := []UserProfile{
		{Username: "alice", Profile: `{"age": 25, "city": "New York", "tags": ["developer", "golang"]}`},
		{Username: "bob", Profile: `{"age": 30, "city": "San Francisco", "tags": ["manager"]}`},
	}
	if err := db.Create(&profiles).Error; err != nil {
		t.Fatalf("Failed to create profiles: %v", err)
	}

	// JSON_EXTRACT
	type Result struct {
		Username string `gorm:"column:username"`
		Age      int    `gorm:"column:age"`
		City     string `gorm:"column:city"`
	}

	var results []Result
	err := db.Raw(`
		SELECT username,
		       JSON_EXTRACT(profile, '$.age') as age,
		       JSON_UNQUOTE(JSON_EXTRACT(profile, '$.city')) as city
		FROM user_profiles
		WHERE JSON_EXTRACT(profile, '$.age') > 20
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("JSON_EXTRACT failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestAdv_JsonModify tests JSON_SET function
func TestAdv_JsonModify(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// Insert JSON data
	profile := UserProfile{Username: "alice", Profile: `{"age": 25, "city": "New York"}`}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// JSON_SET - add/update fields
	err := db.Exec(`
		UPDATE user_profiles
		SET profile = JSON_SET(profile, '$.country', 'USA', '$.age', 26)
		WHERE username = 'alice'
	`).Error

	if err != nil {
		t.Fatalf("JSON_SET failed: %v", err)
	}

	// Verify update
	var country string
	db.Raw(`SELECT JSON_UNQUOTE(JSON_EXTRACT(profile, '$.country')) FROM user_profiles WHERE id = ?`, profile.ID).Scan(&country)
	if country != "USA" {
		t.Errorf("Expected country 'USA', got '%s'", country)
	}

	var age int
	db.Raw(`SELECT JSON_EXTRACT(profile, '$.age') FROM user_profiles WHERE id = ?`, profile.ID).Scan(&age)
	if age != 26 {
		t.Errorf("Expected age 26, got %d", age)
	}
}

// TestAdv_JsonContains tests JSON_CONTAINS function
func TestAdv_JsonContains(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// Insert JSON data
	profiles := []UserProfile{
		{Username: "alice", Profile: `{"skills": ["go", "python", "javascript"]}`},
		{Username: "bob", Profile: `{"skills": ["java", "python"]}`},
	}
	if err := db.Create(&profiles).Error; err != nil {
		t.Fatalf("Failed to create profiles: %v", err)
	}

	// JSON_CONTAINS - find users with specific skill
	var results []UserProfile
	err := db.Raw(`
		SELECT * FROM user_profiles
		WHERE JSON_CONTAINS(profile, '"go"', '$.skills')
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("JSON_CONTAINS failed: %v", err)
	}
	if len(results) != 1 || results[0].Username != "alice" {
		t.Errorf("Expected alice with go skill, got %+v", results)
	}
}

// TestAdv_JsonArray tests JSON_ARRAY function
func TestAdv_JsonArray(t *testing.T) {
	db := getDB()

	// Create JSON array
	var result string
	err := db.Raw(`SELECT JSON_ARRAY(1, 'two', 3.0, NULL)`).Scan(&result).Error
	if err != nil {
		t.Fatalf("JSON_ARRAY failed: %v", err)
	}
	if result != `[1, "two", 3.0, null]` {
		t.Errorf("Unexpected JSON array: %s", result)
	}
}

// TestAdv_JsonObject tests JSON_OBJECT function
func TestAdv_JsonObject(t *testing.T) {
	db := getDB()

	// Create JSON object
	var result string
	err := db.Raw(`SELECT JSON_OBJECT('name', 'Alice', 'age', 25)`).Scan(&result).Error
	if err != nil {
		t.Fatalf("JSON_OBJECT failed: %v", err)
	}
	// JSON object should contain name and age
	if result == "" {
		t.Error("JSON_OBJECT returned empty string")
	}
	t.Logf("JSON_OBJECT result: %s", result)
}

// ==================== Lock Tests ====================

// TestAdv_ForUpdate tests FOR UPDATE locking
func TestAdv_ForUpdate(t *testing.T) {
	tx := TransactionSchema
	setupTable(t, tx.ModelType())
	db := getDB()

	// Insert test data
	trans := Transaction{AccountID: 1, Amount: 100.00, Type: "credit"}
	if err := db.Create(&trans).Error; err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// FOR UPDATE
	sql := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForUpdate().
		ToSQL()

	if sql == "" {
		t.Error("FOR UPDATE SQL should not be empty")
	}
	t.Logf("FOR UPDATE SQL: %s", sql)

	// FOR UPDATE NOWAIT
	sqlNowait := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForUpdate().
		Nowait().
		ToSQL()

	t.Logf("FOR UPDATE NOWAIT SQL: %s", sqlNowait)
}

// TestAdv_ForShare tests FOR SHARE (LOCK IN SHARE MODE) locking
func TestAdv_ForShare(t *testing.T) {
	tx := TransactionSchema
	setupTable(t, tx.ModelType())
	db := getDB()

	// Insert test data
	trans := Transaction{AccountID: 1, Amount: 100.00, Type: "credit"}
	if err := db.Create(&trans).Error; err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// LOCK IN SHARE MODE
	sql := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForShare().
		ToSQL()

	if sql == "" {
		t.Error("LOCK IN SHARE MODE SQL should not be empty")
	}
	t.Logf("LOCK IN SHARE MODE SQL: %s", sql)

	// FOR SHARE SKIP LOCKED
	sqlSkip := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForShare().
		SkipLocked().
		ToSQL()

	t.Logf("FOR SHARE SKIP LOCKED SQL: %s", sqlSkip)
}

// ==================== Large IN Query Tests ====================

// TestAdv_LargeInQuery tests IN query with many values
// Note: gsql doesn't have BatchIn, but we can demonstrate large IN queries
func TestAdv_LargeInQuery(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data
	var products []Product
	for i := 0; i < 100; i++ {
		products = append(products, Product{
			Name:     "Product" + string(rune('A'+i%26)),
			Category: "Category" + string(rune('0'+i%10)),
			Price:    float64(i * 10),
			Stock:    i,
		})
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Collect IDs
	var ids []uint64
	for _, prod := range products {
		ids = append(ids, prod.ID)
	}

	// Use regular IN query with multiple values
	var results []Product
	err := gsql.Select(p.AllFields()...).
		From(&p).
		Where(p.ID.In(ids...)).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Large IN query failed: %v", err)
	}
	if len(results) != 100 {
		t.Errorf("Expected 100 products, got %d", len(results))
	}
}

// ==================== Distinct Tests ====================

// TestAdv_Distinct tests DISTINCT query
func TestAdv_Distinct(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Insert test data with duplicate categories
	products := []Product{
		{Name: "P1", Category: "Electronics", Price: 100, Stock: 10},
		{Name: "P2", Category: "Electronics", Price: 200, Stock: 20},
		{Name: "P3", Category: "Clothing", Price: 50, Stock: 100},
		{Name: "P4", Category: "Clothing", Price: 80, Stock: 50},
		{Name: "P5", Category: "Books", Price: 20, Stock: 200},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// DISTINCT categories
	var categories []string
	err := gsql.Select(p.Category).
		From(&p).
		Distinct().
		Find(db, &categories)

	if err != nil {
		t.Fatalf("DISTINCT failed: %v", err)
	}
	if len(categories) != 3 {
		t.Errorf("Expected 3 distinct categories, got %d", len(categories))
	}
}

// ==================== Exist Tests ====================

// TestAdv_Exist tests existence check
func TestAdv_Exist(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// Initially no products
	exists, err := gsql.Select(p.ID).
		From(&p).
		Where(p.Category.Eq("Electronics")).
		Exist(db)

	if err != nil {
		t.Fatalf("Exist check failed: %v", err)
	}
	if exists {
		t.Error("Expected no Electronics products to exist")
	}

	// Insert product
	product := Product{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Now should exist
	exists, err = gsql.Select(p.ID).
		From(&p).
		Where(p.Category.Eq("Electronics")).
		Exist(db)

	if err != nil {
		t.Fatalf("Exist check failed: %v", err)
	}
	if !exists {
		t.Error("Expected Electronics products to exist")
	}
}

// ==================== Helper Functions ====================

func ptr[T any](v T) *T {
	return &v
}
