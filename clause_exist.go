package gsql

import (
	"github.com/donutnomad/gsql/clause"
)

func Exists(builder *QueryBuilder) Expression {
	return existsClause{
		exists: true,
		expr:   builder.ToExpr(),
	}
}

func NotExists(builder *QueryBuilder) Expression {
	return existsClause{
		exists: false,
		expr:   builder.ToExpr(),
	}
}

type existsClause struct {
	expr   Expression
	exists bool
}

func (e existsClause) Build(builder clause.Builder) {
	if e.exists {
		builder.WriteString(" EXISTS ")
	} else {
		builder.WriteString(" NOT EXISTS ")
	}
	builder.WriteByte('(')
	e.expr.Build(builder)
	builder.WriteByte(')')
}
