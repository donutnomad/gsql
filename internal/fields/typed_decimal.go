package fields

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
)

var _ clause.Expression = (*DecimalExpr[float64])(nil)

// ==================== DecimalExpr 定义 ====================

// DecimalExpr 定点数类型表达式，用于精确的十进制数计算
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
	condFuncSql
	castSql
	formatSql
}

func NewDecimalExpr[T any](expr clause.Expression) DecimalExpr[T] {
	return DecimalExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		condFuncSql:           condFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
	}
}

func (e DecimalExpr[T]) Build(builder clause.Builder) {
	e.numericComparableImpl.Expression.Build(builder)
}

func (e DecimalExpr[T]) ToExpr() field.Expression {
	return e.numericComparableImpl.Expression
}

// As 创建一个别名字段
func (e DecimalExpr[T]) As(alias string) field.IField {
	return field.NewBaseFromSql(e.numericComparableImpl.Expression, alias)
}

// ==================== 算术运算 ====================

// Add 加法 (+)
// SELECT price + tax FROM products;
// SELECT balance + deposit FROM accounts;
func (e DecimalExpr[T]) Add(value any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.addExpr(value))
}

// Sub 减法 (-)
// SELECT price - discount FROM products;
// SELECT balance - withdrawal FROM accounts;
func (e DecimalExpr[T]) Sub(value any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.subExpr(value))
}

// Mul 乘法 (*)
// SELECT price * quantity FROM order_items;
// SELECT rate * hours as total FROM invoices;
func (e DecimalExpr[T]) Mul(value any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.mulExpr(value))
}

// Div 除法 (/)
// SELECT total / count FROM stats;
// SELECT amount / exchange_rate FROM transactions;
func (e DecimalExpr[T]) Div(value any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.divExpr(value))
}

// Neg 取负 (-)
// SELECT -price FROM products;
// SELECT -balance FROM accounts;
func (e DecimalExpr[T]) Neg() DecimalExpr[T] {
	return NewDecimalExpr[T](e.negExpr())
}

// Mod 取模 (MOD)
// SELECT MOD(amount, 100) FROM transactions;
// SELECT * FROM orders WHERE MOD(total, 10) = 0;
func (e DecimalExpr[T]) Mod(value any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.modExpr(value))
}

// ==================== 数学函数 ====================

// Abs 返回数值的绝对值 (ABS)
// SELECT ABS(-10.50); -- 结果为 10.50
// SELECT ABS(balance) FROM accounts;
// SELECT * FROM transactions WHERE ABS(amount) > 1000.00;
func (e DecimalExpr[T]) Abs() DecimalExpr[T] {
	return NewDecimalExpr[T](e.absExpr())
}

// Sign 返回数值的符号 (SIGN)：负数返回-1，零返回0，正数返回1
// SELECT SIGN(-10.50); -- 结果为 -1
// SELECT SIGN(0.00); -- 结果为 0
// SELECT SIGN(balance) FROM accounts;
func (e DecimalExpr[T]) Sign() IntExpr[int8] {
	return NewIntExpr[int8](e.signExpr())
}

// Ceil 向上取整 (CEIL)，返回大于或等于X的最小整数
// SELECT CEIL(4.30); -- 结果为 5
// SELECT CEIL(-4.30); -- 结果为 -4
// SELECT CEIL(price * 1.1) FROM products;
func (e DecimalExpr[T]) Ceil() IntExpr[int64] {
	return NewIntExpr[int64](e.ceilExpr())
}

// Floor 向下取整 (FLOOR)，返回小于或等于X的最大整数
// SELECT FLOOR(4.90); -- 结果为 4
// SELECT FLOOR(-4.30); -- 结果为 -5
// SELECT FLOOR(price * 0.9) FROM products;
func (e DecimalExpr[T]) Floor() IntExpr[int64] {
	return NewIntExpr[int64](e.floorExpr())
}

