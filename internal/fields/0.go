package fields

//go:generate go run -tags gen 0gen.go
//go:generate go run -tags gen 0gen1.go

type Expressions[T any] interface {
	IntField[T] | IntExpr[T] |
		FloatExpr[T] | FloatField[T] |
		StringExpr[T] | StringField[T] |
		DecimalExpr[T] | DecimalField[T] |
		TimeExpr[T] | TimeField[T] |
		DateTimeExpr[T] | DateTimeField[T] |
		DateExpr[T] | DateField[T]
}

type FunctionName string
