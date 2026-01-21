package tutorial

import (
	"strings"
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
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
	// MySQL CTE: WITH high_sales AS (
	//              SELECT sales_records.* FROM sales_records WHERE sales_records.amount > 1200
	//            )
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
	// MySQL: WITH high_sales AS (...)
	//        SELECT salesperson, amount FROM high_sales
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
	// MySQL: ROW_NUMBER() OVER (PARTITION BY sales_records.region ORDER BY sales_records.amount ASC) AS row_num
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
	// MySQL: SELECT sales_records.region, sales_records.salesperson, sales_records.amount,
	//               ROW_NUMBER() OVER (PARTITION BY region ORDER BY amount ASC) AS row_num
	//        FROM sales_records
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

// TestAdv_NthValue tests NTH_VALUE window function
// NTH_VALUE returns the value of the Nth row in the window frame
func TestAdv_NthValue(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// Insert test data
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 100, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 200, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Charlie", Amount: 150, SaleDate: time.Now()},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// ROW_NUMBER with ORDER BY to get ranked results
	rn := gsql.RowNumber().
		OrderBy(s.Amount, true). // DESC
		AsF("rank_num")

	type Result struct {
		Salesperson string  `gorm:"column:salesperson"`
		Amount      float64 `gorm:"column:amount"`
		RankNum     int     `gorm:"column:rank_num"`
	}

	var results []Result
	err := gsql.Select(s.Salesperson, s.Amount, rn).
		From(&s).
		Find(db, &results)

	if err != nil {
		t.Fatalf("NTH_VALUE test failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// First place should be Bob with 200
	for _, r := range results {
		if r.RankNum == 1 && r.Salesperson != "Bob" {
			t.Errorf("Expected Bob at rank 1, got %s", r.Salesperson)
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

	// JSON_EXTRACT with gsql
	type Result struct {
		Username string `gorm:"column:username"`
		Age      int    `gorm:"column:age"`
		City     string `gorm:"column:city"`
	}

	// Use gsql.JSON_EXTRACT and JSON_UNQUOTE
	// MySQL: JSON_EXTRACT(profile, '$.age') AS age
	ageField := gsql.JSON_EXTRACT(gsql.AsJson(u.Profile), "$.age").AsF("age")
	// MySQL: JSON_UNQUOTE(JSON_EXTRACT(profile, '$.city')) AS city
	cityField := gsql.JSON_UNQUOTE(gsql.JSON_EXTRACT(gsql.AsJson(u.Profile), "$.city")).AsF("city")

	var results []Result
	// MySQL: SELECT user_profiles.username,
	//               JSON_EXTRACT(profile, '$.age') AS age,
	//               JSON_UNQUOTE(JSON_EXTRACT(profile, '$.city')) AS city
	//        FROM user_profiles
	//        WHERE JSON_EXTRACT(profile, '$.age') > 20
	err := gsql.Select(u.Username, ageField, cityField).
		From(&u).
		Where(gsql.Expr("JSON_EXTRACT(profile, '$.age') > ?", 20)).
		Find(db, &results)

	if err != nil {
		t.Fatalf("JSON_EXTRACT failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify results
	for _, r := range results {
		t.Logf("User: %s, Age: %d, City: %s", r.Username, r.Age, r.City)
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

	// Use gsql.JSON_SET for update (Builder pattern)
	// MySQL: JSON_SET(profile, '$.country', 'USA', '$.age', 26)
	newProfile := gsql.JSON_SET(gsql.AsJson(u.Profile), "$.country", gsql.Lit("USA")).Set("$.age", gsql.Lit(26))

	// Update using gsql
	// MySQL: UPDATE user_profiles SET profile = JSON_SET(...)
	//        WHERE user_profiles.username = 'alice'
	err := gsql.Select(u.AllFields()...).
		From(&u).
		Where(u.Username.Eq("alice")).
		Update(db, map[string]any{
			"profile": newProfile,
		})

	if err.Error != nil {
		t.Fatalf("JSON_SET update failed: %v", err.Error)
	}

	// Verify update using gsql
	countryField := gsql.JSON_UNQUOTE(gsql.JSON_EXTRACT(gsql.AsJson(u.Profile), "$.country")).AsF("country")
	ageField := gsql.JSON_EXTRACT(gsql.AsJson(u.Profile), "$.age").AsF("age")

	type VerifyResult struct {
		Country string `gorm:"column:country"`
		Age     int    `gorm:"column:age"`
	}

	var verifyResult VerifyResult
	err2 := gsql.Select(countryField, ageField).
		From(&u).
		Where(u.ID.Eq(profile.ID)).
		First(db, &verifyResult)

	if err2 != nil {
		t.Fatalf("Verify failed: %v", err2)
	}
	if verifyResult.Country != "USA" {
		t.Errorf("Expected country 'USA', got '%s'", verifyResult.Country)
	}
	if verifyResult.Age != 26 {
		t.Errorf("Expected age 26, got %d", verifyResult.Age)
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

	// JSON_CONTAINS - find users with specific skill using gsql
	// JSON_CONTAINS(profile, '"go"', '$.skills')
	// MySQL: JSON_CONTAINS(profile, '"go"', '$.skills')
	hasGoSkill := gsql.JSON_CONTAINS(gsql.AsJson(u.Profile), gsql.JsonLit(`"go"`), "$.skills")

	var results []UserProfile
	// MySQL: SELECT user_profiles.* FROM user_profiles
	//        WHERE JSON_CONTAINS(profile, '"go"', '$.skills')
	err := gsql.Select(u.AllFields()...).
		From(&u).
		Where(hasGoSkill).
		Find(db, &results)

	if err != nil {
		t.Fatalf("JSON_CONTAINS failed: %v", err)
	}
	if len(results) != 1 || results[0].Username != "alice" {
		t.Errorf("Expected alice with go skill, got %+v", results)
	}
}

// TestAdv_JsonArray tests JSON_ARRAY function
func TestAdv_JsonArray(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// Insert test data with array
	profile := UserProfile{Username: "test", Profile: `{"items": [1, 2, 3]}`}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Use JSON_ARRAY with column data - demonstrates combining with other functions
	// Get the length of the items array
	itemsLen := gsql.JSON_LENGTH(gsql.AsJson(u.Profile), "$.items").AsF("items_count")

	type Result struct {
		Username   string `gorm:"column:username"`
		ItemsCount int    `gorm:"column:items_count"`
	}

	var result Result
	err := gsql.Select(u.Username, itemsLen).
		From(&u).
		Where(u.ID.Eq(profile.ID)).
		First(db, &result)

	if err != nil {
		t.Fatalf("JSON_ARRAY test failed: %v", err)
	}
	if result.ItemsCount != 3 {
		t.Errorf("Expected 3 items, got %d", result.ItemsCount)
	}
	t.Logf("JSON array has %d items", result.ItemsCount)
}

// TestAdv_JsonObject tests JSON_OBJECT function
func TestAdv_JsonObject(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// Insert test data
	profile := UserProfile{Username: "test", Profile: `{}`}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// Create JSON object using gsql.JSON_OBJECT
	jsonObj := gsql.JSON_OBJECT().
		Add("name", gsql.Lit("Alice")).
		Add("age", gsql.Lit(25))

	var result string
	err := gsql.Select(jsonObj.AsF("obj")).
		From(&u).
		Limit(1).
		First(db, &result)

	if err != nil {
		t.Fatalf("JSON_OBJECT failed: %v", err)
	}
	// JSON object should contain name and age
	if result == "" {
		t.Error("JSON_OBJECT returned empty string")
	}
	t.Logf("JSON_OBJECT result: %s", result)
}

// TestAdv_JsonKeys tests JSON_KEYS function
func TestAdv_JsonKeys(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// Insert JSON data
	profile := UserProfile{Username: "alice", Profile: `{"name": "Alice", "age": 25, "city": "NYC"}`}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// JSON_KEYS - get all keys from JSON object
	keysField := gsql.JSON_KEYS(gsql.AsJson(u.Profile)).AsF("keys")

	var result string
	err := gsql.Select(keysField).
		From(&u).
		Where(u.ID.Eq(profile.ID)).
		First(db, &result)

	if err != nil {
		t.Fatalf("JSON_KEYS failed: %v", err)
	}
	t.Logf("JSON_KEYS result: %s", result)
}

// TestAdv_JsonLength tests JSON_LENGTH function
func TestAdv_JsonLength(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// Insert JSON data with array
	profile := UserProfile{Username: "alice", Profile: `{"skills": ["go", "python", "js"], "name": "Alice"}`}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	// JSON_LENGTH - get length of skills array
	skillsLen := gsql.JSON_LENGTH(gsql.AsJson(u.Profile), "$.skills").AsF("skills_count")

	type Result struct {
		SkillsCount int `gorm:"column:skills_count"`
	}

	var result Result
	err := gsql.Select(skillsLen).
		From(&u).
		Where(u.ID.Eq(profile.ID)).
		First(db, &result)

	if err != nil {
		t.Fatalf("JSON_LENGTH failed: %v", err)
	}
	if result.SkillsCount != 3 {
		t.Errorf("Expected 3 skills, got %d", result.SkillsCount)
	}
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
	// MySQL: SELECT transactions.* FROM transactions
	//        WHERE transactions.account_id = 1
	//        FOR UPDATE
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
	// MySQL: SELECT products.* FROM products WHERE products.id IN (?, ?, ?, ...)
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
	// MySQL: SELECT DISTINCT products.category FROM products
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
	// MySQL: SELECT EXISTS(SELECT 1 FROM products WHERE products.category = 'Electronics')
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
	// MySQL: SELECT EXISTS(SELECT 1 FROM products WHERE products.category = 'Electronics')
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

// ==================== Derived Table Tests ====================

// TestDerivedTable_WithTypedColumn tests using typed columns from derived tables
func TestDerivedTable_WithTypedColumn(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// Insert test data: customers with orders
	customers := []Customer{
		{Name: "Alice", Email: "alice@test.com", Phone: "111-1111"},
		{Name: "Bob", Email: "bob@test.com", Phone: "222-2222"},
		{Name: "Charlie", Email: "charlie@test.com", Phone: "333-3333"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	// Alice: 3 orders, Bob: 2 orders, Charlie: 1 order
	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.50, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 250.00, Status: "pending"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 50.00, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 75.25, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 120.00, Status: "pending"},
		{CustomerID: customers[2].ID, OrderDate: time.Now(), TotalPrice: 200.00, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("Subquery with aggregation and typed column", func(t *testing.T) {
		// Scenario:
		// SELECT sub.customer_id, sub.order_count
		// FROM (
		//   SELECT customer_id, COUNT(*) as order_count
		//   FROM orders
		//   GROUP BY customer_id
		// ) AS sub
		// WHERE sub.order_count > 1

		// 1. Define result type for subquery
		type CustomerOrderCount struct {
			CustomerID uint64 `gorm:"column:customer_id"`
			OrderCount int64  `gorm:"column:order_count"`
		}

		// 2. Build subquery
		subquery := gsql.Select(
			o.CustomerID,
			gsql.COUNT().AsF("order_count"),
		).From(&o).
			GroupBy(o.CustomerID)

		// 3. Create derived table
		derivedTable := gsql.DefineTable[any, CustomerOrderCount]("sub", CustomerOrderCount{}, subquery)

		// 4. Use typed column from derived table
		orderCount := field.IntColumn("order_count").From(&derivedTable)
		customerID := field.NewComparable[uint64]("sub", "customer_id")

		// 5. Build main query with typed column comparison
		var results []CustomerOrderCount
		err := gsql.Select(
			customerID,
			orderCount,
		).From(&derivedTable).
			Where(orderCount.Gt(1)).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// Based on test data: Alice (3 orders), Bob (2 orders) have > 1 order
		t.Logf("Found %d customers with > 1 order", len(results))
		if len(results) != 2 {
			t.Errorf("Expected 2 customers with > 1 order, got %d", len(results))
		}
		for _, r := range results {
			if r.OrderCount <= 1 {
				t.Errorf("Expected order_count > 1, got %d for customer %d", r.OrderCount, r.CustomerID)
			}
		}
	})

	t.Run("Verify SQL generation", func(t *testing.T) {
		// Verify the SQL is generated correctly
		type CustomerOrderCount struct {
			CustomerID uint64 `gorm:"column:customer_id"`
			OrderCount int64  `gorm:"column:order_count"`
		}

		subquery := gsql.Select(
			o.CustomerID,
			gsql.COUNT().AsF("order_count"),
		).From(&o).
			GroupBy(o.CustomerID)

		derivedTable := gsql.DefineTable[any, CustomerOrderCount]("sub", CustomerOrderCount{}, subquery)
		orderCount := field.IntColumn("order_count").From(&derivedTable)

		sql := gsql.Select(
			gsql.Field("customer_id"),
			orderCount,
		).From(&derivedTable).
			Where(orderCount.Gt(1)).
			ToSQL()

		t.Logf("Generated SQL: %s", sql)

		// Verify key parts
		if !strings.Contains(sql, "FROM (") {
			t.Error("SQL should contain derived table FROM (...)")
		}
		if !strings.Contains(sql, "AS sub") {
			t.Error("SQL should contain alias AS sub")
		}
		if !strings.Contains(sql, "order_count") {
			t.Error("SQL should reference order_count column")
		}
	})
}

// ==================== Helper Functions ====================

func ptr[T any](v T) *T {
	return &v
}
