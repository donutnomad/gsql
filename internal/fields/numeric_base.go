package fields

import (
	"fmt"
	"strings"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/clauses2"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/types"
	"github.com/donutnomad/gsql/internal/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// ==================== 空值判断实现 ====================

// pointerExprImpl 用于表达式的空值判断实现
type pointerExprImpl struct {
	clause.Expression
}

func (f pointerExprImpl) IsNull() Condition {
	return Condition{clause.Expr{SQL: "? IS NULL", Vars: []any{f.Expression}}}
}

func (f pointerExprImpl) IsNotNull() Condition {
	return Condition{clause.Expr{SQL: "? IS NOT NULL", Vars: []any{f.Expression}}}
}

// @gen public=Count return=IntExpr[int64]
// Count 计算非NULL值的数量 (COUNT)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT COUNT(id) FROM users;
// SELECT status, COUNT(id) FROM orders GROUP BY status;
func (f pointerExprImpl) countExpr() clause.Expr {
	return clause.Expr{SQL: "COUNT(?)", Vars: []any{f.Expression}}
}

// @gen public=CountDistinct return=IntExpr[int64]
// CountDistinct 计算不重复非NULL值的数量 (COUNT DISTINCT)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT COUNT(DISTINCT status) FROM orders;
// SELECT user_id, COUNT(DISTINCT product_id) FROM cart GROUP BY user_id;
func (f pointerExprImpl) countDistinctExpr() clause.Expr {
	return clause.Expr{SQL: "COUNT(DISTINCT ?)", Vars: []any{f.Expression}}
}

// ==================== 基础表达式方法实现 ====================

// baseExprSql 提供基础表达式方法的实现
// Build, ToExpr, As 方法通过代码生成器生成
// 注意：不嵌入 clause.Expression 以避免与其他嵌入类型冲突
type baseExprSql struct {
	Expr clause.Expression
}

// @gen public=Build void=true
// buildExpr 实现 clause.Expression 接口的 Build 方法
func (b baseExprSql) buildExpr(builder clause.Builder) {
	if lo.IsNil(b.Expr) {
		builder.WriteString("NULL")
	} else {
		b.Expr.Build(builder)
	}
}

// @gen public=ToExpr return=clause.Expression direct=true
// toExprExpr 返回内部的 Expression
func (b baseExprSql) toExprExpr() clause.Expression {
	return b.Expr
}

// @gen public=As return=fieldi.IField direct=true
// asExpr 创建一个别名字段
func (b baseExprSql) asExpr(alias string) fieldi.IField {
	return fieldImpl{
		expr:  b.Expr,
		alias: alias,
	}
}

type fieldImpl struct {
	expr  clause.Expression
	alias string
}

func (f fieldImpl) Build(builder clause.Builder) {
	f.expr.Build(builder)
	builder.WriteString(" AS ")
	builder.WriteQuoted(f.alias)
}

func (f fieldImpl) ToExpr() clause.Expression {
	return f
}

func (f fieldImpl) FullName() string {
	return f.Alias()
}

func (f fieldImpl) Name() string {
	return f.Alias()
}

func (f fieldImpl) As(alias string) fieldi.IField {
	return fieldImpl{
		expr:  f.expr,
		alias: alias,
	}
}

func (f fieldImpl) Alias() string {
	return f.alias
}

// ==================== 基础比较操作实现（等于/不等于/In/NotIn）====================

// baseComparableImpl 基础比较操作实现
// 适用于所有类型（包括字符串），只包含等于、不等于、In、NotIn
type baseComparableImpl[T any] struct {
	clause.Expression
}

func (f baseComparableImpl[T]) Eq(value T) Condition {
	return cond("? = ?", f.Expression, value)
}

func (f baseComparableImpl[T]) EqF(other clause.Expression) Condition {
	return cond("? = ?", f.Expression, other)
}

func (f baseComparableImpl[T]) EqOpt(value mo.Option[T]) Condition {
	if value.IsAbsent() {
		return emptyCondition
	}
	return f.Eq(value.MustGet())
}

func (f baseComparableImpl[T]) Not(value T) Condition {
	return cond("? != ?", f.Expression, value)
}

func (f baseComparableImpl[T]) NotF(other clause.Expression) Condition {
	return cond("? != ?", f.Expression, other)
}

func (f baseComparableImpl[T]) NotOpt(value mo.Option[T]) Condition {
	if value.IsAbsent() {
		return emptyCondition
	}
	return f.Not(value.MustGet())
}

func (f baseComparableImpl[T]) In(values ...T) Condition {
	if len(values) == 0 {
		return emptyCondition
	}
	return cond("? IN ?", f.Expression, values)
}

func (f baseComparableImpl[T]) NotIn(values ...T) Condition {
	if len(values) == 0 {
		return emptyCondition
	}
	return cond("? NOT IN ?", f.Expression, values)
}

// ==================== 数值比较操作的通用实现 ====================

type Condition struct {
	clause.Expression
}

func cond(sql string, vars ...any) Condition {
	return Condition{clause.Expr{SQL: sql, Vars: vars}}
}

var emptyCondition = Condition{clause.Expr{}}

// numericComparableImpl 数值类型的比较操作通用实现
// 适用于 IntExpr, FloatExpr, DecimalExpr
// 嵌入 baseComparableImpl 获得基础比较操作，额外添加大于、小于、Between 等操作
type numericComparableImpl[T any] struct {
	baseComparableImpl[T]
}

func (f numericComparableImpl[T]) Gt(value T) Condition {
	return f.operateValue(value, ">")
}

func (f numericComparableImpl[T]) GtOpt(value mo.Option[T]) Condition {
	if value.IsAbsent() {
		return emptyCondition
	}
	return f.Gt(value.MustGet())
}

func (f numericComparableImpl[T]) GtF(other clause.Expression) Condition {
	return f.operateValue(other, ">")
}

func (f numericComparableImpl[T]) Gte(value T) Condition {
	return f.operateValue(value, ">=")
}

func (f numericComparableImpl[T]) GteOpt(value mo.Option[T]) Condition {
	if value.IsAbsent() {
		return emptyCondition
	}
	return f.Gte(value.MustGet())
}

func (f numericComparableImpl[T]) GteF(other clause.Expression) Condition {
	return f.operateValue(other, ">=")
}

func (f numericComparableImpl[T]) Lt(value T) Condition {
	return f.operateValue(value, "<")
}

func (f numericComparableImpl[T]) LtOpt(value mo.Option[T]) Condition {
	if value.IsAbsent() {
		return emptyCondition
	}
	return f.Lt(value.MustGet())
}

func (f numericComparableImpl[T]) LtF(other clause.Expression) Condition {
	return f.operateValue(other, "<")
}

func (f numericComparableImpl[T]) Lte(value T) Condition {
	return f.operateValue(value, "<=")
}

func (f numericComparableImpl[T]) LteOpt(value mo.Option[T]) Condition {
	if value.IsAbsent() {
		return emptyCondition
	}
	return f.Lte(value.MustGet())
}

func (f numericComparableImpl[T]) LteF(other clause.Expression) Condition {
	return f.operateValue(other, "<=")
}

func (f numericComparableImpl[T]) Between(from, to T) Condition {
	return cond("? BETWEEN ? AND ?", f.Expression, from, to)
}

func (f numericComparableImpl[T]) NotBetween(from, to T) Condition {
	return cond("? NOT BETWEEN ? AND ?", f.Expression, from, to)
}

// BetweenPtr 使用指针参数的范围查询
// 如果 from 或 to 为 nil，则使用 >= 或 <= 替代
func (f numericComparableImpl[T]) BetweenPtr(from, to *T) Condition {
	if from == nil && to == nil {
		return emptyCondition
	}
	if from == nil {
		return f.Lte(*to)
	}
	if to == nil {
		return f.Gte(*from)
	}
	return f.Between(*from, *to)
}

// BetweenOpt 使用 Option 参数的范围查询
func (f numericComparableImpl[T]) BetweenOpt(from, to mo.Option[T]) Condition {
	return f.BetweenPtr(from.ToPointer(), to.ToPointer())
}

// BetweenF 使用字段参数的范围查询
func (f numericComparableImpl[T]) BetweenF(from, to clause.Expression) Condition {
	if from == nil && to == nil {
		return emptyCondition
	}
	if from == nil {
		return f.LteF(to)
	}
	if to == nil {
		return f.GteF(from)
	}
	return cond("? BETWEEN ? AND ?", f.Expression, from, to)
}

// NotBetweenPtr 使用指针参数的范围排除查询
func (f numericComparableImpl[T]) NotBetweenPtr(from, to *T) Condition {
	if from == nil && to == nil {
		return emptyCondition
	}
	if from == nil {
		return f.Gt(*to)
	}
	if to == nil {
		return f.Lt(*from)
	}
	return f.NotBetween(*from, *to)
}

// NotBetweenOpt 使用 Option 参数的范围排除查询
func (f numericComparableImpl[T]) NotBetweenOpt(from, to mo.Option[T]) Condition {
	return f.NotBetweenPtr(from.ToPointer(), to.ToPointer())
}

func (f numericComparableImpl[T]) operateValue(value any, operator string) Condition {
	return f.operateValue2(f.Expression, value, operator)
}

func (f numericComparableImpl[T]) operateValue2(column clause.Expression, value any, operator string) Condition {
	var expr clause.Expression
	switch operator {
	case "=":
		expr = clause.Eq{Column: column, Value: value}
	case "!=":
		expr = clause.Neq{Column: column, Value: value}
	case ">":
		expr = clause.Gt{Column: column, Value: value}
	case ">=":
		expr = clause.Gte{Column: column, Value: value}
	case "<":
		expr = clause.Lt{Column: column, Value: value}
	case "<=":
		expr = clause.Lte{Column: column, Value: value}
	case "IN":
		expr = clause.IN{Column: column, Values: []any{value}}
	case "NOT IN":
		expr = clause.Not(clause.IN{Column: column, Values: []any{value}})
	default:
		panic(fmt.Sprintf("invalid operator %s", operator))
	}
	return Condition{expr}
}

// ==================== 算术运算的 SQL 生成 ====================

// arithmeticSql 生成算术运算的 SQL 表达式
type arithmeticSql struct {
	clause.Expression
}

// @gen public=Add
// Add 加法 (+)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT price + 100 FROM products;
// SELECT users.age + 1 FROM users;
func (a arithmeticSql) addExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? + ?", Vars: []any{a.Expression, value}}
}

