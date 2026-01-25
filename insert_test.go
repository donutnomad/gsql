package gsql_test

import (
	"strings"
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
)

// TestValues 测试泛型 Values 函数
func TestValues(t *testing.T) {
	id := gsql.NewIntField[int64]("", "id")
	name := gsql.NewStringField[string]("", "name")
	count := gsql.NewIntField[int64]("", "count")

	// 测试 Values 函数
	sql := gsql.Select(id.Wrap(gsql.FUNC_VALUES).As("values_id")).From(gsql.TN("test")).ToSQL()
	t.Logf("Values SQL:\n%s", sql)

	if sql != "SELECT VALUES(`id`) AS `values_id` FROM `test`" {
		t.Errorf("期望生成 VALUES() 语法，实际: %s", sql)
	}

	// 测试 Values 的比较方法
	versionCond := count.Wrap(gsql.FUNC_VALUES).GteF(count.ToExpr())
	rowIfExpr := gsql.IF[string](
		versionCond,
		name.Wrap(gsql.FUNC_VALUES),
		name,
	)
	sql2 := gsql.Select(rowIfExpr.As("result")).From(gsql.TN("test")).ToSQL()
	t.Logf("RowIf + Values with comparison SQL:\n%s", sql2)

	if !strings.Contains(sql2, "VALUES(") {
		t.Errorf("期望包含 VALUES()，实际: %s", sql2)
	}
	if !strings.Contains(sql2, ">=") {
		t.Errorf("期望包含 >=，实际: %s", sql2)
	}
}

// TestValues_ComparisonMethods 测试 Values 的各种比较方法
func TestValues_ComparisonMethods(t *testing.T) {
	count := field.NewComparable[int64]("t", "count")

	countF := gsql.NewIntField[int64]("t", "count").Wrap(gsql.FUNC_VALUES)
	testCases := []struct {
		name       string
		buildExpr  func() clause.Expression
		expectLike string
	}{
		{
			name:       "Eq",
			buildExpr:  func() clause.Expression { return countF.Eq(100) },
			expectLike: "VALUES(`t`.`count`) = 100",
		},
		{
			name:       "Not",
			buildExpr:  func() clause.Expression { return countF.Not(100) },
			expectLike: "VALUES(`t`.`count`) != 100",
		},
		{
			name:       "Gt",
			buildExpr:  func() clause.Expression { return countF.Gt(100) },
			expectLike: "VALUES(`t`.`count`) > 100",
		},
		{
			name:       "GteF",
			buildExpr:  func() clause.Expression { return countF.GteF(count.ToExpr()) },
			expectLike: "VALUES(`t`.`count`) >= `t`.`count`",
		},
		{
			name:       "Lt",
			buildExpr:  func() clause.Expression { return countF.Lt(100) },
			expectLike: "VALUES(`t`.`count`) < 100",
		},
		{
			name:       "Lte",
			buildExpr:  func() clause.Expression { return countF.Lte(100) },
			expectLike: "VALUES(`t`.`count`) <= 100",
		},
		{
			name:       "Between",
			buildExpr:  func() clause.Expression { return countF.Between(10, 100) },
			expectLike: "VALUES(`t`.`count`) BETWEEN 10 AND 100",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cond := tc.buildExpr()
			sql := gsql.Select(gsql.Star).From(gsql.TN("test")).Where(cond).ToSQL()
			t.Logf("%s SQL:\n%s", tc.name, sql)

			if !strings.Contains(sql, tc.expectLike) {
				t.Errorf("%s: 期望包含 %q，实际: %s", tc.name, tc.expectLike, sql)
			}
		})
	}
}

