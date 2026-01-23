package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*DecimalExpr[float64])(nil)

// ==================== DecimalExpr 定义 ====================

// DecimalExpr 定点数类型表达式，用于精确的十进制数计算
// @gentype default=[float64]
// 与 FloatExpr 不同，DECIMAL 是精确类型，适合金融计算
// 使用场景：
//   - 价格、金额等需要精确计算的字段
//   - SUM, AVG 等聚合函数处理 DECIMAL 字段的返回值
//   - 派生表中的 DECIMAL 列
type DecimalExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	nullCondFuncSql
	numericCondFuncSql
	castSql
	formatSql
	aggregateSql
	baseExprSql
}

func NewDecimalExpr[T any](expr clause.Expression) DecimalExpr[T] {
	return DecimalExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		numericCondFuncSql:    numericCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
		aggregateSql:          aggregateSql{Expression: expr},
		baseExprSql:           baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e DecimalExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e DecimalExpr[T]) CastSigned() IntExpr[int64] {
	return NewIntExpr[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e DecimalExpr[T]) CastUnsigned() IntExpr[uint64] {
	return NewIntExpr[uint64](e.castUnsignedExpr())
}

// CastDecimal 转换为指定精度的 DECIMAL (CAST AS DECIMAL)
func (e DecimalExpr[T]) CastDecimal(precision, scale int) DecimalExpr[T] {
	return NewDecimalExpr[T](e.castDecimalExpr(precision, scale))
}

// CastFloat 转换为浮点数 (CAST AS DOUBLE)
func (e DecimalExpr[T]) CastFloat() FloatExpr[float64] {
	return NewFloatExpr[float64](e.castDoubleExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e DecimalExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 格式化 ====================

// Format 格式化数字 (FORMAT)
func (e DecimalExpr[T]) Format(decimals int) TextExpr[string] {
	return NewTextExpr[string](e.formatExpr(decimals))
}
