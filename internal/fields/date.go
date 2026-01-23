package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*Date[string])(nil)

// Date 日期类型表达式，用于 DATE 类型字段 (YYYY-MM-DD)
// @gentype default=[string]
// 支持日期比较、日期运算和日期函数
// 使用场景：
//   - DATE 类型字段
//   - CURDATE(), CURRENT_DATE() 等函数的返回值
//   - DATE() 函数提取日期部分的结果
type Date[T any] struct {
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

func NewDate[T any](expr clause.Expression) Date[T] {
	return Date[T]{
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
func (e Date[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
func (e Date[T]) CastDatetime() DateTime[string] {
	return NewDateTime[string](e.castDatetimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e Date[T]) CastChar(length ...int) String[string] {
	return NewString[string](e.castCharExpr(length...))
}
