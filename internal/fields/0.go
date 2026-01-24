package fields

type Cont[T any] interface {
	IntExpr[T] | FloatExpr[T] | StringExpr[T] | DecimalExpr[T] | TimeExpr[T] | DateTimeExpr[T] | DateExpr[T] | Json
}
