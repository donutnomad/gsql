package field

import (
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
	Range[T any]      = types.Range[T]
	IToExpr           = fieldi.IToExpr
	IField            = fieldi.IField
	ExpressionTo      = fieldi.ExpressionTo
	Comparable[T any] = fields.IntField[T]
	BaseFields        = fields.BaseFields
	Pattern[T any]    = fields.StringField[T]
)

func NewComparable[T any](tableName, name string, flags ...types.FieldFlag) Comparable[T] {
	return fields.NewIntField[T](tableName, name, flags...)
}

func NewPattern[T any](tableName, name string, flags ...types.FieldFlag) Pattern[T] {
	return fields.NewStringField[T](tableName, name, flags...)
}
