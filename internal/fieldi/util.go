package fieldi

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/cgg1"
)

var emptyExpression = clause.Expr{}

// EmptyExpression 空表达式，用于跳过条件
var EmptyExpression = emptyExpression

func NewColumnClause(f Base) cgg1.ColumnClause {
	if f.sql != nil {
		return cgg1.ColumnClause{
			Column: clause.Column{
				Alias: f.alias,
				Raw:   true,
			},
			Expr: f.sql,
		}
	}
	return cgg1.ColumnClause{
		Column: clause.Column{
			Table: f.tableName,
			Name:  f.columnName,
			Alias: f.alias,
			Raw:   f.columnName == "*",
		},
	}
}
