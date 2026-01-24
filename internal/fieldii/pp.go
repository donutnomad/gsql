package fieldii

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/types"
)

type Pattern[T any] struct {
	Base
	comparableImpl[T]
	patternImpl[T]
	pointerImpl
}

func NewPattern[T any](tableName, name string, flags ...types.FieldFlag) Pattern[T] {
	b := fieldi.NewBase(tableName, name, flags...)
	return NewPatternFrom[T](*b)
}

func NewPatternFrom[T any](field IField) Pattern[T] {
	base := ifieldToBase(field)
	return Pattern[T]{
		Base:           base,
		comparableImpl: comparableImpl[T]{IField: base},
		patternImpl:    patternImpl[T]{IField: base},
		pointerImpl:    pointerImpl{IField: base},
	}
}

func (f Pattern[T]) WithTable(tableName interface{ TableName() string }, fieldName ...string) Pattern[T] {
	var name = f.Base.ColumnName()
	if len(fieldName) > 0 {
		name = fieldName[0]
	}
	return NewPattern[T](tableName.TableName(), name)
}

func (f Pattern[T]) WithName(name string) Pattern[T] {
	return NewPattern[T](f.Base.TableName(), name)
}

func (f Pattern[T]) WithAlias(alias string) Pattern[T] {
	b := f.Base
	b.SetAlias(alias)
	return NewPatternFrom[T](b)
}

func (f Pattern[T]) FieldType() T {
	var def T
	return def
}

func (f Pattern[T]) Build(builder clause.Builder) {
	f.Base.ToExpr().Build(builder)
}
