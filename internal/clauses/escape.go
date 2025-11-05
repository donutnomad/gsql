package clauses

import "github.com/donutnomad/gsql/clause"

type EscapeClause struct {
	Value  string
	Escape byte
}

func (e EscapeClause) Build(builder clause.Builder) {
	builder.AddVar(builder, e.Value)
	if e.Escape != 0 {
		builder.WriteString(" ESCAPE ")
		builder.AddVar(builder, string(e.Escape))
	}
}