// @gen public=Sub
// Sub 减法 (-)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT price - discount FROM products;
// SELECT stock - sold FROM inventory;
func (a arithmeticSql) subExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? - ?", Vars: []any{a.Expression, value}}
}

// @gen public=Mul
// Mul 乘法 (*)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT price * quantity FROM order_items;
// SELECT users.level * 10 as points FROM users;
func (a arithmeticSql) mulExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? * ?", Vars: []any{a.Expression, value}}
}

// @gen public=Div return=FloatExpr[float64] for=[IntExpr]
// @gen public=Div for=[FloatExpr,DecimalExpr]
// Div 除法 (/)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT total / count FROM stats;
// SELECT points / 100 as level FROM users;
func (a arithmeticSql) divExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? / ?", Vars: []any{a.Expression, value}}
}

// @gen public=Neg
// Neg 取负 (-)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT -price FROM products;
func (a arithmeticSql) negExpr() clause.Expr {
	return clause.Expr{SQL: "-?", Vars: []any{a.Expression}}
}

// @gen public=Mod
// Mod 取模 (MOD)
// 数据库支持: MySQL (PostgreSQL/SQLite 使用 % 操作符)
// SELECT MOD(10, 3); -- 结果为 1
// SELECT MOD(234, 10); -- 结果为 4
// SELECT * FROM users WHERE MOD(id, 2) = 0; -- 偶数ID
func (a arithmeticSql) modExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? MOD ?", Vars: []any{a.Expression, value}}
}

