package clause

import "gorm.io/gorm/clause"

// Expression expression interface
type Expression = clause.Expression

// NegationExpressionBuilder negation expression builder
type NegationExpressionBuilder = clause.NegationExpressionBuilder
type Writer = clause.Writer

// Builder builder interface
type Builder = clause.Builder
type Expr = clause.Expr
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
type OrderBy = clause.OrderBy
type OrderByColumn = clause.OrderByColumn
type GroupBy = clause.GroupBy
type Locking = clause.Locking
type Eq = clause.Eq
type Neq = clause.Neq
type Gt = clause.Gt
type Gte = clause.Gte
type Lt = clause.Lt
type Lte = clause.Lte
type Like = clause.Like
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
