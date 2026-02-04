package tutorial

import (
	"strings"
	"testing"

	"github.com/donutnomad/gsql"
)

// ==================== JSON_TABLE Tests ====================

// TestJsonTable_Basic tests basic JSON_TABLE with simple array
// 展平简单 JSON 数组
func TestJsonTable_Basic(t *testing.T) {
	te := TokenExchangeSchema
	setupTable(t, te.ModelType())
	db := getDB()

	// Insert test data with JSON array
	// exchange_rules: [{"token_symbol":"BTC","rate":1.0},{"token_symbol":"ETH","rate":0.05}]
	exchanges := []TokenExchange{
		{
			Name:          "Exchange A",
			ExchangeRules: `[{"token_symbol":"BTC","rate":1.0},{"token_symbol":"ETH","rate":0.05},{"token_symbol":"USDT","rate":0.00003}]`,
		},
		{
			Name:          "Exchange B",
			ExchangeRules: `[{"token_symbol":"BTC","rate":1.1},{"token_symbol":"SOL","rate":0.008}]`,
		},
	}
	if err := db.Create(&exchanges).Error; err != nil {
		t.Fatalf("Failed to create exchanges: %v", err)
	}

	// Build JSON_TABLE to flatten exchange_rules array
	// SQL: JSON_TABLE(exchange_rules, '$[*]' COLUMNS(
	//        symbol VARCHAR(50) PATH '$.token_symbol',
	//        rate DECIMAL(20,10) PATH '$.rate'
	//      )) AS rules
	jsonTable := gsql.JsonTable(te.ExchangeRules, "$[*]").
		AddColumn("symbol", "VARCHAR(50)", "$.token_symbol").
		AddColumn("rate", "DECIMAL(20,10)", "$.rate").
		As("rules")

	// Create virtual table column references
	symbol := gsql.IntFieldOf[string]("rules", "symbol")
	rate := gsql.IntFieldOf[float64]("rules", "rate")

	type Result struct {
		Name   string  `gorm:"column:name"`
		Symbol string  `gorm:"column:symbol"`
		Rate   float64 `gorm:"column:rate"`
	}

	// MySQL: SELECT token_exchanges.name, rules.symbol, rules.rate
	//        FROM token_exchanges
	//        JOIN JSON_TABLE(exchange_rules, '$[*]' COLUMNS(...)) AS rules
	var results []Result
	err := gsql.Select(te.Name, symbol, rate).
		From(&te).
		Join(gsql.Join(jsonTable).OnEmpty()).
		Find(db, &results)

	if err != nil {
		t.Fatalf("JSON_TABLE basic query failed: %v", err)
	}

	// Should have 5 rows total (3 from Exchange A + 2 from Exchange B)
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	t.Logf("JSON_TABLE basic results: %+v", results)
}