// ==================== 数学函数的 SQL 生成 ====================

// mathFuncSql 生成数学函数的 SQL 表达式
type mathFuncSql struct {
	clause.Expression
}

// @gen public=Abs
// Abs 返回绝对值 (ABS)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT ABS(-10); -- 结果为 10
// SELECT ABS(price - cost) FROM products;
func (m mathFuncSql) absExpr() clause.Expr {
	return clause.Expr{SQL: "ABS(?)", Vars: []any{m.Expression}}
}

// @gen public=Sign return=IntExpr[int8]
// Sign 返回符号 (SIGN)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT SIGN(-10); -- 结果为 -1
// SELECT SIGN(0); -- 结果为 0
// SELECT SIGN(10); -- 结果为 1
func (m mathFuncSql) signExpr() clause.Expr {
	return clause.Expr{SQL: "SIGN(?)", Vars: []any{m.Expression}}
}

// @gen public=Ceil for=[IntExpr]
// @gen public=Ceil return=IntExpr[int64] for=[FloatExpr,DecimalExpr]
// Ceil 向上取整 (CEIL)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT CEIL(1.5); -- 结果为 2
// SELECT CEIL(-1.5); -- 结果为 -1
func (m mathFuncSql) ceilExpr() clause.Expr {
	return clause.Expr{SQL: "CEIL(?)", Vars: []any{m.Expression}}
}

// @gen public=Floor for=[IntExpr]
// @gen public=Floor return=IntExpr[int64] for=[FloatExpr,DecimalExpr]
// Floor 向下取整 (FLOOR)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT FLOOR(1.5); -- 结果为 1
// SELECT FLOOR(-1.5); -- 结果为 -2
func (m mathFuncSql) floorExpr() clause.Expr {
	return clause.Expr{SQL: "FLOOR(?)", Vars: []any{m.Expression}}
}

// @gen public=Round exclude=[IntExpr]
// Round 四舍五入 (ROUND)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT ROUND(1.567); -- 结果为 2
// SELECT ROUND(1.567, 2); -- 结果为 1.57
func (m mathFuncSql) roundExpr(decimals ...int) clause.Expr {
	if len(decimals) > 0 {
		return clause.Expr{SQL: "ROUND(?, ?)", Vars: []any{m.Expression, decimals[0]}}
	}
	return clause.Expr{SQL: "ROUND(?)", Vars: []any{m.Expression}}
}

// @gen public=Truncate exclude=[IntExpr]
// Truncate 截断小数 (TRUNCATE)
// 数据库支持: MySQL (PostgreSQL 使用 TRUNC, SQLite 不支持)
// SELECT TRUNCATE(1.567, 2); -- 结果为 1.56
// SELECT TRUNCATE(1.567, 0); -- 结果为 1
func (m mathFuncSql) truncateExpr(decimals int) clause.Expr {
	return clause.Expr{SQL: "TRUNCATE(?, ?)", Vars: []any{m.Expression, decimals}}
}

// @gen public=Pow return=FloatExpr[T] for=[FloatExpr]
// @gen public=Pow return=FloatExpr[float64] for=[IntExpr,DecimalExpr]
// Pow 幂运算 (POW)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT POW(2, 3); -- 结果为 8
// SELECT POW(price, 2) FROM products;
func (m mathFuncSql) powExpr(exponent float64) clause.Expr {
	return clause.Expr{SQL: "POW(?, ?)", Vars: []any{m.Expression, exponent}}
}

// @gen public=Sqrt return=FloatExpr[T] for=[FloatExpr]
// @gen public=Sqrt return=FloatExpr[float64] for=[IntExpr,DecimalExpr]
// Sqrt 平方根 (SQRT)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT SQRT(16); -- 结果为 4
// SELECT SQRT(variance) FROM stats;
func (m mathFuncSql) sqrtExpr() clause.Expr {
	return clause.Expr{SQL: "SQRT(?)", Vars: []any{m.Expression}}
}

// @gen public=Log return=FloatExpr[T] for=[FloatExpr]
// @gen public=Log return=FloatExpr[float64] for=[IntExpr,DecimalExpr]
// Log 自然对数 (LOG)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT LOG(10); -- 结果为 2.302585...
func (m mathFuncSql) logExpr() clause.Expr {
	return clause.Expr{SQL: "LOG(?)", Vars: []any{m.Expression}}
}

// @gen public=Log10 return=FloatExpr[T] for=[FloatExpr]
// @gen public=Log10 return=FloatExpr[float64] for=[IntExpr,DecimalExpr]
// Log10 以10为底的对数 (LOG10)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT LOG10(100); -- 结果为 2
func (m mathFuncSql) log10Expr() clause.Expr {
	return clause.Expr{SQL: "LOG10(?)", Vars: []any{m.Expression}}
}

// @gen public=Log2 return=FloatExpr[T] for=[FloatExpr]
// @gen public=Log2 return=FloatExpr[float64] for=[IntExpr,DecimalExpr]
// Log2 以2为底的对数 (LOG2)
// 数据库支持: MySQL (PostgreSQL/SQLite 不直接支持)
// SELECT LOG2(8); -- 结果为 3
func (m mathFuncSql) log2Expr() clause.Expr {
	return clause.Expr{SQL: "LOG2(?)", Vars: []any{m.Expression}}
}

// @gen public=Exp return=FloatExpr[T] for=[FloatExpr]
// @gen public=Exp return=FloatExpr[float64] for=[IntExpr,DecimalExpr]
// Exp 指数函数 (EXP)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT EXP(1); -- 结果为 2.718281828...
func (m mathFuncSql) expExpr() clause.Expr {
	return clause.Expr{SQL: "EXP(?)", Vars: []any{m.Expression}}
}

// ==================== 条件函数的 SQL 生成 ====================

