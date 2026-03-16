package fields

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
)

//go:generate go run -tags gen 0gen.go
//go:generate go run -tags gen 0gen1.go

type Expressions[T any] interface {
	IntField[T] | IntExpr[T] |
		FloatExpr[T] | FloatField[T] |
		StringExpr[T] | StringField[T] |
		DecimalExpr[T] | DecimalField[T] |
		TimeExpr[T] | TimeField[T] |
		DateTimeExpr[T] | DateTimeField[T] |
		DateExpr[T] | DateField[T]
}

type FunctionName string

func NewLitExpr[T any](value T) *LitExpr {
	return &LitExpr{
		Expression: clause.Expr{SQL: "?", Vars: []any{value}},
	}
}

type LitExpr struct {
	clause.Expression
}

var _ fieldi.IField = (*LitExpr)(nil)

func (e *LitExpr) As(alias string) fieldi.IField {
	return ScalarOf[any](e.Expression).As(alias)
}

func (e *LitExpr) ToExpr() clause.Expression {
	return e.Expression
}

func (e *LitExpr) FullName() string {
	return ""
}

func (e *LitExpr) Name() string {
	return ""
}

func (e *LitExpr) Alias() string {
	return ""
}
