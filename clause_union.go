package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/types"
	"github.com/samber/lo"
)

// UnionAll 结果集中允许有重复行
func UnionAll(builder ...*QueryBuilder) field.IToExpr {
	exprs := lo.Map(builder, func(item *QueryBuilder, index int) field.Expression {
		return item.ToExpr()
	})
	return ExprTo{Expression: unionClause{
		Exprs:    exprs,
		Distinct: false,
	}}
}

// Union 结果集中不允许有重复行(会造成性能问题), 是UNION DISTINCT的别名
func Union(builder ...*QueryBuilder) field.IToExpr {
	exprs := lo.Map(builder, func(item *QueryBuilder, index int) field.Expression {
		return item.ToExpr()
	})
	return ExprTo{Expression: unionClause{
		Exprs:    exprs,
		Distinct: true,
	}}
}

type unionClause struct {
	Exprs    []field.Expression
	Distinct bool
}

func (u unionClause) Build(builder clause.Builder) {
	writer := &types.SafeWriter{Builder: builder}

	if len(u.Exprs) == 0 {
		return
	}
	if len(u.Exprs) == 1 {
		u.Exprs[0].Build(builder)
		return
	}

	mainSQL := lo.Ternary(u.Distinct, " UNION DISTINCT ", " UNION ALL ")

	for idx, expr := range u.Exprs {
		writer.WriteByte('(')
		expr.Build(builder)
		writer.WriteByte(')')
		if idx != len(u.Exprs)-1 {
			writer.WriteString(mainSQL)
		}
	}
}
