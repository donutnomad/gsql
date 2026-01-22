package field

import (
	"fmt"

	"github.com/donutnomad/gsql/clause"
	"github.com/samber/mo"
)

// ==================== 空值判断实现 ====================

// pointerExprImpl 用于表达式的空值判断实现
type pointerExprImpl struct {
	clause.Expression
}

func (f pointerExprImpl) IsNull() Expression {
	return clause.Expr{SQL: "? IS NULL", Vars: []any{f.Expression}}
}

func (f pointerExprImpl) IsNotNull() Expression {
	return clause.Expr{SQL: "? IS NOT NULL", Vars: []any{f.Expression}}
}

// ==================== 基础比较操作实现（等于/不等于/In/NotIn）====================

// baseComparableImpl 基础比较操作实现
// 适用于所有类型（包括字符串），只包含等于、不等于、In、NotIn
type baseComparableImpl[T any] struct {
	clause.Expression
}

func (f baseComparableImpl[T]) Eq(value T) Expression {
	return clause.Expr{SQL: "? = ?", Vars: []any{f.Expression, value}}
}

func (f baseComparableImpl[T]) EqF(other Expression) Expression {
	return clause.Expr{SQL: "? = ?", Vars: []any{f.Expression, other}}
}

func (f baseComparableImpl[T]) EqOpt(value mo.Option[T]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Eq(value.MustGet())
}

func (f baseComparableImpl[T]) Not(value T) Expression {
	return clause.Expr{SQL: "? != ?", Vars: []any{f.Expression, value}}
}

func (f baseComparableImpl[T]) NotF(other Expression) Expression {
	return clause.Expr{SQL: "? != ?", Vars: []any{f.Expression, other}}
}

func (f baseComparableImpl[T]) NotOpt(value mo.Option[T]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Not(value.MustGet())
}

func (f baseComparableImpl[T]) In(values ...T) Expression {
	if len(values) == 0 {
		return emptyExpression
	}
	return clause.Expr{SQL: "? IN ?", Vars: []any{f.Expression, values}}
}

func (f baseComparableImpl[T]) NotIn(values ...T) Expression {
	if len(values) == 0 {
		return emptyExpression
	}
	return clause.Expr{SQL: "? NOT IN ?", Vars: []any{f.Expression, values}}
}

// ==================== 数值比较操作的通用实现 ====================

// numericComparableImpl 数值类型的比较操作通用实现
// 适用于 IntExprT, FloatExprT, DecimalExprT
// 嵌入 baseComparableImpl 获得基础比较操作，额外添加大于、小于、Between 等操作
type numericComparableImpl[T any] struct {
	baseComparableImpl[T]
}

func (f numericComparableImpl[T]) Gt(value T) Expression {
	return clause.Expr{SQL: "? > ?", Vars: []any{f.Expression, value}}
}

func (f numericComparableImpl[T]) GtOpt(value mo.Option[T]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Gt(value.MustGet())
}

func (f numericComparableImpl[T]) GtF(other Expression) Expression {
	return clause.Expr{SQL: "? > ?", Vars: []any{f.Expression, other}}
}

func (f numericComparableImpl[T]) Gte(value T) Expression {
	return clause.Expr{SQL: "? >= ?", Vars: []any{f.Expression, value}}
}

func (f numericComparableImpl[T]) GteOpt(value mo.Option[T]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Gte(value.MustGet())
}

func (f numericComparableImpl[T]) GteF(other Expression) Expression {
	return clause.Expr{SQL: "? >= ?", Vars: []any{f.Expression, other}}
}

func (f numericComparableImpl[T]) Lt(value T) Expression {
	return clause.Expr{SQL: "? < ?", Vars: []any{f.Expression, value}}
}

func (f numericComparableImpl[T]) LtOpt(value mo.Option[T]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Lt(value.MustGet())
}

func (f numericComparableImpl[T]) LtF(other Expression) Expression {
	return clause.Expr{SQL: "? < ?", Vars: []any{f.Expression, other}}
}

func (f numericComparableImpl[T]) Lte(value T) Expression {
	return clause.Expr{SQL: "? <= ?", Vars: []any{f.Expression, value}}
}

func (f numericComparableImpl[T]) LteOpt(value mo.Option[T]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Lte(value.MustGet())
}

func (f numericComparableImpl[T]) LteF(other Expression) Expression {
	return clause.Expr{SQL: "? <= ?", Vars: []any{f.Expression, other}}
}

