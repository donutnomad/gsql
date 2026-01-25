package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*IntExpr[int64])(nil)

type intExpr[T any] = IntExpr[T]

// ==================== IntExpr 定义 ====================

// IntExpr 整数类型表达式，用于 COUNT 等返回整数的聚合函数
// @gentype default=[int]
// 支持比较操作、算术运算、位运算和数学函数
// 使用场景：
//   - COUNT, COUNT_DISTINCT 等聚合函数的返回值
//   - 派生表中的整数列
//   - 整数字段的算术运算结果
type IntExpr[T any] struct {
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

type IntConstraint interface {
	clause.Expr |
		IntExpr[int] | IntExpr[uint] |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Int creates an IntExpr[int64] from a clause expression.
func Int(expr clause.Expression) IntExpr[int64] {
	return IntOf[int64](expr)
}

// Uint creates an IntExpr[uint64] from a clause expression.
func Uint(expr clause.Expression) IntExpr[uint64] {
	return IntOf[uint64](expr)
}

// IntVal creates an IntExpr from a signed integer literal value.
func IntVal[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | any](val T) IntExpr[T] {
	return IntOf[T](anyToExpr(val))
}

func IntFrom[T any](field interface{ FieldType() T }) IntExpr[T] {
	return IntOf[T](anyToExpr(field))
}

// IntOf creates a generic IntExpr[T] from a clause expression.
func IntOf[T any](expr clause.Expression) IntExpr[T] {
	return IntExpr[T]{
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

// AsFloat 转换为 FloatExpr（不生成 SQL，仅类型转换）
func (e IntExpr[T]) AsFloat() FloatExpr[float64] {
	return FloatOf[float64](e.Unwrap())
}

// AsDecimal 转换为 DecimalExpr（不生成 SQL，仅类型转换）
func (e IntExpr[T]) AsDecimal() DecimalExpr[float64] {
	return DecimalOf[float64](e.Unwrap())
}

// Cast 类型转换 (CAST)
func (e IntExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastFloat 转换为浮点数 (CAST AS DECIMAL)
func (e IntExpr[T]) CastFloat(precision, scale int) FloatExpr[float64] {
	return FloatOf[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e IntExpr[T]) CastChar(length ...int) StringExpr[string] {
	return StringOf[string](e.castCharExpr(length...))
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e IntExpr[T]) CastSigned() IntExpr[int64] {
	return IntOf[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e IntExpr[T]) CastUnsigned() IntExpr[uint64] {
	return IntOf[uint64](e.castUnsignedExpr())
}

// ==================== 字符串转换 ====================

// Hex 转换为十六进制字符串 (HEX)
func (e IntExpr[T]) Hex() StringExpr[string] {
	return StringOf[string](e.hexExpr())
}

// Bin 转换为二进制字符串 (BIN)
func (e IntExpr[T]) Bin() StringExpr[string] {
	return StringOf[string](e.binExpr())
}

// Oct 转换为八进制字符串 (OCT)
func (e IntExpr[T]) Oct() StringExpr[string] {
	return StringOf[string](e.octExpr())
}

// ==================== 网络函数 ====================

// InetNtoa 将整数形式的IP地址转换为点分十进制字符串 (INET_NTOA)
// SELECT INET_NTOA(3232235777); -- 结果为 '192.168.1.1'
// SELECT INET_NTOA(ip_address) FROM access_logs;
func (e IntExpr[T]) InetNtoa() StringExpr[string] {
	return StringOf[string](clause.Expr{
		SQL:  "INET_NTOA(?)",
		Vars: []any{e.Unwrap()},
	})
}

// ==================== 日期/时间转换 ====================

// SecToTime 将秒数转换为时间 (SEC_TO_TIME)
// 数据库支持: MySQL
// SELECT SEC_TO_TIME(5400); -- 返回 '01:30:00'
func (e IntExpr[T]) SecToTime() TimeExpr[string] {
	return TimeOf[string](clause.Expr{
		SQL:  "SEC_TO_TIME(?)",
		Vars: []any{e.Unwrap()},
	})
}

// FromDays 将天数转换为日期 (FROM_DAYS)
// 数据库支持: MySQL
// SELECT FROM_DAYS(739259); -- 返回 '2024-01-15'
func (e IntExpr[T]) FromDays() DateExpr[string] {
	return DateOf[string](clause.Expr{
		SQL:  "FROM_DAYS(?)",
		Vars: []any{e.Unwrap()},
	})
}

// ToDateTime 将 Unix 时间戳转换为 DATETIME 类型 (FROM_UNIXTIME)
// 数据库支持: MySQL
// SELECT FROM_UNIXTIME(1698306600); -- 返回 '2023-10-26 10:30:00'
// SELECT FROM_UNIXTIME(created_at) FROM users;
// 可链式调用日期时间方法，如 .YearExpr(), .Format() 等
func (e IntExpr[T]) ToDateTime() DateTimeExpr[string] {
	return DateTimeOf[string](clause.Expr{
		SQL:  "FROM_UNIXTIME(?)",
		Vars: []any{e.Unwrap()},
	})
}

func (e IntExpr[T]) Unwrap() clause.Expression {
	return e.numericComparableImpl.Expression
}
