package scopes

import (
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/donutnomad/gsql/internal/types"
	"github.com/samber/lo"
)

type Range[T any] = field.Range[T]
type TimeRange = field.Range[time.Time]
type TimestampRange = field.Range[int64]
type SortNameMapping map[string]field.IField
type SortOrder struct {
	Name string
	Asc  bool
}

func (m SortNameMapping) Map(orders []SortOrder, defaultOrder ...SortOrder) []gsql.FieldOrder {
	if len(orders) == 0 {
		orders = defaultOrder
	}
	if len(orders) == 0 {
		return []types.OrderItem{}
	}
	var ret []gsql.FieldOrder
	for _, item := range orders {
		if v, ok := m[item.Name]; ok {
			ret = append(ret, types.NewOrder(v, item.Asc))
		}
	}
	return ret
}

func OrderBy(name string, asc bool) SortOrder {
	return SortOrder{
		Name: name,
		Asc:  asc,
	}
}

// TimeBetween
// opFrom: >=,>,=,<=,<, default: >=
// opTo: >=,>,=,<=,<, default: <
func TimeBetween[F any, ValExpr fields.Expressions[F], Value TimestampRange | TimeRange](
	fieldComparable ValExpr, value Value, op ...string,
) gsql.ScopeFunc {
	var opFrom = ">="
	var opTo = "<"
	if len(op) > 0 {
		opFrom = op[0]
	}
	if len(op) > 1 {
		opTo = op[1]
	}

	var fieldIsTimeStruct = false
	switch any(lo.Empty[F]()).(type) {
	case time.Time:
		fieldIsTimeStruct = true
	case *time.Time:
		fieldIsTimeStruct = true
	}
	var opFunc = func(op string, value *int64) clause.Expression {
		if value == nil {
			return clause.Expr{}
		}
		var left = fieldComparable
		var right clause.Expression = gsql.Lit(*value)
		if fieldIsTimeStruct {
			right = gsql.Int(right).ToDateTime()
		}
		return gsql.Expr("? "+op+" ?", left, right)
	}
	var opFunc2 = func(op string, value *time.Time) clause.Expression {
		if value == nil {
			return clause.Expr{}
		}
		var left any = fieldComparable
		var right = value
		if !fieldIsTimeStruct {
			left = gsql.IntVal(left).ToDateTime()
		}
		return gsql.Expr("? "+op+" ?", left, right)
	}
	return func(b *gsql.Builder) {
		switch v := any(value).(type) {
		case TimestampRange:
			b.Where(opFunc(opFrom, v.From.ToPointer()), opFunc(opTo, v.To.ToPointer()))
		case TimeRange:
			b.Where(opFunc2(opFrom, v.From.ToPointer()), opFunc2(opTo, v.To.ToPointer()))
		}
	}
}

func List[Model any](db gsql.IDB, query *gsql.QueryBuilderG[Model], paginate gsql.Paginate, scopes ...gsql.ScopeFuncG[Model]) ([]*Model, int64, error) {
	total, err := query.Count(db)
	if err != nil {
		return nil, 0, err
	}
	pos, err := query.Paginate(paginate).ScopeG(scopes...).Find(db)
	if err != nil {
		return nil, 0, err
	}
	return pos, total, nil
}

func ListMap[Model any, OUT any](db gsql.IDB, query *gsql.QueryBuilderG[Model], paginate gsql.Paginate, mapper func([]*Model) []*OUT, scopes ...gsql.ScopeFunc) ([]*OUT, int64, error) {
	total, err := query.Count(db)
	if err != nil {
		return nil, 0, err
	}
	pos, err := query.Paginate(paginate).Scope(scopes...).Find(db)
	if err != nil {
		return nil, 0, err
	}
	return mapper(pos), total, nil
}

func ListAndMap[Model any, OUT any](db gsql.IDB, query *gsql.QueryBuilderG[Model], mapper func([]*Model) []*OUT, scopes ...gsql.ScopeFunc) ([]*OUT, int64, error) {
	total, err := query.Count(db)
	if err != nil {
		return nil, 0, err
	}
	pos, err := query.Scope(scopes...).Find(db)
	if err != nil {
		return nil, 0, err
	}
	return mapper(pos), total, nil
}
