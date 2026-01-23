package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*Int[int64])(nil)

// ==================== Int 定义 ====================

// Int 整数类型表达式，用于 COUNT 等返回整数的聚合函数
// @gentype default=[int]
// 支持比较操作、算术运算、位运算和数学函数
// 使用场景：
//   - COUNT, COUNT_DISTINCT 等聚合函数的返回值
//   - 派生表中的整数列
//   - 整数字段的算术运算结果
type Int[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	nullCondFuncSql
	numericCondFuncSql
	castSql
	formatSql
	bitOpSql
	aggregateSql
	baseExprSql
}

func NewInt[T any](expr clause.Expression) Int[T] {
	return Int[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		numericCondFuncSql:    numericCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
		bitOpSql:              bitOpSql{Expression: expr},
		aggregateSql:          aggregateSql{Expression: expr},
		baseExprSql:           baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// AsFloat 转换为 Float（不生成 SQL，仅类型转换）
func (e Int[T]) AsFloat() Float[float64] {
	return NewFloat[float64](e.numericComparableImpl.Expression)
}

// AsDecimal 转换为 Decimal（不生成 SQL，仅类型转换）
func (e Int[T]) AsDecimal() Decimal[float64] {
	return NewDecimal[float64](e.numericComparableImpl.Expression)
}

// Cast 类型转换 (CAST)
func (e Int[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastFloat 转换为浮点数 (CAST AS DECIMAL)
func (e Int[T]) CastFloat(precision, scale int) Float[float64] {
	return NewFloat[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e Int[T]) CastChar(length ...int) String[string] {
	return NewString[string](e.castCharExpr(length...))
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e Int[T]) CastSigned() Int[int64] {
	return NewInt[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e Int[T]) CastUnsigned() Int[uint64] {
	return NewInt[uint64](e.castUnsignedExpr())
}

// ==================== 字符串转换 ====================

// Hex 转换为十六进制字符串 (HEX)
func (e Int[T]) Hex() String[string] {
	return NewString[string](e.hexExpr())
}

// Bin 转换为二进制字符串 (BIN)
func (e Int[T]) Bin() String[string] {
	return NewString[string](e.binExpr())
}

// Oct 转换为八进制字符串 (OCT)
func (e Int[T]) Oct() String[string] {
	return NewString[string](e.octExpr())
}

// ==================== 网络函数 ====================

// InetNtoa 将整数形式的IP地址转换为点分十进制字符串 (INET_NTOA)
// SELECT INET_NTOA(3232235777); -- 结果为 '192.168.1.1'
// SELECT INET_NTOA(ip_address) FROM access_logs;
func (e Int[T]) InetNtoa() String[string] {
	return NewString[string](clause.Expr{
		SQL:  "INET_NTOA(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// ==================== 日期/时间转换 ====================

// SecToTime 将秒数转换为时间 (SEC_TO_TIME)
// 数据库支持: MySQL
// SELECT SEC_TO_TIME(5400); -- 返回 '01:30:00'
func (e Int[T]) SecToTime() Time[string] {
	return NewTime[string](clause.Expr{
		SQL:  "SEC_TO_TIME(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// FromDays 将天数转换为日期 (FROM_DAYS)
// 数据库支持: MySQL
// SELECT FROM_DAYS(739259); -- 返回 '2024-01-15'
func (e Int[T]) FromDays() Date[string] {
	return NewDate[string](clause.Expr{
		SQL:  "FROM_DAYS(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// ToDateTime 将 Unix 时间戳转换为 DATETIME 类型 (FROM_UNIXTIME)
// 数据库支持: MySQL
// SELECT FROM_UNIXTIME(1698306600); -- 返回 '2023-10-26 10:30:00'
// SELECT FROM_UNIXTIME(created_at) FROM users;
// 可链式调用日期时间方法，如 .Year(), .Format() 等
func (e Int[T]) ToDateTime() DateTime[string] {
	return NewDateTime[string](clause.Expr{
		SQL:  "FROM_UNIXTIME(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}
