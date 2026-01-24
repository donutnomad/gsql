package types

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/utils"
)

var EmptyExpression = clause.Expr{}

type SQLVars interface {
	SQLVars() (string, []any)
}

type SQLChecker interface {
	IsEmptySQL() bool
}

var _ field.ExpressionTo = (*LitExpr)(nil)

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

type ExprTo struct {
	clause.Expression
}

func (e ExprTo) AsF(name ...string) field.IField {
	return FieldExpr(e.Expression, utils.Optional(name, ""))
}

func (e ExprTo) ToExpr() field.Expression {
	return e.Expression
}

func FieldExpr(expr field.Expression, alias string) field.IField {
	return field.NewBaseFromSql(expr, alias)
}

func Expr(sql string, args ...any) clause.Expression {
	return clause.Expr{SQL: sql, Vars: args}
}
