package clauses2

import (
	"github.com/donutnomad/gsql/clause"
)

type Exprs []clause.Expression

func (e Exprs) Unwraps() []clause.Expression {
	return e
}

func (e Exprs) Build(builder clause.Builder) {
	for _, item := range e {
		item.Build(builder)
	}
}
