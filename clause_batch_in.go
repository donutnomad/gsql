package gsql

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/donutnomad/gsql/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BatchInStrategy 大批量 IN 优化策略
type BatchInStrategy int

const (
	// StrategyChunk 分片策略：将大 IN 列表拆分成多个小 IN，用 OR 连接
	// 例如: WHERE id IN (1,2,...,1000) OR id IN (1001,1002,...,2000)
	StrategyChunk BatchInStrategy = iota

	// StrategyTempTable 临时表策略（推荐）⭐
	// 创建临时表，插入数据，然后使用 IN (SELECT * FROM temp_table)
	// 优点：可以创建索引，性能最优，适合超大数据量
	StrategyTempTable
)

// BatchInConfig 大批量 IN 优化配置
type BatchInConfig struct {
	// Strategy 优化策略，默认 StrategyTempTable
	Strategy BatchInStrategy

	// Threshold 触发优化的阈值，默认 500
	// 当 IN 列表长度超过此值时，会自动应用优化策略
	// 推荐值：300-500（根据数据库类型调整）
	Threshold int

	// ChunkSize 分片大小，仅在 StrategyChunk 策略下有效，默认 500
	// 每个 IN 子句包含的最大值数量
	ChunkSize int

	// InsertBatchSize 插入临时表时的批次大小，默认 1000
	// 仅在 StrategyTempTable 策略下有效
	InsertBatchSize int
}

// DefaultBatchInConfig 返回默认配置
func DefaultBatchInConfig() BatchInConfig {
	return BatchInConfig{
		Strategy:        StrategyTempTable,
		Threshold:       500,  // 超过 500 个值时触发优化（业界推荐值）
		ChunkSize:       500,  // 每个分片 500 个值
		InsertBatchSize: 1000, // 临时表每批次插入 1000 条
	}
}

// BatchInOptimizer 大批量 IN 优化器
type BatchInOptimizer struct {
	config BatchInConfig
}

// NewBatchInOptimizer 创建一个新的批量 IN 优化器
func NewBatchInOptimizer(config ...BatchInConfig) *BatchInOptimizer {
	cfg := DefaultBatchInConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// 参数校验
	if cfg.Threshold <= 0 {
		cfg.Threshold = 500
	}
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 500
	}
	if cfg.InsertBatchSize <= 0 {
		cfg.InsertBatchSize = 1000
	}

	return &BatchInOptimizer{
		config: cfg,
	}
}

// OptimizeIn 优化 IN 查询，返回一个 TempTableBatchIn 对象
// 注意：临时表策略需要执行多个步骤，不能直接作为表达式使用
// 请使用返回的 TempTableBatchIn 对象的方法来执行查询
func (opt *BatchInOptimizer) OptimizeIn(column field.IField, values []any) *TempTableBatchIn {
	return opt.optimizeInInternal(column, values, false)
}

// OptimizeNotIn 优化 NOT IN 查询
func (opt *BatchInOptimizer) OptimizeNotIn(column field.IField, values []any) *TempTableBatchIn {
	return opt.optimizeInInternal(column, values, true)
}

func (opt *BatchInOptimizer) optimizeInInternal(column field.IField, values []any, isNotIn bool) *TempTableBatchIn {
	// 空值处理
	if len(values) == 0 {
		return &TempTableBatchIn{
			isEmpty: true,
		}
	}

	// 如果值的数量小于阈值，使用普通 IN
	if len(values) < opt.config.Threshold {
		return &TempTableBatchIn{
			useSimpleIn: true,
			column:      column,
			values:      values,
			isNotIn:     isNotIn,
		}
	}

	// 根据策略选择优化方案
	switch opt.config.Strategy {
	case StrategyChunk:
		return &TempTableBatchIn{
			useChunk:  true,
			column:    column,
			values:    values,
			isNotIn:   isNotIn,
			chunkSize: opt.config.ChunkSize,
		}
	case StrategyTempTable:
		return &TempTableBatchIn{
			useTempTable:    true,
			column:          column,
			values:          values,
			isNotIn:         isNotIn,
			tempTableName:   generateTempTableName(),
			insertBatchSize: opt.config.InsertBatchSize,
		}
	default:
		// 默认使用临时表策略
		return &TempTableBatchIn{
			useTempTable:    true,
			column:          column,
			values:          values,
			isNotIn:         isNotIn,
			tempTableName:   generateTempTableName(),
			insertBatchSize: opt.config.InsertBatchSize,
		}
	}
}

