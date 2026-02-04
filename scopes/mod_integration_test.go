package scopes

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 启动 MySQL 8.0 容器
	mysqlContainer, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("test_db"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("password"),
	)
	if err != nil {
		fmt.Printf("Failed to start MySQL container: %v\n", err)
		os.Exit(1)
	}

	// 获取连接字符串
	connStr, err := mysqlContainer.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		fmt.Printf("Failed to get connection string: %v\n", err)
		_ = mysqlContainer.Terminate(ctx)
		os.Exit(1)
	}

	// 连接数据库
	sharedDB, err = gorm.Open(mysql.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		_ = mysqlContainer.Terminate(ctx)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()

	// 清理容器
	_ = mysqlContainer.Terminate(ctx)

	os.Exit(code)
}

// ==================== 测试模型 ====================

// EventLog 事件日志表 - 使用 time.Time 类型的 created_at
type EventLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	EventName string    `gorm:"size:100;not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (EventLog) TableName() string {
	return "event_logs"
}

// EventLogSchema 字段定义
type EventLogSchema struct {
	tableName string
	ID        gsql.IntField[uint64]
	EventName gsql.IntField[string]
	CreatedAt gsql.IntField[time.Time]
}

func (t EventLogSchema) TableName() string {
	return t.tableName
}

func (t EventLogSchema) Alias() string {
	return ""
}

var eventLogSchema = EventLogSchema{
	tableName: "event_logs",
	ID:        gsql.IntFieldOf[uint64]("event_logs", "id"),
	EventName: gsql.IntFieldOf[string]("event_logs", "event_name"),
	CreatedAt: gsql.IntFieldOf[time.Time]("event_logs", "created_at"),
}

// Transaction 交易表 - 使用 int64 类型的 created_at (Unix 时间戳)
type Transaction struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	Amount    int64  `gorm:"not null"`
	CreatedAt int64  `gorm:"not null"` // Unix 时间戳
}

func (Transaction) TableName() string {
	return "transactions"
}

// TransactionSchema 字段定义
type TransactionSchema struct {
	tableName string
	ID        fields.IntField[uint64]
	Amount    fields.IntField[int64]
	CreatedAt fields.IntField[int64]
}

func (t TransactionSchema) TableName() string {
	return t.tableName
}

func (t TransactionSchema) Alias() string {
	return ""
}

var transactionSchema = TransactionSchema{
	tableName: "transactions",
	ID:        gsql.IntFieldOf[uint64]("transactions", "id"),
	Amount:    gsql.IntFieldOf[int64]("transactions", "amount"),
	CreatedAt: gsql.IntFieldOf[int64]("transactions", "created_at"),
}

// ==================== 测试辅助函数 ====================

func setupEventLogTable(t *testing.T) {
	t.Helper()
	err := sharedDB.AutoMigrate(&EventLog{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sharedDB.Migrator().DropTable(&EventLog{})
	})
}

func setupTransactionTable(t *testing.T) {
	t.Helper()
	err := sharedDB.AutoMigrate(&Transaction{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sharedDB.Migrator().DropTable(&Transaction{})
	})
}

// ==================== TimeBetween 集成测试 ====================

// TestTimeBetween_TimeField_WithTimestampRange 测试 time.Time 字段 + TimestampRange 参数
// 这个场景会触发 FROM_UNIXTIME 转换
func TestTimeBetween_TimeField_WithTimestampRange(t *testing.T) {
	setupEventLogTable(t)

	// 准备测试数据
	now := time.Now().UTC()
	events := []EventLog{
		{EventName: "event1", CreatedAt: now.Add(-2 * time.Hour)},
		{EventName: "event2", CreatedAt: now.Add(-1 * time.Hour)},
		{EventName: "event3", CreatedAt: now},
		{EventName: "event4", CreatedAt: now.Add(1 * time.Hour)},
		{EventName: "event5", CreatedAt: now.Add(2 * time.Hour)},
	}
	require.NoError(t, sharedDB.Create(&events).Error)

	// 定义时间范围：1.5小时前 到 0.5小时后
	fromTs := now.Add(-90 * time.Minute).Unix()
	toTs := now.Add(30 * time.Minute).Unix()

	// 使用 TimeBetween 查询
	query := gsql.SelectG[EventLog]().
		From(eventLogSchema).
		Scope(TimeBetween(eventLogSchema.CreatedAt, TimestampRange{
			From: mo.Some(fromTs),
			To:   mo.Some(toTs),
		}))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 应该返回 event2 和 event3
	assert.Len(t, results, 2)
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.EventName
	}
	assert.Contains(t, names, "event2")
	assert.Contains(t, names, "event3")
}

// TestTimeBetween_TimeField_WithTimeRange 测试 time.Time 字段 + TimeRange 参数
func TestTimeBetween_TimeField_WithTimeRange(t *testing.T) {
	setupEventLogTable(t)

	// 准备测试数据
	now := time.Now().UTC()
	events := []EventLog{
		{EventName: "event1", CreatedAt: now.Add(-2 * time.Hour)},
		{EventName: "event2", CreatedAt: now.Add(-1 * time.Hour)},
		{EventName: "event3", CreatedAt: now},
		{EventName: "event4", CreatedAt: now.Add(1 * time.Hour)},
	}
	require.NoError(t, sharedDB.Create(&events).Error)

	// 定义时间范围
	fromTime := now.Add(-90 * time.Minute)
	toTime := now.Add(30 * time.Minute)

	// 使用 TimeBetween 查询
	query := gsql.SelectG[EventLog]().
		From(eventLogSchema).
		Scope(TimeBetween(eventLogSchema.CreatedAt, TimeRange{
			From: mo.Some(fromTime),
			To:   mo.Some(toTime),
		}))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 应该返回 event2 和 event3
	assert.Len(t, results, 2)
}

// TestTimeBetween_IntField_WithTimestampRange 测试 int64 字段 + TimestampRange 参数
func TestTimeBetween_IntField_WithTimestampRange(t *testing.T) {
	setupTransactionTable(t)

	// 准备测试数据
	now := time.Now().Unix()
	transactions := []Transaction{
		{Amount: 100, CreatedAt: now - 7200}, // 2小时前
		{Amount: 200, CreatedAt: now - 3600}, // 1小时前
		{Amount: 300, CreatedAt: now},        // 现在
		{Amount: 400, CreatedAt: now + 3600}, // 1小时后
	}
	require.NoError(t, sharedDB.Create(&transactions).Error)

	// 定义时间范围
	fromTs := now - 5400 // 1.5小时前
	toTs := now + 1800   // 0.5小时后

	// 使用 TimeBetween 查询
	query := gsql.SelectG[Transaction]().
		From(transactionSchema).
		Scope(TimeBetween(transactionSchema.CreatedAt, TimestampRange{
			From: mo.Some(fromTs),
			To:   mo.Some(toTs),
		}))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 应该返回 amount=200 和 amount=300 的记录
	assert.Len(t, results, 2)
	amounts := make([]int64, len(results))
	for i, r := range results {
		amounts[i] = r.Amount
	}
	assert.Contains(t, amounts, int64(200))
	assert.Contains(t, amounts, int64(300))
}

// TestTimeBetween_IntField_WithTimeRange 测试 int64 字段 + TimeRange 参数
// 这个场景会触发 FROM_UNIXTIME 转换
func TestTimeBetween_IntField_WithTimeRange(t *testing.T) {
	setupTransactionTable(t)

	// 准备测试数据
	now := time.Now()
	transactions := []Transaction{
		{Amount: 100, CreatedAt: now.Add(-2 * time.Hour).Unix()},
		{Amount: 200, CreatedAt: now.Add(-1 * time.Hour).Unix()},
		{Amount: 300, CreatedAt: now.Unix()},
		{Amount: 400, CreatedAt: now.Add(1 * time.Hour).Unix()},
	}
	require.NoError(t, sharedDB.Create(&transactions).Error)

	// 定义时间范围
	fromTime := now.Add(-90 * time.Minute)
	toTime := now.Add(30 * time.Minute)

	// 使用 TimeBetween 查询
	query := gsql.SelectG[Transaction]().
		From(transactionSchema).
		Scope(TimeBetween(transactionSchema.CreatedAt, TimeRange{
			From: mo.Some(fromTime),
			To:   mo.Some(toTime),
		}))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 应该返回 amount=200 和 amount=300 的记录
	assert.Len(t, results, 2)
}

// TestTimeBetween_CustomOperators 测试自定义操作符
func TestTimeBetween_CustomOperators(t *testing.T) {
	setupEventLogTable(t)

	// 准备测试数据 - 使用整秒时间避免毫秒精度问题
	now := time.Now().UTC().Truncate(time.Second)
	events := []EventLog{
		{EventName: "event1", CreatedAt: now.Add(-2 * time.Hour)},
		{EventName: "event2", CreatedAt: now.Add(-1 * time.Hour)},
		{EventName: "event3", CreatedAt: now.Add(-30 * time.Minute)}, // 在范围内
		{EventName: "event4", CreatedAt: now.Add(1 * time.Hour)},
	}
	require.NoError(t, sharedDB.Create(&events).Error)

	// 时间范围: -1小时 到 现在
	// 操作符: > 和 <=
	// 预期: event3 (-30分钟) 满足 > (-1小时) 且 <= (现在)
	fromTs := now.Add(-1 * time.Hour).Unix()
	toTs := now.Unix()

	// 使用自定义操作符 > 和 <=
	query := gsql.SelectG[EventLog]().
		From(eventLogSchema).
		Scope(TimeBetween(eventLogSchema.CreatedAt, TimestampRange{
			From: mo.Some(fromTs),
			To:   mo.Some(toTs),
		}, ">", "<="))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 使用 > 和 <= 应该只返回 event3
	assert.Len(t, results, 1)
	if len(results) > 0 {
		assert.Equal(t, "event3", results[0].EventName)
	}
}

// TestTimeBetween_OnlyFrom 测试只有 From 没有 To
func TestTimeBetween_OnlyFrom(t *testing.T) {
	setupEventLogTable(t)

	// 准备测试数据
	now := time.Now().UTC()
	events := []EventLog{
		{EventName: "event1", CreatedAt: now.Add(-2 * time.Hour)},
		{EventName: "event2", CreatedAt: now.Add(-1 * time.Hour)},
		{EventName: "event3", CreatedAt: now},
	}
	require.NoError(t, sharedDB.Create(&events).Error)

	fromTs := now.Add(-90 * time.Minute).Unix()

	// 只设置 From
	query := gsql.SelectG[EventLog]().
		From(eventLogSchema).
		Scope(TimeBetween(eventLogSchema.CreatedAt, TimestampRange{
			From: mo.Some(fromTs),
		}))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 应该返回 event2 和 event3
	assert.Len(t, results, 2)
}

// TestTimeBetween_OnlyTo 测试只有 To 没有 From
func TestTimeBetween_OnlyTo(t *testing.T) {
	setupEventLogTable(t)

	// 准备测试数据
	now := time.Now().UTC()
	events := []EventLog{
		{EventName: "event1", CreatedAt: now.Add(-2 * time.Hour)},
		{EventName: "event2", CreatedAt: now.Add(-1 * time.Hour)},
		{EventName: "event3", CreatedAt: now},
	}
	require.NoError(t, sharedDB.Create(&events).Error)

	toTs := now.Add(-90 * time.Minute).Unix()

	// 只设置 To
	query := gsql.SelectG[EventLog]().
		From(eventLogSchema).
		Scope(TimeBetween(eventLogSchema.CreatedAt, TimestampRange{
			To: mo.Some(toTs),
		}))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 应该只返回 event1
	assert.Len(t, results, 1)
	assert.Equal(t, "event1", results[0].EventName)
}

// TestTimeBetween_EmptyRange 测试空范围（From 和 To 都没有设置）
func TestTimeBetween_EmptyRange(t *testing.T) {
	setupEventLogTable(t)

	// 准备测试数据
	now := time.Now().UTC()
	events := []EventLog{
		{EventName: "event1", CreatedAt: now.Add(-1 * time.Hour)},
		{EventName: "event2", CreatedAt: now},
	}
	require.NoError(t, sharedDB.Create(&events).Error)

	// 空范围
	query := gsql.SelectG[EventLog]().
		From(eventLogSchema).
		Scope(TimeBetween(eventLogSchema.CreatedAt, TimestampRange{}))

	results, err := query.Find(sharedDB)
	require.NoError(t, err)

	// 应该返回所有记录
	assert.Len(t, results, 2)
}
