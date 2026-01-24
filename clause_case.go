package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/cgg2"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/donutnomad/gsql/internal/types"
)

type CaseBuilder[Result fields.Expressions[V], V any] struct {
	value     clause.Expression // CASE value WHEN ... (可选，简单 CASE 表达式)
	whenPairs []whenPair
	elseValue Result
}

type whenPair struct {
	condition clause.Expression // WHEN condition (或简单 CASE 中的比较值)
	result    clause.Expression // THEN result
}

// Case 创建 CASE 表达式构建器（搜索式 CASE）
// 用法: gsql.Case().When(cond1, val1).When(cond2, val2).Else(val3)
func Case[V any, Result fields.Expressions[V]]() *CaseBuilder[Result, V] {
	return &CaseBuilder[Result, V]{}
}

func CaseString() *CaseBuilder[fields.StringExpr[string], string] {
	return Case[string, fields.StringExpr[string]]()
}

// CaseValue 创建简单 CASE 表达式构建器（简单 CASE）
// 用法: gsql.CaseValue(field).When(val1, result1).When(val2, result2).Else(defaultVal)
func CaseValue(value clause.Expression) *CaseBuilder {
	return &CaseBuilder{value: value}
}

// When 添加 WHEN ... THEN ... 子句
// 对于搜索式 CASE：condition 是布尔表达式
// 对于简单 CASE：condition 是与 value 比较的值
func (c *CaseBuilder[Result, V]) When(condition fields.Condition, result Result) *CaseBuilder[Result, V] {
	c.whenPairs = append(c.whenPairs, whenPair{
		condition: condition,
		result:    result,
	})
	return c
}

// Else 设置 ELSE 子句（可选）
func (c *CaseBuilder[Result, V]) Else(value Result) *CaseExpr {
	c.elseValue = value
	return &CaseExpr{ExprTo{Expression: caseClause{c}}}
}

// End 结束 CASE 表达式，返回可用作字段的表达式
func (c *CaseBuilder[Result, V]) End() *CaseExpr {
	return &CaseExpr{ExprTo{Expression: caseClause{c}}}
}

type CaseExpr struct {
	cgg2.ExprTo
}

var _ clause.Expression = (*caseClause)(nil)

type caseClause struct {
	*CaseBuilder
}

func (c caseClause) Build(builder clause.Builder) {
	writer := &types.SafeWriter{Builder: builder}

	writer.WriteString("CASE")

	// 简单 CASE 表达式：CASE value WHEN ...
	if c.value != nil {
		writer.WriteByte(' ')
		writer.AddVar(writer, c.value)
	}

	// WHEN ... THEN ... 子句
	for _, pair := range c.whenPairs {
		writer.WriteString(" WHEN ")
		writer.AddVar(writer, pair.condition)
		writer.WriteString(" THEN ")
		writer.AddVar(writer, pair.result)
	}

	// ELSE 子句（可选）
	if c.elseValue != nil {
		writer.WriteString(" ELSE ")
		writer.AddVar(writer, c.elseValue)
	}

	writer.WriteString(" END")
}
