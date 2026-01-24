package fields

type Expressions[T any] interface {
	IntField[T] | IntExpr[T] | FloatExpr[T] | StringExpr[T] | DecimalExpr[T] | TimeExpr[T] | DateTimeExpr[T] | DateExpr[T]
}
