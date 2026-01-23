package fields

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
)

var _ clause.Expression = (*FloatExpr[float64])(nil)

// ==================== FloatExpr 定义 ====================

// FloatExpr 浮点类型表达式，用于 AVG, SUM 等返回浮点数的聚合函数
// 支持比较操作、算术运算和数学函数
// 使用场景：
//   - AVG, SUM 等聚合函数的返回值
//   - 派生表中的浮点列
//   - 浮点字段的算术运算结果
type FloatExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	nullCondFuncSql
	numericCondFuncSql
	castSql
	formatSql
	trigFuncSql
}

// NewFloatExpr 创建一个新的 FloatExpr 实例
func NewFloatExpr[T any](expr clause.Expression) FloatExpr[T] {
	return FloatExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		numericCondFuncSql:    numericCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
		trigFuncSql:           trigFuncSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e FloatExpr[T]) Build(builder clause.Builder) {
	e.numericComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e FloatExpr[T]) ToExpr() clause.Expression {
	return e.numericComparableImpl.Expression
}

// As 创建一个别名字段
func (e FloatExpr[T]) As(alias string) field.IField {
	return field.NewBaseFromSql(e.numericComparableImpl.Expression, alias)
}

// ==================== 数学函数 (特殊方法) ====================

// Sign 返回数值的符号 (SIGN)：负数返回-1，零返回0，正数返回1
// SELECT SIGN(-10.5); -- 结果为 -1
// SELECT SIGN(0); -- 结果为 0
// SELECT SIGN(balance) FROM accounts;
func (e FloatExpr[T]) Sign() IntExpr[int8] {
	return NewIntExpr[int8](e.signExpr())
}

// Ceil 向上取整 (CEIL)，返回大于或等于X的最小整数
// SELECT CEIL(4.3); -- 结果为 5
// SELECT CEIL(-4.3); -- 结果为 -4
// SELECT CEIL(price * 1.1) FROM products;
func (e FloatExpr[T]) Ceil() IntExpr[int64] {
	return NewIntExpr[int64](e.ceilExpr())
}

// Floor 向下取整 (FLOOR)，返回小于或等于X的最大整数
// SELECT FLOOR(4.9); -- 结果为 4
// SELECT FLOOR(-4.3); -- 结果为 -5
// SELECT FLOOR(price * 0.9) FROM products;
func (e FloatExpr[T]) Floor() IntExpr[int64] {
	return NewIntExpr[int64](e.floorExpr())
}

// Sqrt 返回X的平方根 (SQRT)，X必须为非负数
// SELECT SQRT(4); -- 结果为 2
// SELECT SQRT(2); -- 结果为 1.4142...
// SELECT SQRT(area) as side_length FROM squares;
func (e FloatExpr[T]) Sqrt() FloatExpr[T] {
	return NewFloatExpr[T](e.sqrtExpr())
}

// Log 自然对数 (LOG)
func (e FloatExpr[T]) Log() FloatExpr[T] {
	return NewFloatExpr[T](e.logExpr())
}

// Log10 以10为底的对数 (LOG10)
func (e FloatExpr[T]) Log10() FloatExpr[T] {
	return NewFloatExpr[T](e.log10Expr())
}

// Log2 以2为底的对数 (LOG2)
func (e FloatExpr[T]) Log2() FloatExpr[T] {
	return NewFloatExpr[T](e.log2Expr())
}

// Exp 自然指数 (EXP)
func (e FloatExpr[T]) Exp() FloatExpr[T] {
	return NewFloatExpr[T](e.expExpr())
}

// ==================== 三角函数 ====================

// Sin 正弦 (SIN)
func (e FloatExpr[T]) Sin() FloatExpr[T] {
	return NewFloatExpr[T](e.sinExpr())
}

// Cos 余弦 (COS)
func (e FloatExpr[T]) Cos() FloatExpr[T] {
	return NewFloatExpr[T](e.cosExpr())
}

// Tan 正切 (TAN)
func (e FloatExpr[T]) Tan() FloatExpr[T] {
	return NewFloatExpr[T](e.tanExpr())
}

// Asin 反正弦 (ASIN)
func (e FloatExpr[T]) Asin() FloatExpr[T] {
	return NewFloatExpr[T](e.asinExpr())
}

// Acos 反余弦 (ACOS)
func (e FloatExpr[T]) Acos() FloatExpr[T] {
	return NewFloatExpr[T](e.acosExpr())
}

// Atan 反正切 (ATAN)
func (e FloatExpr[T]) Atan() FloatExpr[T] {
	return NewFloatExpr[T](e.atanExpr())
}

// Radians 角度转弧度 (RADIANS)
func (e FloatExpr[T]) Radians() FloatExpr[T] {
	return NewFloatExpr[T](e.radiansExpr())
}

// Degrees 弧度转角度 (DEGREES)
func (e FloatExpr[T]) Degrees() FloatExpr[T] {
	return NewFloatExpr[T](e.degreesExpr())
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e FloatExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e FloatExpr[T]) CastSigned() IntExpr[int64] {
	return NewIntExpr[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e FloatExpr[T]) CastUnsigned() IntExpr[uint64] {
	return NewIntExpr[uint64](e.castUnsignedExpr())
}

// CastDecimal 转换为指定精度的小数 (CAST AS DECIMAL)
func (e FloatExpr[T]) CastDecimal(precision, scale int) DecimalExpr[float64] {
	return NewDecimalExpr[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e FloatExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// ==================== 格式化 ====================

// Format 格式化数字 (FORMAT)
func (e FloatExpr[T]) Format(decimals int) TextExpr[string] {
	return NewTextExpr[string](e.formatExpr(decimals))
}

// ==================== 聚合函数 ====================

// Sum 计算数值的总和 (SUM)
// SELECT SUM(price) FROM products;
// SELECT category, SUM(amount) FROM orders GROUP BY category;
func (e FloatExpr[T]) Sum() FloatExpr[T] {
	return NewFloatExpr[T](clause.Expr{
		SQL:  "SUM(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Avg 计算数值的平均值 (AVG)
// SELECT AVG(price) FROM products;
// SELECT category, AVG(rating) FROM reviews GROUP BY category;
func (e FloatExpr[T]) Avg() FloatExpr[T] {
	return NewFloatExpr[T](clause.Expr{
		SQL:  "AVG(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Max 返回最大值 (MAX)
// SELECT MAX(price) FROM products;
// SELECT category, MAX(temperature) FROM readings GROUP BY category;
func (e FloatExpr[T]) Max() FloatExpr[T] {
	return NewFloatExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Min 返回最小值 (MIN)
// SELECT MIN(price) FROM products;
// SELECT category, MIN(temperature) FROM readings GROUP BY category;
func (e FloatExpr[T]) Min() FloatExpr[T] {
	return NewFloatExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}
