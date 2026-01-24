package fields

import (
	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*Scalar[any])(nil)

// ==================== Scalar 定义 ====================

// Scalar 标量类型表达式，用于没有专门类型覆盖的其他类型
// @gentype default=[any]
// 支持基础比较操作：Eq, Not, In, NotIn
// 使用场景：
//   - bool 类型字段
//   - 枚举类型字段
//   - 其他没有专门类型的字段
type Scalar[T any] struct {
	baseComparableImpl[T] // Eq, Not, In, NotIn
	pointerExprImpl       // IsNull, IsNotNull
	nullCondFuncSql       // IfNull, Coalesce, NullIf
	baseExprSql           // Build, ToExpr, As
}

func NewScalar[T any](expr clause.Expression) Scalar[T] {
	return Scalar[T]{
		baseComparableImpl: baseComparableImpl[T]{Expression: expr},
		pointerExprImpl:    pointerExprImpl{Expression: expr},
		nullCondFuncSql:    nullCondFuncSql{Expression: expr},
		baseExprSql:        baseExprSql{Expr: expr},
	}
}
