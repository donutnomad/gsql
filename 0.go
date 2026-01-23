package gsql

import "sync/atomic"

//go:generate go run ./cmd/export-types -src ./internal/fields -dst ./typed_field.go -pkg gsql

// MySQLVersion 表示 MySQL 版本
type MySQLVersion int

const (
	// MySQLVersionDefault 默认版本，使用旧的 VALUES() 语法
	MySQLVersionDefault MySQLVersion = iota
	// MySQLVersion8020 MySQL 8.0.20+，使用新的行别名语法
	MySQLVersion8020
)

var globalMySQLVersion atomic.Int32

// SetMySQLVersion 设置全局 MySQL 版本
// 这会影响 InsertValue 函数生成的 SQL 语法：
//   - MySQLVersionDefault: 生成 VALUES(column) 语法
//   - MySQLVersion8020: 生成 _new.column 语法（需要配合 AS _new 行别名）
func SetMySQLVersion(v MySQLVersion) {
	globalMySQLVersion.Store(int32(v))
}

// GetMySQLVersion 获取当前全局 MySQL 版本设置
func GetMySQLVersion() MySQLVersion {
	return MySQLVersion(globalMySQLVersion.Load())
}

// insertRowAlias 是 MySQL 8.0.20+ 语法中使用的行别名
const insertRowAlias = "_new"
