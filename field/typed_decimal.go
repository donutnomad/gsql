package field

import (
	"github.com/donutnomad/gsql/clause"
)

// ==================== DecimalExprT 定义 ====================

// DecimalExprT 定点数类型表达式，用于精确的十进制数计算
// 与 FloatExprT 不同，DECIMAL 是精确类型，适合金融计算
// 使用场景：
//   - 价格、金额等需要精确计算的字段
//   - SUM, AVG 等聚合函数处理 DECIMAL 字段的返回值
//   - 派生表中的 DECIMAL 列
type DecimalExprT[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	condFuncSql
	castSql
	formatSql
}

// NewDecimalExprT 创建一个新的 DecimalExprT 实例
func NewDecimalExprT[T any](expr clause.Expression) DecimalExprT[T] {
	return DecimalExprT[T]{
		numericComparableImpl: numericComparableImpl[T]{Expression: expr},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		condFuncSql:           condFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e DecimalExprT[T]) Build(builder clause.Builder) {
	e.numericComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e DecimalExprT[T]) ToExpr() Expression {
	return e.numericComparableImpl.Expression
}

// As 创建一个别名字段
func (e DecimalExprT[T]) As(alias string) IField {
	return NewBaseFromSql(e.numericComparableImpl.Expression, alias)
}

// ==================== 算术运算 ====================

// Add 加法 (+)
// SELECT price + tax FROM products;
// SELECT balance + deposit FROM accounts;
func (e DecimalExprT[T]) Add(value any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.addExpr(value))
}

// Sub 减法 (-)
// SELECT price - discount FROM products;
// SELECT balance - withdrawal FROM accounts;
func (e DecimalExprT[T]) Sub(value any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.subExpr(value))
}

// Mul 乘法 (*)
// SELECT price * quantity FROM order_items;
// SELECT rate * hours as total FROM invoices;
func (e DecimalExprT[T]) Mul(value any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.mulExpr(value))
}

// Div 除法 (/)
// SELECT total / count FROM stats;
// SELECT amount / exchange_rate FROM transactions;
func (e DecimalExprT[T]) Div(value any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.divExpr(value))
}

// Neg 取负 (-)
// SELECT -price FROM products;
// SELECT -balance FROM accounts;
func (e DecimalExprT[T]) Neg() DecimalExprT[T] {
	return NewDecimalExprT[T](e.negExpr())
}

// Mod 取模 (MOD)
// SELECT MOD(amount, 100) FROM transactions;
// SELECT * FROM orders WHERE MOD(total, 10) = 0;
func (e DecimalExprT[T]) Mod(value any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.modExpr(value))
}

// ==================== 数学函数 ====================

// Abs 返回数值的绝对值 (ABS)
// SELECT ABS(-10.50); -- 结果为 10.50
// SELECT ABS(balance) FROM accounts;
// SELECT * FROM transactions WHERE ABS(amount) > 1000.00;
func (e DecimalExprT[T]) Abs() DecimalExprT[T] {
	return NewDecimalExprT[T](e.absExpr())
}

// Sign 返回数值的符号 (SIGN)：负数返回-1，零返回0，正数返回1
// SELECT SIGN(-10.50); -- 结果为 -1
// SELECT SIGN(0.00); -- 结果为 0
// SELECT SIGN(balance) FROM accounts;
func (e DecimalExprT[T]) Sign() IntExprT[int8] {
	return NewIntExprT[int8](e.signExpr())
}

// Ceil 向上取整 (CEIL)，返回大于或等于X的最小整数
// SELECT CEIL(4.30); -- 结果为 5
// SELECT CEIL(-4.30); -- 结果为 -4
// SELECT CEIL(price * 1.1) FROM products;
func (e DecimalExprT[T]) Ceil() IntExprT[int64] {
	return NewIntExprT[int64](e.ceilExpr())
}