// TestJsonTable_NestedPath tests JSON_TABLE with nested path
// 核心场景: 展平嵌套 JSON 数组 $.flow[*].members[*]
func TestJsonTable_NestedPath(t *testing.T) {
	wj := WorkflowJobSchema
	setupTable(t, wj.ModelType())
	db := getDB()

	// Insert test data with nested JSON structure
	// step2: {"flow":[{"step":"approval","members":[{"ext_id":1001},{"ext_id":1002}]},{"step":"review","members":[{"ext_id":2001}]}]}
	jobs := []WorkflowJob{
		{
			BusinessID: 100,
			RefType:    "contract",
			Status:     "pending",
			Step2:      `{"flow":[{"step":"approval","members":[{"ext_id":1001},{"ext_id":1002}]},{"step":"review","members":[{"ext_id":2001}]}]}`,
		},
		{
			BusinessID: 101,
			RefType:    "contract",
			Status:     "completed",
			Step2:      `{"flow":[{"step":"final","members":[{"ext_id":1001},{"ext_id":3001}]}]}`,
		},
		{
			BusinessID: 102,
			RefType:    "invoice",
			Status:     "pending",
			Step2:      `{"flow":[{"step":"check","members":[{"ext_id":4001}]}]}`,
		},
	}
	if err := db.Create(&jobs).Error; err != nil {
		t.Fatalf("Failed to create workflow jobs: %v", err)
	}

	// Build JSON_TABLE with nested path
	// SQL: JSON_TABLE(step2, '$.flow[*].members[*]' COLUMNS(
	//        ext_id BIGINT UNSIGNED PATH '$.ext_id'
	//      )) AS flow_member
	jsonTable := gsql.JsonTable(wj.Step2, "$.flow[*].members[*]").
		AddColumn("ext_id", "BIGINT UNSIGNED", "$.ext_id").
		As("flow_member")

	// Create virtual table column reference
	extID := gsql.IntFieldOf[uint64]("flow_member", "ext_id")

	type Result struct {
		ID     uint64 `gorm:"column:id"`
		Status string `gorm:"column:status"`
	}

	t.Run("Find jobs where user 1001 is a member", func(t *testing.T) {
		// MySQL: SELECT DISTINCT workflow_jobs.id, workflow_jobs.status
		//        FROM workflow_jobs
		//        JOIN JSON_TABLE(step2, '$.flow[*].members[*]' COLUMNS(...)) AS flow_member
		//        WHERE flow_member.ext_id = 1001
		var results []Result
		err := gsql.Select(wj.ID, wj.Status).
			From(&wj).
			Join(gsql.Join(jsonTable).OnEmpty()).
			Where(extID.Eq(1001)).
			Distinct().
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// User 1001 is in job 1 (pending) and job 2 (completed)
		if len(results) != 2 {
			t.Errorf("Expected 2 jobs with user 1001, got %d", len(results))
		}
		t.Logf("Jobs with user 1001: %+v", results)
	})

	t.Run("Find pending jobs where user 1001 is a member", func(t *testing.T) {
		// MySQL: SELECT DISTINCT workflow_jobs.id, workflow_jobs.status
		//        FROM workflow_jobs
		//        JOIN JSON_TABLE(...) AS flow_member
		//        WHERE flow_member.ext_id = 1001 AND workflow_jobs.status = 'pending'
		var results []Result
		err := gsql.Select(wj.ID, wj.Status).
			From(&wj).
			Join(gsql.Join(jsonTable).OnEmpty()).
			Where(extID.Eq(1001), wj.Status.Eq("pending")).
			Distinct().
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// Only job 1 is pending with user 1001
		if len(results) != 1 {
			t.Errorf("Expected 1 pending job with user 1001, got %d", len(results))
		}
		t.Logf("Pending jobs with user 1001: %+v", results)
	})

	t.Run("Find jobs with multiple member IDs", func(t *testing.T) {
		// MySQL: SELECT DISTINCT workflow_jobs.id, workflow_jobs.status
		//        FROM workflow_jobs
		//        JOIN JSON_TABLE(...) AS flow_member
		//        WHERE flow_member.ext_id IN (1001, 2001, 4001)
		var results []Result
		err := gsql.Select(wj.ID, wj.Status).
			From(&wj).
			Join(gsql.Join(jsonTable).OnEmpty()).
			Where(extID.In(1001, 2001, 4001)).
			Distinct().
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// All 3 jobs should match
		if len(results) != 3 {
			t.Errorf("Expected 3 jobs, got %d", len(results))
		}
		t.Logf("Jobs with users 1001/2001/4001: %+v", results)
	})
}

