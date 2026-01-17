package field

// NumericExpr 数值表达式接口，支持比较操作
// 用于聚合函数（如 COUNT, SUM）返回的表达式，使其可以直接进行比较操作
// 例如: gsql.COUNT().Gt(5), gsql.SUM(amount).Gte(1000)
type NumericExpr interface {
	Expression
	ExpressionTo
	// Gt 大于 (>)
	Gt(value any) Expression
	// Gte 大于等于 (>=)
	Gte(value any) Expression
	// Lt 小于 (<)
	Lt(value any) Expression
	// Lte 小于等于 (<=)
	Lte(value any) Expression
	// Eq 等于 (=)
	Eq(value any) Expression
	// Not 不等于 (!=)
	Not(value any) Expression
	// Between 在范围内 (BETWEEN ... AND ...)
	Between(from, to any) Expression
	// NotBetween 不在范围内 (NOT BETWEEN ... AND ...)
	NotBetween(from, to any) Expression
	// In 在列表中 (IN (...))
	In(values ...any) Expression
	// NotIn 不在列表中 (NOT IN (...))
	NotIn(values ...any) Expression
}
