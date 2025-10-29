package gsql

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/donutnomad/gsql/field"
	"gorm.io/gorm/clause"
)

// 模拟字段类型用于测试
type mockField struct {
	tableName  string
	columnName string
}

func (m mockField) ToExpr() field.Expression {
	return clause.Expr{
		SQL: "`" + m.tableName + "`.`" + m.columnName + "`",
	}
}

func (m mockField) IsExpr() bool {
	return false
}

func (m mockField) ToColumn() clause.Column {
	return clause.Column{
		Table: m.tableName,
		Name:  m.columnName,
	}
}

func (m mockField) Name() string {
	return m.columnName
}

func (m mockField) FullName() string {
	if m.tableName != "" {
		return m.tableName + "." + m.columnName
	}
	return m.columnName
}

func (m mockField) As(alias string) field.IField {
	return m
}

func (m mockField) AsPrefix(prefix string) field.IField {
	return m
}

func (m mockField) AsSuffix(suffix string) field.IField {
	return m
}

func (m mockField) Alias() string {
	return ""
}

func TestBatchInOptimizer_SmallList(t *testing.T) {
	t.Log("\n========== 测试 1: 小列表（低于阈值）==========")

	optimizer := NewBatchInOptimizer()
	col := mockField{tableName: "users", columnName: "id"}

	// 测试小列表（低于阈值）
	values := make([]any, 100)
	for i := 0; i < 100; i++ {
		values[i] = i + 1
	}

	batchIn := optimizer.OptimizeIn(col, values)

	// 小列表应该使用简单 IN
	if !batchIn.useSimpleIn {
		t.Error("小列表应该使用简单 IN 策略")
	}

	expr := batchIn.ToExpression()
	sql := buildExprToSQL(expr)

	t.Logf("小列表 IN (100 个值):\n%s\n", truncate(sql, 200))
	t.Log("✓ 使用普通 IN 策略")
}

func TestBatchInOptimizer_ChunkStrategy(t *testing.T) {
	t.Log("\n========== 测试 2: 分片策略 ==========")

	config := BatchInConfig{
		Strategy:  StrategyChunk,
		Threshold: 500,
		ChunkSize: 500,
	}
	optimizer := NewBatchInOptimizer(config)
	col := mockField{tableName: "users", columnName: "id"}

	// 生成 2000 个值
	values := make([]any, 2000)
	for i := 0; i < 2000; i++ {
		values[i] = i + 1
	}

	t.Log("\n--- IN 查询 (2000 个值) ---")
	batchIn := optimizer.OptimizeIn(col, values)
	expr := batchIn.ToExpression()
	sql := buildExprToSQL(expr)
	t.Logf("分片策略 IN:\n%s\n", truncate(sql, 500))

	t.Log("\n--- NOT IN 查询 (2000 个值) ---")
	batchNotIn := optimizer.OptimizeNotIn(col, values)
	notInExpr := batchNotIn.ToExpression()
	notInSQL := buildExprToSQL(notInExpr)
	t.Logf("分片策略 NOT IN:\n%s\n", truncate(notInSQL, 500))
}

func TestBatchInOptimizer_TempTableStrategy(t *testing.T) {
	t.Log("\n========== 测试 3: 临时表策略 ==========")

	config := BatchInConfig{
		Strategy:        StrategyTempTable,
		Threshold:       500,
		InsertBatchSize: 1000,
	}
	optimizer := NewBatchInOptimizer(config)
	col := mockField{tableName: "users", columnName: "id"}

	// 生成 2000 个值
	values := make([]any, 2000)
	for i := 0; i < 2000; i++ {
		values[i] = i + 1
	}

	t.Log("\n--- 临时表 IN 查询 (2000 个值) ---")
	batchIn := optimizer.OptimizeIn(col, values)

	if !batchIn.useTempTable {
		t.Error("应该使用临时表策略")
	}

	t.Logf("临时表名: %s", batchIn.tempTableName)

	// 检查临时表名是否符合格式
	matched, _ := regexp.MatchString(`^tmp_batch_in_[0-9a-f]{16}$`, batchIn.tempTableName)
	if !matched {
		t.Errorf("临时表名格式不正确: %s", batchIn.tempTableName)
	}

	// 构建临时表表达式（模拟）
	expr := batchIn.buildTempTableExpression()
	sql := buildExprToSQL(expr)
	t.Logf("临时表查询表达式:\n%s\n", sql)

	t.Log("\n--- 临时表 NOT IN 查询 ---")
	batchNotIn := optimizer.OptimizeNotIn(col, values)
	notInExpr := batchNotIn.buildTempTableExpression()
	notInSQL := buildExprToSQL(notInExpr)
	t.Logf("临时表 NOT IN 表达式:\n%s\n", notInSQL)
}

