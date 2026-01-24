package fields

import (
	"testing"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

// ==================== 比较操作测试 ====================

func TestIntExprT_Eq(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "count", Vars: nil})
	result := expr.Eq(10)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", e.SQL)
	assert.Equal(t, 10, e.Vars[1])
}

func TestIntExprT_EqOpt(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "count", Vars: nil})

	result := expr.EqOpt(mo.Some(10))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", e.SQL)

	result2 := expr.EqOpt(mo.None[int]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestIntExprT_Gt(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.Gt(60)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? > ?", e.SQL)
}

func TestIntExprT_In(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "status", Vars: nil})
	result := expr.In(1, 2, 3)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IN ?", e.SQL)

	// 空列表返回空表达式
	result2 := expr.In()
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestIntExprT_Between(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "age", Vars: nil})
	result := expr.Between(18, 65)

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? BETWEEN ? AND ?", e.SQL)
}

func TestIntExprT_BetweenPtr(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "age", Vars: nil})

	// 两个值都有
	from, to := 18, 65
	result := expr.BetweenPtr(&from, &to)
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? BETWEEN ? AND ?", e.SQL)

	// 只有 from
	result2 := expr.BetweenPtr(&from, nil)
	e2, ok := result2.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? >= ?", e2.SQL)

	// 只有 to
	result3 := expr.BetweenPtr(nil, &to)
	e3, ok := result3.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? <= ?", e3.SQL)

	// 都为 nil
	result4 := expr.BetweenPtr(nil, nil)
	assert.Equal(t, field.EmptyExpression, result4)
}

func TestIntExprT_BetweenOpt(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "age", Vars: nil})

	result := expr.BetweenOpt(mo.Some(18), mo.Some(65))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? BETWEEN ? AND ?", e.SQL)

	result2 := expr.BetweenOpt(mo.None[int](), mo.None[int]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestIntExprT_BetweenF(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "age", Vars: nil})
	from := clause.Expr{SQL: "min_age", Vars: nil}
	to := clause.Expr{SQL: "max_age", Vars: nil}

	result := expr.BetweenF(from, to)
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? BETWEEN ? AND ?", e.SQL)
}

// ==================== 算术运算测试 ====================

func TestIntExprT_Add(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Add(100)

	e := result.Eq(200)
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", exprResult.SQL)
}

func TestIntExprT_Sub(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "price", Vars: nil})
	result := expr.Sub(50)

	e := result.Gt(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Mul(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "quantity", Vars: nil})
	result := expr.Mul(10)

	e := result.Lte(1000)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Div(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "total", Vars: nil})
	result := expr.Div(2)

	e := result.Gte(50)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_IntDiv(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.IntDiv(3)

	e := result.Eq(3)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Mod(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.Mod(3)

	e := result.Eq(1)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Neg(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "balance", Vars: nil})
	result := expr.Neg()

	e := result.Lt(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 位运算测试 ====================

func TestIntExprT_BitAnd(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "flags", Vars: nil})
	result := expr.BitAnd(0x01)

	e := result.Eq(1)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_BitOr(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "flags", Vars: nil})
	result := expr.BitOr(0x02)

	e := result.Gt(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_BitXor(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "flags", Vars: nil})
	result := expr.BitXor(0xFF)

	e := result.Not(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_BitNot(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "flags", Vars: nil})
	result := expr.BitNot()

	e := result.Lt(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_LeftShift(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.LeftShift(2)

	e := result.Eq(40)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_RightShift(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.RightShift(2)

	e := result.Eq(2)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 数学函数测试 ====================

func TestIntExprT_Abs(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "balance", Vars: nil})
	result := expr.Abs()

	e := result.Gte(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Sign(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "balance", Vars: nil})
	result := expr.Sign()

	e := result.In(-1, 0, 1)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Ceil(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.Ceil()

	e := result.Gte(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Floor(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.Floor()

	e := result.Gte(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Pow(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "base", Vars: nil})
	result := expr.Pow(2)

	// Pow 返回 Float
	e := result.Gt(0.0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Sqrt(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "area", Vars: nil})
	result := expr.Sqrt()

	// Sqrt 返回 Float
	e := result.Gte(0.0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 类型转换测试 ====================

func TestIntExprT_Cast(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "id", Vars: nil})
	result := expr.Cast("CHAR")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "CAST(? AS CHAR)", e.SQL)
}

func TestIntExprT_CastFloat(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "price", Vars: nil})
	result := expr.CastFloat(10, 2)

	e := result.Gt(0.0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_CastChar(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "id", Vars: nil})
	result := expr.CastChar()

	e := result.Like("1%")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_CastSigned(t *testing.T) {
	expr := NewInt[uint](clause.Expr{SQL: "value", Vars: nil})
	result := expr.CastSigned()

	e := result.Lt(int64(0))
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_CastUnsigned(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.CastUnsigned()

	e := result.Gte(uint64(0))
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 字符串转换测试 ====================

func TestIntExprT_Hex(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.Hex()

	e := result.Eq("FF")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Bin(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.Bin()

	e := result.Like("1%")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Oct(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "value", Vars: nil})
	result := expr.Oct()

	e := result.Eq("12")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 条件函数测试 ====================

func TestIntExprT_IfNull(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.IfNull(0)

	e := result.Gte(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Coalesce(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.Coalesce(0)

	e := result.Gte(0)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_NullIf(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.NullIf(0)

	e := result.IsNull()
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Greatest(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "a", Vars: nil})
	result := expr.Greatest(10, 20)

	e := result.Eq(20)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestIntExprT_Least(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "a", Vars: nil})
	result := expr.Least(10, 20)

	e := result.Eq(10)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 空值判断测试 ====================

func TestIntExprT_IsNull(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.IsNull()

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IS NULL", e.SQL)
}

func TestIntExprT_IsNotNull(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "score", Vars: nil})
	result := expr.IsNotNull()

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IS NOT NULL", e.SQL)
}

// ==================== 链式调用测试 ====================

func TestIntExprT_Chaining(t *testing.T) {
	expr := NewInt[int](clause.Expr{SQL: "price", Vars: nil})

	// 链式调用: (price + 100) * 2
	result := expr.Add(100).Mul(2)
	e := result.Gt(200)
	_, ok := e.(clause.Expr)
	assert.True(t, ok)

	// 链式调用: ABS(balance) > 0
	balance := NewInt[int](clause.Expr{SQL: "balance", Vars: nil})
	result2 := balance.Abs()
	e2 := result2.Gte(0)
	_, ok = e2.(clause.Expr)
	assert.True(t, ok)
}

// ==================== 泛型类型安全性测试 ====================

func TestIntExprT_TypeSafety(t *testing.T) {
	// int 类型
	intExpr := NewInt[int](clause.Expr{SQL: "count", Vars: nil})
	_ = intExpr.Eq(10)
	_ = intExpr.In(1, 2, 3)
	_ = intExpr.Between(0, 100)

	// int64 类型
	int64Expr := NewInt[int64](clause.Expr{SQL: "big_count", Vars: nil})
	_ = int64Expr.Eq(int64(10000000000))
	_ = int64Expr.In(int64(1), int64(2))

	// uint 类型
	uintExpr := NewInt[uint](clause.Expr{SQL: "positive", Vars: nil})
	_ = uintExpr.Eq(uint(10))
	_ = uintExpr.Gte(uint(0))
}