// nullCondFuncSql 生成空值处理函数的 SQL 表达式
// 适用于所有类型（包括 StringExpr）
type nullCondFuncSql struct {
	clause.Expression
}

// @gen public=IfNull
// IfNull 如果表达式为NULL则返回默认值
// 内部使用 COALESCE 实现，等价于 Coalesce(defaultValue)
func (c nullCondFuncSql) ifNullExpr(defaultValue any) clause.Expr {
	return c.coalesceExpr(defaultValue)
}

// @gen public=Coalesce
// Coalesce 返回参数列表中第一个非NULL的值 (COALESCE)
// 数据库支持: MySQL, PostgreSQL, SQLite (SQL 标准函数)
// SELECT COALESCE(nickname, username, 'Anonymous') FROM users;
func (c nullCondFuncSql) coalesceExpr(values ...any) clause.Expr {
	allArgs := []any{c.Expression}
	placeholders := "?"
	for _, v := range values {
		placeholders += ", ?"
		allArgs = append(allArgs, v)
	}
	return clause.Expr{SQL: "COALESCE(" + placeholders + ")", Vars: allArgs}
}

// @gen public=NullIf
// NullIf 如果两个表达式相等则返回NULL，否则返回第一个表达式 (NULLIF)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT NULLIF(username, ") FROM users; -- 空字符串转为NULL
func (c nullCondFuncSql) nullifExpr(value any) clause.Expr {
	return clause.Expr{SQL: "NULLIF(?, ?)", Vars: []any{c.Expression, value}}
}

// numericCondFuncSql 生成数值比较函数的 SQL 表达式
// 仅适用于数值类型（IntExpr, FloatExpr, DecimalExpr）
type numericCondFuncSql struct {
	clause.Expression
}

// @gen public=Greatest
// Greatest 返回参数列表中的最大值 (GREATEST)
// 数据库支持: MySQL, PostgreSQL (SQLite 不支持)
// SELECT GREATEST(10, 20, 30); -- 返回 30
// SELECT GREATEST(price, min_price) FROM products;
func (c numericCondFuncSql) greatestExpr(values ...any) clause.Expr {
	allArgs := []any{c.Expression}
	placeholders := "?"
	for _, v := range values {
		placeholders += ", ?"
		allArgs = append(allArgs, v)
	}
	return clause.Expr{SQL: "GREATEST(" + placeholders + ")", Vars: allArgs}
}

// @gen public=Least
// Least 返回参数列表中的最小值 (LEAST)
// 数据库支持: MySQL, PostgreSQL (SQLite 不支持)
// SELECT LEAST(10, 20, 30); -- 返回 10
// SELECT LEAST(price, max_price) FROM products;
func (c numericCondFuncSql) leastExpr(values ...any) clause.Expr {
	allArgs := []any{c.Expression}
	placeholders := "?"
	for _, v := range values {
		placeholders += ", ?"
		allArgs = append(allArgs, v)
	}
	return clause.Expr{SQL: "LEAST(" + placeholders + ")", Vars: allArgs}
}

// ==================== 类型转换的 SQL 生成 ====================

// castSql 生成类型转换的 SQL 表达式
type castSql struct {
	clause.Expression
}

func (c castSql) castExpr(targetType string) clause.Expr {
	return clause.Expr{SQL: "CAST(? AS " + targetType + ")", Vars: []any{c.Expression}}
}

func (c castSql) castSignedExpr() clause.Expr {
	return clause.Expr{SQL: "CAST(? AS SIGNED)", Vars: []any{c.Expression}}
}

func (c castSql) castUnsignedExpr() clause.Expr {
	return clause.Expr{SQL: "CAST(? AS UNSIGNED)", Vars: []any{c.Expression}}
}

func (c castSql) castDecimalExpr(precision, scale int) clause.Expr {
	return clause.Expr{SQL: fmt.Sprintf("CAST(? AS DECIMAL(%d, %d))", precision, scale), Vars: []any{c.Expression}}
}

func (c castSql) castCharExpr(length ...int) clause.Expr {
	if len(length) > 0 {
		return clause.Expr{SQL: fmt.Sprintf("CAST(? AS CHAR(%d))", length[0]), Vars: []any{c.Expression}}
	}
	return clause.Expr{SQL: "CAST(? AS CHAR)", Vars: []any{c.Expression}}
}

func (c castSql) castDoubleExpr() clause.Expr {
	return clause.Expr{SQL: "CAST(? AS DOUBLE)", Vars: []any{c.Expression}}
}

func (c castSql) castDateExpr() clause.Expr {
	return clause.Expr{SQL: "CAST(? AS DATE)", Vars: []any{c.Expression}}
}

func (c castSql) castDatetimeExpr() clause.Expr {
	return clause.Expr{SQL: "CAST(? AS DATETIME)", Vars: []any{c.Expression}}
}

func (c castSql) castTimeExpr() clause.Expr {
	return clause.Expr{SQL: "CAST(? AS TIME)", Vars: []any{c.Expression}}
}

// ==================== 格式化的 SQL 生成 ====================

// formatSql 生成格式化的 SQL 表达式
type formatSql struct {
	clause.Expression
}

func (f formatSql) formatExpr(decimals int) clause.Expr {
	return clause.Expr{SQL: "FORMAT(?, ?)", Vars: []any{f.Expression, decimals}}
}

func (f formatSql) hexExpr() clause.Expr {
	return clause.Expr{SQL: "HEX(?)", Vars: []any{f.Expression}}
}

func (f formatSql) binExpr() clause.Expr {
	return clause.Expr{SQL: "BIN(?)", Vars: []any{f.Expression}}
}

func (f formatSql) octExpr() clause.Expr {
	return clause.Expr{SQL: "OCT(?)", Vars: []any{f.Expression}}
}

// ==================== 三角函数的 SQL 生成 ====================

// trigFuncSql 生成三角函数的 SQL 表达式
type trigFuncSql struct {
	clause.Expression
}