func TestBatchInOptimizer_EmptyValues(t *testing.T) {
	t.Log("\n========== 测试 4: 空值列表 ==========")

	optimizer := NewBatchInOptimizer()
	col := mockField{tableName: "users", columnName: "id"}

	values := []any{}

	batchIn := optimizer.OptimizeIn(col, values)

	if !batchIn.isEmpty {
		t.Error("空值列表应该标记为 isEmpty")
	}

	expr := batchIn.ToExpression()
	sql := buildExprToSQL(expr)

	t.Logf("空值 IN:\n%s\n", sql)
	t.Log("✓ 返回空表达式")
}

func TestBatchInOptimizer_DifferentThresholds(t *testing.T) {
	t.Log("\n========== 测试 5: 不同阈值对比 ==========")

	col := mockField{tableName: "users", columnName: "id"}
	values := make([]any, 1500)
	for i := 0; i < 1500; i++ {
		values[i] = i + 1
	}

	thresholds := []int{500, 1000, 2000}

	for _, threshold := range thresholds {
		config := BatchInConfig{
			Strategy:        StrategyTempTable,
			Threshold:       threshold,
			InsertBatchSize: 1000,
		}
		optimizer := NewBatchInOptimizer(config)
		batchIn := optimizer.OptimizeIn(col, values)

		t.Logf("\n阈值 %d (1500 个值):", threshold)
		if threshold < 1500 {
			t.Logf("  ✓ 触发优化，使用临时表策略")
			t.Logf("  临时表名: %s", batchIn.tempTableName)
		} else {
			t.Logf("  ✗ 未触发优化，使用普通 IN")
		}
	}
}

func TestBatchInOptimizer_TempTableName(t *testing.T) {
	t.Log("\n========== 测试 6: 临时表名生成 ==========")

	// 生成多个临时表名，确保唯一性
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := generateTempTableName()
		if names[name] {
			t.Errorf("生成了重复的临时表名: %s", name)
		}
		names[name] = true

		// 检查格式
		matched, _ := regexp.MatchString(`^tmp_batch_in_[0-9a-f]{16}$`, name)
		if !matched {
			t.Errorf("临时表名格式不正确: %s", name)
		}
	}

	t.Logf("✓ 生成了 %d 个唯一的临时表名", len(names))
	t.Logf("示例: %v", getSomeKeys(names, 5))
}

func TestBatchInOptimizer_ColumnTypeInference(t *testing.T) {
	t.Log("\n========== 测试 7: 列类型推断 ==========")

	col := mockField{tableName: "users", columnName: "id"}

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"int64", int64(123), "BIGINT"},
		{"int", int(123), "BIGINT"},
		{"uint64", uint64(123), "BIGINT UNSIGNED"},
		{"float64", float64(123.45), "DOUBLE"},
		{"string", "test", "VARCHAR(255)"},
		{"bool", true, "BOOLEAN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optimizer := NewBatchInOptimizer(BatchInConfig{
				Strategy:  StrategyTempTable,
				Threshold: 1,
			})
			batchIn := optimizer.OptimizeIn(col, []any{tt.value})
			colType := batchIn.inferColumnType()

			t.Logf("值类型: %T, 推断列类型: %s", tt.value, colType)

			if colType != tt.expected {
				t.Errorf("列类型推断错误: 期望 %s, 得到 %s", tt.expected, colType)
			}
		})
	}
}

func TestBatchInOptimizer_PerformanceComparison(t *testing.T) {
	t.Log("\n========== 测试 8: 策略对比 ==========")

	col := mockField{tableName: "users", columnName: "id"}

	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		values := make([]any, size)
		for i := 0; i < size; i++ {
			values[i] = i + 1
		}

		t.Logf("\n--- %d 个值 ---", size)

		// 1. 普通 IN（低于阈值时）
		normalConfig := BatchInConfig{
			Strategy:  StrategyTempTable,
			Threshold: 100000, // 高阈值，强制使用简单 IN
		}
		normalOpt := NewBatchInOptimizer(normalConfig)
		normalBatchIn := normalOpt.OptimizeIn(col, values)
		normalExpr := normalBatchIn.ToExpression()
		normalSQL := buildExprToSQL(normalExpr)
		t.Logf("  普通 IN:       %6d 字符", len(normalSQL))

		// 2. 分片策略
		chunkConfig := BatchInConfig{
			Strategy:  StrategyChunk,
			Threshold: 500,
			ChunkSize: 1000,
		}
		chunkOpt := NewBatchInOptimizer(chunkConfig)
		chunkBatchIn := chunkOpt.OptimizeIn(col, values)
		chunkExpr := chunkBatchIn.ToExpression()
		chunkSQL := buildExprToSQL(chunkExpr)
		t.Logf("  分片策略:      %6d 字符", len(chunkSQL))

		// 3. 临时表策略
		tempTableConfig := BatchInConfig{
			Strategy:        StrategyTempTable,
			Threshold:       500,
			InsertBatchSize: 1000,
		}
		tempTableOpt := NewBatchInOptimizer(tempTableConfig)
		tempTableBatchIn := tempTableOpt.OptimizeIn(col, values)
		if size >= 500 {
			tempTableExpr := tempTableBatchIn.buildTempTableExpression()
			tempTableSQL := buildExprToSQL(tempTableExpr)
			t.Logf("  临时表策略:    %6d 字符 (查询表达式)", len(tempTableSQL))
			t.Logf("               临时表名: %s", tempTableBatchIn.tempTableName)
		}
	}
}

