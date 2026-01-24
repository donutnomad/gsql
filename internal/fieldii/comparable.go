package fieldii

import (
	"fmt"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/clauses"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type ComparableImpl[T any] = comparableImpl[T]

// =, !=, IN, NOT IN, >, >=, <, <=
type comparableImpl[T any] struct {
	fieldi.IField
}

func (f comparableImpl[T]) Eq(value T) clause.Expression {
	return f.operateValue(value, "=")
}

func (f comparableImpl[T]) EqOpt(value mo.Option[T]) clause.Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Eq(value.MustGet())
}

func (f comparableImpl[T]) EqF(other fieldi.IField) clause.Expression {
	return f.operateField(other, "=")
}

func (f comparableImpl[T]) Not(value T) clause.Expression {
	return f.operateValue(value, "!=")
}

func (f comparableImpl[T]) NotOpt(value mo.Option[T]) clause.Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Not(value.MustGet())
}

func (f comparableImpl[T]) NotF(other fieldi.IField) clause.Expression {
	return f.operateField(other, "!=")
}

func (f comparableImpl[T]) In(values ...T) clause.Expression {
	if len(values) == 0 {
		return emptyExpression
	}
	return f.operateValue(lo.ToAnySlice(values), "IN")
}

func (f comparableImpl[T]) NotIn(values ...T) clause.Expression {
	if len(values) == 0 {
		return emptyExpression
	}
	return f.operateValue(lo.ToAnySlice(values), "NOT IN")
}

func (f comparableImpl[T]) Gt(value T) clause.Expression {
	return f.operateValue(value, ">")
}

func (f comparableImpl[T]) GtOpt(value mo.Option[T]) clause.Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Gt(value.MustGet())
}

func (f comparableImpl[T]) GtF(other fieldi.IField) clause.Expression {
	return f.operateField(other, ">")
}

func (f comparableImpl[T]) Gte(value T) clause.Expression {
	return f.operateValue(value, ">=")
}

func (f comparableImpl[T]) GteOpt(value mo.Option[T]) clause.Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Gte(value.MustGet())
}

func (f comparableImpl[T]) GteF(other fieldi.IField) clause.Expression {
	return f.operateField(other, ">=")
}

func (f comparableImpl[T]) Lt(value T) clause.Expression {
	return f.operateValue(value, "<")
}

func (f comparableImpl[T]) LtOpt(value mo.Option[T]) clause.Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Lt(value.MustGet())
}

func (f comparableImpl[T]) LtF(other fieldi.IField) clause.Expression {
	return f.operateField(other, "<")
}

func (f comparableImpl[T]) Lte(value T) clause.Expression {
	return f.operateValue(value, "<=")
}

func (f comparableImpl[T]) LteOpt(value mo.Option[T]) clause.Expression {
	if value.IsAbsent() {
		return emptyExpression
	}
	return f.Lte(value.MustGet())
}

func (f comparableImpl[T]) LteF(other fieldi.IField) clause.Expression {
	return f.operateField(other, "<=")
}

// Between
// opFrom: >=,>,=,<=,<, default: >=
// opTo: >=,>,=,<=,<, default: <
func (f comparableImpl[T]) Between(from, to *T, op ...string) clause.Expression {
	return f.BetweenOpt(fieldi.Range[T]{
		From: mo.PointerToOption(from),
		To:   mo.PointerToOption(to),
	}, op...)
}

// BetweenF
// opFrom: >=,>,=,<=,<, default: >=
// opTo: >=,>,=,<=,<, default: <
func (f comparableImpl[T]) BetweenF(from, to fieldi.IField, op ...string) clause.Expression {
	var opFrom = ">="
	var opTo = "<"
	if len(op) > 0 {
		opFrom = op[0]
	}
	if len(op) > 1 {
		opTo = op[1]
	}
	var opFunc = func(op string, value fieldi.IField) clause.Expression {
		if value == nil {
			return nil
		}
		switch op {
		case ">=":
			return f.GteF(value)
		case ">":
			return f.GtF(value)
		case "=":
			return f.EqF(value)
		case "<=":
			return f.LteF(value)
		case "<":
			return f.LtF(value)
		default:
			panic(fmt.Sprintf("invalid operation %q", op))
		}
	}
	return clause.And(notNilExpr(opFunc(opFrom, from), opFunc(opTo, to))...)
}

// BetweenOpt
// opFrom: >=,>,=,<=,<, default: >=
// opTo: >=,>,=,<=,<, default: <
func (f comparableImpl[T]) BetweenOpt(rng interface {
	FromValue() *T
	ToValue() *T
}, op ...string) clause.Expression {
	var opFrom = ">="
	var opTo = "<"
	if len(op) > 0 {
		opFrom = op[0]
	}
	if len(op) > 1 {
		opTo = op[1]
	}
	var opFunc = func(op string, val *T) clause.Expression {
		if val == nil {
			return nil
		}
		value := mo.Some(*val)
		switch op {
		case ">=":
			return f.GteOpt(value)
		case ">":
			return f.GtOpt(value)
		case "=":
			return f.EqOpt(value)
		case "<=":
			return f.LteOpt(value)
		case "<":
			return f.LtOpt(value)
		default:
			panic(fmt.Sprintf("invalid operation %q", op))
		}
	}
	return clause.And(notNilExpr(opFunc(opFrom, rng.FromValue()), opFunc(opTo, rng.ToValue()))...)
}

func (f comparableImpl[T]) operateField(other fieldi.IField, operator string) clause.Expression {
	return f.operateValue(other.ToExpr(), operator)
}

func (f comparableImpl[T]) operateValue(value any, operator string) clause.Expression {
	return f.operateValue2(f.ToColumn(), value, operator)
}

func (f comparableImpl[T]) operateValue2(column any, value any, operator string) clause.Expression {
	var expr clause.Expression
	switch operator {
	case "=":
		expr = clause.Eq{Column: column, Value: value}
	case "!=":
		expr = clause.Neq{Column: column, Value: value}
	case ">":
		expr = clause.Gt{Column: column, Value: value}
	case ">=":
		expr = clause.Gte{Column: column, Value: value}
	case "<":
		expr = clause.Lt{Column: column, Value: value}
	case "<=":
		expr = clause.Lte{Column: column, Value: value}
	case "IN":
		expr = clauses.IN{Column: column, Values: []any{value}}
	case "NOT IN":
		expr = clause.Not(clauses.IN{Column: column, Values: []any{value}})
	default:
		panic(fmt.Sprintf("invalid operator %s", operator))
	}
	return expr
}
