package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*TimestampExpr[int64])(nil)

// TimestampExpr 时间戳类型表达式，用于 TIMESTAMP 类型字段
// TIMESTAMP 类型在存储时会转换为 UTC，读取时转换为当前时区
// 使用场景：
//   - TIMESTAMP 类型字段
//   - 需要时区感知的时间记录
type TimestampExpr[T any] struct {
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

// NewTimestampExpr 创建一个新的 TimestampExpr 实例
func NewTimestampExpr[T any](expr clause.Expression) TimestampExpr[T] {
	return TimestampExpr[T]{
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

// ToDateTime 将Unix时间戳转换为DATETIME类型 (FROM_UNIXTIME)
// SELECT FROM_UNIXTIME(1698306600);
// SELECT FROM_UNIXTIME(users.created_at);
// SELECT FROM_UNIXTIME(users.created_at).Format('%Y年%m月%d日');
func (e TimestampExpr[T]) ToDateTime() DateTimeExpr[string] {
	return NewDateTimeExpr[string](clause.Expr{
		SQL:  "FROM_UNIXTIME(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e TimestampExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastDate 转换为 DATE 类型 (CAST AS DATE)
func (e TimestampExpr[T]) CastDate() DateExpr[string] {
	return NewDateExpr[string](e.castDateExpr())
}

// CastDatetime 转换为 DATETIME 类型 (CAST AS DATETIME)
func (e TimestampExpr[T]) CastDatetime() DateTimeExpr[string] {
	return NewDateTimeExpr[string](e.castDatetimeExpr())
}

// CastTime 转换为 TIME 类型 (CAST AS TIME)
func (e TimestampExpr[T]) CastTime() TimeExpr[string] {
	return NewTimeExpr[string](e.castTimeExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e TimestampExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}