// ============ 辅助函数 ============

// buildExprToSQL 将表达式转换为 SQL 字符串
func buildExprToSQL(expr field.Expression) string {
	if expr == nil {
		return ""
	}

	// 检查是否是空表达式
	if isEmpty(expr) {
		return "(empty expression)"
	}

	builder := &testBuilder{
		sql: "",
	}
	expr.Build(builder)
	return builder.sql
}

func isEmpty(expr field.Expression) bool {
	builder := &testBuilder{sql: ""}
	expr.Build(builder)
	return builder.sql == ""
}

// testBuilder 测试用的 SQL 构建器
type testBuilder struct {
	sql     string
	varIdx  int
	dialect string
}

func (b *testBuilder) WriteQuoted(v interface{}) {
	switch val := v.(type) {
	case string:
		b.sql += "`" + val + "`"
	case clause.Column:
		if val.Table != "" {
			b.sql += "`" + val.Table + "`."
		}
		b.sql += "`" + val.Name + "`"
	default:
		b.sql += fmt.Sprintf("`%v`", v)
	}
}

func (b *testBuilder) WriteString(s string) (int, error) {
	b.sql += s
	return len(s), nil
}

func (b *testBuilder) WriteByte(c byte) error {
	b.sql += string(c)
	return nil
}

func (b *testBuilder) AddVar(writer clause.Writer, vars ...interface{}) {
	for i, v := range vars {
		if i > 0 {
			b.sql += ", "
		}
		b.varIdx++
		// 根据类型格式化变量
		switch val := v.(type) {
		case string:
			b.sql += "'" + val + "'"
		case int, int8, int16, int32, int64:
			b.sql += fmt.Sprintf("%v", val)
		case uint, uint8, uint16, uint32, uint64:
			b.sql += fmt.Sprintf("%v", val)
		case float32, float64:
			b.sql += fmt.Sprintf("%v", val)
		case clause.Expression:
			val.Build(b)
		case []any:
			// 处理数组
			for j, item := range val {
				if j > 0 {
					b.sql += ", "
				}
				b.AddVar(b, item)
			}
		default:
			b.sql += fmt.Sprintf("'%v'", val)
		}
	}
}

func (b *testBuilder) AddError(err error) error {
	return err
}

func (b *testBuilder) Quote(s string) string {
	return "`" + s + "`"
}

// truncate 截断过长的 SQL 用于显示
func truncate(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "... (截断)"
}

// getSomeKeys 获取 map 的一些 key
func getSomeKeys(m map[string]bool, n int) []string {
	keys := make([]string, 0, n)
	for k := range m {
		if len(keys) >= n {
			break
		}
		keys = append(keys, k)
	}
	return keys
}

// ============ Benchmark 测试 ============

func BenchmarkBatchIn_Simple_1000(b *testing.B) {
	col := mockField{tableName: "users", columnName: "id"}
	values := make([]any, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = i + 1
	}

	config := BatchInConfig{
		Strategy:  StrategyTempTable,
		Threshold: 100000, // 强制使用简单 IN
	}
	optimizer := NewBatchInOptimizer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batchIn := optimizer.OptimizeIn(col, values)
		expr := batchIn.ToExpression()
		_ = buildExprToSQL(expr)
	}
}

func BenchmarkBatchIn_Chunk_1000(b *testing.B) {
	col := mockField{tableName: "users", columnName: "id"}
	values := make([]any, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = i + 1
	}

	config := BatchInConfig{
		Strategy:  StrategyChunk,
		Threshold: 500,
		ChunkSize: 500,
	}
	optimizer := NewBatchInOptimizer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batchIn := optimizer.OptimizeIn(col, values)
		expr := batchIn.ToExpression()
		_ = buildExprToSQL(expr)
	}
}

func BenchmarkBatchIn_TempTable_1000(b *testing.B) {
	col := mockField{tableName: "users", columnName: "id"}
	values := make([]any, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = i + 1
	}

	config := BatchInConfig{
		Strategy:        StrategyTempTable,
		Threshold:       500,
		InsertBatchSize: 1000,
	}
	optimizer := NewBatchInOptimizer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batchIn := optimizer.OptimizeIn(col, values)
		expr := batchIn.buildTempTableExpression()
		_ = buildExprToSQL(expr)
	}
}

func BenchmarkBatchIn_TempTable_10000(b *testing.B) {
	col := mockField{tableName: "users", columnName: "id"}
	values := make([]any, 10000)
	for i := 0; i < 10000; i++ {
		values[i] = i + 1
	}

	config := BatchInConfig{
		Strategy:        StrategyTempTable,
		Threshold:       500,
		InsertBatchSize: 1000,
	}
	optimizer := NewBatchInOptimizer(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batchIn := optimizer.OptimizeIn(col, values)
		expr := batchIn.buildTempTableExpression()
		_ = buildExprToSQL(expr)
	}
}
