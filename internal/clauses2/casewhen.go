package clauses2

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/types"
	"github.com/samber/lo"
)

type SimpleCaseData struct {
	expression clause.Expression
	values     []clause.Expression
	results    []clause.Expression
	elseResult clause.Expression
}

func NewSimpleCaseData(expression clause.Expression, values []clause.Expression, results []clause.Expression, elseResult clause.Expression) *SimpleCaseData {
	return &SimpleCaseData{expression: expression, values: values, results: results, elseResult: elseResult}
}

type WhenPairAny struct {
	condition clause.Expression // WHEN condition (或简单 CASE 中的比较值)
	result    clause.Expression // THEN result
}

func NewWhenPairAny(condition clause.Expression, result clause.Expression) *WhenPairAny {
	return &WhenPairAny{condition: condition, result: result}
}

type SearchCaseData struct {
	whenPairs []WhenPairAny
	elseValue clause.Expression
}

func NewSearchCaseData(whenPairs []WhenPairAny, elseValue clause.Expression) *SearchCaseData {
	return &SearchCaseData{whenPairs: whenPairs, elseValue: elseValue}
}

type CaseWhenExpr struct {
	Simple *SimpleCaseData
	Search *SearchCaseData
}

var _ clause.Expression = (*CaseWhenExpr)(nil)

func (c CaseWhenExpr) Build(builder clause.Builder) {
	writer := &types.SafeWriter{Builder: builder}

	writer.WriteString("CASE")

	if c.Simple != nil {
		// 简单 CASE: CASE expression WHEN value1 THEN result1 ...
		writer.WriteString(" ")
		writer.AddVar(writer, c.Simple.expression)
		for i, value := range c.Simple.values {
			writer.WriteString(" WHEN ")
			writer.AddVar(writer, value)
			writer.WriteString(" THEN ")
			writer.AddVar(writer, c.Simple.results[i])
		}
		// ELSE 子句（可选）
		if !lo.IsNil(c.Simple.elseResult) {
			writer.WriteString(" ELSE ")
			writer.AddVar(writer, c.Simple.elseResult)
		}
	} else if c.Search != nil {
		// 搜索式 CASE: CASE WHEN condition1 THEN result1 ...
		for _, pair := range c.Search.whenPairs {
			writer.WriteString(" WHEN ")
			writer.AddVar(writer, pair.condition)
			writer.WriteString(" THEN ")
			writer.AddVar(writer, pair.result)
		}
		// ELSE 子句（可选）
		if c.Search.elseValue != nil {
			writer.WriteString(" ELSE ")
			writer.AddVar(writer, c.Search.elseValue)
		}
	}

	writer.WriteString(" END")
}