// @gen public=Sin return=FloatExpr[T] for=[FloatExpr]
// Sin 正弦 (SIN)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT SIN(0); -- 结果为 0
// SELECT SIN(PI()/2); -- 结果为 1
func (t trigFuncSql) sinExpr() clause.Expr {
	return clause.Expr{SQL: "SIN(?)", Vars: []any{t.Expression}}
}

// @gen public=Cos return=FloatExpr[T] for=[FloatExpr]
// Cos 余弦 (COS)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT COS(0); -- 结果为 1
// SELECT COS(PI()); -- 结果为 -1
func (t trigFuncSql) cosExpr() clause.Expr {
	return clause.Expr{SQL: "COS(?)", Vars: []any{t.Expression}}
}

// @gen public=Tan return=FloatExpr[T] for=[FloatExpr]
// Tan 正切 (TAN)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT TAN(0); -- 结果为 0
// SELECT TAN(PI()/4); -- 结果约为 1
func (t trigFuncSql) tanExpr() clause.Expr {
	return clause.Expr{SQL: "TAN(?)", Vars: []any{t.Expression}}
}

// @gen public=Asin return=FloatExpr[T] for=[FloatExpr]
// Asin 反正弦 (ASIN)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT ASIN(0); -- 结果为 0
// SELECT ASIN(1); -- 结果为 PI()/2
func (t trigFuncSql) asinExpr() clause.Expr {
	return clause.Expr{SQL: "ASIN(?)", Vars: []any{t.Expression}}
}

// @gen public=Acos return=FloatExpr[T] for=[FloatExpr]
// Acos 反余弦 (ACOS)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT ACOS(1); -- 结果为 0
// SELECT ACOS(0); -- 结果为 PI()/2
func (t trigFuncSql) acosExpr() clause.Expr {
	return clause.Expr{SQL: "ACOS(?)", Vars: []any{t.Expression}}
}

// @gen public=Atan return=FloatExpr[T] for=[FloatExpr]
// Atan 反正切 (ATAN)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT ATAN(0); -- 结果为 0
// SELECT ATAN(1); -- 结果为 PI()/4
func (t trigFuncSql) atanExpr() clause.Expr {
	return clause.Expr{SQL: "ATAN(?)", Vars: []any{t.Expression}}
}

// @gen public=Radians return=FloatExpr[T] for=[FloatExpr]
// Radians 角度转弧度 (RADIANS)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT RADIANS(180); -- 结果为 PI()
// SELECT RADIANS(90); -- 结果为 PI()/2
func (t trigFuncSql) radiansExpr() clause.Expr {
	return clause.Expr{SQL: "RADIANS(?)", Vars: []any{t.Expression}}
}

// @gen public=Degrees return=FloatExpr[T] for=[FloatExpr]
// Degrees 弧度转角度 (DEGREES)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT DEGREES(PI()); -- 结果为 180
// SELECT DEGREES(PI()/2); -- 结果为 90
func (t trigFuncSql) degreesExpr() clause.Expr {
	return clause.Expr{SQL: "DEGREES(?)", Vars: []any{t.Expression}}
}

// ==================== 位运算的 SQL 生成 ====================

// bitOpSql 生成位运算的 SQL 表达式
type bitOpSql struct {
	clause.Expression
}

// @gen public=BitAnd for=[IntExpr]
// BitAnd 按位与 (&)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT 5 & 3; -- 结果为 1
// SELECT flags & 0x0F FROM settings;
func (b bitOpSql) bitAndExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? & ?", Vars: []any{b.Expression, value}}
}

// @gen public=BitOr for=[IntExpr]
// BitOr 按位或 (|)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT 5 | 3; -- 结果为 7
// SELECT flags | 0x10 FROM settings;
func (b bitOpSql) bitOrExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? | ?", Vars: []any{b.Expression, value}}
}

// @gen public=BitXor for=[IntExpr]
// BitXor 按位异或 (^)
// 数据库支持: MySQL, PostgreSQL (使用 #), SQLite
// SELECT 5 ^ 3; -- 结果为 6
// SELECT flags ^ mask FROM settings;
func (b bitOpSql) bitXorExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? ^ ?", Vars: []any{b.Expression, value}}
}

// @gen public=BitNot for=[IntExpr]
// BitNot 按位取反 (~)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT ~5; -- 结果为 -6 (有符号整数)
// SELECT ~flags FROM settings;
func (b bitOpSql) bitNotExpr() clause.Expr {
	return clause.Expr{SQL: "~?", Vars: []any{b.Expression}}
}

// @gen public=LeftShift for=[IntExpr]
// LeftShift 左移 (<<)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT 1 << 4; -- 结果为 16
// SELECT value << 2 FROM data;
func (b bitOpSql) leftShiftExpr(n int) clause.Expr {
	return clause.Expr{SQL: "? << ?", Vars: []any{b.Expression, n}}
}

// @gen public=RightShift for=[IntExpr]
// RightShift 右移 (>>)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT 16 >> 2; -- 结果为 4
// SELECT value >> 1 FROM data;
func (b bitOpSql) rightShiftExpr(n int) clause.Expr {
	return clause.Expr{SQL: "? >> ?", Vars: []any{b.Expression, n}}
}

// @gen public=IntDiv for=[IntExpr]
// IntDiv 整数除法 (DIV)
// 数据库支持: MySQL (PostgreSQL/SQLite 使用 / 或 TRUNC)
// SELECT 10 DIV 3; -- 结果为 3
// SELECT total DIV page_size as pages FROM posts;
func (b bitOpSql) intDivExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? DIV ?", Vars: []any{b.Expression, value}}
}

// ==================== 聚合函数的 SQL 生成 ====================

// aggregateSql 生成聚合函数的 SQL 表达式
type aggregateSql struct {
	clause.Expression
}

// @gen public=Sum return=DecimalExpr[T] for=[IntExpr]
// @gen public=Sum for=[FloatExpr,DecimalExpr]
// Sum 计算数值的总和 (SUM)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT SUM(quantity) FROM orders;
// SELECT user_id, SUM(points) FROM transactions GROUP BY user_id;
func (a aggregateSql) sumExpr() clause.Expr {
	return clause.Expr{SQL: "SUM(?)", Vars: []any{a.Expression}}
}

