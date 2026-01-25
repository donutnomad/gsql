package fields

import "github.com/donutnomad/gsql/clause"

func anyToExpr[T any](val T) clause.Expression {
	var expr clause.Expression
	switch v := any(val).(type) {
	case clause.Expression:
		expr = v
	case interface{ Unwrap() clause.Expression }:
		expr = v.Unwrap()
	case interface{ ToExpr() clause.Expression }:
		expr = v.ToExpr()
	default:
		expr = NewLitExpr(val)
	}
	return expr
}
