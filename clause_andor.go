package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/types"
)

var empty = clause.Expr{}

func And(exprs ...Expression) Condition {
	exprs = filterExpr(exprs...)
	if len(exprs) == 0 {
		return Condition{Expression: empty}
	}
	if len(exprs) == 1 {
		if _, ok := exprs[0].(clause.OrConditions); !ok {
			return Condition{Expression: exprs[0]}
		}
	}

	var and clause.AndConditions
	for _, expr := range exprs {
		if v, ok := expr.(clause.AndConditions); ok {
			and.Exprs = append(and.Exprs, v.Exprs...)
		} else {
			and.Exprs = append(and.Exprs, expr)
		}
	}

	return Condition{Expression: and}
}

func Or(exprs ...Expression) Condition {
	exprs = filterExpr(exprs...)
	if len(exprs) == 0 {
		return Condition{Expression: empty}
	}
	var or clause.OrConditions
	for _, expr := range exprs {
		if v, ok := expr.(clause.OrConditions); ok {
			or.Exprs = append(or.Exprs, v.Exprs...)
		} else {
			or.Exprs = append(or.Exprs, expr)
		}
	}
	return Condition{Expression: or}
}

func filterExpr(input ...Expression) []Expression {
	var output = make([]Expression, 0, len(input))
	for _, expr := range input {
		if v, ok := expr.(types.SQLChecker); ok && v.IsEmptySQL() {
			continue
		}
		output = append(output, expr)
	}
	return output
}