// @gen public=Avg return=FloatExpr[float64]
// Avg 计算数值的平均值 (AVG)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT AVG(score) FROM students;
// SELECT class_id, AVG(grade) FROM exams GROUP BY class_id;
func (a aggregateSql) avgExpr() clause.Expr {
	return clause.Expr{SQL: "AVG(?)", Vars: []any{a.Expression}}
}

// @gen public=Max
// Max 返回最大值 (MAX)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT MAX(price) FROM products;
// SELECT category, MAX(stock) FROM inventory GROUP BY category;
func (a aggregateSql) maxExpr() clause.Expr {
	return clause.Expr{SQL: "MAX(?)", Vars: []any{a.Expression}}
}

// @gen public=Min
// Min 返回最小值 (MIN)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT MIN(price) FROM products;
// SELECT category, MIN(stock) FROM inventory GROUP BY category;
func (a aggregateSql) minExpr() clause.Expr {
	return clause.Expr{SQL: "MIN(?)", Vars: []any{a.Expression}}
}

// ==================== 日期提取函数的 SQL 生成 ====================

// dateExtractSql 生成日期提取函数的 SQL 表达式
// 适用于 DateExpr, DateTimeExpr
type dateExtractSql struct {
	clause.Expression
}

// @gen public=Year return=IntExpr[int]
// YearExpr 提取年份部分 (YEAR)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT YEAR(date_column) FROM table;
func (d dateExtractSql) yearExpr() clause.Expr {
	return clause.Expr{SQL: "YEAR(?)", Vars: []any{d.Expression}}
}

// @gen public=Month return=IntExpr[int]
// Month 提取月份部分 (MONTH)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT MONTH(date_column) FROM table;
func (d dateExtractSql) monthExpr() clause.Expr {
	return clause.Expr{SQL: "MONTH(?)", Vars: []any{d.Expression}}
}

// @gen public=Day return=IntExpr[int]
// Day 提取天数部分 (DAY)
// 数据库支持: MySQL, PostgreSQL, SQLite
// SELECT DAY(date_column) FROM table;
func (d dateExtractSql) dayExpr() clause.Expr {
	return clause.Expr{SQL: "DAY(?)", Vars: []any{d.Expression}}
}

// @gen public=DayOfMonth return=IntExpr[int]
// DayOfMonth 提取一月中的天数 (DAYOFMONTH)
// 数据库支持: MySQL
// 与 DAY() 等价
func (d dateExtractSql) dayOfMonthExpr() clause.Expr {
	return clause.Expr{SQL: "DAYOFMONTH(?)", Vars: []any{d.Expression}}
}

// @gen public=DayOfWeek return=IntExpr[int]
// DayOfWeek 返回一周中的索引 (DAYOFWEEK)
// 数据库支持: MySQL
// 1=周日, 2=周一, ..., 7=周六
func (d dateExtractSql) dayOfWeekExpr() clause.Expr {
	return clause.Expr{SQL: "DAYOFWEEK(?)", Vars: []any{d.Expression}}
}

// @gen public=DayOfYear return=IntExpr[int]
// DayOfYear 返回一年中的天数 (DAYOFYEAR)
// 数据库支持: MySQL
// 范围: 1-366
func (d dateExtractSql) dayOfYearExpr() clause.Expr {
	return clause.Expr{SQL: "DAYOFYEAR(?)", Vars: []any{d.Expression}}
}

// @gen public=Week return=IntExpr[int]
// Week 提取周数 (WEEK)
// 数据库支持: MySQL
// 范围: 0-53
func (d dateExtractSql) weekExpr() clause.Expr {
	return clause.Expr{SQL: "WEEK(?)", Vars: []any{d.Expression}}
}

// @gen public=WeekOfYear return=IntExpr[int]
// WeekOfYear 提取周数 (WEEKOFYEAR)
// 数据库支持: MySQL
// 范围: 1-53，相当于 WEEK(date, 3)
func (d dateExtractSql) weekOfYearExpr() clause.Expr {
	return clause.Expr{SQL: "WEEKOFYEAR(?)", Vars: []any{d.Expression}}
}

// @gen public=Quarter return=IntExpr[int]
// Quarter 提取季度 (QUARTER)
// 数据库支持: MySQL
// 范围: 1-4
func (d dateExtractSql) quarterExpr() clause.Expr {
	return clause.Expr{SQL: "QUARTER(?)", Vars: []any{d.Expression}}
}

// @gen public=LastDay return=DateExpr[string]
// LastDay 返回指定日期所在月份的最后一天 (LAST_DAY)
// 数据库支持: MySQL
// SELECT LAST_DAY('2024-02-15'); -- 返回 '2024-02-29'
func (d dateExtractSql) lastDayExpr() clause.Expr {
	return clause.Expr{SQL: "LAST_DAY(?)", Vars: []any{d.Expression}}
}

// @gen public=DayName return=StringExpr[string]
// DayName 返回日期的星期名称 (DAYNAME)
// 数据库支持: MySQL
// SELECT DAYNAME('2024-01-15'); -- 返回 'Monday'
func (d dateExtractSql) dayNameExpr() clause.Expr {
	return clause.Expr{SQL: "DAYNAME(?)", Vars: []any{d.Expression}}
}

// @gen public=MonthName return=StringExpr[string]
// MonthName 返回日期的月份名称 (MONTHNAME)
// 数据库支持: MySQL
// SELECT MONTHNAME('2024-01-15'); -- 返回 'January'
func (d dateExtractSql) monthNameExpr() clause.Expr {
	return clause.Expr{SQL: "MONTHNAME(?)", Vars: []any{d.Expression}}
}

