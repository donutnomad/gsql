package fields

import (
	"github.com/donutnomad/gsql/field"
)

func IntColumn(name string) IntColumnBuilder {
	return IntColumnBuilder{name: name}
}

type IntColumnBuilder struct {
	name string
}

func (b IntColumnBuilder) From(source interface{ TableName() string }) IntField[int64] {
	return NewIntField[int64](source.TableName(), b.name)
}

func FloatColumn(name string) FloatColumnBuilder {
	return FloatColumnBuilder{name: name}
}

type FloatColumnBuilder struct {
	name string
}

func (b FloatColumnBuilder) From(source interface{ TableName() string }) FloatField[float64] {
	return NewFloatField[float64](source.TableName(), b.name)
}

func StringColumn(name string) StringColumnBuilder {
	return StringColumnBuilder{name: name}
}

type StringColumnBuilder struct {
	name string
}

func (b StringColumnBuilder) From(source interface{ TableName() string }) StringField[string] {
	return NewStringField[string](source.TableName(), b.name)
}

func BoolColumn(name string) BoolColumnBuilder {
	return BoolColumnBuilder{name: name}
}

type BoolColumnBuilder struct {
	name string
}

func (b BoolColumnBuilder) From(source interface{ TableName() string }) field.Comparable[bool] {
	return field.NewComparable[bool](source.TableName(), b.name)
}

func DateTimeColumn(name string) DateTimeColumnBuilder {
	return DateTimeColumnBuilder{name: name}
}

type DateTimeColumnBuilder struct {
	name string
}

func (b DateTimeColumnBuilder) From(source interface{ TableName() string }) DateTimeField[string] {
	return NewDateTimeField[string](source.TableName(), b.name)
}

func DateColumn(name string) DateColumnBuilder {
	return DateColumnBuilder{name: name}
}

type DateColumnBuilder struct {
	name string
}

func (b DateColumnBuilder) From(source interface{ TableName() string }) DateField[string] {
	return NewDateField[string](source.TableName(), b.name)
}

func TimeColumn(name string) TimeColumnBuilder {
	return TimeColumnBuilder{name: name}
}

type TimeColumnBuilder struct {
	name string
}

func (b TimeColumnBuilder) From(source interface{ TableName() string }) TimeField[string] {
	return NewTimeField[string](source.TableName(), b.name)
}

func DecimalColumn(name string) DecimalColumnBuilder {
	return DecimalColumnBuilder{name: name}
}

type DecimalColumnBuilder struct {
	name string
}

func (b DecimalColumnBuilder) From(source interface{ TableName() string }) DecimalField[float64] {
	return NewDecimalField[float64](source.TableName(), b.name)
}

func Column[T any](name string) ColumnBuilder[T] {
	return ColumnBuilder[T]{name: name}
}

type ColumnBuilder[T any] struct {
	name string
}

func (b ColumnBuilder[T]) From(source interface{ TableName() string }) field.Comparable[T] {
	return field.NewComparable[T](source.TableName(), b.name)
}
