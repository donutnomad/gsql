package clauses

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/utils"
)

// IN Whether a value is within a set of values
type IN struct {
	Column any
	Values []any
}

func (in IN) Build(builder clause.Builder) {
	writeString := func(str string) {
		_, _ = builder.WriteString(str)
	}
	builder.WriteQuoted(in.Column)

	switch len(in.Values) {
	case 0:
		writeString(" IN (NULL)")
	case 1:
		if _, ok := in.Values[0].([]any); !ok {
			writeString(" = ")
			builder.AddVar(builder, in.Values[0])
			break
		}

		fallthrough
	default:
		writeString(" IN ")
		utils.AddVarAutoBracket(builder, in.Values)
	}
}

func (in IN) NegationBuild(builder clause.Builder) {
	writeString := func(str string) {
		_, _ = builder.WriteString(str)
	}

	builder.WriteQuoted(in.Column)
	switch len(in.Values) {
	case 0:
		writeString(" IS NOT NULL")
	case 1:
		if _, ok := in.Values[0].([]any); !ok {
			writeString(" <> ")
			builder.AddVar(builder, in.Values[0])
			break
		}

		fallthrough
	default:
		writeString(" NOT IN ")
		utils.AddVarAutoBracket(builder, in.Values)
	}
}
