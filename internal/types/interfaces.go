package types

type IFieldType[T any] interface {
	FieldType() T
}
