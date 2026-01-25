package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*YearExpr[int64])(nil)

type yearExpr[T any] = YearExpr[T]

// YearExpr 年份类型表达式，用于 YEAR 类型字段
// @gentype default=[int]
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

// Year creates a YearExpr[int64] from a clause expression.
func Year(expr clause.Expression) YearExpr[int64] {
	return YearOf[int64](expr)
}

// YearE creates a YearExpr[int64] from raw SQL with optional variables.
func YearE(sql string, vars ...any) YearExpr[int64] {
	return Year(clause.Expr{SQL: sql, Vars: vars})
}

// YearVal creates a YearExpr from an integer literal value.
func YearVal[T ~int | ~int16 | ~int32 | ~int64 | any](val T) YearExpr[T] {
	return YearOf[T](anyToExpr(val))
}

func YearFromField[T any, Expr Expressions[T]](field Expr) YearExpr[T] {
	return YearOf[T](anyToExpr(field))
}

// YearOf creates a generic YearExpr[T] from a clause expression.
func YearOf[T any](expr clause.Expression) YearExpr[T] {
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
	return IntOf[int64](e.castSignedExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e YearExpr[T]) CastChar(length ...int) StringExpr[string] {
	return StringOf[string](e.castCharExpr(length...))
}

func (e YearExpr[T]) Unwrap() clause.Expression {
	return e.numericComparableImpl.Expression
}
