package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fields"
)

var Sys = sysFunction{}

// ==================== sysFunction 系统函数 ====================

// sysFunction 系统函数结构体，用于获取当前时间等系统级函数
type sysFunction struct {
	dbType DbType
}

func (s sysFunction) Use(db DbType) sysFunction {
	return sysFunction{dbType: db}
}

// UnixTimestamp 返回当前Unix时间戳（秒）
// 数据库支持: MySQL, SQLite, PostgresSQL
// Example:
// - MySQL: 1769153092
// - SQLite: 1769153092
// - PostgresSQL: 1769153092
func (s sysFunction) UnixTimestamp() fields.IntExpr[int64] {
	switch s.dbType {
	case PostgresSQL:
		return fields.NewIntExpr[int64](expr("EXTRACT(EPOCH FROM NOW())::BIGINT"))
	case SQLite:
		return fields.NewIntExpr[int64](expr("strftime('%s', 'now')"))
	default:
		return fields.NewIntExpr[int64](expr("UNIX_TIMESTAMP()"))
	}
}

// CurrentTimestamp 获取系统时区 日期和时间 等于 Now()
// 数据库支持: MySQL, SQLite, PostgresSQL
// Example:
// - MySQL: 2026-01-23 07:17:18
// - SQLite: 2026-01-23 07:17:18
// - PostgresSQL: 2026-01-23 07:17:18.039624+00 (TIMESTAMP WITH TIME ZONE, Use NOW()::TIMESTAMP->2026-01-23 07:18:34.039624)
func (sysFunction) CurrentTimestamp() fields.DateTimeExpr[string] {
	return fields.NewDateTimeExpr[string](expr("CURRENT_TIMESTAMP"))
}

// Now 获取系统时区 日期和时间 等于 CurrentTimestamp()
// 数据库支持: MySQL, SQLite, PostgresSQL
// Example:
// - MySQL: 2026-01-23 07:17:18
// - SQLite: 2026-01-23 07:17:18
// - PostgresSQL: 2026-01-23 07:17:18.039624+00 (TIMESTAMP WITH TIME ZONE, Use NOW()::TIMESTAMP->2026-01-23 07:18:34.039624)
func (s sysFunction) Now() fields.DateTimeExpr[string] {
	if s.dbType == SQLite {
		return fields.NewDateTimeExpr[string](expr("datetime('now')"))
	}
	return fields.NewDateTimeExpr[string](expr("NOW()"))
}

// UtcTimestamp 获取UTC时区 日期和时间
// 数据库支持: MySQL, SQLite, PostgresSQL
// Example:
// - MySQL: 2026-01-23 07:39:41
// - SQLite: 2026-01-23 07:39:41
// - PostgresSQL: 2026-01-23 07:39:41.13382
func (s sysFunction) UtcTimestamp() fields.DateTimeExpr[string] {
	switch s.dbType {
	case PostgresSQL:
		return fields.NewDateTimeExpr[string](expr("(NOW() AT TIME ZONE 'UTC')"))
	case SQLite:
		return fields.NewDateTimeExpr[string](expr("datetime('now')"))
	default:
		return fields.NewDateTimeExpr[string](expr("UTC_TIMESTAMP()"))
	}
}

// CurrentDate 获取系统时区 年月日 (YYYY-MM-DD)
// 数据库支持: MySQL, SQLite, PostgresSQL
// Example:
// - MySQL: 2026-01-23
// - SQLite: 2026-01-23
// - PostgresSQL: 2026-01-23
func (sysFunction) CurrentDate() fields.DateExpr[string] {
	return fields.NewDateExpr[string](expr("CURRENT_DATE"))
}

// CurrentTime 获取系统时区 时分秒 (HH:MM:SS)
// 数据库支持: MySQL, SQLite, PostgresSQL
// Example:
// - MySQL: 07:35:39
// - SQLite: 07:35:39
// - PostgresSQL: 07:35:39.589506+00
func (sysFunction) CurrentTime() fields.TimeExpr[string] {
	return fields.NewTimeExpr[string](expr("CURRENT_TIME"))
}

// ==================== 系统信息函数 ====================

// Database 返回当前使用的数据库名，如果未选择数据库则返回NULL (DATABASE)
// 数据库支持: MySQL, PostgresSQL
// Example:
// - MySQL: my_database
// - PostgresSQL: postgres
func (s sysFunction) Database() fields.TextExpr[string] {
	if s.dbType == PostgresSQL {
		return fields.NewTextExpr[string](expr("current_database()"))
	}
	return fields.NewTextExpr[string](expr("DATABASE()"))
}

// User 返回当前用户名 (USER)
// 数据库支持: MySQL, PostgresSQL
// - MySQL: root@127.0.0.1
// - PostgresSQL: postgres
func (s sysFunction) User() fields.TextExpr[string] {
	if s.dbType == PostgresSQL {
		return fields.NewTextExpr[string](expr("USER"))
	}
	return fields.NewTextExpr[string](expr("USER()"))
}

// CurrentUser 返回当前用户名 (CURRENT_USER)
// 数据库支持: MySQL, PostgresSQL
// - MySQL: root@%
// - PostgresSQL: postgres
func (s sysFunction) CurrentUser() fields.TextExpr[string] {
	if s.dbType == PostgresSQL {
		return fields.NewTextExpr[string](expr("CURRENT_USER"))
	}
	return fields.NewTextExpr[string](expr("CURRENT_USER()"))
}

// Version 返回数据库服务器的版本号 (VERSION)
// 数据库支持: MySQL, PostgresSQL, SQLite
// Example:
// - PostgresSQL: PostgresSQL 17.6 on aarch64-unknown-linux-gnu, compiled by gcc (GCC) 13.2.0, 64-bit
// - MySQL: 8.0.41
// - SQLite: 3.37.2
func (s sysFunction) Version() fields.TextExpr[string] {
	if s.dbType == SQLite {
		return fields.NewTextExpr[string](expr("SQLITE_VERSION()"))
	}
	return fields.NewTextExpr[string](expr("VERSION()"))
}

// UUID 生成一个符合RFC 4122标准的通用唯一标识符（36字符的字符串）(UUID)
// 数据库支持: MySQL, PostgresSQL (PostgresSQL 需要 uuid-ossp 扩展或使用 gen_random_uuid())
// Example:
// - MySQL: 5e881457-f82a-11f0-a148-ee233f69b7a1
func (sysFunction) UUID() fields.TextExpr[string] {
	return fields.NewTextExpr[string](expr("UUID()"))
}

// ==================== 已过时的方法 ====================

// CurDate
// 数据库支持: MySQL
// Deprecated: 请使用 CurrentDate 代替，以获得更好的跨数据库兼容性
func (sysFunction) CurDate() fields.DateExpr[string] {
	return fields.NewDateExpr[string](expr("CURDATE()"))
}

// CurTime
// 数据库支持: MySQL
// Deprecated: 请使用 CurrentTime 代替，以获得更好的跨数据库兼容性
func (sysFunction) CurTime() fields.TimeExpr[string] {
	return fields.NewTimeExpr[string](expr("CURTIME()"))
}

func expr(sql string, vars ...any) clause.Expr {
	return clause.Expr{SQL: sql, Vars: vars}
}
