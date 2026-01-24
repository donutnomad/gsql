package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*DateTimeExpr[string])(nil)

// DateTimeExpr 日期时间类型表达式，用于 DATETIME 类型字段 (YYYY-MM-DD HH:MM:SS)
// @gentype default=[string]
// 支持日期时间比较、运算和提取函数
// 使用场景：
//   - DATETIME 类型字段
//   - NOW(), CURRENT_TIMESTAMP() 等函数的返回值
//   - FROM_UNIXTIME() 函数的返回值
type DateTimeExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	nullCondFuncSql
	castSql
	aggregateSql
	dateExtractSql
	timeExtractSql
	dateIntervalSql
	dateDiffSql
	timeDiffSql
	timestampDiffSql
	dateFormatSql
	dateConversionSql
	unixTimestampSql
	baseExprSql
}

// DateTime creates a DateTimeExpr[string] from a clause expression.
func DateTime(expr clause.Expression) DateTimeExpr[string] {
	return DateTimeOf[string](expr)
}

// DateTimeE creates a DateTimeExpr[string] from raw SQL with optional variables.
func DateTimeE(sql string, vars ...any) DateTimeExpr[string] {
	return DateTime(clause.Expr{SQL: sql, Vars: vars})
}

// DateTimeOf creates a generic DateTimeExpr[T] from a clause expression.
func DateTimeOf[T any](expr clause.Expression) DateTimeExpr[T] {
	return DateTimeExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		aggregateSql:          aggregateSql{Expression: expr},
		dateExtractSql:        dateExtractSql{Expression: expr},
		timeExtractSql:        timeExtractSql{Expression: expr},
		dateIntervalSql:       dateIntervalSql{Expression: expr},
		dateDiffSql:           dateDiffSql{Expression: expr},
		timeDiffSql:           timeDiffSql{Expression: expr},
		timestampDiffSql:      timestampDiffSql{Expression: expr},
		dateFormatSql:         dateFormatSql{Expression: expr},
		dateConversionSql:     dateConversionSql{Expression: expr},
		unixTimestampSql:      unixTimestampSql{Expression: expr},
		baseExprSql:           baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e DateTimeExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastDate 转换为 DATE 类型 (CAST AS DATE)
func (e DateTimeExpr[T]) CastDate() DateExpr[string] {
	return DateOf[string](e.castDateExpr())
}

// CastTime 转换为 TIME 类型 (CAST AS TIME)
func (e DateTimeExpr[T]) CastTime() TimeExpr[string] {
	return TimeOf[string](e.castTimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e DateTimeExpr[T]) CastChar(length ...int) StringExpr[string] {
	return StringOf[string](e.castCharExpr(length...))
}
