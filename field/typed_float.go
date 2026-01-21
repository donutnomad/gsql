package field

import (
	"github.com/donutnomad/gsql/clause"
)

// ==================== FloatExprT 定义 ====================

// FloatExprT 浮点类型表达式，用于 AVG, SUM 等返回浮点数的聚合函数
// 支持比较操作、算术运算和数学函数
// 使用场景：
//   - AVG, SUM 等聚合函数的返回值
//   - 派生表中的浮点列
//   - 浮点字段的算术运算结果
type FloatExprT[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	condFuncSql
	castSql
	formatSql
	trigFuncSql
}

// NewFloatExprT 创建一个新的 FloatExprT 实例
func NewFloatExprT[T any](expr clause.Expression) FloatExprT[T] {
	return FloatExprT[T]{
		numericComparableImpl: numericComparableImpl[T]{Expression: expr},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		condFuncSql:           condFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
		trigFuncSql:           trigFuncSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e FloatExprT[T]) Build(builder clause.Builder) {
	e.numericComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e FloatExprT[T]) ToExpr() Expression {
	return e.numericComparableImpl.Expression
}

// As 创建一个别名字段
func (e FloatExprT[T]) As(alias string) IField {
	return NewBaseFromSql(e.numericComparableImpl.Expression, alias)
}

// ==================== 算术运算 ====================

// Add 加法 (+)
// SELECT price + 100 FROM products;
// SELECT users.balance + deposit FROM users;
func (e FloatExprT[T]) Add(value any) FloatExprT[T] {
	return NewFloatExprT[T](e.addExpr(value))
}

// Sub 减法 (-)
// SELECT price - discount FROM products;
// SELECT balance - withdrawal FROM accounts;
func (e FloatExprT[T]) Sub(value any) FloatExprT[T] {
	return NewFloatExprT[T](e.subExpr(value))
}

// Mul 乘法 (*)
// SELECT price * quantity FROM order_items;
// SELECT rate * hours as total FROM timesheets;
func (e FloatExprT[T]) Mul(value any) FloatExprT[T] {
	return NewFloatExprT[T](e.mulExpr(value))
}

// Div 除法 (/)
// SELECT total / count FROM stats;
// SELECT amount / rate as quantity FROM orders;
func (e FloatExprT[T]) Div(value any) FloatExprT[T] {
	return NewFloatExprT[T](e.divExpr(value))
}

// Neg 取负 (-)
// SELECT -price FROM products;
// SELECT -balance FROM accounts;
func (e FloatExprT[T]) Neg() FloatExprT[T] {
	return NewFloatExprT[T](e.negExpr())
}

// ==================== 数学函数 ====================

// Abs 返回数值的绝对值 (ABS)
// SELECT ABS(-10.5); -- 结果为 10.5
// SELECT ABS(users.balance) FROM users;
// SELECT * FROM transactions WHERE ABS(amount) > 1000;
func (e FloatExprT[T]) Abs() FloatExprT[T] {
	return NewFloatExprT[T](e.absExpr())
}

// Sign 返回数值的符号 (SIGN)：负数返回-1，零返回0，正数返回1
// SELECT SIGN(-10.5); -- 结果为 -1
// SELECT SIGN(0); -- 结果为 0
// SELECT SIGN(balance) FROM accounts;
func (e FloatExprT[T]) Sign() IntExprT[int8] {
	return NewIntExprT[int8](e.signExpr())
}

// Ceil 向上取整 (CEIL)，返回大于或等于X的最小整数
// SELECT CEIL(4.3); -- 结果为 5
// SELECT CEIL(-4.3); -- 结果为 -4
// SELECT CEIL(price * 1.1) FROM products;
func (e FloatExprT[T]) Ceil() IntExprT[int64] {
	return NewIntExprT[int64](e.ceilExpr())
}

// Floor 向下取整 (FLOOR)，返回小于或等于X的最大整数
// SELECT FLOOR(4.9); -- 结果为 4
// SELECT FLOOR(-4.3); -- 结果为 -5
// SELECT FLOOR(price * 0.9) FROM products;
func (e FloatExprT[T]) Floor() IntExprT[int64] {
	return NewIntExprT[int64](e.floorExpr())
}

// Round 四舍五入 (ROUND) 到指定小数位数，默认四舍五入到整数
// SELECT ROUND(4.567); -- 结果为 5
// SELECT ROUND(4.567, 2); -- 结果为 4.57
// SELECT ROUND(price, 2) FROM products;
// SELECT ROUND(123.456, -1); -- 结果为 120
func (e FloatExprT[T]) Round(decimals ...int) FloatExprT[T] {
	return NewFloatExprT[T](e.roundExpr(decimals...))
}

// Truncate 截断数值到指定小数位数 (TRUNCATE)，不进行四舍五入
// SELECT TRUNCATE(4.567, 2); -- 结果为 4.56
// SELECT TRUNCATE(4.567, 0); -- 结果为 4
// SELECT TRUNCATE(123.456, -1); -- 结果为 120
// SELECT TRUNCATE(price, 2) FROM products;
func (e FloatExprT[T]) Truncate(decimals int) FloatExprT[T] {
	return NewFloatExprT[T](e.truncateExpr(decimals))
}

// Pow 返回X的Y次幂 (POW)
// SELECT POW(2, 3); -- 结果为 8
// SELECT POW(10, 2); -- 结果为 100
// SELECT POW(5, -1); -- 结果为 0.2
// SELECT SQRT(POW(x2 - x1, 2) + POW(y2 - y1, 2)) as distance FROM points;
func (e FloatExprT[T]) Pow(exponent float64) FloatExprT[T] {
	return NewFloatExprT[T](e.powExpr(exponent))
}

// Sqrt 返回X的平方根 (SQRT)，X必须为非负数
// SELECT SQRT(4); -- 结果为 2
// SELECT SQRT(2); -- 结果为 1.4142...
// SELECT SQRT(area) as side_length FROM squares;
func (e FloatExprT[T]) Sqrt() FloatExprT[T] {
	return NewFloatExprT[T](e.sqrtExpr())
}

// Log 自然对数 (LOG)
func (e FloatExprT[T]) Log() FloatExprT[T] {
	return NewFloatExprT[T](e.logExpr())
}

// Log10 以10为底的对数 (LOG10)
func (e FloatExprT[T]) Log10() FloatExprT[T] {
	return NewFloatExprT[T](e.log10Expr())
}

// Log2 以2为底的对数 (LOG2)
func (e FloatExprT[T]) Log2() FloatExprT[T] {
	return NewFloatExprT[T](e.log2Expr())
}

// Exp 自然指数 (EXP)
func (e FloatExprT[T]) Exp() FloatExprT[T] {
	return NewFloatExprT[T](e.expExpr())
}

// ==================== 三角函数 ====================

// Sin 正弦 (SIN)
func (e FloatExprT[T]) Sin() FloatExprT[T] {
	return NewFloatExprT[T](e.sinExpr())
}

// Cos 余弦 (COS)
func (e FloatExprT[T]) Cos() FloatExprT[T] {
	return NewFloatExprT[T](e.cosExpr())
}

// Tan 正切 (TAN)
func (e FloatExprT[T]) Tan() FloatExprT[T] {
	return NewFloatExprT[T](e.tanExpr())
}

// Asin 反正弦 (ASIN)
func (e FloatExprT[T]) Asin() FloatExprT[T] {
	return NewFloatExprT[T](e.asinExpr())
}

// Acos 反余弦 (ACOS)
func (e FloatExprT[T]) Acos() FloatExprT[T] {
	return NewFloatExprT[T](e.acosExpr())
}

// Atan 反正切 (ATAN)
func (e FloatExprT[T]) Atan() FloatExprT[T] {
	return NewFloatExprT[T](e.atanExpr())
}

// Radians 角度转弧度 (RADIANS)
func (e FloatExprT[T]) Radians() FloatExprT[T] {
	return NewFloatExprT[T](e.radiansExpr())
}

// Degrees 弧度转角度 (DEGREES)
func (e FloatExprT[T]) Degrees() FloatExprT[T] {
	return NewFloatExprT[T](e.degreesExpr())
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e FloatExprT[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e FloatExprT[T]) CastSigned() IntExprT[int64] {
	return NewIntExprT[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e FloatExprT[T]) CastUnsigned() IntExprT[uint64] {
	return NewIntExprT[uint64](e.castUnsignedExpr())
}

// CastDecimal 转换为指定精度的小数 (CAST AS DECIMAL)
func (e FloatExprT[T]) CastDecimal(precision, scale int) DecimalExprT[float64] {
	return NewDecimalExprT[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e FloatExprT[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 条件函数 ====================

// IfNull 如果为NULL则返回默认值 (IFNULL)
func (e FloatExprT[T]) IfNull(defaultValue T) FloatExprT[T] {
	return NewFloatExprT[T](e.ifNullExpr(defaultValue))
}

// Coalesce 返回第一个非NULL值 (COALESCE)
func (e FloatExprT[T]) Coalesce(values ...any) FloatExprT[T] {
	return NewFloatExprT[T](e.coalesceExpr(values...))
}

// Nullif 如果两个值相等则返回NULL (NULLIF)
func (e FloatExprT[T]) Nullif(value T) FloatExprT[T] {
	return NewFloatExprT[T](e.nullifExpr(value))
}

// Greatest 返回最大值 (GREATEST)
func (e FloatExprT[T]) Greatest(values ...any) FloatExprT[T] {
	return NewFloatExprT[T](e.greatestExpr(values...))
}

// Least 返回最小值 (LEAST)
func (e FloatExprT[T]) Least(values ...any) FloatExprT[T] {
	return NewFloatExprT[T](e.leastExpr(values...))
}

// ==================== 格式化 ====================

// Format 格式化数字 (FORMAT)
func (e FloatExprT[T]) Format(decimals int) TextExpr[string] {
	return NewTextExpr[string](e.formatExpr(decimals))
}