// TestSet_Function 测试 Set 函数
func TestSet_Function(t *testing.T) {
	count := gsql.NewIntField[int64]("", "count")
	version := gsql.NewIntField[int64]("", "version")

	// 测试简单的 Set（使用 Values）
	assignment := gsql.Set(count, count.Wrap(gsql.FUNC_VALUES))
	t.Logf("Assignment Column: %s", assignment.Column.Name())

	// 测试条件 Set (使用 RowIf + Values)
	versionCond := version.Wrap(gsql.FUNC_VALUES).GteF(version.ToExpr())
	assignment2 := gsql.Set(count,
		gsql.IF[int64](
			versionCond,
			count.Wrap(gsql.FUNC_VALUES),
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
	ID                    gsql.IntField[int64]
	ConsumerGroup         gsql.StringField[string]
	LastConsumedMessageID gsql.IntField[int64]
	GenerationID          gsql.IntField[int64]
	CreatedAt             gsql.DateTimeField[time.Time]
	UpdatedAt             gsql.DateTimeField[time.Time]
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
		ID:                    gsql.NewIntField[int64](tableName, "id"),
		ConsumerGroup:         gsql.NewStringField[string](tableName, "consumer_group"),
		LastConsumedMessageID: gsql.NewIntField[int64](tableName, "last_consumed_message_id"),
		GenerationID:          gsql.NewIntField[int64](tableName, "generation_id"),
		CreatedAt:             gsql.NewDateTimeField[time.Time](tableName, "created_at"),
		UpdatedAt:             gsql.NewDateTimeField[time.Time](tableName, "updated_at"),
	}
}

// TestDuplicateUpdateExpr_ConditionalUpdate 测试条件更新场景
func TestDuplicateUpdateExpr_ConditionalUpdate(t *testing.T) {
	table := NewMessageConsumerProgressTable()

	row := MessageConsumerProgress{
		ID:                    1,
		ConsumerGroup:         "test-group",
		LastConsumedMessageID: 100,
		GenerationID:          5,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// 条件表达式：新版本号 >= 现有版本号（使用 Values）
	versionCondition := table.GenerationID.Wrap(gsql.FUNC_VALUES).GteF(table.GenerationID.ToExpr())

	// 构建 INSERT ... ON DUPLICATE KEY UPDATE 语句
	builder := gsql.InsertInto(table).
		Value(row).
		DuplicateUpdateExpr(
			gsql.Set(table.LastConsumedMessageID,
				gsql.IF[int64](
					versionCondition,
					table.LastConsumedMessageID.Wrap(gsql.FUNC_VALUES),
					table.LastConsumedMessageID,
				),
			),
			gsql.Set(table.GenerationID,
				gsql.IF[int64](
					versionCondition,
					table.GenerationID.Wrap(gsql.FUNC_VALUES),
					table.GenerationID,
				),
			),
			gsql.Set(table.UpdatedAt,
				gsql.IF[time.Time](
					versionCondition,
					table.UpdatedAt.Wrap(gsql.FUNC_VALUES),
					table.UpdatedAt,
				),
			),
		)

	if builder == nil {
		t.Fatal("builder should not be nil")
	}

	sql := builder.ToSQL()
	t.Logf("DuplicateUpdateExpr SQL:\n%s", sql)

	// 验证生成的 SQL 包含预期的语法
	if !strings.Contains(sql, "VALUES(") {
		t.Errorf("期望包含 VALUES() 语法，实际: %s", sql)
	}
	if !strings.Contains(sql, ">=") {
		t.Errorf("期望包含 >= 比较，实际: %s", sql)
	}
}

// TestDuplicateUpdate_Simple 测试简单的 DuplicateUpdate
func TestDuplicateUpdate_Simple(t *testing.T) {
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

// TestValues_InSelect 测试 Values 在 SELECT 中的使用
func TestValues_InSelect(t *testing.T) {
	id := gsql.NewIntField[int64]("t", "id")
	name := gsql.NewStringField[string]("t", "name")
	version := gsql.NewIntField[int64]("t", "version")

	// 测试简单的 Values
	sql1 := gsql.Select(id.Wrap(gsql.FUNC_VALUES).As("result")).From(gsql.TN("test")).ToSQL()
	t.Logf("Simple Values SQL:\n%s", sql1)
	if !strings.Contains(sql1, "VALUES(") {
		t.Errorf("期望包含 VALUES()，实际: %s", sql1)
	}

	// 测试 RowIf + Values 组合
	versionCond := version.Wrap(gsql.FUNC_VALUES).GtF(version.ToExpr())
	rowIfExpr := gsql.IF[string](
		versionCond,
		name.Wrap(gsql.FUNC_VALUES),
		name,
	)
	sql2 := gsql.Select(rowIfExpr.As("result")).From(gsql.TN("test")).ToSQL()
	t.Logf("RowIf + Values SQL:\n%s", sql2)
	if !strings.Contains(sql2, "VALUES(") {
		t.Errorf("期望包含 VALUES()，实际: %s", sql2)
	}
}