func (f numericComparableImpl[T]) Between(from, to T) Expression {
	return clause.Expr{SQL: "? BETWEEN ? AND ?", Vars: []any{f.Expression, from, to}}
}

func (f numericComparableImpl[T]) NotBetween(from, to T) Expression {
	return clause.Expr{SQL: "? NOT BETWEEN ? AND ?", Vars: []any{f.Expression, from, to}}
}

// ==================== 算术运算的 SQL 生成 ====================

// arithmeticSql 生成算术运算的 SQL 表达式
type arithmeticSql struct {
	clause.Expression
}

func (a arithmeticSql) addExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? + ?", Vars: []any{a.Expression, value}}
}

func (a arithmeticSql) subExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? - ?", Vars: []any{a.Expression, value}}
}

func (a arithmeticSql) mulExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? * ?", Vars: []any{a.Expression, value}}
}

func (a arithmeticSql) divExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? / ?", Vars: []any{a.Expression, value}}
}

func (a arithmeticSql) negExpr() clause.Expr {
	return clause.Expr{SQL: "-?", Vars: []any{a.Expression}}
}

func (a arithmeticSql) modExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? MOD ?", Vars: []any{a.Expression, value}}
}

// ==================== 数学函数的 SQL 生成 ====================

// mathFuncSql 生成数学函数的 SQL 表达式
type mathFuncSql struct {
	clause.Expression
}