// Floor 向下取整 (FLOOR)，返回小于或等于X的最大整数
// SELECT FLOOR(4.90); -- 结果为 4
// SELECT FLOOR(-4.30); -- 结果为 -5
// SELECT FLOOR(price * 0.9) FROM products;
func (e DecimalExprT[T]) Floor() IntExprT[int64] {
	return NewIntExprT[int64](e.floorExpr())
}

// Round 四舍五入 (ROUND) 到指定小数位数，默认四舍五入到整数
// SELECT ROUND(4.567); -- 结果为 5
// SELECT ROUND(4.567, 2); -- 结果为 4.57
// SELECT ROUND(price, 2) FROM products;
func (e DecimalExprT[T]) Round(decimals ...int) DecimalExprT[T] {
	return NewDecimalExprT[T](e.roundExpr(decimals...))
}

// Truncate 截断数值到指定小数位数 (TRUNCATE)，不进行四舍五入
// SELECT TRUNCATE(4.567, 2); -- 结果为 4.56
// SELECT TRUNCATE(4.567, 0); -- 结果为 4
// SELECT TRUNCATE(price, 2) FROM products;
func (e DecimalExprT[T]) Truncate(decimals int) DecimalExprT[T] {
	return NewDecimalExprT[T](e.truncateExpr(decimals))
}

// Pow 返回X的Y次幂 (POW)
// SELECT POW(2.5, 3); -- 结果为 15.625
// SELECT POW(10, 2); -- 结果为 100
// SELECT POW(rate, years) FROM investments;
func (e DecimalExprT[T]) Pow(exponent float64) DecimalExprT[T] {
	return NewDecimalExprT[T](e.powExpr(exponent))
}

// Sqrt 返回X的平方根 (SQRT)，X必须为非负数
// SELECT SQRT(4.00); -- 结果为 2.00
// SELECT SQRT(2.00); -- 结果为 1.4142...
// SELECT SQRT(area) as side_length FROM plots;
func (e DecimalExprT[T]) Sqrt() DecimalExprT[T] {
	return NewDecimalExprT[T](e.sqrtExpr())
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e DecimalExprT[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e DecimalExprT[T]) CastSigned() IntExprT[int64] {
	return NewIntExprT[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e DecimalExprT[T]) CastUnsigned() IntExprT[uint64] {
	return NewIntExprT[uint64](e.castUnsignedExpr())
}

// CastDecimal 转换为指定精度的 DECIMAL (CAST AS DECIMAL)
func (e DecimalExprT[T]) CastDecimal(precision, scale int) DecimalExprT[T] {
	return NewDecimalExprT[T](e.castDecimalExpr(precision, scale))
}

// CastFloat 转换为浮点数 (CAST AS DOUBLE)
func (e DecimalExprT[T]) CastFloat() FloatExprT[float64] {
	return NewFloatExprT[float64](e.castDoubleExpr())
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e DecimalExprT[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 条件函数 ====================

// IfNull 如果为NULL则返回默认值 (IFNULL)
func (e DecimalExprT[T]) IfNull(defaultValue T) DecimalExprT[T] {
	return NewDecimalExprT[T](e.ifNullExpr(defaultValue))
}

// Coalesce 返回第一个非NULL值 (COALESCE)
func (e DecimalExprT[T]) Coalesce(values ...any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.coalesceExpr(values...))
}

// Nullif 如果两个值相等则返回NULL (NULLIF)
func (e DecimalExprT[T]) Nullif(value T) DecimalExprT[T] {
	return NewDecimalExprT[T](e.nullifExpr(value))
}

// Greatest 返回最大值 (GREATEST)
func (e DecimalExprT[T]) Greatest(values ...any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.greatestExpr(values...))
}

// Least 返回最小值 (LEAST)
func (e DecimalExprT[T]) Least(values ...any) DecimalExprT[T] {
	return NewDecimalExprT[T](e.leastExpr(values...))
}

// ==================== 格式化 ====================

// Format 格式化数字 (FORMAT)
func (e DecimalExprT[T]) Format(decimals int) TextExpr[string] {
	return NewTextExpr[string](e.formatExpr(decimals))
}
