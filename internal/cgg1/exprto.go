package cgg1

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/fields"
)

type ExprTo struct {
	clause.Expression
}

func (e ExprTo) AsF(name ...string) fieldi.IField {
	var n = "_default"
	if len(name) > 0 {
		n = name[0]
	}
	return fields.ScalarOf[any](e.Expression).As(n)
}

func (e ExprTo) ToExpr() clause.Expression {
	return e.Expression
}
