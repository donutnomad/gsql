package fields

import (
	"database/sql"
	"time"

	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*TimeExpr[string])(nil)

type timeExpr[T any] = TimeExpr[T]

// TimeExpr 时间类型表达式，用于 TIME 类型字段 (HH:MM:SS)
// @gentype default=[string]
// 支持时间比较、运算和提取函数
// 使用场景：
//   - TIME 类型字段
//   - CURTIME(), CURRENT_TIME() 等函数的返回值
//   - TIME() 函数提取时间部分的结果
type TimeExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	nullCondFuncSql
	castSql
	aggregateSql
	timeExtractSql
	dateIntervalSql
	timeDiffSql
	timeFormatSql
	baseExprSql
}

// Time creates a TimeExpr[string] from a clause expression.
func Time(expr clause.Expression) TimeExpr[time.Time] {
	return TimeOf[time.Time](expr)
}

// TimeE creates a TimeExpr[string] from raw SQL with optional variables.
func TimeE(sql string, vars ...any) TimeExpr[string] {
	return TimeOf[string](clause.Expr{SQL: sql, Vars: vars})
}

// TimeVal creates a TimeExpr from a time literal value.
func TimeVal[T ~string | time.Time | *time.Time | sql.NullTime | any](val T) TimeExpr[T] {
	return TimeOf[T](anyToExpr(val))
}

func TimeFrom[T any](field interface{ FieldType() T }) TimeExpr[T] {
	return TimeOf[T](anyToExpr(field))
}

// TimeOf creates a generic TimeExpr[T] from a clause expression.
func TimeOf[T any](expr clause.Expression) TimeExpr[T] {
	return TimeExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		aggregateSql:          aggregateSql{Expression: expr},
		timeExtractSql:        timeExtractSql{Expression: expr},
		dateIntervalSql:       dateIntervalSql{Expression: expr},
		timeDiffSql:           timeDiffSql{Expression: expr},
		timeFormatSql:         timeFormatSql{Expression: expr},
		baseExprSql:           baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e TimeExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
// 注意：TIME 转 DATETIME 时，日期部分为当前日期
func (e TimeExpr[T]) CastDatetime() DateTimeExpr[string] {
	return DateTimeOf[string](e.castDatetimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e TimeExpr[T]) CastChar(length ...int) StringExpr[string] {
	return StringOf[string](e.castCharExpr(length...))
}

func (e TimeExpr[T]) Unwrap() clause.Expression {
	return e.numericComparableImpl.Expression
}