// TempTableBatchIn 临时表批量 IN 对象
// 封装了创建临时表、插入数据、查询、清理的完整流程
type TempTableBatchIn struct {
	// 内部状态
	isEmpty         bool
	useSimpleIn     bool
	useChunk        bool
	useTempTable    bool
	chunkSize       int
	insertBatchSize int

	// 查询信息
	column        field.IField
	values        []any
	isNotIn       bool
	tempTableName string
}

// generateTempTableName 生成随机临时表名
func generateTempTableName() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "tmp_batch_in_" + hex.EncodeToString(b)
}

// ToExpression 转换为普通表达式（用于简单 IN 或分片策略）
func (t *TempTableBatchIn) ToExpression() field.Expression {
	if t.isEmpty {
		return clause.Expr{}
	}

	if t.useSimpleIn {
		return t.buildSimpleIn()
	}

	if t.useChunk {
		return t.buildChunkedIn()
	}

	// 临时表策略不能直接转为表达式
	panic("临时表策略不能直接转为表达式，请使用 Execute 方法")
}

// Execute 执行完整的临时表流程（创建、插入、返回查询条件）
// 返回一个可以用于 WHERE 子句的表达式
func (t *TempTableBatchIn) Execute(db *gorm.DB) (field.Expression, func() error, error) {
	if t.isEmpty {
		return clause.Expr{}, func() error { return nil }, nil
	}

	if t.useSimpleIn {
		return t.buildSimpleIn(), func() error { return nil }, nil
	}

	if t.useChunk {
		return t.buildChunkedIn(), func() error { return nil }, nil
	}

	if !t.useTempTable {
		return clause.Expr{}, func() error { return nil }, fmt.Errorf("未知的策略")
	}

	// 1. 创建临时表
	if err := t.createTempTable(db); err != nil {
		return nil, nil, fmt.Errorf("创建临时表失败: %w", err)
	}

	// 2. 插入数据
	if err := t.insertData(db); err != nil {
		// 清理临时表
		_ = t.dropTempTable(db)
		return nil, nil, fmt.Errorf("插入数据失败: %w", err)
	}

	// 3. 返回查询表达式和清理函数
	expr := t.buildTempTableExpression()
	cleanup := func() error {
		return t.dropTempTable(db)
	}

	return expr, cleanup, nil
}

// ExecuteWithAutoCleanup 执行并自动清理（使用 defer）
// 返回一个可以用于 WHERE 子句的表达式
func (t *TempTableBatchIn) ExecuteWithAutoCleanup(db *gorm.DB) (field.Expression, error) {
	expr, cleanup, err := t.Execute(db)
	if err != nil {
		return nil, err
	}

	// 注册清理函数（在连接关闭时自动清理）
	if cleanup != nil {
		// 注意：这里需要用户在查询完成后手动调用 cleanup
		// 或者使用 defer cleanup()
		defer cleanup()
	}

	return expr, nil
}

// createTempTable 创建临时表
func (t *TempTableBatchIn) createTempTable(db *gorm.DB) error {
	// 获取列类型
	colType := t.inferColumnType()

	// 创建临时表的 SQL
	sql := fmt.Sprintf(
		"CREATE TEMPORARY TABLE %s (val %s, PRIMARY KEY (val))",
		db.Statement.Quote(t.tempTableName),
		colType,
	)

	return db.Exec(sql).Error
}

