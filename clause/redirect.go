package clause

import (
	"database/sql"
	"database/sql/driver"
	"reflect"

	"gorm.io/gorm/clause"
)

// Expression expression interface
type Expression = clause.Expression

// NegationExpressionBuilder negation expression builder
type NegationExpressionBuilder = clause.NegationExpressionBuilder
type Writer = clause.Writer

// Builder builder interface
type Builder = clause.Builder

type RawExpr = clause.Expr

//type Expr = clause.Expr

// Expr raw expression
type Expr struct {
	SQL                string
	Vars               []any
	WithoutParentheses bool
}

func (expr Expr) Compat() clause.Expr {
	return clause.Expr{
		SQL:                expr.SQL,
		Vars:               expr.Vars,
		WithoutParentheses: expr.WithoutParentheses,
	}
}

func (expr Expr) SQLVars() (string, []any) {
	return expr.SQL, expr.Vars
}

func (expr Expr) IsEmptySQL() bool {
	return len(expr.SQL) == 0
}

// Build build raw expression
func (expr Expr) Build(builder Builder) {
	var (
		afterParenthesis bool
		idx              int
	)
	if len(expr.Vars) == 1 {
		val := expr.Vars[0]
		v0, ok := val.(interface{ ToExpr() clause.Expression })
		if ok {
			val = v0.ToExpr()
		}
		if v, ok := val.(Expr); ok {
			if v.SQL == "?" {
				expr.Vars = v.Vars
			}
		}
	}

	for _, v := range []byte(expr.SQL) {
		if v == '?' && len(expr.Vars) > idx {
			if afterParenthesis || expr.WithoutParentheses {
				if _, ok := expr.Vars[idx].(driver.Valuer); ok {
					builder.AddVar(builder, expr.Vars[idx])
				} else {
					switch rv := reflect.ValueOf(expr.Vars[idx]); rv.Kind() {
					case reflect.Slice, reflect.Array:
						if rv.Len() == 0 {
							builder.AddVar(builder, nil)
						} else {
							for i := 0; i < rv.Len(); i++ {
								if i > 0 {
									builder.WriteByte(',')
								}
								builder.AddVar(builder, rv.Index(i).Interface())
							}
						}
					default:
						builder.AddVar(builder, expr.Vars[idx])
					}
				}
			} else {
				builder.AddVar(builder, expr.Vars[idx])
			}

			idx++
		} else {
			if v == '(' {
				afterParenthesis = true
			} else {
				afterParenthesis = false
			}
			builder.WriteByte(v)
		}
	}

	if idx < len(expr.Vars) {
		for _, v := range expr.Vars[idx:] {
			builder.AddVar(builder, sql.NamedArg{Value: v})
		}
	}
}

type OrConditions = clause.OrConditions
type AndConditions = clause.AndConditions
type NotConditions = clause.NotConditions
type IN = clause.IN
type Column = clause.Column
type Clause = clause.Clause
type Insert = clause.Insert
type OnConflict = clause.OnConflict
type Set = clause.Set
type Select = clause.Select
type Interface = clause.Interface
type CommaExpression = clause.CommaExpression
type Table = clause.Table
type ClauseBuilder = clause.ClauseBuilder
type From = clause.From
type Where = clause.Where
type Join = clause.Join
type Limit = clause.Limit
type Locking = clause.Locking
type NamedExpr = clause.NamedExpr
type JoinType = clause.JoinType

var LockingOptionsNoWait = clause.LockingOptionsNoWait
var LockingStrengthUpdate = clause.LockingStrengthUpdate
var LockingStrengthShare = clause.LockingStrengthShare
var LockingOptionsSkipLocked = clause.LockingOptionsSkipLocked

var CurrentTable = clause.CurrentTable
var PrimaryKey = clause.PrimaryKey

func Not(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}
	if len(exprs) == 1 {
		if andCondition, ok := exprs[0].(AndConditions); ok {
			exprs = andCondition.Exprs
		}
	}
	return NotConditions{Exprs: exprs}
}

func And(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}

	if len(exprs) == 1 {
		if _, ok := exprs[0].(OrConditions); !ok {
			return exprs[0]
		}
	}

	return AndConditions{Exprs: exprs}
}

func Or(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}
	return OrConditions{Exprs: exprs}
}
