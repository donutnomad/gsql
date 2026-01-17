package tutorial

import (
	"context"
	"fmt"
	"os"
	"testing"

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
		tcmysql.WithDatabase("tutorial"),
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
		Logger: logger.Default.LogMode(logger.Silent),
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

// setupTable 创建表并在测试结束后清理
func setupTable[T any](t *testing.T, model *T) {
	t.Helper()
	err := sharedDB.AutoMigrate(model)
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}
	t.Cleanup(func() {
		_ = sharedDB.Migrator().DropTable(model)
	})
}

// getDB 返回共享的数据库连接
func getDB() *gorm.DB {
	return sharedDB
}