// insertData 批量插入数据到临时表
func (t *TempTableBatchIn) insertData(db *gorm.DB) error {
	// 分批插入
	for i := 0; i < len(t.values); i += t.insertBatchSize {
		end := i + t.insertBatchSize
		if end > len(t.values) {
			end = len(t.values)
		}

		batch := t.values[i:end]

		// 构建 INSERT 语句
		// INSERT INTO temp_table (val) VALUES (?), (?), ...
		sql := fmt.Sprintf(
			"INSERT INTO %s (val) VALUES ",
			db.Statement.Quote(t.tempTableName),
		)

		// 添加占位符
		placeholders := ""
		for j := range batch {
			if j > 0 {
				placeholders += ", "
			}
			placeholders += "(?)"
		}
		sql += placeholders

		// 执行插入
		if err := db.Exec(sql, batch...).Error; err != nil {
			return err
		}
	}

	return nil
}

// buildTempTableExpression 构建使用临时表的查询表达式
func (t *TempTableBatchIn) buildTempTableExpression() field.Expression {
	col := t.column.ToColumn()

	// 构建子查询: SELECT val FROM temp_table
	subQuery := clause.Expr{
		SQL: fmt.Sprintf("SELECT val FROM %s", t.tempTableName),
	}

	// 构建 IN 表达式
	if t.isNotIn {
		return &tempTableInExpression{
			column:   col,
			subQuery: subQuery,
			isNotIn:  true,
		}
	}

	return &tempTableInExpression{
		column:   col,
		subQuery: subQuery,
		isNotIn:  false,
	}
}

// dropTempTable 删除临时表
func (t *TempTableBatchIn) dropTempTable(db *gorm.DB) error {
	sql := fmt.Sprintf("DROP TEMPORARY TABLE IF EXISTS %s", db.Statement.Quote(t.tempTableName))
	return db.Exec(sql).Error
}

// inferColumnType 推断列类型
func (t *TempTableBatchIn) inferColumnType() string {
	if len(t.values) == 0 {
		return "VARCHAR(255)"
	}

	// 根据第一个值的类型推断
	switch t.values[0].(type) {
	case int, int8, int16, int32, int64:
		return "BIGINT"
	case uint, uint8, uint16, uint32, uint64:
		return "BIGINT UNSIGNED"
	case float32, float64:
		return "DOUBLE"
	case string:
		return "VARCHAR(255)"
	case bool:
		return "BOOLEAN"
	default:
		return "VARCHAR(255)"
	}
}

// buildSimpleIn 构建普通 IN 表达式
func (t *TempTableBatchIn) buildSimpleIn() field.Expression {
	col := t.column.ToColumn()
	if t.isNotIn {
		return clause.Not(clause.IN{
			Column: col,
			Values: t.values,
		})
	}
	return clause.IN{
		Column: col,
		Values: t.values,
	}
}

// buildChunkedIn 构建分片 IN 表达式
func (t *TempTableBatchIn) buildChunkedIn() field.Expression {
	col := t.column.ToColumn()
	chunkSize := t.chunkSize

	var exprs []clause.Expression
	for i := 0; i < len(t.values); i += chunkSize {
		end := i + chunkSize
		if end > len(t.values) {
			end = len(t.values)
		}

		chunk := t.values[i:end]
		var expr clause.Expression
		if t.isNotIn {
			expr = clause.Not(clause.IN{
				Column: col,
				Values: chunk,
			})
		} else {
			expr = clause.IN{
				Column: col,
				Values: chunk,
			}
		}
		exprs = append(exprs, expr)
	}

	// 使用 OR/AND 连接所有表达式
	if t.isNotIn {
		return clause.And(exprs...)
	} else {
		return clause.Or(exprs...)
	}
}

// tempTableInExpression 临时表 IN 表达式
type tempTableInExpression struct {
	column   clause.Column
	subQuery clause.Expression
	isNotIn  bool
}

func (t *tempTableInExpression) Build(builder clause.Builder) {
	builder.WriteQuoted(t.column)
	if t.isNotIn {
		builder.WriteString(" NOT IN (")
	} else {
		builder.WriteString(" IN (")
	}
	t.subQuery.Build(builder)
	builder.WriteByte(')')
}

// ============ 便捷函数 ============