// TestJsonTable_WithFilter tests JSON_TABLE combined with WHERE conditions
func TestJsonTable_WithFilter(t *testing.T) {
	te := TokenExchangeSchema
	setupTable(t, te.ModelType())
	db := getDB()

	// Insert test data
	exchanges := []TokenExchange{
		{
			Name:          "Major Exchange",
			ExchangeRules: `[{"token_symbol":"BTC","rate":1.0,"active":true},{"token_symbol":"ETH","rate":0.05,"active":true}]`,
		},
		{
			Name:          "Minor Exchange",
			ExchangeRules: `[{"token_symbol":"DOGE","rate":0.00001,"active":false}]`,
		},
	}
	if err := db.Create(&exchanges).Error; err != nil {
		t.Fatalf("Failed to create exchanges: %v", err)
	}

	// JSON_TABLE with filter on rate
	jsonTable := gsql.JsonTable(te.ExchangeRules, "$[*]").
		AddColumn("symbol", "VARCHAR(50)", "$.token_symbol").
		AddColumn("rate", "DECIMAL(20,10)", "$.rate").
		As("rules")

	symbol := gsql.IntFieldOf[string]("rules", "symbol")
	rate := gsql.IntFieldOf[float64]("rules", "rate")

	type Result struct {
		Name   string  `gorm:"column:name"`
		Symbol string  `gorm:"column:symbol"`
		Rate   float64 `gorm:"column:rate"`
	}

	// Find tokens with rate > 0.01
	var results []Result
	err := gsql.Select(te.Name, symbol, rate).
		From(&te).
		Join(gsql.Join(jsonTable).OnEmpty()).
		Where(rate.Gt(0.01)).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Only BTC (1.0) and ETH (0.05) have rate > 0.01
	if len(results) != 2 {
		t.Errorf("Expected 2 results with rate > 0.01, got %d", len(results))
	}
	t.Logf("Tokens with rate > 0.01: %+v", results)
}

// TestJsonTable_OnEmptyOnError tests JSON_TABLE with ON EMPTY and ON ERROR handling
func TestJsonTable_OnEmptyOnError(t *testing.T) {
	te := TokenExchangeSchema
	setupTable(t, te.ModelType())
	db := getDB()

	// Insert test data with some missing fields
	exchanges := []TokenExchange{
		{
			Name:          "Complete",
			ExchangeRules: `[{"token_symbol":"BTC","rate":1.0}]`,
		},
		{
			Name:          "Missing Rate",
			ExchangeRules: `[{"token_symbol":"ETH"}]`, // rate is missing
		},
	}
	if err := db.Create(&exchanges).Error; err != nil {
		t.Fatalf("Failed to create exchanges: %v", err)
	}

	// JSON_TABLE with DEFAULT on empty
	// When rate is missing, use default value 0
	jsonTable := gsql.JsonTable(te.ExchangeRules, "$[*]").
		AddColumn("symbol", "VARCHAR(50)", "$.token_symbol").
		AddColumn("rate", "DECIMAL(20,10)", "$.rate", "NULL", "NULL"). // NULL ON EMPTY, NULL ON ERROR
		As("rules")

	symbol := gsql.IntFieldOf[string]("rules", "symbol")
	rate := gsql.IntFieldOf[float64]("rules", "rate")

	type Result struct {
		Name   string   `gorm:"column:name"`
		Symbol string   `gorm:"column:symbol"`
		Rate   *float64 `gorm:"column:rate"` // Nullable
	}

	var results []Result
	err := gsql.Select(te.Name, symbol, rate).
		From(&te).
		Join(gsql.Join(jsonTable).OnEmpty()).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify ETH has NULL rate
	for _, r := range results {
		if r.Symbol == "ETH" && r.Rate != nil {
			t.Errorf("Expected NULL rate for ETH, got %v", *r.Rate)
		}
		if r.Symbol == "BTC" && (r.Rate == nil || *r.Rate != 1.0) {
			t.Errorf("Expected rate 1.0 for BTC, got %v", r.Rate)
		}
	}
	t.Logf("Results with NULL handling: %+v", results)
}

