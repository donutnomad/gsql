package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*YearExpr[int64])(nil)

// @gentype default=[int]
// YearExpr 年份类型表达式，用于 YEAR 类型字段
// YEAR 类型存储年份值，范围通常是 1901-2155
// 使用场景：
//   - YEAR 类型字段
//   - YEAR() 函数提取年份的结果
type YearExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	nullCondFuncSql
	castSql
	aggregateSql
	baseExprSql
}

// NewYearExpr 创建一个新的 YearExpr 实例
func NewYearExpr[T any](expr clause.Expression) YearExpr[T] {
	return YearExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		aggregateSql:          aggregateSql{Expression: expr},
		baseExprSql:           baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e YearExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e YearExpr[T]) CastSigned() IntExpr[int64] {
	return NewIntExpr[int64](e.castSignedExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e YearExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}
