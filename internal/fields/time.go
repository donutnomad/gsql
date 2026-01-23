package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*Time[string])(nil)

// Time 时间类型表达式，用于 TIME 类型字段 (HH:MM:SS)
// @gentype default=[string]
// 支持时间比较、运算和提取函数
// 使用场景：
//   - TIME 类型字段
//   - CURTIME(), CURRENT_TIME() 等函数的返回值
//   - TIME() 函数提取时间部分的结果
type Time[T any] struct {
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

func NewTime[T any](expr clause.Expression) Time[T] {
	return Time[T]{
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
func (e Time[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
// 注意：TIME 转 DATETIME 时，日期部分为当前日期
func (e Time[T]) CastDatetime() DateTime[string] {
	return NewDateTime[string](e.castDatetimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e Time[T]) CastChar(length ...int) String[string] {
	return NewString[string](e.castCharExpr(length...))
}