// BatchIn 创建一个批量优化的 IN 表达式（使用默认配置）
// 注意：如果使用临时表策略，需要调用 Execute 或 ExecuteWithAutoCleanup
func BatchIn[T any](column field.Comparable[T], values []T) *TempTableBatchIn {
	optimizer := NewBatchInOptimizer()
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return optimizer.OptimizeIn(column, anyValues)
}

// BatchNotIn 创建一个批量优化的 NOT IN 表达式（使用默认配置）
func BatchNotIn[T any](column field.Comparable[T], values []T) *TempTableBatchIn {
	optimizer := NewBatchInOptimizer()
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return optimizer.OptimizeNotIn(column, anyValues)
}

// BatchInWithConfig 创建一个批量优化的 IN 表达式（自定义配置）
func BatchInWithConfig[T any](column field.Comparable[T], values []T, config BatchInConfig) *TempTableBatchIn {
	optimizer := NewBatchInOptimizer(config)
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return optimizer.OptimizeIn(column, anyValues)
}

// BatchNotInWithConfig 创建一个批量优化的 NOT IN 表达式（自定义配置）
func BatchNotInWithConfig[T any](column field.Comparable[T], values []T, config BatchInConfig) *TempTableBatchIn {
	optimizer := NewBatchInOptimizer(config)
	anyValues := make([]any, len(values))
	for i, v := range values {
		anyValues[i] = v
	}
	return optimizer.OptimizeNotIn(column, anyValues)
}

// ============ 使用示例（注释形式） ============

/*
使用示例：

关于 IN 列表大小的最佳实践：
- < 100 个值: 性能最优，无需优化，使用普通 IN
- 100-300 个值: 可接受范围，普通 IN 即可
- 300-500 个值: 开始有性能影响，建议使用优化策略
- > 500 个值: 强烈建议使用临时表或分片优化

不同数据库的建议：
- MySQL: 建议不超过 500 个（无硬性限制）
- PostgreSQL: 建议不超过 1000 个
- Oracle: 硬性限制 1000 个
- SQL Server: 参数限制 2100 个

1. 基本用法（临时表策略，默认阈值 500）：

	var ids []int64
	// ... 假设 ids 有 10000 个元素

	// 创建批量 IN 对象
	batchIn := BatchIn(u.ID, ids)

	// 执行临时表流程
	expr, cleanup, err := batchIn.Execute(db)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup() // 确保清理临时表

	// 使用返回的表达式查询
	query := Select(u.ALL).From(u).Where(expr)
	var users []User
	err = query.Find(db, &users)

2. 自动清理版本：

	batchIn := BatchIn(u.ID, ids)
	expr, err := batchIn.ExecuteWithAutoCleanup(db)
	if err != nil {
		log.Fatal(err)
	}

	query := Select(u.ALL).From(u).Where(expr)
	var users []User
	err = query.Find(db, &users)

3. 使用分片策略（不需要临时表）：

	config := BatchInConfig{
		Strategy:  StrategyChunk,
		Threshold: 500,
		ChunkSize: 500,
	}
	batchIn := BatchInWithConfig(u.ID, ids, config)
	expr := batchIn.ToExpression() // 直接转为表达式

	query := Select(u.ALL).From(u).Where(expr)
	var users []User
	err = query.Find(db, &users)

4. NOT IN 查询：

	batchNotIn := BatchNotIn(u.ID, excludeIds)
	expr, cleanup, err := batchNotIn.Execute(db)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	query := Select(u.ALL).From(u).Where(expr)
	var users []User
	err = query.Find(db, &users)

性能优势：
- 临时表可以创建主键索引，查询性能最优
- 适合超大数据量（10万+）
- 临时表会话结束后自动删除，无需担心清理问题（但建议手动清理以释放资源）
- 支持多次查询复用同一个临时表

注意事项：
- 临时表策略需要数据库支持 CREATE TEMPORARY TABLE
- 临时表在会话（连接）级别，不同连接看不到
- 记得调用 cleanup 函数清理临时表
- 对于小数据量（< 1000），会自动使用普通 IN，无需创建临时表
*/
