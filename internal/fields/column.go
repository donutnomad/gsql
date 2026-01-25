package fields

func BoolColumn(name string) ScalarColumnBuilder[bool] {
	return ScalarColumnBuilder[bool]{name: name}
}
func Column[T any](name string) ScalarColumnBuilder[T] {
	return ScalarColumnBuilder[T]{name: name}
}
