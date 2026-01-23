package gsql_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 启动 MySQL 8.0.20+ 容器（支持新旧两种语法）
	mysqlContainer, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("test_insert"),
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
	testDB, err = gorm.Open(mysql.Open(connStr), &gorm.Config{
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

// ConsumerProgress 测试用的模型
type ConsumerProgress struct {
	ID                    int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ConsumerGroup         string    `gorm:"column:consumer_group;type:varchar(255);uniqueIndex"`
	LastConsumedMessageID int64     `gorm:"column:last_consumed_message_id"`
	GenerationID          int64     `gorm:"column:generation_id"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
}

func (ConsumerProgress) TableName() string {
	return "consumer_progress"
}

// ConsumerProgressTable 字段定义
type ConsumerProgressTable struct {
	ID                    field.Comparable[int64]
	ConsumerGroup         field.Pattern[string]
	LastConsumedMessageID field.Comparable[int64]
	GenerationID          field.Comparable[int64]
	CreatedAt             field.Comparable[time.Time]
	UpdatedAt             field.Comparable[time.Time]
}

func (ConsumerProgressTable) TableName() string {
	return "consumer_progress"
}

func (ConsumerProgressTable) ModelType() ConsumerProgress {
	return ConsumerProgress{}
}

func NewConsumerProgressTable() ConsumerProgressTable {
	tableName := "consumer_progress"
	return ConsumerProgressTable{
		ID:                    field.NewComparable[int64](tableName, "id"),
		ConsumerGroup:         field.NewPattern[string](tableName, "consumer_group"),
		LastConsumedMessageID: field.NewComparable[int64](tableName, "last_consumed_message_id"),
		GenerationID:          field.NewComparable[int64](tableName, "generation_id"),
		CreatedAt:             field.NewComparable[time.Time](tableName, "created_at"),
		UpdatedAt:             field.NewComparable[time.Time](tableName, "updated_at"),
	}
}

// TestIntegration_DuplicateUpdateExpr_OldSyntax 测试旧语法 VALUES() 在真实数据库中的执行
func TestIntegration_DuplicateUpdateExpr_OldSyntax(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)
	defer gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	// 创建表
	err := testDB.AutoMigrate(&ConsumerProgress{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}
	t.Cleanup(func() {
		_ = testDB.Migrator().DropTable(&ConsumerProgress{})
	})

	table := NewConsumerProgressTable()
	now := time.Now().Truncate(time.Second)

	// 第一次插入
	row1 := ConsumerProgress{
		ConsumerGroup:         "group-old-syntax",
		LastConsumedMessageID: 100,
		GenerationID:          1,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err = gsql.InsertInto(table).
		Value(row1).
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
		).Exec(testDB)

	if err != nil {
		t.Fatalf("First insert failed: %v", err)
	}

	// 验证第一次插入
	var result1 ConsumerProgress
	testDB.Where("consumer_group = ?", "group-old-syntax").First(&result1)
	if result1.LastConsumedMessageID != 100 || result1.GenerationID != 1 {
		t.Errorf("First insert: expected messageID=100, generationID=1, got messageID=%d, generationID=%d",
			result1.LastConsumedMessageID, result1.GenerationID)
	}
	t.Logf("First insert result: messageID=%d, generationID=%d", result1.LastConsumedMessageID, result1.GenerationID)

	// 第二次插入（更高的 generationID，应该更新）
	row2 := ConsumerProgress{
		ConsumerGroup:         "group-old-syntax",
		LastConsumedMessageID: 200,
		GenerationID:          2,
		CreatedAt:             now,
		UpdatedAt:             now.Add(time.Hour),
	}

	err = gsql.InsertInto(table).
		Value(row2).
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
		).Exec(testDB)

	if err != nil {
		t.Fatalf("Second insert failed: %v", err)
	}

	// 验证更新成功（generationID=2 > generationID=1）
	var result2 ConsumerProgress
	testDB.Where("consumer_group = ?", "group-old-syntax").First(&result2)
	if result2.LastConsumedMessageID != 200 || result2.GenerationID != 2 {
		t.Errorf("Second insert: expected messageID=200, generationID=2, got messageID=%d, generationID=%d",
			result2.LastConsumedMessageID, result2.GenerationID)
	}
	t.Logf("Second insert result (should update): messageID=%d, generationID=%d", result2.LastConsumedMessageID, result2.GenerationID)

	// 第三次插入（更低的 generationID，不应该更新）
	row3 := ConsumerProgress{
		ConsumerGroup:         "group-old-syntax",
		LastConsumedMessageID: 50,
		GenerationID:          1,
		CreatedAt:             now,
		UpdatedAt:             now.Add(2 * time.Hour),
	}

	err = gsql.InsertInto(table).
		Value(row3).
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
		).Exec(testDB)

	if err != nil {
		t.Fatalf("Third insert failed: %v", err)
	}

	// 验证没有更新（generationID=1 < generationID=2）
	var result3 ConsumerProgress
	testDB.Where("consumer_group = ?", "group-old-syntax").First(&result3)
	if result3.LastConsumedMessageID != 200 || result3.GenerationID != 2 {
		t.Errorf("Third insert: expected no update (messageID=200, generationID=2), got messageID=%d, generationID=%d",
			result3.LastConsumedMessageID, result3.GenerationID)
	}
	t.Logf("Third insert result (should NOT update): messageID=%d, generationID=%d", result3.LastConsumedMessageID, result3.GenerationID)
}

// TestIntegration_DuplicateUpdateExpr_NewSyntax 测试新语法 _new.column 在真实数据库中的执行
func TestIntegration_DuplicateUpdateExpr_NewSyntax(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersion8020)
	defer gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	// 创建表
	err := testDB.AutoMigrate(&ConsumerProgress{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}
	t.Cleanup(func() {
		_ = testDB.Migrator().DropTable(&ConsumerProgress{})
	})

	table := NewConsumerProgressTable()
	now := time.Now().Truncate(time.Second)

	// 第一次插入
	row1 := ConsumerProgress{
		ConsumerGroup:         "group-new-syntax",
		LastConsumedMessageID: 100,
		GenerationID:          1,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err = gsql.InsertInto(table).
		Value(row1).
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
		).Exec(testDB)

	if err != nil {
		t.Fatalf("First insert failed: %v", err)
	}

	// 验证第一次插入
	var result1 ConsumerProgress
	testDB.Where("consumer_group = ?", "group-new-syntax").First(&result1)
	if result1.LastConsumedMessageID != 100 || result1.GenerationID != 1 {
		t.Errorf("First insert: expected messageID=100, generationID=1, got messageID=%d, generationID=%d",
			result1.LastConsumedMessageID, result1.GenerationID)
	}
	t.Logf("First insert result: messageID=%d, generationID=%d", result1.LastConsumedMessageID, result1.GenerationID)

	// 第二次插入（更高的 generationID，应该更新）
	row2 := ConsumerProgress{
		ConsumerGroup:         "group-new-syntax",
		LastConsumedMessageID: 200,
		GenerationID:          2,
		CreatedAt:             now,
		UpdatedAt:             now.Add(time.Hour),
	}

	err = gsql.InsertInto(table).
		Value(row2).
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
		).Exec(testDB)

	if err != nil {
		t.Fatalf("Second insert failed: %v", err)
	}

	// 验证更新成功（generationID=2 > generationID=1）
	var result2 ConsumerProgress
	testDB.Where("consumer_group = ?", "group-new-syntax").First(&result2)
	if result2.LastConsumedMessageID != 200 || result2.GenerationID != 2 {
		t.Errorf("Second insert: expected messageID=200, generationID=2, got messageID=%d, generationID=%d",
			result2.LastConsumedMessageID, result2.GenerationID)
	}
	t.Logf("Second insert result (should update): messageID=%d, generationID=%d", result2.LastConsumedMessageID, result2.GenerationID)

	// 第三次插入（更低的 generationID，不应该更新）
	row3 := ConsumerProgress{
		ConsumerGroup:         "group-new-syntax",
		LastConsumedMessageID: 50,
		GenerationID:          1,
		CreatedAt:             now,
		UpdatedAt:             now.Add(2 * time.Hour),
	}

	err = gsql.InsertInto(table).
		Value(row3).
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
		).Exec(testDB)

	if err != nil {
		t.Fatalf("Third insert failed: %v", err)
	}

	// 验证没有更新（generationID=1 < generationID=2）
	var result3 ConsumerProgress
	testDB.Where("consumer_group = ?", "group-new-syntax").First(&result3)
	if result3.LastConsumedMessageID != 200 || result3.GenerationID != 2 {
		t.Errorf("Third insert: expected no update (messageID=200, generationID=2), got messageID=%d, generationID=%d",
			result3.LastConsumedMessageID, result3.GenerationID)
	}
	t.Logf("Third insert result (should NOT update): messageID=%d, generationID=%d", result3.LastConsumedMessageID, result3.GenerationID)
}

// TestIntegration_SimpleDuplicateUpdate 测试简单的 DuplicateUpdate
func TestIntegration_SimpleDuplicateUpdate(t *testing.T) {
	gsql.SetMySQLVersion(gsql.MySQLVersionDefault)

	// 创建表
	err := testDB.AutoMigrate(&ConsumerProgress{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}
	t.Cleanup(func() {
		_ = testDB.Migrator().DropTable(&ConsumerProgress{})
	})

	table := NewConsumerProgressTable()
	now := time.Now().Truncate(time.Second)

	// 第一次插入
	row1 := ConsumerProgress{
		ConsumerGroup:         "group-simple",
		LastConsumedMessageID: 100,
		GenerationID:          1,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err = gsql.InsertInto(table).
		Value(row1).
		DuplicateUpdate(table.LastConsumedMessageID, table.GenerationID, table.UpdatedAt).
		Exec(testDB)

	if err != nil {
		t.Fatalf("First insert failed: %v", err)
	}

	// 第二次插入（应该更新所有指定的列）
	row2 := ConsumerProgress{
		ConsumerGroup:         "group-simple",
		LastConsumedMessageID: 200,
		GenerationID:          2,
		CreatedAt:             now,
		UpdatedAt:             now.Add(time.Hour),
	}

	err = gsql.InsertInto(table).
		Value(row2).
		DuplicateUpdate(table.LastConsumedMessageID, table.GenerationID, table.UpdatedAt).
		Exec(testDB)

	if err != nil {
		t.Fatalf("Second insert failed: %v", err)
	}

	// 验证更新
	var result ConsumerProgress
	testDB.Where("consumer_group = ?", "group-simple").First(&result)
	if result.LastConsumedMessageID != 200 || result.GenerationID != 2 {
		t.Errorf("Simple DuplicateUpdate: expected messageID=200, generationID=2, got messageID=%d, generationID=%d",
			result.LastConsumedMessageID, result.GenerationID)
	}
	t.Logf("Simple DuplicateUpdate result: messageID=%d, generationID=%d", result.LastConsumedMessageID, result.GenerationID)
}
