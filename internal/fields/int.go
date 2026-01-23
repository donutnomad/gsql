package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*IntExpr[int64])(nil)

// ==================== IntExpr 定义 ====================

// @gentype default=[int]
// IntExpr 整数类型表达式，用于 COUNT 等返回整数的聚合函数
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

func NewIntExpr[T any](expr clause.Expression) IntExpr[T] {
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

// Cast 类型转换 (CAST)
func (e IntExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastFloat 转换为浮点数 (CAST AS DECIMAL)
func (e IntExpr[T]) CastFloat(precision, scale int) FloatExpr[float64] {
	return NewFloatExpr[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e IntExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e IntExpr[T]) CastSigned() IntExpr[int64] {
	return NewIntExpr[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e IntExpr[T]) CastUnsigned() IntExpr[uint64] {
	return NewIntExpr[uint64](e.castUnsignedExpr())
}

// ==================== 字符串转换 ====================

// Hex 转换为十六进制字符串 (HEX)
func (e IntExpr[T]) Hex() TextExpr[string] {
	return NewTextExpr[string](e.hexExpr())
}

// Bin 转换为二进制字符串 (BIN)
func (e IntExpr[T]) Bin() TextExpr[string] {
	return NewTextExpr[string](e.binExpr())
}

// Oct 转换为八进制字符串 (OCT)
func (e IntExpr[T]) Oct() TextExpr[string] {
	return NewTextExpr[string](e.octExpr())
}

// ==================== 网络函数 ====================

// InetNtoa 将整数形式的IP地址转换为点分十进制字符串 (INET_NTOA)
// SELECT INET_NTOA(3232235777); -- 结果为 '192.168.1.1'
// SELECT INET_NTOA(ip_address) FROM access_logs;
func (e IntExpr[T]) InetNtoa() TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "INET_NTOA(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}