// Round 四舍五入 (ROUND) 到指定小数位数，默认四舍五入到整数
// SELECT ROUND(4.567); -- 结果为 5
// SELECT ROUND(4.567, 2); -- 结果为 4.57
// SELECT ROUND(price, 2) FROM products;
func (e DecimalExpr[T]) Round(decimals ...int) DecimalExpr[T] {
	return NewDecimalExpr[T](e.roundExpr(decimals...))
}

// Truncate 截断数值到指定小数位数 (TRUNCATE)，不进行四舍五入
// SELECT TRUNCATE(4.567, 2); -- 结果为 4.56
// SELECT TRUNCATE(4.567, 0); -- 结果为 4
// SELECT TRUNCATE(price, 2) FROM products;
func (e DecimalExpr[T]) Truncate(decimals int) DecimalExpr[T] {
	return NewDecimalExpr[T](e.truncateExpr(decimals))
}

// Pow 返回X的Y次幂 (POW)
// SELECT POW(2.5, 3); -- 结果为 15.625
// SELECT POW(10, 2); -- 结果为 100
// SELECT POW(rate, years) FROM investments;
func (e DecimalExpr[T]) Pow(exponent float64) DecimalExpr[T] {
	return NewDecimalExpr[T](e.powExpr(exponent))
}

// Sqrt 返回X的平方根 (SQRT)，X必须为非负数
// SELECT SQRT(4.00); -- 结果为 2.00
// SELECT SQRT(2.00); -- 结果为 1.4142...
// SELECT SQRT(area) as side_length FROM plots;
func (e DecimalExpr[T]) Sqrt() DecimalExpr[T] {
	return NewDecimalExpr[T](e.sqrtExpr())
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e DecimalExpr[T]) Cast(targetType string) field.Expression {
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

// ==================== 条件函数 ====================

// IfNull 如果为NULL则返回默认值 (IFNULL)
func (e DecimalExpr[T]) IfNull(defaultValue T) DecimalExpr[T] {
	return NewDecimalExpr[T](e.ifNullExpr(defaultValue))
}

// Coalesce 返回第一个非NULL值 (COALESCE)
func (e DecimalExpr[T]) Coalesce(values ...any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.coalesceExpr(values...))
}

// Nullif 如果两个值相等则返回NULL (NULLIF)
func (e DecimalExpr[T]) Nullif(value T) DecimalExpr[T] {
	return NewDecimalExpr[T](e.nullifExpr(value))
}

// Greatest 返回最大值 (GREATEST)
func (e DecimalExpr[T]) Greatest(values ...any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.greatestExpr(values...))
}

// Least 返回最小值 (LEAST)
func (e DecimalExpr[T]) Least(values ...any) DecimalExpr[T] {
	return NewDecimalExpr[T](e.leastExpr(values...))
}

// ==================== 格式化 ====================

// Format 格式化数字 (FORMAT)
func (e DecimalExpr[T]) Format(decimals int) TextExpr[string] {
	return NewTextExpr[string](e.formatExpr(decimals))
}

// ==================== 聚合函数 ====================

// Sum 计算数值的总和 (SUM)
// SELECT SUM(amount) FROM transactions;
// SELECT category, SUM(price) FROM products GROUP BY category;
func (e DecimalExpr[T]) Sum() DecimalExpr[T] {
	return NewDecimalExpr[T](clause.Expr{
		SQL:  "SUM(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Avg 计算数值的平均值 (AVG)
// SELECT AVG(price) FROM products;
// SELECT category, AVG(amount) FROM transactions GROUP BY category;
func (e DecimalExpr[T]) Avg() DecimalExpr[T] {
	return NewDecimalExpr[T](clause.Expr{
		SQL:  "AVG(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Max 返回最大值 (MAX)
// SELECT MAX(price) FROM products;
// SELECT category, MAX(amount) FROM transactions GROUP BY category;
func (e DecimalExpr[T]) Max() DecimalExpr[T] {
	return NewDecimalExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Min 返回最小值 (MIN)
// SELECT MIN(price) FROM products;
// SELECT category, MIN(amount) FROM transactions GROUP BY category;
func (e DecimalExpr[T]) Min() DecimalExpr[T] {
	return NewDecimalExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}