// @gen public=ToDays return=IntExpr[int]
// ToDays 将日期转换为天数（从公元0年开始）(TO_DAYS)
// 数据库支持: MySQL
// SELECT TO_DAYS('2024-01-15'); -- 返回 739259
func (d dateExtractSql) toDaysExpr() clause.Expr {
	return clause.Expr{SQL: "TO_DAYS(?)", Vars: []any{d.Expression}}
}

// ==================== 时间提取函数的 SQL 生成 ====================

// timeExtractSql 生成时间提取函数的 SQL 表达式
// 适用于 DateTimeExpr, TimeExpr
type timeExtractSql struct {
	clause.Expression
}

// @gen public=Hour return=IntExpr[int]
// Hour 提取小时部分 (HOUR)
// 数据库支持: MySQL, PostgreSQL, SQLite
// 范围: 0-23
func (t timeExtractSql) hourExpr() clause.Expr {
	return clause.Expr{SQL: "HOUR(?)", Vars: []any{t.Expression}}
}

// @gen public=Minute return=IntExpr[int]
// Minute 提取分钟部分 (MINUTE)
// 数据库支持: MySQL, PostgreSQL, SQLite
// 范围: 0-59
func (t timeExtractSql) minuteExpr() clause.Expr {
	return clause.Expr{SQL: "MINUTE(?)", Vars: []any{t.Expression}}
}

// @gen public=Second return=IntExpr[int]
// Second 提取秒数部分 (SECOND)
// 数据库支持: MySQL, PostgreSQL, SQLite
// 范围: 0-59
func (t timeExtractSql) secondExpr() clause.Expr {
	return clause.Expr{SQL: "SECOND(?)", Vars: []any{t.Expression}}
}

// @gen public=Microsecond return=IntExpr[int]
// Microsecond 提取微秒部分 (MICROSECOND)
// 数据库支持: MySQL
// 范围: 0-999999
func (t timeExtractSql) microsecondExpr() clause.Expr {
	return clause.Expr{SQL: "MICROSECOND(?)", Vars: []any{t.Expression}}
}

// @gen public=TimeToSec return=IntExpr[int]
// TimeToSec 将时间转换为秒数 (TIME_TO_SEC)
// 数据库支持: MySQL
// SELECT TIME_TO_SEC('01:30:00'); -- 返回 5400
func (t timeExtractSql) timeToSecExpr() clause.Expr {
	return clause.Expr{SQL: "TIME_TO_SEC(?)", Vars: []any{t.Expression}}
}

// ==================== 日期时间运算的 SQL 生成 ====================

// dateIntervalSql 生成日期时间间隔运算的 SQL 表达式
// 适用于 DateExpr, DateTimeExpr, TimeExpr
type dateIntervalSql struct {
	clause.Expression
}

// @gen public=AddInterval
// AddInterval 在日期/时间上增加时间间隔 (DATE_ADD)
// 数据库支持: MySQL
// interval 格式: "1 DAY", "2 MONTH", "1 YEAR" 等
// 支持单位: MICROSECOND, SECOND, MINUTE, HOUR, DAY, WEEK, MONTH, QUARTER, YEAR
// SELECT DATE_ADD(date_column, INTERVAL 1 DAY) FROM table;
func (d dateIntervalSql) addIntervalExpr(interval string) clause.Expr {
	safeInterval := parseInterval(interval, "AddInterval")
	return clause.Expr{
		SQL:  fmt.Sprintf("DATE_ADD(?, INTERVAL %s)", safeInterval),
		Vars: []any{d.Expression},
	}
}

// @gen public=SubInterval
// SubInterval 从日期/时间中减去时间间隔 (DATE_SUB)
// 数据库支持: MySQL
// interval 格式: "1 DAY", "2 MONTH", "1 YEAR" 等
// SELECT DATE_SUB(date_column, INTERVAL 1 MONTH) FROM table;
func (d dateIntervalSql) subIntervalExpr(interval string) clause.Expr {
	safeInterval := parseInterval(interval, "SubInterval")
	return clause.Expr{
		SQL:  fmt.Sprintf("DATE_SUB(?, INTERVAL %s)", safeInterval),
		Vars: []any{d.Expression},
	}
}

// dateDiffSql 生成日期差值计算的 SQL 表达式
// 适用于 DateExpr, DateTimeExpr
type dateDiffSql struct {
	clause.Expression
}

// @gen public=DateDiff return=IntExpr[int]
// DateDiff 计算与另一个日期的差值（天数）(DATEDIFF)
// 数据库支持: MySQL
// 返回 this - other 的天数
// SELECT DATEDIFF(end_date, start_date) FROM events;
func (d dateDiffSql) dateDiffExpr(other clause.Expression) clause.Expr {
	return clause.Expr{SQL: "DATEDIFF(?, ?)", Vars: []any{d.Expression, other}}
}

// timeDiffSql 生成时间差值计算的 SQL 表达式
// 适用于 DateTimeExpr, TimeExpr
type timeDiffSql struct {
	clause.Expression
}

// @gen public=TimeDiff return=TimeExpr[T]
// TimeDiff 计算与另一个时间的差值 (TIMEDIFF)
// 数据库支持: MySQL
// SELECT TIMEDIFF(end_time, start_time) FROM events;
func (t timeDiffSql) timeDiffExpr(other clause.Expression) clause.Expr {
	return clause.Expr{SQL: "TIMEDIFF(?, ?)", Vars: []any{t.Expression, other}}
}

// timestampDiffSql 生成时间戳差值计算的 SQL 表达式
// 适用于 DateTimeExpr
type timestampDiffSql struct {
	clause.Expression
}

// @gen public=TimestampDiff return=IntExpr[int64]
// TimestampDiff 计算与另一个日期时间的差值（指定单位）(TIMESTAMPDIFF)
// 数据库支持: MySQL
// unit: MICROSECOND, SECOND, MINUTE, HOUR, DAY, WEEK, MONTH, QUARTER, YEAR
// SELECT TIMESTAMPDIFF(DAY, start_date, end_date) FROM events;
func (t timestampDiffSql) timestampDiffExpr(unit string, other clause.Expression) clause.Expr {
	unit = strings.ToUpper(strings.TrimSpace(unit))
	if !allowedIntervalUnits[unit] {
		panic(fmt.Sprintf("TimestampDiff: invalid unit: %s", unit))
	}
	return clause.Expr{
		SQL:  fmt.Sprintf("TIMESTAMPDIFF(%s, ?, ?)", unit),
		Vars: []any{other, t.Expression},
	}
}

