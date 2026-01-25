package clauses

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/clauses2"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/samber/lo"
)

var Cases = struct {
	String    func() *SearchCaseBuilder[fields.StringExpr[string], string]
	Int       func() *SearchCaseBuilder[fields.IntExpr[int], int]
	Int8      func() *SearchCaseBuilder[fields.IntExpr[int8], int8]
	Int16     func() *SearchCaseBuilder[fields.IntExpr[int16], int16]
	Int32     func() *SearchCaseBuilder[fields.IntExpr[int32], int32]
	Int64     func() *SearchCaseBuilder[fields.IntExpr[int64], int64]
	Uint      func() *SearchCaseBuilder[fields.IntExpr[uint64], uint64]
	Uint8     func() *SearchCaseBuilder[fields.IntExpr[uint8], uint8]
	Uint16    func() *SearchCaseBuilder[fields.IntExpr[uint16], uint16]
	Uint32    func() *SearchCaseBuilder[fields.IntExpr[uint32], uint32]
	Uint64    func() *SearchCaseBuilder[fields.IntExpr[uint64], uint64]
	Float     func() *SearchCaseBuilder[fields.FloatExpr[float64], float64]
	Float32   func() *SearchCaseBuilder[fields.FloatExpr[float32], float32]
	Decimal   func() *SearchCaseBuilder[fields.DecimalExpr[float64], float64]
	Decimal32 func() *SearchCaseBuilder[fields.DecimalExpr[float32], float32]
}{
	String:    Case[string, fields.StringExpr[string]],
	Int:       Case[int, fields.IntExpr[int]],
	Uint:      Case[uint64, fields.IntExpr[uint64]],
	Int8:      Case[int8, fields.IntExpr[int8]],
	Int16:     Case[int16, fields.IntExpr[int16]],
	Int32:     Case[int32, fields.IntExpr[int32]],
	Int64:     Case[int64, fields.IntExpr[int64]],
	Uint8:     Case[uint8, fields.IntExpr[uint8]],
	Uint16:    Case[uint16, fields.IntExpr[uint16]],
	Uint32:    Case[uint32, fields.IntExpr[uint32]],
	Uint64:    Case[uint64, fields.IntExpr[uint64]],
	Float:     Case[float64, fields.FloatExpr[float64]],
	Float32:   Case[float32, fields.FloatExpr[float32]],
	Decimal:   Case[float64, fields.DecimalExpr[float64]],
	Decimal32: Case[float32, fields.DecimalExpr[float32]],
}

// Case 创建搜索式 CASE 表达式构建器
//
// 搜索式 CASE 语法:
//
//	CASE
//	    WHEN condition1 THEN result1
//	    WHEN condition2 THEN result2
//	    [ELSE default_result]
//	END
//
// 用法:
//
//	Case[int, gsql.IntExpr[int]]().
//	    When(user.Age.Gte(18), gsql.IntVal(1)).
//	    When(user.Age.Lt(18), gsql.IntVal(0)).
//	    Else(gsql.IntVal(-1))
func Case[R any, ResultExpr fields.Expressions[R]]() *SearchCaseBuilder[ResultExpr, R] {
	return &SearchCaseBuilder[ResultExpr, R]{}
}

// CaseValue 创建简单 CASE 表达式构建器（简单 CASE）
// 用法: gsql.CaseValue(target).When(cond1val, val1).When(cond2val, val2).Else(defaultVal)
// CASE expression
// WHEN value1 THEN result1
// WHEN value2 THEN result2
// [ELSE default_result]
// END
func CaseValue[V any, R any, ResultExpr fields.Expressions[R], ValExpr fields.Expressions[V]](expression ValExpr) *SimpleCaseBuilder[ValExpr, V, ResultExpr, R] {
	return &SimpleCaseBuilder[ValExpr, V, ResultExpr, R]{expression: expression}
}

///////////////////////////// Simple CASE-WHEN //////////////////////////

type SimpleCaseBuilder[ValExpr fields.Expressions[V], V any, ResultExpr fields.Expressions[R], R any] struct {
	expression ValExpr
	values     []ValExpr
	results    []ResultExpr
	elseResult *ResultExpr
}

