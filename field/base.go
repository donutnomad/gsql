package field

import (
	"github.com/donutnomad/gsql/internal/cgg1"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/donutnomad/gsql/internal/fieldii"
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
	FieldFlag          = types.FieldFlag
	IToExpr            = fieldi.IToExpr
	IField             = fieldi.IField
	IPointer           = fieldi.IPointer
	IPattern[T any]    = fieldi.IPattern[T]
	IComparable[T any] = fieldi.IComparable[T]
	Range[T any]       = fieldi.Range[T]
	Expression         = fieldi.Expression
	Comparable[T any]  = fieldii.Comparable[T]
	ExpressionTo       = fieldi.ExpressionTo
	IFieldType[T any]  = fieldi.IFieldType[T]
	Base               = fieldi.Base
	BaseFields         = fieldi.BaseFields
	PatternImpl[T any] = fieldii.PatternImpl[T]
	PointerImpl        = fieldii.PointerImpl
	Pattern[T any]     = fieldii.Pattern[T]
)

// EmptyExpression 空表达式，用于跳过条件
var EmptyExpression = types.EmptyExpression

func NewBase(tableName, name string, flags ...FieldFlag) *Base {
	return fieldi.NewBase(tableName, name, flags...)
}

func NewBaseFromSql(expr Expression, alias string) *Base {
	return fieldi.NewBaseFromSql(expr, alias)
}

// Deprecated: 移除
func NewColumnClause(f Base) cgg1.ColumnClause {
	return fieldi.NewColumnClause(f)
}

func NewPattern[T any](tableName, name string, flags ...FieldFlag) Pattern[T] {
	return fieldii.NewPattern[T](tableName, name, flags...)
}

func NewPatternFrom[T any](field IField) Pattern[T] {
	return fieldii.NewPatternFrom[T](field)
}

func NewComparable[T any](tableName, name string, flags ...FieldFlag) Comparable[T] {
	return fieldii.NewComparable[T](tableName, name, flags...)
}

func NewComparableFrom[T any](field IField) Comparable[T] {
	return fieldii.NewComparableFrom[T](field)
}

// IFieldToBase 将 IField 转换为 Base
// 用于 internal/fields 包中创建类型化字段
func IFieldToBase(f IField) Base {
	return fieldi.IFieldToBase(f)
}
