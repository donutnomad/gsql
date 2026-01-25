package types

import "github.com/samber/mo"

type Range[T any] struct {
	From mo.Option[T]
	To   mo.Option[T]
}

func (r Range[T]) FromValue() *T {
	return r.From.ToPointer()
}

func (r Range[T]) ToValue() *T {
	return r.To.ToPointer()
}
