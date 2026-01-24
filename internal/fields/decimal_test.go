package fields

import (
	"testing"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

// ==================== 比较操作测试 ====================

func TestDecimalExprT_Eq(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Eq(99.99)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", e.SQL)
	assert.Equal(t, 99.99, e.Vars[1])
}

func TestDecimalExprT_EqOpt(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})

	result := expr.EqOpt(mo.Some(99.99))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", e.SQL)

	result2 := expr.EqOpt(mo.None[float64]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestDecimalExprT_Gt(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Gt(100.00)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? > ?", e.SQL)
}

func TestDecimalExprT_In(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.In(9.99, 19.99, 29.99)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IN ?", e.SQL)

	result2 := expr.In()
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestDecimalExprT_Between(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Between(10.00, 100.00)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? BETWEEN ? AND ?", e.SQL)
}

// ==================== 算术运算测试 ====================

func TestDecimalExprT_Add(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Add(10.00)

	e := result.Gt(100.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Sub(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Sub(5.00)

	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Mul(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "quantity", Vars: nil})
	result := expr.Mul(9.99)

	e := result.Lte(1000.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Div(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "total", Vars: nil})
	result := expr.Div(2.00)

	e := result.Gt(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Neg(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "balance", Vars: nil})
	result := expr.Neg()

	e := result.Lt(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 数学函数测试 ====================

func TestDecimalExprT_Abs(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "balance", Vars: nil})
	result := expr.Abs()

	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Sign(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "balance", Vars: nil})
	result := expr.Sign()

	e := result.In(-1, 0, 1)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Ceil(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Ceil()

	e := result.Gte(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Floor(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Floor()

	e := result.Gte(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Round(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})

	result := expr.Round()
	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)

	result2 := expr.Round(2)
	e2 := result2.Gte(0.00)
	_, ok = e2.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Truncate(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Truncate(2)

	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Pow(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "base", Vars: nil})
	result := expr.Pow(2)

	e := result.Gt(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Sqrt(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "area", Vars: nil})
	result := expr.Sqrt()

	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Mod(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "total", Vars: nil})
	result := expr.Mod(100.00)

	e := result.Lt(100.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 类型转换测试 ====================

func TestDecimalExprT_Cast(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Cast("SIGNED")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "CAST(? AS SIGNED)", e.SQL)
}

func TestDecimalExprT_CastSigned(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.CastSigned()

	e := result.Gte(int64(0))
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_CastUnsigned(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.CastUnsigned()

	e := result.Gte(uint64(0))
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_CastDecimal(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "value", Vars: nil})
	result := expr.CastDecimal(10, 2)

	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_CastFloat(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.CastFloat()

	e := result.Gt(0.0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_CastChar(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.CastChar()

	e := result.Like("99%")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 条件函数测试 ====================

func TestDecimalExprT_IfNull(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.IfNull(0.00)

	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Coalesce(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Coalesce(0.00)

	e := result.Gte(0.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_NullIf(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.NullIf(0.00)

	e := result.IsNull()
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Greatest(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "a", Vars: nil})
	result := expr.Greatest(10.00, 20.00)

	e := result.Eq(20.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestDecimalExprT_Least(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "a", Vars: nil})
	result := expr.Least(10.00, 20.00)

	e := result.Eq(10.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 格式化测试 ====================

func TestDecimalExprT_Format(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Format(2)

	e := result.Like("1,234%")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 空值判断测试 ====================

func TestDecimalExprT_IsNull(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.IsNull()

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IS NULL", e.SQL)
}

func TestDecimalExprT_IsNotNull(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	result := expr.IsNotNull()

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IS NOT NULL", e.SQL)
}

// ==================== 链式调用测试 ====================

func TestDecimalExprT_Chaining(t *testing.T) {
	expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})

	// 链式调用: (price + tax) * quantity
	result := expr.Add(10.00).Mul(2)
	e := result.Gt(100.00)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)

	// 链式调用: ROUND(ABS(balance), 2)
	balance := DecimalOf[float64](clause.Expr{SQL: "balance", Vars: nil})
	result2 := balance.Abs().Round(2)
	e2 := result2.Gte(0.00)
	_, ok = e2.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 泛型类型安全性测试 ====================

func TestDecimalExprT_TypeSafety(t *testing.T) {
	// float64 类型 (常用)
	f64Expr := DecimalOf[float64](clause.Expr{SQL: "price", Vars: nil})
	_ = f64Expr.Eq(99.99)
	_ = f64Expr.In(9.99, 19.99, 29.99)
	_ = f64Expr.Between(0.00, 100.00)

	// float32 类型
	f32Expr := DecimalOf[float32](clause.Expr{SQL: "small_price", Vars: nil})
	_ = f32Expr.Eq(float32(9.99))
	_ = f32Expr.Gte(float32(0))

	// string 类型 (用于自定义 DecimalExpr 类型如 shopspring/decimal)
	strExpr := DecimalOf[string](clause.Expr{SQL: "big_decimal", Vars: nil})
	_ = strExpr.Eq("123456789.123456789")
}
