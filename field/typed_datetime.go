package field

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/donutnomad/gsql/clause"
)

// 允许的时间间隔单位
var allowedIntervalUnits = map[string]bool{
	"MICROSECOND": true,
	"SECOND":      true,
	"MINUTE":      true,
	"HOUR":        true,
	"DAY":         true,
	"WEEK":        true,
	"MONTH":       true,
	"QUARTER":     true,
	"YEAR":        true,
}

// parseInterval 解析并验证时间间隔格式
func parseInterval(interval string, funcName string) string {
	parts := strings.Fields(interval)
	if len(parts) != 2 {
		panic(fmt.Sprintf("%s: invalid interval format, expected '<number> <unit>' (e.g., '1 DAY')", funcName))
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(fmt.Sprintf("%s: interval value must be a number, got: %s", funcName, parts[0]))
	}

	unit := strings.ToUpper(parts[1])
	if !allowedIntervalUnits[unit] {
		panic(fmt.Sprintf("%s: invalid interval unit: %s", funcName, unit))
	}

	return fmt.Sprintf("%d %s", num, unit)
}

var _ clause.Expression = (*DateExpr[string])(nil)

// ==================== DateExpr 定义 ====================

// DateExpr 日期类型表达式，用于 DATE 类型字段 (YYYY-MM-DD)
// 支持日期比较、日期运算和日期函数
// 使用场景：
//   - DATE 类型字段
//   - CURDATE(), CURRENT_DATE() 等函数的返回值
//   - DATE() 函数提取日期部分的结果
type DateExpr[T any] struct {
	baseComparableImpl[T]
	pointerExprImpl
	castSql
}

