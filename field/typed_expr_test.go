package field

import (
	"testing"

	"github.com/donutnomad/gsql/clause"
	"github.com/stretchr/testify/assert"
)

func TestIntExpr_Gt(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.Gt(5)

	// 验证生成的 SQL
	assert.NotNil(t, result)
}

func TestIntExpr_Gte(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.Gte(5)

	assert.NotNil(t, result)
}

func TestIntExpr_Lt(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.Lt(5)

	assert.NotNil(t, result)
}

func TestIntExpr_Lte(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.Lte(5)

	assert.NotNil(t, result)
}

func TestIntExpr_Eq(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.Eq(5)

	assert.NotNil(t, result)
}

func TestIntExpr_Not(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.Not(5)

	assert.NotNil(t, result)
}

func TestIntExpr_Between(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.Between(1, 10)

	assert.NotNil(t, result)
}

func TestIntExpr_NotBetween(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.NotBetween(1, 10)

	assert.NotNil(t, result)
}

func TestIntExpr_In(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.In(1, 2, 3)

	assert.NotNil(t, result)
}

func TestIntExpr_NotIn(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	result := expr.NotIn(1, 2, 3)

	assert.NotNil(t, result)
}

func TestIntExpr_AsF(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})
	field := expr.AsF("total")

	assert.NotNil(t, field)
	assert.Equal(t, "total", field.Name())
}

func TestIntExpr_ToExpr(t *testing.T) {
	baseExpr := clause.Expr{SQL: "COUNT(*)"}
	expr := NewIntExpr(baseExpr)
	result := expr.ToExpr()

	assert.NotNil(t, result)
}

func TestIntExpr_ImplementsNumericExpr(t *testing.T) {
	expr := NewIntExpr(clause.Expr{SQL: "COUNT(*)"})

	// 验证 IntExpr 实现了 NumericExpr 接口
	var _ NumericExpr = expr
}

// FloatExpr 测试

func TestFloatExpr_Gt(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.Gt(99.99)

	assert.NotNil(t, result)
}

func TestFloatExpr_Gte(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.Gte(99.99)

	assert.NotNil(t, result)
}

func TestFloatExpr_Lt(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.Lt(99.99)

	assert.NotNil(t, result)
}

func TestFloatExpr_Lte(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.Lte(99.99)

	assert.NotNil(t, result)
}

func TestFloatExpr_Eq(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.Eq(99.99)

	assert.NotNil(t, result)
}

func TestFloatExpr_Not(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.Not(99.99)

	assert.NotNil(t, result)
}

func TestFloatExpr_Between(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.Between(10.0, 100.0)

	assert.NotNil(t, result)
}

func TestFloatExpr_NotBetween(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.NotBetween(10.0, 100.0)

	assert.NotNil(t, result)
}

func TestFloatExpr_In(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.In(10.0, 20.0, 30.0)

	assert.NotNil(t, result)
}

func TestFloatExpr_NotIn(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	result := expr.NotIn(10.0, 20.0, 30.0)

	assert.NotNil(t, result)
}

func TestFloatExpr_AsF(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})
	field := expr.AsF("avg_price")

	assert.NotNil(t, field)
	assert.Equal(t, "avg_price", field.Name())
}

func TestFloatExpr_ToExpr(t *testing.T) {
	baseExpr := clause.Expr{SQL: "AVG(price)"}
	expr := NewFloatExpr(baseExpr)
	result := expr.ToExpr()

	assert.NotNil(t, result)
}

func TestFloatExpr_ImplementsNumericExpr(t *testing.T) {
	expr := NewFloatExpr(clause.Expr{SQL: "AVG(price)"})

	// 验证 FloatExpr 实现了 NumericExpr 接口
	var _ NumericExpr = expr
}

// StringExpr 测试

func TestStringExpr_Like(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.Like("%test%")

	assert.NotNil(t, result)
}

func TestStringExpr_Like_WithEscape(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.Like("%test\\%value%", '\\')

	assert.NotNil(t, result)
}

func TestStringExpr_NotLike(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.NotLike("%test%")

	assert.NotNil(t, result)
}

func TestStringExpr_NotLike_WithEscape(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.NotLike("%test\\%value%", '\\')

	assert.NotNil(t, result)
}

func TestStringExpr_Contains(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.Contains("test")

	assert.NotNil(t, result)
}

func TestStringExpr_HasPrefix(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.HasPrefix("test")

	assert.NotNil(t, result)
}

func TestStringExpr_HasSuffix(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.HasSuffix("test")

	assert.NotNil(t, result)
}

func TestStringExpr_Eq(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MIN(name)"})
	result := expr.Eq("test")

	assert.NotNil(t, result)
}

func TestStringExpr_Gt(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.Gt("abc")

	assert.NotNil(t, result)
}

func TestStringExpr_In(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	result := expr.In("a", "b", "c")

	assert.NotNil(t, result)
}

func TestStringExpr_AsF(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})
	field := expr.AsF("max_name")

	assert.NotNil(t, field)
	assert.Equal(t, "max_name", field.Name())
}

func TestStringExpr_ToExpr(t *testing.T) {
	baseExpr := clause.Expr{SQL: "MAX(name)"}
	expr := NewStringExpr(baseExpr)
	result := expr.ToExpr()

	assert.NotNil(t, result)
}

func TestStringExpr_ImplementsNumericExpr(t *testing.T) {
	expr := NewStringExpr(clause.Expr{SQL: "MAX(name)"})

	// 验证 StringExpr 实现了 NumericExpr 接口（因为它继承了 numericExprBase）
	var _ NumericExpr = expr
}
