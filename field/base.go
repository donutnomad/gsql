package field

import (
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/donutnomad/gsql/internal/types"
)

const (
	FlagNone          = types.FlagNone
	FlagPrimaryKey    = types.FlagPrimaryKey
	FlagUniqueIndex   = types.FlagUniqueIndex
	FlagIndex         = types.FlagIndex
	FlagAutoIncrement = types.FlagAutoIncrement
)

type (
	FieldFlag    = types.FieldFlag
	Range[T any] = types.Range[T]
	IToExpr      = fieldi.IToExpr
	IField       = fieldi.IField
	ExpressionTo = fieldi.ExpressionTo
	BaseFields   = fields.BaseFields
)

type (
	// Deprecated: 使用 gsql.IntField 替代
	Comparable[T any] = fields.IntField[T]
	// Deprecated: 使用 gsql.StringField 替代
	Pattern[T any] = fields.StringField[T]
)

// Deprecated: 使用 gsql.IntFieldOf 替代
func NewComparable[T any](tableName, name string, flags ...types.FieldFlag) fields.IntField[T] {
	return fields.IntFieldOf[T](tableName, name, flags...)
}

// Deprecated: 使用 gsql.StringFieldOf 替代
func NewPattern[T any](tableName, name string, flags ...types.FieldFlag) fields.StringField[T] {
	return fields.StringFieldOf[T](tableName, name, flags...)
}
