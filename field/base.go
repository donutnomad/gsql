package field

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/donutnomad/gsql/internal/types"
)

const (
	FlagNone          FieldFlag = types.FlagNone
	FlagPrimaryKey    FieldFlag = types.FlagPrimaryKey
	FlagUniqueIndex   FieldFlag = types.FlagUniqueIndex
	FlagIndex         FieldFlag = types.FlagIndex
	FlagAutoIncrement FieldFlag = types.FlagAutoIncrement
)

type (
	FieldFlag         = types.FieldFlag
	IToExpr           = fieldi.IToExpr
	IField            = fieldi.IField
	Range[T any]      = types.Range[T]
	Comparable[T any] = fields.IntField[T]
	ExpressionTo      = fieldi.ExpressionTo
	Base              = fieldi.Base
	BaseFields        = fieldi.BaseFields
	Pattern[T any]    = fields.StringField[T]
)

func NewBase(tableName, name string, flags ...FieldFlag) *Base {
	return fieldi.NewBase(tableName, name, flags...)
}

func NewBaseFromSql(expr clause.Expression, alias string) *Base {
	return fieldi.NewBaseFromSql(expr, alias)
}

func NewComparable[T any](tableName, name string, flags ...types.FieldFlag) Comparable[T] {
	return fields.NewIntField[T](tableName, name, flags...)
}

func NewPattern[T any](tableName, name string, flags ...types.FieldFlag) Pattern[T] {
	return fields.NewStringField[T](tableName, name, flags...)
}
