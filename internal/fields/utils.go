package fields

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/samber/lo"
)

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

func CastExpr[Expr interface{ ExprType() R }, R any](v clause.Expression) Expr {
	switch any(lo.Empty[Expr]()).(type) {
	case StringExpr[R]:
		return any(StringOf[R](v)).(Expr)
	case IntExpr[R]:
		return any(IntOf[R](v)).(Expr)
	case FloatExpr[R]:
		return any(FloatOf[R](v)).(Expr)
	case DecimalExpr[R]:
		return any(DecimalOf[R](v)).(Expr)
	case DateTimeExpr[R]:
		return any(DateTimeOf[R](v)).(Expr)
	case DateExpr[R]:
		return any(DateOf[R](v)).(Expr)
	case TimeExpr[R]:
		return any(TimeOf[R](v)).(Expr)
	case YearExpr[R]:
		return any(YearOf[R](v)).(Expr)
	case ScalarExpr[R]:
		return any(ScalarOf[R](v)).(Expr)
	default:
		panic("CastExpr: unsupported Expr type")
	}
}
