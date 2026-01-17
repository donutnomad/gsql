package field

import (
	"github.com/donutnomad/gsql/clause"
)

// NumericExpr 数值表达式接口，支持比较操作
// 用于聚合函数（如 COUNT, SUM）返回的表达式，使其可以直接进行比较操作
// 例如: gsql.COUNT().Gt(5), gsql.SUM(amount).Gte(1000)
type NumericExpr interface {
	Expression
	ExpressionTo
	// Gt 大于 (>)
	Gt(value any) Expression
	// Gte 大于等于 (>=)
	Gte(value any) Expression
	// Lt 小于 (<)
	Lt(value any) Expression
	// Lte 小于等于 (<=)
	Lte(value any) Expression
	// Eq 等于 (=)
	Eq(value any) Expression
	// Not 不等于 (!=)
	Not(value any) Expression
	// Between 在范围内 (BETWEEN ... AND ...)
	Between(from, to any) Expression
	// NotBetween 不在范围内 (NOT BETWEEN ... AND ...)
	NotBetween(from, to any) Expression
	// In 在列表中 (IN (...))
	In(values ...any) Expression
	// NotIn 不在列表中 (NOT IN (...))
	NotIn(values ...any) Expression
}

// IntExpr 整数类型表达式，用于 COUNT 等返回整数的聚合函数
type IntExpr struct {
	clause.Expression
}

// NewIntExpr 创建一个新的 IntExpr 实例
func NewIntExpr(expr clause.Expression) IntExpr {
	return IntExpr{Expression: expr}
}

// Build 实现 clause.Expression 接口
func (e IntExpr) Build(builder clause.Builder) {
	e.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e IntExpr) ToExpr() Expression {
	return e.Expression
}

// AsF 将表达式转换为带别名的字段
func (e IntExpr) AsF(name ...string) IField {
	alias := ""
	if len(name) > 0 {
		alias = name[0]
	}
	return NewBaseFromSql(e.Expression, alias)
}

// Gt 大于 (>)
func (e IntExpr) Gt(value any) Expression {
	return clause.Expr{
		SQL:  "? > ?",
		Vars: []any{e.Expression, value},
	}
}

// Gte 大于等于 (>=)
func (e IntExpr) Gte(value any) Expression {
	return clause.Expr{
		SQL:  "? >= ?",
		Vars: []any{e.Expression, value},
	}
}

// Lt 小于 (<)
func (e IntExpr) Lt(value any) Expression {
	return clause.Expr{
		SQL:  "? < ?",
		Vars: []any{e.Expression, value},
	}
}

// Lte 小于等于 (<=)
func (e IntExpr) Lte(value any) Expression {
	return clause.Expr{
		SQL:  "? <= ?",
		Vars: []any{e.Expression, value},
	}
}

// Eq 等于 (=)
func (e IntExpr) Eq(value any) Expression {
	return clause.Expr{
		SQL:  "? = ?",
		Vars: []any{e.Expression, value},
	}
}

// Not 不等于 (!=)
func (e IntExpr) Not(value any) Expression {
	return clause.Expr{
		SQL:  "? != ?",
		Vars: []any{e.Expression, value},
	}
}

// Between 在范围内 (BETWEEN ... AND ...)
func (e IntExpr) Between(from, to any) Expression {
	return clause.Expr{
		SQL:  "? BETWEEN ? AND ?",
		Vars: []any{e.Expression, from, to},
	}
}

// NotBetween 不在范围内 (NOT BETWEEN ... AND ...)
func (e IntExpr) NotBetween(from, to any) Expression {
	return clause.Expr{
		SQL:  "? NOT BETWEEN ? AND ?",
		Vars: []any{e.Expression, from, to},
	}
}

// In 在列表中 (IN (...))
func (e IntExpr) In(values ...any) Expression {
	return clause.Expr{
		SQL:  "? IN ?",
		Vars: []any{e.Expression, values},
	}
}

// NotIn 不在列表中 (NOT IN (...))
func (e IntExpr) NotIn(values ...any) Expression {
	return clause.Expr{
		SQL:  "? NOT IN ?",
		Vars: []any{e.Expression, values},
	}
}
