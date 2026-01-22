package field

import (
	"github.com/donutnomad/gsql/clause"
)

// ==================== IntExprT 定义 ====================

// IntExprT 整数类型表达式，用于 COUNT 等返回整数的聚合函数
// 支持比较操作、算术运算、位运算和数学函数
// 使用场景：
//   - COUNT, COUNT_DISTINCT 等聚合函数的返回值
//   - 派生表中的整数列
//   - 整数字段的算术运算结果
type IntExprT[T any] struct {
	numericComparableImpl[T]
	pointerExprImpl
	arithmeticSql
	mathFuncSql
	condFuncSql
	castSql
	formatSql
	bitOpSql
}

// NewIntExprT 创建一个新的 IntExprT 实例
func NewIntExprT[T any](expr clause.Expression) IntExprT[T] {
	return IntExprT[T]{
		numericComparableImpl: numericComparableImpl[T]{Expression: expr},
		pointerExprImpl:       pointerExprImpl{Expression: expr},
		arithmeticSql:         arithmeticSql{Expression: expr},
		mathFuncSql:           mathFuncSql{Expression: expr},
		condFuncSql:           condFuncSql{Expression: expr},
		castSql:               castSql{Expression: expr},
		formatSql:             formatSql{Expression: expr},
		bitOpSql:              bitOpSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e IntExprT[T]) Build(builder clause.Builder) {
	e.numericComparableImpl.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e IntExprT[T]) ToExpr() Expression {
	return e.numericComparableImpl.Expression
}

// As 创建一个别名字段
func (e IntExprT[T]) As(alias string) IField {
	return NewBaseFromSql(e.numericComparableImpl.Expression, alias)
}

// ==================== 算术运算 ====================

// Add 加法 (+)
// SELECT price + 100 FROM products;
// SELECT users.age + 1 FROM users;
func (e IntExprT[T]) Add(value any) IntExprT[T] {
	return NewIntExprT[T](e.addExpr(value))
}

// Sub 减法 (-)
// SELECT price - discount FROM products;
// SELECT stock - sold FROM inventory;
func (e IntExprT[T]) Sub(value any) IntExprT[T] {
	return NewIntExprT[T](e.subExpr(value))
}

// Mul 乘法 (*)
// SELECT price * quantity FROM order_items;
// SELECT users.level * 10 as points FROM users;
func (e IntExprT[T]) Mul(value any) IntExprT[T] {
	return NewIntExprT[T](e.mulExpr(value))
}

// Div 除法 (/)
// SELECT total / count FROM stats;
// SELECT points / 100 as level FROM users;
func (e IntExprT[T]) Div(value any) IntExprT[T] {
	return NewIntExprT[T](e.divExpr(value))
}

// IntDiv 整数除法 (DIV)
// SELECT 10 DIV 3; -- 结果为 3
// SELECT total DIV page_size as pages FROM posts;
func (e IntExprT[T]) IntDiv(value any) IntExprT[T] {
	return NewIntExprT[T](e.intDivExpr(value))
}

// Mod 取模 (MOD)
// SELECT MOD(10, 3); -- 结果为 1
// SELECT MOD(234, 10); -- 结果为 4
// SELECT * FROM users WHERE MOD(id, 2) = 0; -- 偶数ID
func (e IntExprT[T]) Mod(value any) IntExprT[T] {
	return NewIntExprT[T](e.modExpr(value))
}

// Neg 取负 (-)
// SELECT -price FROM products;
// SELECT -balance FROM accounts;
func (e IntExprT[T]) Neg() IntExprT[T] {
	return NewIntExprT[T](e.negExpr())
}

// ==================== 位运算 ====================

// BitAnd 按位与 (&)
func (e IntExprT[T]) BitAnd(value any) IntExprT[T] {
	return NewIntExprT[T](e.bitAndExpr(value))
}

// BitOr 按位或 (|)
func (e IntExprT[T]) BitOr(value any) IntExprT[T] {
	return NewIntExprT[T](e.bitOrExpr(value))
}

// BitXor 按位异或 (^)
func (e IntExprT[T]) BitXor(value any) IntExprT[T] {
	return NewIntExprT[T](e.bitXorExpr(value))
}

// BitNot 按位取反 (~)
func (e IntExprT[T]) BitNot() IntExprT[T] {
	return NewIntExprT[T](e.bitNotExpr())
}

// LeftShift 左移 (<<)
func (e IntExprT[T]) LeftShift(n int) IntExprT[T] {
	return NewIntExprT[T](e.leftShiftExpr(n))
}

// RightShift 右移 (>>)
func (e IntExprT[T]) RightShift(n int) IntExprT[T] {
	return NewIntExprT[T](e.rightShiftExpr(n))
}

// ==================== 数学函数 ====================

// Abs 返回数值的绝对值 (ABS)
// SELECT ABS(-10); -- 结果为 10
// SELECT ABS(10); -- 结果为 10
// SELECT ABS(users.balance) FROM users;
// SELECT * FROM transactions WHERE ABS(amount) > 1000;
func (e IntExprT[T]) Abs() IntExprT[T] {
	return NewIntExprT[T](e.absExpr())
}

// Sign 返回数值的符号 (SIGN)：负数返回-1，零返回0，正数返回1
// SELECT SIGN(-10); -- 结果为 -1
// SELECT SIGN(0); -- 结果为 0
// SELECT SIGN(10); -- 结果为 1
// SELECT SIGN(balance) FROM accounts;
func (e IntExprT[T]) Sign() IntExprT[int8] {
	return NewIntExprT[int8](e.signExpr())
}

// Ceil 向上取整 (CEIL)，返回大于或等于X的最小整数
// SELECT CEIL(4.3); -- 结果为 5
// SELECT CEIL(-4.3); -- 结果为 -4
// SELECT CEIL(price * 1.1) FROM products;
func (e IntExprT[T]) Ceil() IntExprT[T] {
	return NewIntExprT[T](e.ceilExpr())
}

// Floor 向下取整 (FLOOR)，返回小于或等于X的最大整数
// SELECT FLOOR(4.9); -- 结果为 4
// SELECT FLOOR(-4.3); -- 结果为 -5
// SELECT FLOOR(price * 0.9) FROM products;
func (e IntExprT[T]) Floor() IntExprT[T] {
	return NewIntExprT[T](e.floorExpr())
}

// Round 四舍五入 (ROUND) 到指定小数位数，默认四舍五入到整数
// SELECT ROUND(4.567); -- 结果为 5
// SELECT ROUND(4.567, 2); -- 结果为 4.57
// SELECT ROUND(price, 2) FROM products;
// SELECT ROUND(123.456, -1); -- 结果为 120
func (e IntExprT[T]) Round(decimals ...int) IntExprT[T] {
	return NewIntExprT[T](e.roundExpr(decimals...))
}

// Pow 返回X的Y次幂 (POW)
// SELECT POW(2, 3); -- 结果为 8
// SELECT POW(10, 2); -- 结果为 100
// SELECT POW(users.level, 2) FROM users;
func (e IntExprT[T]) Pow(exponent int) FloatExprT[float64] {
	return NewFloatExprT[float64](e.powExpr(float64(exponent)))
}

// Sqrt 返回X的平方根 (SQRT)，X必须为非负数
// SELECT SQRT(4); -- 结果为 2
// SELECT SQRT(16); -- 结果为 4
// SELECT SQRT(area) as side_length FROM squares;
func (e IntExprT[T]) Sqrt() FloatExprT[float64] {
	return NewFloatExprT[float64](e.sqrtExpr())
}

// Log 自然对数 (LOG)
func (e IntExprT[T]) Log() FloatExprT[float64] {
	return NewFloatExprT[float64](e.logExpr())
}

// Log10 以10为底的对数 (LOG10)
func (e IntExprT[T]) Log10() FloatExprT[float64] {
	return NewFloatExprT[float64](e.log10Expr())
}

// Log2 以2为底的对数 (LOG2)
func (e IntExprT[T]) Log2() FloatExprT[float64] {
	return NewFloatExprT[float64](e.log2Expr())
}

// Exp 自然指数 (EXP)
func (e IntExprT[T]) Exp() FloatExprT[float64] {
	return NewFloatExprT[float64](e.expExpr())
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
func (e IntExprT[T]) Cast(targetType string) Expression {
	return e.castExpr(targetType)
}

// CastFloat 转换为浮点数 (CAST AS DECIMAL)
func (e IntExprT[T]) CastFloat(precision, scale int) FloatExprT[float64] {
	return NewFloatExprT[float64](e.castDecimalExpr(precision, scale))
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e IntExprT[T]) CastChar(length ...int) TextExpr[string] {
	return NewTextExpr[string](e.castCharExpr(length...))
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e IntExprT[T]) CastSigned() IntExprT[int64] {
	return NewIntExprT[int64](e.castSignedExpr())
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e IntExprT[T]) CastUnsigned() IntExprT[uint64] {
	return NewIntExprT[uint64](e.castUnsignedExpr())
}

// ==================== 字符串转换 ====================

// Hex 转换为十六进制字符串 (HEX)
func (e IntExprT[T]) Hex() TextExpr[string] {
	return NewTextExpr[string](e.hexExpr())
}

// Bin 转换为二进制字符串 (BIN)
func (e IntExprT[T]) Bin() TextExpr[string] {
	return NewTextExpr[string](e.binExpr())
}

// Oct 转换为八进制字符串 (OCT)
func (e IntExprT[T]) Oct() TextExpr[string] {
	return NewTextExpr[string](e.octExpr())
}

// ==================== 条件函数 ====================

// IfNull 如果为NULL则返回默认值 (IFNULL)
func (e IntExprT[T]) IfNull(defaultValue T) IntExprT[T] {
	return NewIntExprT[T](e.ifNullExpr(defaultValue))
}

// Coalesce 返回第一个非NULL值 (COALESCE)
func (e IntExprT[T]) Coalesce(values ...any) IntExprT[T] {
	return NewIntExprT[T](e.coalesceExpr(values...))
}

// Nullif 如果两个值相等则返回NULL (NULLIF)
func (e IntExprT[T]) Nullif(value T) IntExprT[T] {
	return NewIntExprT[T](e.nullifExpr(value))
}

// Greatest 返回最大值 (GREATEST)
func (e IntExprT[T]) Greatest(values ...any) IntExprT[T] {
	return NewIntExprT[T](e.greatestExpr(values...))
}

// Least 返回最小值 (LEAST)
func (e IntExprT[T]) Least(values ...any) IntExprT[T] {
	return NewIntExprT[T](e.leastExpr(values...))
}

// ==================== 聚合函数 ====================

// Sum 计算数值的总和 (SUM)
// SELECT SUM(quantity) FROM orders;
// SELECT user_id, SUM(points) FROM transactions GROUP BY user_id;
func (e IntExprT[T]) Sum() IntExprT[T] {
	return NewIntExprT[T](clause.Expr{
		SQL:  "SUM(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Avg 计算数值的平均值 (AVG)
// SELECT AVG(score) FROM students;
// SELECT class_id, AVG(grade) FROM exams GROUP BY class_id;
func (e IntExprT[T]) Avg() FloatExprT[float64] {
	return NewFloatExprT[float64](clause.Expr{
		SQL:  "AVG(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Max 返回最大值 (MAX)
// SELECT MAX(price) FROM products;
// SELECT category, MAX(stock) FROM inventory GROUP BY category;
func (e IntExprT[T]) Max() IntExprT[T] {
	return NewIntExprT[T](clause.Expr{
		SQL:  "MAX(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// Min 返回最小值 (MIN)
// SELECT MIN(price) FROM products;
// SELECT category, MIN(stock) FROM inventory GROUP BY category;
func (e IntExprT[T]) Min() IntExprT[T] {
	return NewIntExprT[T](clause.Expr{
		SQL:  "MIN(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}

// ==================== 网络函数 ====================

// InetNtoa 将整数形式的IP地址转换为点分十进制字符串 (INET_NTOA)
// SELECT INET_NTOA(3232235777); -- 结果为 '192.168.1.1'
// SELECT INET_NTOA(ip_address) FROM access_logs;
func (e IntExprT[T]) InetNtoa() TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "INET_NTOA(?)",
		Vars: []any{e.numericComparableImpl.Expression},
	})
}