// TestJsonTable_SQLGeneration tests SQL generation for JSON_TABLE
func TestJsonTable_SQLGeneration(t *testing.T) {
	wj := WorkflowJobSchema

	t.Run("Basic JSON_TABLE SQL", func(t *testing.T) {
		jsonTable := gsql.JsonTable(wj.Step2, "$.flow[*].members[*]").
			AddColumn("ext_id", "BIGINT UNSIGNED", "$.ext_id").
			As("flow_member")

		extID := gsql.IntFieldOf[uint64]("flow_member", "ext_id")

		sql := gsql.Select(wj.ID, wj.Status).
			From(&wj).
			Join(gsql.Join(jsonTable).OnEmpty()).
			Where(extID.Eq(1001)).
			Distinct().
			ToSQL()

		t.Logf("Generated SQL: %s", sql)

		// Verify key parts
		if !strings.Contains(sql, "JSON_TABLE") {
			t.Error("SQL should contain JSON_TABLE")
		}
		if !strings.Contains(sql, "$.flow[*].members[*]") {
			t.Error("SQL should contain nested path $.flow[*].members[*]")
		}
		if !strings.Contains(sql, "flow_member") {
			t.Error("SQL should contain alias flow_member")
		}
		if !strings.Contains(sql, "DISTINCT") {
			t.Error("SQL should contain DISTINCT")
		}
	})

	t.Run("JSON_TABLE with ON EMPTY", func(t *testing.T) {
		te := TokenExchangeSchema

		jsonTable := gsql.JsonTable(te.ExchangeRules, "$[*]").
			AddColumn("symbol", "VARCHAR(50)", "$.token_symbol").
			AddColumn("rate", "DECIMAL(20,10)", "$.rate", "NULL"). // NULL ON EMPTY
			As("rules")

		symbol := gsql.IntFieldOf[string]("rules", "symbol")

		sql := gsql.Select(te.Name, symbol).
			From(&te).
			Join(gsql.Join(jsonTable).OnEmpty()).
			ToSQL()

		t.Logf("Generated SQL with ON EMPTY: %s", sql)

		if !strings.Contains(sql, "ON EMPTY") {
			t.Error("SQL should contain ON EMPTY")
		}
	})

	t.Run("JSON_TABLE with default value", func(t *testing.T) {
		te := TokenExchangeSchema

		jsonTable := gsql.JsonTable(te.ExchangeRules, "$[*]").
			AddColumn("symbol", "VARCHAR(50)", "$.token_symbol").
			AddColumn("rate", "DECIMAL(20,10)", "$.rate", "0"). // DEFAULT 0 ON EMPTY
			As("rules")

		symbol := gsql.IntFieldOf[string]("rules", "symbol")

		sql := gsql.Select(te.Name, symbol).
			From(&te).
			Join(gsql.Join(jsonTable).OnEmpty()).
			ToSQL()

		t.Logf("Generated SQL with default: %s", sql)

		if !strings.Contains(sql, "DEFAULT") {
			t.Error("SQL should contain DEFAULT")
		}
		if !strings.Contains(sql, "ON EMPTY") {
			t.Error("SQL should contain ON EMPTY")
		}
	})
}

// TestJsonTable_Aggregation tests JSON_TABLE with aggregation functions
func TestJsonTable_Aggregation(t *testing.T) {
	wj := WorkflowJobSchema
	setupTable(t, wj.ModelType())
	db := getDB()

	// Insert test data
	jobs := []WorkflowJob{
		{
			BusinessID: 100,
			RefType:    "contract",
			Status:     "pending",
			Step2:      `{"flow":[{"members":[{"ext_id":1001},{"ext_id":1002}]}]}`,
		},
		{
			BusinessID: 101,
			RefType:    "contract",
			Status:     "completed",
			Step2:      `{"flow":[{"members":[{"ext_id":1001}]}]}`,
		},
	}
	if err := db.Create(&jobs).Error; err != nil {
		t.Fatalf("Failed to create jobs: %v", err)
	}

	// Count how many times each user appears across all jobs
	jsonTable := gsql.JsonTable(wj.Step2, "$.flow[*].members[*]").
		AddColumn("ext_id", "BIGINT UNSIGNED", "$.ext_id").
		As("flow_member")

	extID := gsql.IntFieldOf[uint64]("flow_member", "ext_id")

	type Result struct {
		ExtID uint64 `gorm:"column:ext_id"`
		Count int64  `gorm:"column:count"`
	}

	var results []Result
	err := gsql.Select(extID, gsql.COUNT().As("count")).
		From(&wj).
		Join(gsql.Join(jsonTable).OnEmpty()).
		GroupBy(extID).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	t.Logf("User appearance counts: %+v", results)

	// User 1001 appears in 2 jobs, user 1002 appears in 1 job
	for _, r := range results {
		if r.ExtID == 1001 && r.Count != 2 {
			t.Errorf("Expected user 1001 to appear 2 times, got %d", r.Count)
		}
		if r.ExtID == 1002 && r.Count != 1 {
			t.Errorf("Expected user 1002 to appear 1 time, got %d", r.Count)
		}
	}
}
