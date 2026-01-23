package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*Float[float64])(nil)

// ==================== Float 定义 ====================

// Float 浮点类型表达式，用于 AVG, SUM 等返回浮点数的聚合函数
// @gentype default=[float64]
// 支持比较操作、算术运算和数学函数
// 使用场景：
//   - AVG, SUM 等聚合函数的返回值
//   - 派生表中的浮点列
//   - 浮点字段的算术运算结果
type Float[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	nullCondFuncSql
	numericCondFuncSql
	castSql
	formatSql
	trigFuncSql
	aggregateSql
	baseExprSql
}

func NewFloat[T any](expr clause.Expression) Float[T] {
	return Float[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		numericCondFuncSql:    numericCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
		trigFuncSql:           trigFuncSql{Expression: expr},
		aggregateSql:          aggregateSql{Expression: expr},
		baseExprSql:           baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e Float[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e Float[T]) CastSigned() Int[int64] {
	return NewInt[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e Float[T]) CastUnsigned() Int[uint64] {
	return NewInt[uint64](e.castUnsignedExpr())
}

// CastDecimal 转换为指定精度的小数 (CAST AS DECIMAL)
func (e Float[T]) CastDecimal(precision, scale int) Decimal[float64] {
	return NewDecimal[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e Float[T]) CastChar(length ...int) String[string] {
	return NewString[string](e.castCharExpr(length...))
}

// ==================== 格式化 ====================

// Format 格式化数字 (FORMAT)
func (e Float[T]) Format(decimals int) String[string] {
	return NewString[string](e.formatExpr(decimals))
}
