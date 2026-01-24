package fieldii

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/types"
	"github.com/donutnomad/gsql/internal/utils"
)

type Comparable[T any] struct {
	Base
	comparableImpl[T]
	pointerImpl
}

func NewComparable[T any](tableName, name string, flags ...types.FieldFlag) Comparable[T] {
	b := fieldi.NewBase(tableName, name, flags...)
	return NewComparableFrom[T](b)
}

// Deprecated: 使用 NewComparableFrom 替代。
func NewComparableWithField[T any](field IField) Comparable[T] {
	return NewComparableFrom[T](field)
}

func NewComparableFrom[T any](field IField) Comparable[T] {
	base := ifieldToBase(field)
	return Comparable[T]{
		Base:           base,
		comparableImpl: comparableImpl[T]{IField: base},
		pointerImpl:    pointerImpl{IField: base},
	}
}

func (f Comparable[T]) FieldType() T {
	var def T
	return def
}

func (f Comparable[T]) Build(builder clause.Builder) {
	f.Base.ToExpr().Build(builder)
}

func (f Comparable[T]) WithTable(tableName interface{ TableName() string }, fieldNames ...string) Comparable[T] {
	return NewComparable[T](tableName.TableName(), utils.Optional(fieldNames, f.Base.Name()))
}

func (f Comparable[T]) WithName(fieldName string) Comparable[T] {
	return NewComparable[T](f.Base.TableName(), fieldName)
}

func (f Comparable[T]) WithAlias(alias string) Comparable[T] {
	b := f.Base
	b.SetAlias(alias)
	return NewComparableFrom[T](b)
}