func (c *SimpleCaseBuilder[ValExpr, V, ResultExpr, R]) asAny() clauses2.CaseWhenExpr {
	values := lo.Map(c.values, func(v ValExpr, _ int) clause.Expression {
		return any(v).(clause.Expression)
	})
	results := lo.Map(c.results, func(r ResultExpr, _ int) clause.Expression {
		return any(r).(clause.Expression)
	})
	var elseResult clause.Expression
	if any(c.elseResult) != nil {
		elseResult = any(c.elseResult).(clause.Expression)
	}
	return clauses2.CaseWhenExpr{
		Simple: clauses2.NewSimpleCaseData(any(c.expression).(clause.Expression),
			values, results, elseResult,
		),
	}
}

// When 添加 WHEN ... THEN ... 子句
// 对于搜索式 CASE：condition 是布尔表达式
// 对于简单 CASE：condition 是与 value 比较的值
func (c *SimpleCaseBuilder[ValExpr, V, ResultExpr, R]) When(value ValExpr, result ResultExpr) *SimpleCaseBuilder[ValExpr, V, ResultExpr, R] {
	c.values = append(c.values, value)
	c.results = append(c.results, result)
	return c
}

// Else 设置 ELSE 子句（可选）
func (c *SimpleCaseBuilder[ValExpr, V, ResultExpr, R]) Else(result ResultExpr) fields.ScalarExpr[R] {
	c.elseResult = &result
	return fields.ScalarOf[R](c.asAny())
}

// End 结束 CASE 表达式，返回可用作字段的表达式
func (c *SimpleCaseBuilder[ValExpr, V, ResultExpr, R]) End() fields.ScalarExpr[R] {
	return fields.ScalarOf[R](c.asAny())
}

///////////////////////////// CASE-WHEN //////////////////////////

type SearchCaseBuilder[ResultExpr fields.Expressions[R], R any] struct {
	whenPairs []whenPair[ResultExpr, R]
	elseValue *ResultExpr
}

func (c *SearchCaseBuilder[ResultExpr, R]) asAny() clauses2.CaseWhenExpr {
	var elseValue clause.Expression
	if c.elseValue != nil {
		elseValue = any(*c.elseValue).(clause.Expression)
	}
	return clauses2.CaseWhenExpr{
		Search: clauses2.NewSearchCaseData(
			lo.Map(c.whenPairs, func(w whenPair[ResultExpr, R], _ int) clauses2.WhenPairAny {
				return w.asAny()
			}),
			elseValue,
		),
	}
}

// When 添加 WHEN ... THEN ... 子句
// 对于搜索式 CASE：condition 是布尔表达式
// 对于简单 CASE：condition 是与 value 比较的值
func (c *SearchCaseBuilder[ResultExpr, R]) When(condition fields.Condition, val ResultExpr) *SearchCaseBuilder[ResultExpr, R] {
	c.whenPairs = append(c.whenPairs, whenPair[ResultExpr, R]{
		condition: condition,
		result:    val,
	})
	return c
}

// Else 设置 ELSE 子句（可选）
func (c *SearchCaseBuilder[ResultExpr, R]) Else(value ResultExpr) fields.ScalarExpr[R] {
	c.elseValue = &value
	return fields.ScalarOf[R](c.asAny())
}

// End 结束 CASE 表达式，返回可用作字段的表达式
func (c *SearchCaseBuilder[ResultExpr, R]) End() fields.ScalarExpr[R] {
	return fields.ScalarOf[R](c.asAny())
}

// //////////////////////// builder //////////////////////////

type whenPair[ResultExpr fields.Expressions[R], R any] struct {
	condition clause.Expression // WHEN condition (或简单 CASE 中的比较值)
	result    ResultExpr        // THEN result
}

func (w *whenPair[ResultExpr, R]) asAny() clauses2.WhenPairAny {
	return *clauses2.NewWhenPairAny(w.condition, any(w.result).(clause.Expression))
}