func (m mathFuncSql) absExpr() clause.Expr {
	return clause.Expr{SQL: "ABS(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) signExpr() clause.Expr {
	return clause.Expr{SQL: "SIGN(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) ceilExpr() clause.Expr {
	return clause.Expr{SQL: "CEIL(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) floorExpr() clause.Expr {
	return clause.Expr{SQL: "FLOOR(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) roundExpr(decimals ...int) clause.Expr {
	if len(decimals) > 0 {
		return clause.Expr{SQL: "ROUND(?, ?)", Vars: []any{m.Expression, decimals[0]}}
	}
	return clause.Expr{SQL: "ROUND(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) truncateExpr(decimals int) clause.Expr {
	return clause.Expr{SQL: "TRUNCATE(?, ?)", Vars: []any{m.Expression, decimals}}
}

func (m mathFuncSql) powExpr(exponent float64) clause.Expr {
	return clause.Expr{SQL: "POW(?, ?)", Vars: []any{m.Expression, exponent}}
}

func (m mathFuncSql) sqrtExpr() clause.Expr {
	return clause.Expr{SQL: "SQRT(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) logExpr() clause.Expr {
	return clause.Expr{SQL: "LOG(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) log10Expr() clause.Expr {
	return clause.Expr{SQL: "LOG10(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) log2Expr() clause.Expr {
	return clause.Expr{SQL: "LOG2(?)", Vars: []any{m.Expression}}
}

func (m mathFuncSql) expExpr() clause.Expr {
	return clause.Expr{SQL: "EXP(?)", Vars: []any{m.Expression}}
}

// ==================== 条件函数的 SQL 生成 ====================

// condFuncSql 生成条件函数的 SQL 表达式
type condFuncSql struct {
	clause.Expression
}

func (c condFuncSql) ifNullExpr(defaultValue any) clause.Expr {
	return clause.Expr{SQL: "IFNULL(?, ?)", Vars: []any{c.Expression, defaultValue}}
}

func (c condFuncSql) coalesceExpr(values ...any) clause.Expr {
	allArgs := []any{c.Expression}
	placeholders := "?"
	for _, v := range values {
		placeholders += ", ?"
		allArgs = append(allArgs, v)
	}
	return clause.Expr{SQL: "COALESCE(" + placeholders + ")", Vars: allArgs}
}

func (c condFuncSql) nullifExpr(value any) clause.Expr {
	return clause.Expr{SQL: "NULLIF(?, ?)", Vars: []any{c.Expression, value}}
}

func (c condFuncSql) greatestExpr(values ...any) clause.Expr {
	allArgs := []any{c.Expression}
	placeholders := "?"
	for _, v := range values {
		placeholders += ", ?"
		allArgs = append(allArgs, v)
	}
	return clause.Expr{SQL: "GREATEST(" + placeholders + ")", Vars: allArgs}
}

func (c condFuncSql) leastExpr(values ...any) clause.Expr {
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

func (t trigFuncSql) sinExpr() clause.Expr {
	return clause.Expr{SQL: "SIN(?)", Vars: []any{t.Expression}}
}

func (t trigFuncSql) cosExpr() clause.Expr {
	return clause.Expr{SQL: "COS(?)", Vars: []any{t.Expression}}
}

func (t trigFuncSql) tanExpr() clause.Expr {
	return clause.Expr{SQL: "TAN(?)", Vars: []any{t.Expression}}
}

func (t trigFuncSql) asinExpr() clause.Expr {
	return clause.Expr{SQL: "ASIN(?)", Vars: []any{t.Expression}}
}

func (t trigFuncSql) acosExpr() clause.Expr {
	return clause.Expr{SQL: "ACOS(?)", Vars: []any{t.Expression}}
}

func (t trigFuncSql) atanExpr() clause.Expr {
	return clause.Expr{SQL: "ATAN(?)", Vars: []any{t.Expression}}
}

func (t trigFuncSql) radiansExpr() clause.Expr {
	return clause.Expr{SQL: "RADIANS(?)", Vars: []any{t.Expression}}
}

func (t trigFuncSql) degreesExpr() clause.Expr {
	return clause.Expr{SQL: "DEGREES(?)", Vars: []any{t.Expression}}
}

// ==================== 位运算的 SQL 生成 ====================

// bitOpSql 生成位运算的 SQL 表达式
type bitOpSql struct {
	clause.Expression
}

func (b bitOpSql) bitAndExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? & ?", Vars: []any{b.Expression, value}}
}

func (b bitOpSql) bitOrExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? | ?", Vars: []any{b.Expression, value}}
}

func (b bitOpSql) bitXorExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? ^ ?", Vars: []any{b.Expression, value}}
}

func (b bitOpSql) bitNotExpr() clause.Expr {
	return clause.Expr{SQL: "~?", Vars: []any{b.Expression}}
}

func (b bitOpSql) leftShiftExpr(n int) clause.Expr {
	return clause.Expr{SQL: "? << ?", Vars: []any{b.Expression, n}}
}

func (b bitOpSql) rightShiftExpr(n int) clause.Expr {
	return clause.Expr{SQL: "? >> ?", Vars: []any{b.Expression, n}}
}

func (b bitOpSql) intDivExpr(value any) clause.Expr {
	return clause.Expr{SQL: "? DIV ?", Vars: []any{b.Expression, value}}
}

// ==================== 模式匹配的 SQL 生成 ====================

// patternExprImpl 用于表达式的模式匹配实现
// 适用于 TextExpr
type patternExprImpl[T any] struct {
	clause.Expression
}

func (f patternExprImpl[T]) Like(value T, escape ...byte) Expression {
	if len(escape) > 0 {
		return clause.Expr{
			SQL:  "? LIKE ? ESCAPE ?",
			Vars: []any{f.Expression, value, string(escape[0])},
		}
	}
	return clause.Expr{SQL: "? LIKE ?", Vars: []any{f.Expression, value}}
}

func (f patternExprImpl[T]) LikeOpt(value mo.Option[T], escape ...byte) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Like(value.MustGet(), escape...)
}

func (f patternExprImpl[T]) NotLike(value T, escape ...byte) Expression {
	if len(escape) > 0 {
		return clause.Expr{
			SQL:  "? NOT LIKE ? ESCAPE ?",
			Vars: []any{f.Expression, value, string(escape[0])},
		}
	}
	return clause.Expr{SQL: "? NOT LIKE ?", Vars: []any{f.Expression, value}}
}

func (f patternExprImpl[T]) NotLikeOpt(value mo.Option[T], escape ...byte) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.NotLike(value.MustGet(), escape...)
}

func (f patternExprImpl[T]) Contains(value string) Expression {
	return clause.Expr{SQL: "? LIKE ?", Vars: []any{f.Expression, "%" + value + "%"}}
}

func (f patternExprImpl[T]) ContainsOpt(value mo.Option[string]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Contains(value.MustGet())
}

func (f patternExprImpl[T]) HasPrefix(value string) Expression {
	return clause.Expr{SQL: "? LIKE ?", Vars: []any{f.Expression, value + "%"}}
}

func (f patternExprImpl[T]) HasPrefixOpt(value mo.Option[string]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.HasPrefix(value.MustGet())
}

func (f patternExprImpl[T]) HasSuffix(value string) Expression {
	return clause.Expr{SQL: "? LIKE ?", Vars: []any{f.Expression, "%" + value}}
}

func (f patternExprImpl[T]) HasSuffixOpt(value mo.Option[string]) Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.HasSuffix(value.MustGet())
}
