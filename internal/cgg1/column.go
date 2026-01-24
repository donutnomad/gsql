package cgg1

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/utils"
)

type ColumnClause struct {
	clause.Column
	Expr clause.Expression
}

func (v ColumnClause) Build(builder clause.Builder) {
	writer := builder
	write := func(raw bool, str string) {
		if raw {
			writer.WriteString(str)
		} else {
			writer.WriteQuoted(str)
		}
	}

	if v.Expr != nil {
		utils.WriteExpr(builder, v.Expr)
		if v.Alias != "" {
			writer.WriteString(" AS ")
			write(false, v.Alias)
		}
	} else {
		writer.WriteQuoted(v.Column)
	}
}
