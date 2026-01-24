package cgg2

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/utils"
)

type ExprTo struct {
	clause.Expression
}

func (e ExprTo) AsF(name ...string) fieldi.IField {
	return FieldExpr(e.Expression, utils.Optional(name, ""))
}

func (e ExprTo) ToExpr() clause.Expression {
	return e.Expression
}

func FieldExpr(expr clause.Expression, alias string) fieldi.IField {
	return fieldi.NewBaseFromSql(expr, alias)
}

var _ fieldi.ExpressionTo = (*LitExpr)(nil)

type LitExpr struct {
	ExprTo
}

func NewLitExpr[T any](value T) *LitExpr {
	return &LitExpr{
		ExprTo: ExprTo{
			Expression: Expr("?", value),
		},
	}
}

func Expr(sql string, args ...any) clause.Expression {
	return clause.Expr{SQL: sql, Vars: args}
}
