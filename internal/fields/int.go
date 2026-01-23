package fields

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
)

var _ clause.Expression = (*IntExpr[int64])(nil)

// ==================== IntExpr 定义 ====================

// IntExpr 整数类型表达式，用于 COUNT 等返回整数的聚合函数
// 支持比较操作、算术运算、位运算和数学函数
// 使用场景：
//   - COUNT, COUNT_DISTINCT 等聚合函数的返回值
//   - 派生表中的整数列
//   - 整数字段的算术运算结果
type IntExpr[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	nullCondFuncSql
	numericCondFuncSql
	castSql
	formatSql
	bitOpSql
}

func NewIntExpr[T any](expr clause.Expression) IntExpr[T] {
	return IntExpr[T]{
		numericComparableImpl: numericComparableImpl[T]{baseComparableImpl[T]{expr}},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		nullCondFuncSql:       nullCondFuncSql{Expression: expr},
		numericCondFuncSql:    numericCondFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
		bitOpSql:              bitOpSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e IntExpr[T]) Build(builder clause.Builder) {
	e.numericComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e IntExpr[T]) ToExpr() clause.Expression {
	return e.numericComparableImpl.Expression
}

// As 创建一个别名字段
func (e IntExpr[T]) As(alias string) field.IField {
	return field.NewBaseFromSql(e.numericComparableImpl.Expression, alias)
}

// ==================== 算术运算 (特殊方法) ====================

// IntDiv 整数除法 (DIV)
// SELECT 10 DIV 3; -- 结果为 3
// SELECT total DIV page_size as pages FROM posts;
func (e IntExpr[T]) IntDiv(value any) IntExpr[T] {
	return NewIntExpr[T](e.intDivExpr(value))
}

// ==================== 位运算 ====================

// BitAnd 按位与 (&)
func (e IntExpr[T]) BitAnd(value any) IntExpr[T] {
	return NewIntExpr[T](e.bitAndExpr(value))
}

// BitOr 按位或 (|)
func (e IntExpr[T]) BitOr(value any) IntExpr[T] {
	return NewIntExpr[T](e.bitOrExpr(value))
}

// BitXor 按位异或 (^)
func (e IntExpr[T]) BitXor(value any) IntExpr[T] {
	return NewIntExpr[T](e.bitXorExpr(value))
}

// BitNot 按位取反 (~)
func (e IntExpr[T]) BitNot() IntExpr[T] {
	return NewIntExpr[T](e.bitNotExpr())
}

// LeftShift 左移 (<<)
func (e IntExpr[T]) LeftShift(n int) IntExpr[T] {
	return NewIntExpr[T](e.leftShiftExpr(n))
}

// RightShift 右移 (>>)
func (e IntExpr[T]) RightShift(n int) IntExpr[T] {
	return NewIntExpr[T](e.rightShiftExpr(n))
}

// ==================== 数学函数 (特殊方法) ====================

// Sign 返回数值的符号 (SIGN)：负数返回-1，零返回0，正数返回1
// SELECT SIGN(-10); -- 结果为 -1
// SELECT SIGN(0); -- 结果为 0
// SELECT SIGN(10); -- 结果为 1
// SELECT SIGN(balance) FROM accounts;
func (e IntExpr[T]) Sign() IntExpr[int8] {
	return NewIntExpr[int8](e.signExpr())
}

// Ceil 向上取整 (CEIL)，返回大于或等于X的最小整数
// SELECT CEIL(4.3); -- 结果为 5
// SELECT CEIL(-4.3); -- 结果为 -4
// SELECT CEIL(price * 1.1) FROM products;
func (e IntExpr[T]) Ceil() IntExpr[T] {
	return NewIntExpr[T](e.ceilExpr())
}

// Floor 向下取整 (FLOOR)，返回小于或等于X的最大整数
// SELECT FLOOR(4.9); -- 结果为 4
// SELECT FLOOR(-4.3); -- 结果为 -5
// SELECT FLOOR(price * 0.9) FROM products;
func (e IntExpr[T]) Floor() IntExpr[T] {
	return NewIntExpr[T](e.floorExpr())
}

// Pow 返回X的Y次幂 (POW)
// SELECT POW(2, 3); -- 结果为 8
// SELECT POW(10, 2); -- 结果为 100
// SELECT POW(users.level, 2) FROM users;
func (e IntExpr[T]) Pow(exponent int) FloatExpr[float64] {
	return NewFloatExpr[float64](e.powExpr(float64(exponent)))
}

// Sqrt 返回X的平方根 (SQRT)，X必须为非负数
// SELECT SQRT(4); -- 结果为 2
// SELECT SQRT(16); -- 结果为 4
// SELECT SQRT(area) as side_length FROM squares;
func (e IntExpr[T]) Sqrt() FloatExpr[float64] {
	return NewFloatExpr[float64](e.sqrtExpr())
}

// Log 自然对数 (LOG)
func (e IntExpr[T]) Log() FloatExpr[float64] {
	return NewFloatExpr[float64](e.logExpr())
}

// Log10 以10为底的对数 (LOG10)
func (e IntExpr[T]) Log10() FloatExpr[float64] {
	return NewFloatExpr[float64](e.log10Expr())
}

// Log2 以2为底的对数 (LOG2)
func (e IntExpr[T]) Log2() FloatExpr[float64] {
	return NewFloatExpr[float64](e.log2Expr())
}

// Exp 自然指数 (EXP)
func (e IntExpr[T]) Exp() FloatExpr[float64] {
	return NewFloatExpr[float64](e.expExpr())
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e IntExpr[T]) Cast(targetType string) clause.Expression {
	return e.castExpr(targetType)
}

// CastFloat 转换为浮点数 (CAST AS DECIMAL)
func (e IntExpr[T]) CastFloat(precision, scale int) FloatExpr[float64] {
	return NewFloatExpr[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e IntExpr[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e IntExpr[T]) CastSigned() IntExpr[int64] {
	return NewIntExpr[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e IntExpr[T]) CastUnsigned() IntExpr[uint64] {
	return NewIntExpr[uint64](e.castUnsignedExpr())
}

// ==================== 字符串转换 ====================

// Hex 转换为十六进制字符串 (HEX)
func (e IntExpr[T]) Hex() TextExpr[string] {
	return NewTextExpr[string](e.hexExpr())
}

// Bin 转换为二进制字符串 (BIN)
func (e IntExpr[T]) Bin() TextExpr[string] {
	return NewTextExpr[string](e.binExpr())
}

// Oct 转换为八进制字符串 (OCT)
func (e IntExpr[T]) Oct() TextExpr[string] {
	return NewTextExpr[string](e.octExpr())
}

// ==================== 聚合函数 ====================

// Sum 计算数值的总和 (SUM)
// SELECT SUM(quantity) FROM orders;
// SELECT user_id, SUM(points) FROM transactions GROUP BY user_id;
func (e IntExpr[T]) Sum() IntExpr[T] {
	return NewIntExpr[T](clause.Expr{
		SQL:  "SUM(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Avg 计算数值的平均值 (AVG)
// SELECT AVG(score) FROM students;
// SELECT class_id, AVG(grade) FROM exams GROUP BY class_id;
func (e IntExpr[T]) Avg() FloatExpr[float64] {
	return NewFloatExpr[float64](clause.Expr{
		SQL:  "AVG(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Max 返回最大值 (MAX)
// SELECT MAX(price) FROM products;
// SELECT category, MAX(stock) FROM inventory GROUP BY category;
func (e IntExpr[T]) Max() IntExpr[T] {
	return NewIntExpr[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Min 返回最小值 (MIN)
// SELECT MIN(price) FROM products;
// SELECT category, MIN(stock) FROM inventory GROUP BY category;
func (e IntExpr[T]) Min() IntExpr[T] {
	return NewIntExpr[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// ==================== 网络函数 ====================

// InetNtoa 将整数形式的IP地址转换为点分十进制字符串 (INET_NTOA)
// SELECT INET_NTOA(3232235777); -- 结果为 '192.168.1.1'
// SELECT INET_NTOA(ip_address) FROM access_logs;
func (e IntExpr[T]) InetNtoa() TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "INET_NTOA(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}
