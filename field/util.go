package field

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/clauses"
	"github.com/samber/lo"
)

var emptyExpression = clause.Expr{}

// EmptyExpression 空表达式，用于跳过条件
var EmptyExpression = emptyExpression

func NewColumnClause(f Base) clauses.ColumnClause {
	if f.sql != nil {
		return clauses.ColumnClause{
			Column: clause.Column{
				Alias: f.alias,
				Raw:   true,
			},
			Expr: f.sql,
		}
	}
	return clauses.ColumnClause{
		Column: clause.Column{
			Table: f.tableName,
			Name:  f.columnName,
			Alias: f.alias,
			Raw:   f.columnName == "*",
		},
	}
}

func notNilExpr(input ...clause.Expression) []clause.Expression {
	return lo.Filter(input, func(item clause.Expression, index int) bool {
		return item != nil
	})
}