// dateFormatSql 生成日期格式化的 SQL 表达式
// 适用于 DateExpr, DateTimeExpr
type dateFormatSql struct {
	clause.Expression
}

// @gen public=Format return=StringExpr[string]
// DateFormat 格式化日期为字符串 (DATE_FORMAT)
// 数据库支持: MySQL
// SELECT DATE_FORMAT(date_column, '%Y年%m月%d日') FROM table;
func (d dateFormatSql) dateFormatExpr(format string) clause.Expr {
	return clause.Expr{SQL: "DATE_FORMAT(?, ?)", Vars: []any{d.Expression, format}}
}

// timeFormatSql 生成时间格式化的 SQL 表达式
// 适用于 TimeExpr
type timeFormatSql struct {
	clause.Expression
}

// @gen public=Format return=StringExpr[string]
// TimeFormat 格式化时间为字符串 (TIME_FORMAT)
// 数据库支持: MySQL
// SELECT TIME_FORMAT(time_column, '%H:%i:%s') FROM table;
func (t timeFormatSql) timeFormatExpr(format string) clause.Expr {
	return clause.Expr{SQL: "TIME_FORMAT(?, ?)", Vars: []any{t.Expression, format}}
}

// dateConversionSql 生成日期时间转换的 SQL 表达式
// 适用于 DateTimeExpr
type dateConversionSql struct {
	clause.Expression
}

// @gen public=Date return=DateExpr[string]
// DateExpr 提取日期部分 (DATE)
// 数据库支持: MySQL
// SELECT DATE(datetime_column) FROM table;
func (d dateConversionSql) extractDateExpr() clause.Expr {
	return clause.Expr{SQL: "DATE(?)", Vars: []any{d.Expression}}
}

// @gen public=Time return=TimeExpr[string]
// TimeExpr 提取时间部分 (TIME)
// 数据库支持: MySQL
// SELECT TIME(datetime_column) FROM table;
func (d dateConversionSql) extractTimeExpr() clause.Expr {
	return clause.Expr{SQL: "TIME(?)", Vars: []any{d.Expression}}
}

// unixTimestampSql 生成 Unix 时间戳转换的 SQL 表达式
// 适用于 DateExpr, DateTimeExpr
type unixTimestampSql struct {
	clause.Expression
}

// @gen public=UnixTimestamp return=IntExpr[int64]
// UnixTimestamp 转换为 Unix 时间戳（秒）(UNIX_TIMESTAMP)
// 数据库支持: MySQL
// SELECT UNIX_TIMESTAMP(date_column) FROM table;
func (u unixTimestampSql) unixTimestampExpr() clause.Expr {
	return clause.Expr{SQL: "UNIX_TIMESTAMP(?)", Vars: []any{u.Expression}}
}

// ==================== 模式匹配的 SQL 生成 ====================

// patternExprImpl 用于表达式的模式匹配实现
// 适用于 StringExpr
type patternExprImpl[T any] struct {
	clause.Expression
}

func (f patternExprImpl[T]) Like(value T, escape ...byte) clause.Expression {
	return f.operateValue(value, "LIKE", utils.Optional(escape, 0))
}

func (f patternExprImpl[T]) LikeOpt(value mo.Option[T], escape ...byte) clause.Expression {
	if value.IsAbsent() {
		return types.EmptyExpression
	}
	return f.Like(value.MustGet(), escape...)
}

func (f patternExprImpl[T]) NotLike(value T, escape ...byte) clause.Expression {
	return f.operateValue(value, "NOT LIKE", utils.Optional(escape, 0))
}

func (f patternExprImpl[T]) NotLikeOpt(value mo.Option[T], escape ...byte) clause.Expression {
	if value.IsAbsent() {
		return types.EmptyExpression
	}
	return f.NotLike(value.MustGet(), escape...)
}

func (f patternExprImpl[T]) Contains(value string) clause.Expression {
	expr := clause.Expr{SQL: "?", Vars: []any{"%" + value + "%"}}
	return f.operateValue(expr, "LIKE", 0)
}

func (f patternExprImpl[T]) ContainsOpt(value mo.Option[string]) clause.Expression {
	if value.IsAbsent() {
		return types.EmptyExpression
	}
	return f.Contains(value.MustGet())
}

func (f patternExprImpl[T]) HasPrefix(value string) clause.Expression {
	expr := clause.Expr{SQL: "?", Vars: []any{value + "%"}}
	return f.operateValue(expr, "LIKE", 0)
}

func (f patternExprImpl[T]) HasPrefixOpt(value mo.Option[string]) clause.Expression {
	if value.IsAbsent() {
		return types.EmptyExpression
	}
	return f.HasPrefix(value.MustGet())
}

func (f patternExprImpl[T]) HasSuffix(value string) clause.Expression {
	expr := clause.Expr{SQL: "?", Vars: []any{"%" + value}}
	return f.operateValue(expr, "LIKE", 0)
}

func (f patternExprImpl[T]) HasSuffixOpt(value mo.Option[string]) clause.Expression {
	if value.IsAbsent() {
		return types.EmptyExpression
	}
	return f.HasSuffix(value.MustGet())
}

func (f patternExprImpl[T]) operateValue(value any, operator string, escape byte) clause.Expression {
	var expr clause.Expression = clause.Like{
		Column: f.Expression,
		Value: clauses2.EscapeClause{
			Value:  value,
			Escape: escape,
		},
	}
	if strings.HasPrefix(operator, "NOT") {
		expr = clause.Not(expr)
	}
	return expr
}
