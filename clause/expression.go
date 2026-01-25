package clause

import (
	"database/sql/driver"
	"reflect"
)

// Eq equal to for where
type Eq struct {
	Column Expression
	Value  any
}

func (eq Eq) Build(builder Builder) {
	eq.Column.Build(builder)

	switch eq.Value.(type) {
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []any:
		rv := reflect.ValueOf(eq.Value)
		if rv.Len() == 0 {
			builder.WriteString(" IN (NULL)")
		} else {
			builder.WriteString(" IN (")
			for i := 0; i < rv.Len(); i++ {
				if i > 0 {
					builder.WriteByte(',')
				}
				builder.AddVar(builder, rv.Index(i).Interface())
			}
			builder.WriteByte(')')
		}
	default:
		if eqNil(eq.Value) {
			builder.WriteString(" IS NULL")
		} else {
			builder.WriteString(" = ")
			builder.AddVar(builder, eq.Value)
		}
	}
}

func (eq Eq) NegationBuild(builder Builder) {
	Neq(eq).Build(builder)
}

// Neq not equal to for where
type Neq Eq

func (neq Neq) Build(builder Builder) {
	neq.Column.Build(builder)

	switch neq.Value.(type) {
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []any:
		builder.WriteString(" NOT IN (")
		rv := reflect.ValueOf(neq.Value)
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.AddVar(builder, rv.Index(i).Interface())
		}
		builder.WriteByte(')')
	default:
		if eqNil(neq.Value) {
			builder.WriteString(" IS NOT NULL")
		} else {
			builder.WriteString(" <> ")
			builder.AddVar(builder, neq.Value)
		}
	}
}

func (neq Neq) NegationBuild(builder Builder) {
	Eq(neq).Build(builder)
}

// Like whether string matches regular expression
type Like Eq

func (like Like) Build(builder Builder) {
	like.Column.Build(builder)
	builder.WriteString(" LIKE ")
	builder.AddVar(builder, like.Value)
}

func (like Like) NegationBuild(builder Builder) {
	builder.WriteQuoted(like.Column)
	builder.WriteString(" NOT LIKE ")
	builder.AddVar(builder, like.Value)
}

// Gt greater than for where
type Gt Eq

func (gt Gt) Build(builder Builder) {
	gt.Column.Build(builder)
	builder.WriteString(" > ")
	builder.AddVar(builder, gt.Value)
}

func (gt Gt) NegationBuild(builder Builder) {
	Lte(gt).Build(builder)
}

// Gte greater than or equal to for where
type Gte Eq

func (gte Gte) Build(builder Builder) {
	gte.Column.Build(builder)
	builder.WriteString(" >= ")
	builder.AddVar(builder, gte.Value)
}

func (gte Gte) NegationBuild(builder Builder) {
	Lt(gte).Build(builder)
}

// Lt less than for where
type Lt Eq

func (lt Lt) Build(builder Builder) {
	lt.Column.Build(builder)
	builder.WriteString(" < ")
	builder.AddVar(builder, lt.Value)
}

func (lt Lt) NegationBuild(builder Builder) {
	Gte(lt).Build(builder)
}

// Lte less than or equal to for where
type Lte Eq

func (lte Lte) Build(builder Builder) {
	lte.Column.Build(builder)
	builder.WriteString(" <= ")
	builder.AddVar(builder, lte.Value)
}

func (lte Lte) NegationBuild(builder Builder) {
	Gt(lte).Build(builder)
}

func eqNil(value any) bool {
	if valuer, ok := value.(driver.Valuer); ok && !eqNilReflect(valuer) {
		value, _ = valuer.Value()
	}

	return value == nil || eqNilReflect(value)
}

func eqNilReflect(value any) bool {
	reflectValue := reflect.ValueOf(value)
	return reflectValue.Kind() == reflect.Ptr && reflectValue.IsNil()
}