// NewDateExpr 创建一个新的 DateExpr 实例
func NewDateExpr[T any](expr clause.Expression) DateExpr[T] {
	return DateExpr[T]{
		baseComparableImpl: baseComparableImpl[T]{Expression: expr},
		pointerExprImpl:    pointerExprImpl{Expression: expr},
		castSql:            castSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e DateExpr[T]) Build(builder clause.Builder) {
	e.baseComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e DateExpr[T]) ToExpr() Expression {
	return e.baseComparableImpl.Expression
}

// As 创建一个别名字段
func (e DateExpr[T]) As(alias string) IField {
	return NewBaseFromSql(e.baseComparableImpl.Expression, alias)
}

// ==================== 日期比较 ====================

// Gt 大于 (>)
func (e DateExpr[T]) Gt(value T) clause.Expression {
	return clause.Expr{SQL: "? > ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Gte 大于等于 (>=)
func (e DateExpr[T]) Gte(value T) clause.Expression {
	return clause.Expr{SQL: "? >= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lt 小于 (<)
func (e DateExpr[T]) Lt(value T) clause.Expression {
	return clause.Expr{SQL: "? < ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lte 小于等于 (<=)
func (e DateExpr[T]) Lte(value T) clause.Expression {
	return clause.Expr{SQL: "? <= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Between 范围查询 (BETWEEN ... AND ...)
func (e DateExpr[T]) Between(start, end T) clause.Expression {
	return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// NotBetween 范围外查询 (NOT BETWEEN ... AND ...)
func (e DateExpr[T]) NotBetween(start, end T) clause.Expression {
	return clause.Expr{SQL: "? NOT BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// ==================== 日期提取函数 ====================

// Year 提取年份部分 (YEAR)
// SELECT YEAR(date_column) FROM table;
func (e DateExpr[T]) Year() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "YEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Month 提取月份部分 (MONTH)
// SELECT MONTH(date_column) FROM table;
func (e DateExpr[T]) Month() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MONTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Day 提取天数部分 (DAY)
// SELECT DAY(date_column) FROM table;
func (e DateExpr[T]) Day() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAY(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfMonth 提取一月中的天数 (DAYOFMONTH)，是 Day 的同义词
func (e DateExpr[T]) DayOfMonth() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFMONTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfWeek 返回一周中的索引 (DAYOFWEEK)
// 1=周日, 2=周一, ..., 7=周六
func (e DateExpr[T]) DayOfWeek() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFWEEK(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfYear 返回一年中的天数 (DAYOFYEAR)
// 范围: 1-366
func (e DateExpr[T]) DayOfYear() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFYEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Week 提取周数 (WEEK)
// 范围: 0-53
func (e DateExpr[T]) Week() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "WEEK(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// WeekOfYear 提取周数 (WEEKOFYEAR)
// 范围: 1-53，相当于 WEEK(date, 3)
func (e DateExpr[T]) WeekOfYear() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "WEEKOFYEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Quarter 提取季度 (QUARTER)
// 范围: 1-4
func (e DateExpr[T]) Quarter() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "QUARTER(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 日期运算 ====================

// AddInterval 在日期上增加时间间隔 (DATE_ADD)
// interval 格式: "1 DAY", "2 MONTH", "1 YEAR" 等
// 支持单位: MICROSECOND, SECOND, MINUTE, HOUR, DAY, WEEK, MONTH, QUARTER, YEAR
func (e DateExpr[T]) AddInterval(interval string) DateExpr[T] {
	safeInterval := parseInterval(interval, "AddInterval")
	return NewDateExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_ADD(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// SubInterval 从日期中减去时间间隔 (DATE_SUB)
// interval 格式: "1 DAY", "2 MONTH", "1 YEAR" 等
func (e DateExpr[T]) SubInterval(interval string) DateExpr[T] {
	safeInterval := parseInterval(interval, "SubInterval")
	return NewDateExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_SUB(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DateDiff 计算与另一个日期的差值（天数）(DATEDIFF)
// 返回 this - other 的天数
func (e DateExpr[T]) DateDiff(other Expression) IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DATEDIFF(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, other},
	})
}

// ==================== 日期格式化 ====================

// Format 格式化日期为字符串 (DATE_FORMAT)
// SELECT DATE_FORMAT(date_column, '%Y年%m月%d日') FROM table;
func (e DateExpr[T]) Format(format string) TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "DATE_FORMAT(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, format},
	})
}

// ==================== 日期转换 ====================

// UnixTimestamp 转换为 Unix 时间戳（秒）(UNIX_TIMESTAMP)
func (e DateExpr[T]) UnixTimestamp() IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "UNIX_TIMESTAMP(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e DateExpr[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
func (e DateExpr[T]) CastDatetime() DateTimeExpr[string] {
	return NewDateTimeExpr[string](e.castDatetimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e DateExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 聚合函数 ====================

// Max 返回最大日期 (MAX)
func (e DateExpr[T]) Max() DateExpr[T] {
	return NewDateExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Min 返回最小日期 (MIN)
func (e DateExpr[T]) Min() DateExpr[T] {
	return NewDateExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

var _ clause.Expression = (*DateTimeExpr[string])(nil)

// ==================== DateTimeExpr 定义 ====================

// DateTimeExpr 日期时间类型表达式，用于 DATETIME 类型字段 (YYYY-MM-DD HH:MM:SS)
// 支持日期时间比较、运算和提取函数
// 使用场景：
//   - DATETIME 类型字段
//   - NOW(), CURRENT_TIMESTAMP() 等函数的返回值
//   - FROM_UNIXTIME() 函数的返回值
type DateTimeExpr[T any] struct {
	baseComparableImpl[T]
	pointerExprImpl
	castSql
}

// NewDateTimeExpr 创建一个新的 DateTimeExpr 实例
func NewDateTimeExpr[T any](expr clause.Expression) DateTimeExpr[T] {
	return DateTimeExpr[T]{
		baseComparableImpl: baseComparableImpl[T]{Expression: expr},
		pointerExprImpl:    pointerExprImpl{Expression: expr},
		castSql:            castSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e DateTimeExpr[T]) Build(builder clause.Builder) {
	e.baseComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e DateTimeExpr[T]) ToExpr() Expression {
	return e.baseComparableImpl.Expression
}

// As 创建一个别名字段
func (e DateTimeExpr[T]) As(alias string) IField {
	return NewBaseFromSql(e.baseComparableImpl.Expression, alias)
}

// ==================== 日期时间比较 ====================

// Gt 大于 (>)
func (e DateTimeExpr[T]) Gt(value T) clause.Expression {
	return clause.Expr{SQL: "? > ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Gte 大于等于 (>=)
func (e DateTimeExpr[T]) Gte(value T) clause.Expression {
	return clause.Expr{SQL: "? >= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lt 小于 (<)
func (e DateTimeExpr[T]) Lt(value T) clause.Expression {
	return clause.Expr{SQL: "? < ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lte 小于等于 (<=)
func (e DateTimeExpr[T]) Lte(value T) clause.Expression {
	return clause.Expr{SQL: "? <= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Between 范围查询 (BETWEEN ... AND ...)
func (e DateTimeExpr[T]) Between(start, end T) clause.Expression {
	return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// NotBetween 范围外查询 (NOT BETWEEN ... AND ...)
func (e DateTimeExpr[T]) NotBetween(start, end T) clause.Expression {
	return clause.Expr{SQL: "? NOT BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// ==================== 日期提取函数 ====================

// Year 提取年份部分 (YEAR)
func (e DateTimeExpr[T]) Year() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "YEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Month 提取月份部分 (MONTH)
func (e DateTimeExpr[T]) Month() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MONTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Day 提取天数部分 (DAY)
func (e DateTimeExpr[T]) Day() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAY(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfMonth 提取一月中的天数 (DAYOFMONTH)
func (e DateTimeExpr[T]) DayOfMonth() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFMONTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfWeek 返回一周中的索引 (DAYOFWEEK)
func (e DateTimeExpr[T]) DayOfWeek() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFWEEK(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfYear 返回一年中的天数 (DAYOFYEAR)
func (e DateTimeExpr[T]) DayOfYear() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFYEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Week 提取周数 (WEEK)
func (e DateTimeExpr[T]) Week() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "WEEK(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// WeekOfYear 提取周数 (WEEKOFYEAR)
func (e DateTimeExpr[T]) WeekOfYear() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "WEEKOFYEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Quarter 提取季度 (QUARTER)
func (e DateTimeExpr[T]) Quarter() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "QUARTER(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 时间提取函数 ====================

// Hour 提取小时部分 (HOUR)
// 范围: 0-23
func (e DateTimeExpr[T]) Hour() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "HOUR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Minute 提取分钟部分 (MINUTE)
// 范围: 0-59
func (e DateTimeExpr[T]) Minute() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MINUTE(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Second 提取秒数部分 (SECOND)
// 范围: 0-59
func (e DateTimeExpr[T]) Second() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "SECOND(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Microsecond 提取微秒部分 (MICROSECOND)
// 范围: 0-999999
func (e DateTimeExpr[T]) Microsecond() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MICROSECOND(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 日期时间运算 ====================

// AddInterval 在日期时间上增加时间间隔 (DATE_ADD)
func (e DateTimeExpr[T]) AddInterval(interval string) DateTimeExpr[T] {
	safeInterval := parseInterval(interval, "AddInterval")
	return NewDateTimeExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_ADD(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// SubInterval 从日期时间中减去时间间隔 (DATE_SUB)
func (e DateTimeExpr[T]) SubInterval(interval string) DateTimeExpr[T] {
	safeInterval := parseInterval(interval, "SubInterval")
	return NewDateTimeExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_SUB(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DateDiff 计算与另一个日期的差值（天数）(DATEDIFF)
func (e DateTimeExpr[T]) DateDiff(other Expression) IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DATEDIFF(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, other},
	})
}

// TimeDiff 计算与另一个时间的差值 (TIMEDIFF)
// 返回 TimeExpr 格式
func (e DateTimeExpr[T]) TimeDiff(other Expression) TimeExpr[string] {
	return NewTimeExpr[string](clause.Expr{
		SQL:  "TIMEDIFF(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, other},
	})
}

// TimestampDiff 计算与另一个日期时间的差值（指定单位）(TIMESTAMPDIFF)
// unit: MICROSECOND, SECOND, MINUTE, HOUR, DAY, WEEK, MONTH, QUARTER, YEAR
func (e DateTimeExpr[T]) TimestampDiff(unit string, other Expression) IntExprT[int64] {
	unit = strings.ToUpper(strings.TrimSpace(unit))
	if !allowedIntervalUnits[unit] {
		panic(fmt.Sprintf("TimestampDiff: invalid unit: %s", unit))
	}
	return NewIntExprT[int64](clause.Expr{
		SQL:  fmt.Sprintf("TIMESTAMPDIFF(%s, ?, ?)", unit),
		Vars: []any{other, e.baseComparableImpl.Expression},
	})
}

// ==================== 日期时间格式化 ====================

// Format 格式化日期时间为字符串 (DATE_FORMAT)
func (e DateTimeExpr[T]) Format(format string) TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "DATE_FORMAT(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, format},
	})
}

// ==================== 日期时间转换 ====================

// Date 提取日期部分 (DATE)
func (e DateTimeExpr[T]) Date() DateExpr[string] {
	return NewDateExpr[string](clause.Expr{
		SQL:  "DATE(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Time 提取时间部分 (TIME)
func (e DateTimeExpr[T]) Time() TimeExpr[string] {
	return NewTimeExpr[string](clause.Expr{
		SQL:  "TIME(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// UnixTimestamp 转换为 Unix 时间戳（秒）(UNIX_TIMESTAMP)
func (e DateTimeExpr[T]) UnixTimestamp() IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "UNIX_TIMESTAMP(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e DateTimeExpr[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastDate 转换为 DATE 类型 (CAST AS DATE)
func (e DateTimeExpr[T]) CastDate() DateExpr[string] {
	return NewDateExpr[string](e.castDateExpr())
}

// CastTime 转换为 TIME 类型 (CAST AS TIME)
func (e DateTimeExpr[T]) CastTime() TimeExpr[string] {
	return NewTimeExpr[string](e.castTimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e DateTimeExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 聚合函数 ====================

// Max 返回最大日期时间 (MAX)
func (e DateTimeExpr[T]) Max() DateTimeExpr[T] {
	return NewDateTimeExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Min 返回最小日期时间 (MIN)
func (e DateTimeExpr[T]) Min() DateTimeExpr[T] {
	return NewDateTimeExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

var _ clause.Expression = (*TimeExpr[string])(nil)

// ==================== TimeExpr 定义 ====================

// TimeExpr 时间类型表达式，用于 TIME 类型字段 (HH:MM:SS)
// 支持时间比较、运算和提取函数
// 使用场景：
//   - TIME 类型字段
//   - CURTIME(), CURRENT_TIME() 等函数的返回值
//   - TIME() 函数提取时间部分的结果

type TimeExpr[T any] struct {
	baseComparableImpl[T]
	pointerExprImpl
	castSql
}

// NewTimeExpr 创建一个新的 TimeExpr 实例
func NewTimeExpr[T any](expr clause.Expression) TimeExpr[T] {
	return TimeExpr[T]{
		baseComparableImpl: baseComparableImpl[T]{Expression: expr},
		pointerExprImpl:    pointerExprImpl{Expression: expr},
		castSql:            castSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e TimeExpr[T]) Build(builder clause.Builder) {
	e.baseComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e TimeExpr[T]) ToExpr() Expression {
	return e.baseComparableImpl.Expression
}

// As 创建一个别名字段
func (e TimeExpr[T]) As(alias string) IField {
	return NewBaseFromSql(e.baseComparableImpl.Expression, alias)
}

// ==================== 时间比较 ====================

// Gt 大于 (>)
func (e TimeExpr[T]) Gt(value T) clause.Expression {
	return clause.Expr{SQL: "? > ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Gte 大于等于 (>=)
func (e TimeExpr[T]) Gte(value T) clause.Expression {
	return clause.Expr{SQL: "? >= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lt 小于 (<)
func (e TimeExpr[T]) Lt(value T) clause.Expression {
	return clause.Expr{SQL: "? < ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lte 小于等于 (<=)
func (e TimeExpr[T]) Lte(value T) clause.Expression {
	return clause.Expr{SQL: "? <= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Between 范围查询 (BETWEEN ... AND ...)
func (e TimeExpr[T]) Between(start, end T) clause.Expression {
	return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// NotBetween 范围外查询 (NOT BETWEEN ... AND ...)
func (e TimeExpr[T]) NotBetween(start, end T) clause.Expression {
	return clause.Expr{SQL: "? NOT BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// ==================== 时间提取函数 ====================

// Hour 提取小时部分 (HOUR)
func (e TimeExpr[T]) Hour() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "HOUR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Minute 提取分钟部分 (MINUTE)
func (e TimeExpr[T]) Minute() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MINUTE(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Second 提取秒数部分 (SECOND)
func (e TimeExpr[T]) Second() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "SECOND(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Microsecond 提取微秒部分 (MICROSECOND)
func (e TimeExpr[T]) Microsecond() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MICROSECOND(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 时间运算 ====================

// AddInterval 在时间上增加时间间隔 (ADDTIME 或 DATE_ADD)
func (e TimeExpr[T]) AddInterval(interval string) TimeExpr[T] {
	safeInterval := parseInterval(interval, "AddInterval")
	return NewTimeExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_ADD(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// SubInterval 从时间中减去时间间隔 (SUBTIME 或 DATE_SUB)
func (e TimeExpr[T]) SubInterval(interval string) TimeExpr[T] {
	safeInterval := parseInterval(interval, "SubInterval")
	return NewTimeExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_SUB(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// TimeDiff 计算与另一个时间的差值 (TIMEDIFF)
func (e TimeExpr[T]) TimeDiff(other Expression) TimeExpr[T] {
	return NewTimeExpr[T](clause.Expr{
		SQL:  "TIMEDIFF(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, other},
	})
}

// ==================== 时间格式化 ====================

// Format 格式化时间为字符串 (TIME_FORMAT)
func (e TimeExpr[T]) Format(format string) TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "TIME_FORMAT(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, format},
	})
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e TimeExpr[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
// 注意：TIME 转 DATETIME 时，日期部分为当前日期
func (e TimeExpr[T]) CastDatetime() DateTimeExpr[string] {
	return NewDateTimeExpr[string](e.castDatetimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e TimeExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 聚合函数 ====================

// Max 返回最大时间 (MAX)
func (e TimeExpr[T]) Max() TimeExpr[T] {
	return NewTimeExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Min 返回最小时间 (MIN)
func (e TimeExpr[T]) Min() TimeExpr[T] {
	return NewTimeExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

var _ clause.Expression = (*TimestampExpr[int64])(nil)

// ==================== TimestampExpr 定义 ====================

// TimestampExpr 时间戳类型表达式，用于 TIMESTAMP 类型字段
// TIMESTAMP 类型在存储时会转换为 UTC，读取时转换为当前时区
// 使用场景：
//   - TIMESTAMP 类型字段
//   - 需要时区感知的时间记录
type TimestampExpr[T any] struct {
	baseComparableImpl[T]
	pointerExprImpl
	castSql
}

// NewTimestampExpr 创建一个新的 TimestampExpr 实例
func NewTimestampExpr[T any](expr clause.Expression) TimestampExpr[T] {
	return TimestampExpr[T]{
		baseComparableImpl: baseComparableImpl[T]{Expression: expr},
		pointerExprImpl:    pointerExprImpl{Expression: expr},
		castSql:            castSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e TimestampExpr[T]) Build(builder clause.Builder) {
	e.baseComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e TimestampExpr[T]) ToExpr() Expression {
	return e.baseComparableImpl.Expression
}

// As 创建一个别名字段
func (e TimestampExpr[T]) As(alias string) IField {
	return NewBaseFromSql(e.baseComparableImpl.Expression, alias)
}

// ==================== 时间戳比较 ====================

// Gt 大于 (>)
func (e TimestampExpr[T]) Gt(value T) clause.Expression {
	return clause.Expr{SQL: "? > ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Gte 大于等于 (>=)
func (e TimestampExpr[T]) Gte(value T) clause.Expression {
	return clause.Expr{SQL: "? >= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lt 小于 (<)
func (e TimestampExpr[T]) Lt(value T) clause.Expression {
	return clause.Expr{SQL: "? < ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Lte 小于等于 (<=)
func (e TimestampExpr[T]) Lte(value T) clause.Expression {
	return clause.Expr{SQL: "? <= ?", Vars: []any{e.baseComparableImpl.Expression, value}}
}

// Between 范围查询 (BETWEEN ... AND ...)
func (e TimestampExpr[T]) Between(start, end T) clause.Expression {
	return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// NotBetween 范围外查询 (NOT BETWEEN ... AND ...)
func (e TimestampExpr[T]) NotBetween(start, end T) clause.Expression {
	return clause.Expr{SQL: "? NOT BETWEEN ? AND ?", Vars: []any{e.baseComparableImpl.Expression, start, end}}
}

// ==================== 日期提取函数 ====================

// Year 提取年份部分 (YEAR)
func (e TimestampExpr[T]) Year() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "YEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Month 提取月份部分 (MONTH)
func (e TimestampExpr[T]) Month() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MONTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Day 提取天数部分 (DAY)
func (e TimestampExpr[T]) Day() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAY(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfMonth 提取一月中的天数 (DAYOFMONTH)
func (e TimestampExpr[T]) DayOfMonth() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFMONTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfWeek 返回一周中的索引 (DAYOFWEEK)
func (e TimestampExpr[T]) DayOfWeek() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFWEEK(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DayOfYear 返回一年中的天数 (DAYOFYEAR)
func (e TimestampExpr[T]) DayOfYear() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DAYOFYEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Week 提取周数 (WEEK)
func (e TimestampExpr[T]) Week() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "WEEK(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// WeekOfYear 提取周数 (WEEKOFYEAR)
func (e TimestampExpr[T]) WeekOfYear() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "WEEKOFYEAR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Quarter 提取季度 (QUARTER)
func (e TimestampExpr[T]) Quarter() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "QUARTER(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 时间提取函数 ====================

// Hour 提取小时部分 (HOUR)
func (e TimestampExpr[T]) Hour() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "HOUR(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Minute 提取分钟部分 (MINUTE)
func (e TimestampExpr[T]) Minute() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MINUTE(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Second 提取秒数部分 (SECOND)
func (e TimestampExpr[T]) Second() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "SECOND(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Microsecond 提取微秒部分 (MICROSECOND)
func (e TimestampExpr[T]) Microsecond() IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "MICROSECOND(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 时间戳运算 ====================

// AddInterval 在时间戳上增加时间间隔 (DATE_ADD)
func (e TimestampExpr[T]) AddInterval(interval string) TimestampExpr[T] {
	safeInterval := parseInterval(interval, "AddInterval")
	return NewTimestampExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_ADD(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// SubInterval 从时间戳中减去时间间隔 (DATE_SUB)
func (e TimestampExpr[T]) SubInterval(interval string) TimestampExpr[T] {
	safeInterval := parseInterval(interval, "SubInterval")
	return NewTimestampExpr[T](clause.Expr{
		SQL:  fmt.Sprintf("DATE_SUB(?, INTERVAL %s)", safeInterval),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// DateDiff 计算与另一个日期的差值（天数）(DATEDIFF)
func (e TimestampExpr[T]) DateDiff(other Expression) IntExprT[int] {
	return NewIntExprT[int](clause.Expr{
		SQL:  "DATEDIFF(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, other},
	})
}

// TimeDiff 计算与另一个时间的差值 (TIMEDIFF)
func (e TimestampExpr[T]) TimeDiff(other Expression) TimeExpr[string] {
	return NewTimeExpr[string](clause.Expr{
		SQL:  "TIMEDIFF(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, other},
	})
}

// TimestampDiff 计算与另一个日期时间的差值（指定单位）(TIMESTAMPDIFF)
func (e TimestampExpr[T]) TimestampDiff(unit string, other Expression) IntExprT[int64] {
	unit = strings.ToUpper(strings.TrimSpace(unit))
	if !allowedIntervalUnits[unit] {
		panic(fmt.Sprintf("TimestampDiff: invalid unit: %s", unit))
	}
	return NewIntExprT[int64](clause.Expr{
		SQL:  fmt.Sprintf("TIMESTAMPDIFF(%s, ?, ?)", unit),
		Vars: []any{other, e.baseComparableImpl.Expression},
	})
}

// ==================== 时间戳格式化 ====================

// Format 格式化时间戳为字符串 (DATE_FORMAT)
func (e TimestampExpr[T]) Format(format string) TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "DATE_FORMAT(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, format},
	})
}

// ==================== 时间戳转换 ====================

// Date 提取日期部分 (DATE)
func (e TimestampExpr[T]) Date() DateExpr[string] {
	return NewDateExpr[string](clause.Expr{
		SQL:  "DATE(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Time 提取时间部分 (TIME)
func (e TimestampExpr[T]) Time() TimeExpr[string] {
	return NewTimeExpr[string](clause.Expr{
		SQL:  "TIME(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// UnixTimestamp 转换为 Unix 时间戳（秒）(UNIX_TIMESTAMP)
func (e TimestampExpr[T]) UnixTimestamp() IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "UNIX_TIMESTAMP(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e TimestampExpr[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastDate 转换为 DATE 类型 (CAST AS DATE)
func (e TimestampExpr[T]) CastDate() DateExpr[string] {
	return NewDateExpr[string](e.castDateExpr())
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
func (e TimestampExpr[T]) CastDatetime() DateTimeExpr[string] {
	return NewDateTimeExpr[string](e.castDatetimeExpr())
}

// CastTime 转换为 TIME 类型 (CAST AS TIME)
func (e TimestampExpr[T]) CastTime() TimeExpr[string] {
	return NewTimeExpr[string](e.castTimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e TimestampExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 聚合函数 ====================

// Max 返回最大时间戳 (MAX)
func (e TimestampExpr[T]) Max() TimestampExpr[T] {
	return NewTimestampExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Min 返回最小时间戳 (MIN)
func (e TimestampExpr[T]) Min() TimestampExpr[T] {
	return NewTimestampExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

var _ clause.Expression = (*YearExpr[int64])(nil)

// ==================== YearExpr 定义 ====================

// YearExpr 年份类型表达式，用于 YEAR 类型字段
// YEAR 类型存储年份值，范围通常是 1901-2155
// 使用场景：
//   - YEAR 类型字段
//   - YEAR() 函数提取年份的结果
type YearExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	castSql
}

// NewYearExpr 创建一个新的 YearExpr 实例
func NewYearExpr[T any](expr clause.Expression) YearExpr[T] {
	return YearExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		castSql:               castSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e YearExpr[T]) Build(builder clause.Builder) {
	e.numericComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e YearExpr[T]) ToExpr() Expression {
	return e.numericComparableImpl.Expression
}

// As 创建一个别名字段
func (e YearExpr[T]) As(alias string) IField {
	return NewBaseFromSql(e.numericComparableImpl.Expression, alias)
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e YearExpr[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e YearExpr[T]) CastSigned() IntExprT[int64] {
	return NewIntExprT[int64](e.castSignedExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e YearExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 聚合函数 ====================

// Max 返回最大年份 (MAX)
func (e YearExpr[T]) Max() YearExpr[T] {
	return NewYearExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Min 返回最小年份 (MIN)
func (e YearExpr[T]) Min() YearExpr[T] {
	return NewYearExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}
