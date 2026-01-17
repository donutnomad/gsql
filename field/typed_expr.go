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

// numericExprBase 数值表达式的公共实现
type numericExprBase struct {
	clause.Expression
}

// Build 实现 clause.Expression 接口
func (e numericExprBase) Build(builder clause.Builder) {
	e.Expression.Build(builder)
}

// ToExpr 转换为 Expression
func (e numericExprBase) ToExpr() Expression {
	return e.Expression
}

// AsF 将表达式转换为带别名的字段
func (e numericExprBase) AsF(name ...string) IField {
	alias := ""
	if len(name) > 0 {
		alias = name[0]
	}
	return NewBaseFromSql(e.Expression, alias)
}

// Gt 大于 (>)
func (e numericExprBase) Gt(value any) Expression {
	return clause.Expr{
		SQL:  "? > ?",
		Vars: []any{e.Expression, value},
	}
}

// Gte 大于等于 (>=)
func (e numericExprBase) Gte(value any) Expression {
	return clause.Expr{
		SQL:  "? >= ?",
		Vars: []any{e.Expression, value},
	}
}

// Lt 小于 (<)
func (e numericExprBase) Lt(value any) Expression {
	return clause.Expr{
		SQL:  "? < ?",
		Vars: []any{e.Expression, value},
	}
}

// Lte 小于等于 (<=)
func (e numericExprBase) Lte(value any) Expression {
	return clause.Expr{
		SQL:  "? <= ?",
		Vars: []any{e.Expression, value},
	}
}

// Eq 等于 (=)
func (e numericExprBase) Eq(value any) Expression {
	return clause.Expr{
		SQL:  "? = ?",
		Vars: []any{e.Expression, value},
	}
}

// Not 不等于 (!=)
func (e numericExprBase) Not(value any) Expression {
	return clause.Expr{
		SQL:  "? != ?",
		Vars: []any{e.Expression, value},
	}
}

// Between 在范围内 (BETWEEN ... AND ...)
func (e numericExprBase) Between(from, to any) Expression {
	return clause.Expr{
		SQL:  "? BETWEEN ? AND ?",
		Vars: []any{e.Expression, from, to},
	}
}

// NotBetween 不在范围内 (NOT BETWEEN ... AND ...)
func (e numericExprBase) NotBetween(from, to any) Expression {
	return clause.Expr{
		SQL:  "? NOT BETWEEN ? AND ?",
		Vars: []any{e.Expression, from, to},
	}
}

// In 在列表中 (IN (...))
func (e numericExprBase) In(values ...any) Expression {
	return clause.Expr{
		SQL:  "? IN ?",
		Vars: []any{e.Expression, values},
	}
}

// NotIn 不在列表中 (NOT IN (...))
func (e numericExprBase) NotIn(values ...any) Expression {
	return clause.Expr{
		SQL:  "? NOT IN ?",
		Vars: []any{e.Expression, values},
	}
}

// IntExpr 整数类型表达式，用于 COUNT 等返回整数的聚合函数
type IntExpr struct {
	numericExprBase
}

// NewIntExpr 创建一个新的 IntExpr 实例
func NewIntExpr(expr clause.Expression) IntExpr {
	return IntExpr{numericExprBase{Expression: expr}}
}

// FloatExpr 浮点类型表达式，用于 AVG, SUM 等返回浮点数的聚合函数
type FloatExpr struct {
	numericExprBase
}

// NewFloatExpr 创建一个新的 FloatExpr 实例
func NewFloatExpr(expr clause.Expression) FloatExpr {
	return FloatExpr{numericExprBase{Expression: expr}}
}
