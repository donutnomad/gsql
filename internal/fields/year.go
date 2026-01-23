package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*Year[int64])(nil)

// Year 年份类型表达式，用于 YEAR 类型字段
// @gentype default=[int]
// YEAR 类型存储年份值，范围通常是 1901-2155
// 使用场景：
//   - YEAR 类型字段
//   - YEAR() 函数提取年份的结果
type Year[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	nullCondFuncSql
	castSql
	aggregateSql
	baseExprSql
}

func NewYear[T any](expr clause.Expression) Year[T] {
	return Year[T]{
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
func (e Year[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e Year[T]) CastSigned() Int[int64] {
	return NewInt[int64](e.castSignedExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e Year[T]) CastChar(length ...int) String[string] {
	return NewString[string](e.castCharExpr(length...))
}
