package fieldii

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/samber/lo"
)

type Base = fieldi.Base
type IField = fieldi.IField

var emptyExpression = clause.Expr{}

func notNilExpr(input ...clause.Expression) []clause.Expression {
	return lo.Filter(input, func(item clause.Expression, index int) bool {
		return item != nil
	})
}

func ifieldToBase(field IField) Base {
	var base Base
	if v, ok := field.(Base); ok {
		base = v
	} else if v, ok := field.(*Base); ok {
		base = *v
	} else {
		base = *fieldi.NewBaseFromSql(field.ToExpr(), "")
	}
	return base
}
