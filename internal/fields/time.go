package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*TimeExpr[string])(nil)

// TimeExpr 时间类型表达式，用于 TIME 类型字段 (HH:MM:SS)
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

// NewTimeExpr 创建一个新的 TimeExpr 实例
func NewTimeExpr[T any](expr clause.Expression) TimeExpr[T] {
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
	return NewDateTimeExpr[string](e.castDatetimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e TimeExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}
