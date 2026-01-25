package fields

import (
	"database/sql"
	"time"

	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*DateExpr[string])(nil)

type dateExpr[T any] = DateExpr[T]

// DateExpr 日期类型表达式，用于 DATE 类型字段 (YYYY-MM-DD)
// @gentype default=[string]
// 支持日期比较、日期运算和日期函数
// 使用场景：
//   - DATE 类型字段
//   - CURDATE(), CURRENT_DATE() 等函数的返回值
//   - DATE() 函数提取日期部分的结果
type DateExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	nullCondFuncSql
	castSql
	aggregateSql
	dateExtractSql
	dateIntervalSql
	dateDiffSql
	dateFormatSql
	unixTimestampSql
	baseExprSql
}

// Date creates a DateExpr[string] from a clause expression.
func Date(expr clause.Expression) DateExpr[time.Time] {
	return DateOf[time.Time](expr)
}

// DateE creates a DateExpr[string] from raw SQL with optional variables.
func DateE(sql string, vars ...any) DateExpr[string] {
	return DateOf[string](clause.Expr{SQL: sql, Vars: vars})
}

// DateVal creates a DateExpr from a date literal value.
func DateVal[T ~string | time.Time | *time.Time | sql.NullTime | any](val T) DateExpr[T] {
	return DateOf[T](anyToExpr(val))
}

// DateOf creates a generic DateExpr[T] from a clause expression.
func DateOf[T any](expr clause.Expression) DateExpr[T] {
	return DateExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		aggregateSql:          aggregateSql{Expression: expr},
		dateExtractSql:        dateExtractSql{Expression: expr},
		dateIntervalSql:       dateIntervalSql{Expression: expr},
		dateDiffSql:           dateDiffSql{Expression: expr},
		dateFormatSql:         dateFormatSql{Expression: expr},
		unixTimestampSql:      unixTimestampSql{Expression: expr},
		baseExprSql:           baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e DateExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
func (e DateExpr[T]) CastDatetime() DateTimeExpr[string] {
	return DateTimeOf[string](e.castDatetimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e DateExpr[T]) CastChar(length ...int) StringExpr[string] {
	return StringOf[string](e.castCharExpr(length...))
}

func (e DateExpr[T]) Unwrap() clause.Expression {
	return e.numericComparableImpl.Expression
}
