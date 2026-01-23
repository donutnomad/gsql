package gsql_test

import (
	"strings"
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
)

// TestRowValue_OldSyntax 测试旧语法 VALUES()
func TestRowValue_OldSyntax(t *testing.T) {
	// 确保使用旧语法
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	id := field.NewComparable[int64]("", "id")
	name := field.NewPattern[string]("", "name")
	count := field.NewComparable[int64]("", "count")

	// 测试简单的 RowValue 函数
	valuesExpr := gsql.RowValue(id)
	sql := gsql.Select(valuesExpr.AsF("values_id")).From(gsql.TN("test")).ToSQL()
	t.Logf("RowValue 函数 (旧语法) SQL:\n%s", sql)

	if sql != "SELECT VALUES(`id`) AS `values_id` FROM `test`" {
		t.Errorf("期望生成 VALUES() 语法，实际: %s", sql)
	}

	// 测试 IF + RowValue 组合
	ifExpr := gsql.IF(
		gsql.Expr("? >= ?", gsql.RowValue(count), count),
		gsql.RowValue(name),
		name,
	)
	sql2 := gsql.Select(ifExpr.AsF("result")).From(gsql.TN("test")).ToSQL()
	t.Logf("IF + RowValue 组合 (旧语法) SQL:\n%s", sql2)
}

// TestRowValue_NewSyntax 测试 MySQL 8.0.20+ 新语法
func TestRowValue_NewSyntax(t *testing.T) {
	// 设置为 MySQL 8.0.20+ 模式
	gsql.SetMySQLVersion(gsql.MySQLVersion8020)
	defer gsql.SetMySQLVersion(gsql.MySQLVersionDefault) // 恢复默认

	id := field.NewComparable[int64]("", "id")
	name := field.NewPattern[string]("", "name")
	count := field.NewComparable[int64]("", "count")

	// 测试简单的 RowValue 函数
	valuesExpr := gsql.RowValue(id)
	sql := gsql.Select(valuesExpr.AsF("values_id")).From(gsql.TN("test")).ToSQL()
	t.Logf("RowValue 函数 (新语法) SQL:\n%s", sql)

	if sql != "SELECT `_new`.`id` AS `values_id` FROM `test`" {
		t.Errorf("期望生成 _new.column 语法，实际: %s", sql)
	}

	// 测试 IF + RowValue 组合
	ifExpr := gsql.IF(
		gsql.Expr("? >= ?", gsql.RowValue(count), count),
		gsql.RowValue(name),
		name,
	)
	sql2 := gsql.Select(ifExpr.AsF("result")).From(gsql.TN("test")).ToSQL()
	t.Logf("IF + RowValue 组合 (新语法) SQL:\n%s", sql2)
}

// TestSet_Function 测试 Set 函数
func TestSet_Function(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	count := field.NewComparable[int64]("", "count")
	version := field.NewComparable[int64]("", "version")

	// 测试简单的 Set
	assignment := gsql.Set(count, gsql.RowValue(count))
	t.Logf("Assignment Column: %s", assignment.Column.Name())

	// 测试条件 Set
	assignment2 := gsql.Set(count,
		gsql.IF(
			gsql.Expr("? > ?", gsql.RowValue(version), version),
			gsql.RowValue(count),
			count,
		),
	)
	t.Logf("Conditional Assignment Column: %s", assignment2.Column.Name())
}

