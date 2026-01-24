package types

import (
	"github.com/donutnomad/gsql/clause"
)

var EmptyExpression = clause.Expr{}

type SQLVars interface {
	SQLVars() (string, []any)
}

type SQLChecker interface {
	IsEmptySQL() bool
}