// 模拟的消息消费者进度表模型
type MessageConsumerProgress struct {
	ID                    int64     `gorm:"column:id;primaryKey"`
	ConsumerGroup         string    `gorm:"column:consumer_group"`
	LastConsumedMessageID int64     `gorm:"column:last_consumed_message_id"`
	GenerationID          int64     `gorm:"column:generation_id"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
}

func (MessageConsumerProgress) TableName() string {
	return "message_consumer_progress"
}

// MessageConsumerProgressTable 字段定义
type MessageConsumerProgressTable struct {
	ID                    field.Comparable[int64]
	ConsumerGroup         field.Pattern[string]
	LastConsumedMessageID field.Comparable[int64]
	GenerationID          field.Comparable[int64]
	CreatedAt             field.Comparable[time.Time]
	UpdatedAt             field.Comparable[time.Time]
}

func (MessageConsumerProgressTable) TableName() string {
	return "message_consumer_progress"
}

func (MessageConsumerProgressTable) ModelType() MessageConsumerProgress {
	return MessageConsumerProgress{}
}

func NewMessageConsumerProgressTable() MessageConsumerProgressTable {
	tableName := "message_consumer_progress"
	return MessageConsumerProgressTable{
		ID:                    field.NewComparable[int64](tableName, "id"),
		ConsumerGroup:         field.NewPattern[string](tableName, "consumer_group"),
		LastConsumedMessageID: field.NewComparable[int64](tableName, "last_consumed_message_id"),
		GenerationID:          field.NewComparable[int64](tableName, "generation_id"),
		CreatedAt:             field.NewComparable[time.Time](tableName, "created_at"),
		UpdatedAt:             field.NewComparable[time.Time](tableName, "updated_at"),
	}
}

// TestDuplicateUpdateExpr_ConditionalUpdate 测试条件更新场景
func TestDuplicateUpdateExpr_ConditionalUpdate(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	table := NewMessageConsumerProgressTable()

	row := MessageConsumerProgress{
		ID:                    1,
		ConsumerGroup:         "test-group",
		LastConsumedMessageID: 100,
		GenerationID:          5,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// 构建 INSERT ... ON DUPLICATE KEY UPDATE 语句
	builder := gsql.InsertInto(table).
		Value(row).
		DuplicateUpdateExpr(
			gsql.Set(table.LastConsumedMessageID,
				gsql.IF(
					gsql.Expr("? >= ?", gsql.RowValue(table.GenerationID), table.GenerationID),
					gsql.RowValue(table.LastConsumedMessageID),
					table.LastConsumedMessageID,
				),
			),
			gsql.Set(table.GenerationID,
				gsql.IF(
					gsql.Expr("? >= ?", gsql.RowValue(table.GenerationID), table.GenerationID),
					gsql.RowValue(table.GenerationID),
					table.GenerationID,
				),
			),
			gsql.Set(table.UpdatedAt,
				gsql.IF(
					gsql.Expr("? >= ?", gsql.RowValue(table.GenerationID), table.GenerationID),
					gsql.RowValue(table.UpdatedAt),
					table.UpdatedAt,
				),
			),
		)

	if builder == nil {
		t.Fatal("builder should not be nil")
	}

	sql := builder.ToSQL()
	t.Logf("DuplicateUpdateExpr SQL (旧语法 VALUES()):\n%s", sql)

	// 验证生成的 SQL 包含预期的语法
	if !strings.Contains(sql, "VALUES(") {
		t.Errorf("期望包含 VALUES() 语法，实际: %s", sql)
	}
}

// TestDuplicateUpdateExpr_NewSyntax 测试 MySQL 8.0.20+ 语法
func TestDuplicateUpdateExpr_NewSyntax(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersion8020)
	defer gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	table := NewMessageConsumerProgressTable()

	row := MessageConsumerProgress{
		ID:                    1,
		ConsumerGroup:         "test-group",
		LastConsumedMessageID: 100,
		GenerationID:          5,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	builder := gsql.InsertInto(table).
		Value(row).
		DuplicateUpdateExpr(
			gsql.Set(table.LastConsumedMessageID,
				gsql.IF(
					gsql.Expr("? >= ?", gsql.RowValue(table.GenerationID), table.GenerationID),
					gsql.RowValue(table.LastConsumedMessageID),
					table.LastConsumedMessageID,
				),
			),
			gsql.Set(table.GenerationID,
				gsql.IF(
					gsql.Expr("? >= ?", gsql.RowValue(table.GenerationID), table.GenerationID),
					gsql.RowValue(table.GenerationID),
					table.GenerationID,
				),
			),
		)

	if builder == nil {
		t.Fatal("builder should not be nil")
	}

	sql := builder.ToSQL()
	t.Logf("DuplicateUpdateExpr SQL (新语法 MySQL 8.0.20+):\n%s", sql)

	// 验证生成的 SQL 包含预期的语法
	if !strings.Contains(sql, "`_new`.") {
		t.Errorf("期望包含 `_new`. 语法，实际: %s", sql)
	}
}

// TestDuplicateUpdate_Simple 测试简单的 DuplicateUpdate
func TestDuplicateUpdate_Simple(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	table := NewMessageConsumerProgressTable()

	row := MessageConsumerProgress{
		ID:            1,
		ConsumerGroup: "test-group",
	}

	// 简单的列更新
	builder := gsql.InsertInto(table).
		Value(row).
		DuplicateUpdate(table.LastConsumedMessageID, table.GenerationID)

	if builder == nil {
		t.Fatal("builder should not be nil")
	}

	t.Log("Simple DuplicateUpdate builder created successfully")
}

// TestVALUES_Deprecated 测试 VALUES 函数（已弃用，但仍然可用）
func TestVALUES_Deprecated(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	id := field.NewComparable[int64]("", "id")

	// VALUES 和 InsertValue 函数应该和 RowValue 行为相同
	valuesExpr := gsql.VALUES(id)
	insertValueExpr := gsql.InsertValue(id)
	rowValueExpr := gsql.RowValue(id)

	sql1 := gsql.Select(valuesExpr.AsF("v1")).From(gsql.TN("test")).ToSQL()
	sql2 := gsql.Select(insertValueExpr.AsF("v1")).From(gsql.TN("test")).ToSQL()
	sql3 := gsql.Select(rowValueExpr.AsF("v1")).From(gsql.TN("test")).ToSQL()

	if sql1 != sql2 || sql2 != sql3 {
		t.Errorf("VALUES、InsertValue 和 RowValue 应该生成相同的 SQL\nVALUES: %s\nInsertValue: %s\nRowValue: %s", sql1, sql2, sql3)
	}

	t.Logf("VALUES、InsertValue (deprecated) 和 RowValue 生成相同的 SQL: %s", sql1)
}

// TestMySQLVersion_Switch 测试版本切换
func TestMySQLVersion_Switch(t *testing.T) {
	id := field.NewComparable[int64]("", "id")

	// 测试默认版本
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)
	if gsql.GetMySQLVersion() != gsql.MySQLVersionDefault {
		t.Error("GetMySQLVersion 返回值错误")
	}

	sql1 := gsql.Select(gsql.RowValue(id).AsF("v")).From(gsql.TN("test")).ToSQL()
	t.Logf("默认版本 SQL: %s", sql1)

	// 切换到 MySQL 8.0.20+
	gsql.SetMySQLVersion(gsql.MySQLVersion8020)
	if gsql.GetMySQLVersion() != gsql.MySQLVersion8020 {
		t.Error("GetMySQLVersion 返回值错误")
	}

	sql2 := gsql.Select(gsql.RowValue(id).AsF("v")).From(gsql.TN("test")).ToSQL()
	t.Logf("MySQL 8.0.20+ SQL: %s", sql2)

	// 恢复默认
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	if sql1 == sql2 {
		t.Error("两种版本应该生成不同的 SQL")
	}
}

// TestRowValue_InSelect 测试 RowValue 在 SELECT 中的使用（验证表达式构建）
func TestRowValue_InSelect(t *testing.T) {
	id := field.NewComparable[int64]("t", "id")
	name := field.NewPattern[string]("t", "name")
	version := field.NewComparable[int64]("t", "version")

	testCases := []struct {
		name       string
		version    gsql.MySQLVersion
		buildExpr  func() field.ExpressionTo
		expectLike string // 期望包含的子字符串
	}{
		{
			name:       "Simple RowValue (旧语法)",
			version:    gsql.MySQLVersionDefault,
			buildExpr:  func() field.ExpressionTo { return gsql.RowValue(id) },
			expectLike: "VALUES(",
		},
		{
			name:       "Simple RowValue (新语法)",
			version:    gsql.MySQLVersion8020,
			buildExpr:  func() field.ExpressionTo { return gsql.RowValue(id) },
			expectLike: "`_new`.",
		},
		{
			name:    "RowValue with IF (旧语法)",
			version: gsql.MySQLVersionDefault,
			buildExpr: func() field.ExpressionTo {
				return gsql.IF(
					gsql.Expr("? > ?", gsql.RowValue(version), version),
					gsql.RowValue(name),
					name,
				)
			},
			expectLike: "VALUES(",
		},
		{
			name:    "RowValue with IF (新语法)",
			version: gsql.MySQLVersion8020,
			buildExpr: func() field.ExpressionTo {
				return gsql.IF(
					gsql.Expr("? > ?", gsql.RowValue(version), version),
					gsql.RowValue(name),
					name,
				)
			},
			expectLike: "`_new`.",
		},
	}

	for _, tc := range testCases {
		gsql.SetMySQLVersion(tc.version)
		expr := tc.buildExpr()
		sql := gsql.Select(expr.AsF("result")).From(gsql.TN("test")).ToSQL()
		t.Logf("%s SQL:\n%s\n", tc.name, sql)

		if !strings.Contains(sql, tc.expectLike) {
			t.Errorf("%s: 期望包含 %q，实际: %s", tc.name, tc.expectLike, sql)
		}
	}

	// 恢复默认
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)
}
