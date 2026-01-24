package fields

import (
	"testing"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

func TestTextExpr_Like(t *testing.T) {
	expr := String(clause.Expr{SQL: "UPPER(name)", Vars: nil})
	result := expr.Like("JOHN%")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Len(t, e.Vars, 2)
	assert.Equal(t, "JOHN%", e.Vars[1])
}

func TestTextExpr_LikeWithEscape(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Like("100%", '\\')

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ? ESCAPE ?", e.SQL)
	assert.Len(t, e.Vars, 3)
	assert.Equal(t, "\\", e.Vars[2])
}

func TestTextExpr_LikeOpt(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})

	// 有值时
	result := expr.LikeOpt(mo.Some("JOHN%"))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)

	// 无值时返回空表达式
	result2 := expr.LikeOpt(mo.None[string]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestTextExpr_NotLike(t *testing.T) {
	expr := String(clause.Expr{SQL: "LOWER(email)", Vars: nil})
	result := expr.NotLike("%test%")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? NOT LIKE ?", e.SQL)
	assert.Len(t, e.Vars, 2)
}

func TestTextExpr_Contains(t *testing.T) {
	expr := String(clause.Expr{SQL: "description", Vars: nil})
	result := expr.Contains("keyword")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Equal(t, "%keyword%", e.Vars[1])
}

func TestTextExpr_ContainsOpt(t *testing.T) {
	expr := String(clause.Expr{SQL: "description", Vars: nil})

	result := expr.ContainsOpt(mo.Some("test"))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "%test%", e.Vars[1])

	result2 := expr.ContainsOpt(mo.None[string]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestTextExpr_HasPrefix(t *testing.T) {
	expr := String(clause.Expr{SQL: "code", Vars: nil})
	result := expr.HasPrefix("PRE_")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Equal(t, "PRE_%", e.Vars[1])
}

func TestTextExpr_HasPrefixOpt(t *testing.T) {
	expr := String(clause.Expr{SQL: "code", Vars: nil})

	result := expr.HasPrefixOpt(mo.Some("PRE_"))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "PRE_%", e.Vars[1])

	result2 := expr.HasPrefixOpt(mo.None[string]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestTextExpr_HasSuffix(t *testing.T) {
	expr := String(clause.Expr{SQL: "filename", Vars: nil})
	result := expr.HasSuffix(".txt")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Equal(t, "%.txt", e.Vars[1])
}

func TestTextExpr_HasSuffixOpt(t *testing.T) {
	expr := String(clause.Expr{SQL: "filename", Vars: nil})

	result := expr.HasSuffixOpt(mo.Some(".txt"))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "%.txt", e.Vars[1])

	result2 := expr.HasSuffixOpt(mo.None[string]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestTextExpr_Eq(t *testing.T) {
	expr := String(clause.Expr{SQL: "UPPER(name)", Vars: nil})
	result := expr.Eq("JOHN")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", e.SQL)
	assert.Equal(t, "JOHN", e.Vars[1])
}

func TestTextExpr_EqOpt(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})

	result := expr.EqOpt(mo.Some("test"))
	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", e.SQL)

	result2 := expr.EqOpt(mo.None[string]())
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestTextExpr_Not(t *testing.T) {
	expr := String(clause.Expr{SQL: "status", Vars: nil})
	result := expr.Not("deleted")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? != ?", e.SQL)
	assert.Equal(t, "deleted", e.Vars[1])
}

func TestTextExpr_In(t *testing.T) {
	expr := String(clause.Expr{SQL: "status", Vars: nil})
	result := expr.In("active", "pending")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IN ?", e.SQL)

	// 空列表返回空表达式
	result2 := expr.In()
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestTextExpr_NotIn(t *testing.T) {
	expr := String(clause.Expr{SQL: "status", Vars: nil})
	result := expr.NotIn("deleted", "archived")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? NOT IN ?", e.SQL)

	// 空列表返回空表达式
	result2 := expr.NotIn()
	assert.Equal(t, field.EmptyExpression, result2)
}

func TestTextExpr_IsNull(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.IsNull()

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IS NULL", e.SQL)
}

func TestTextExpr_IsNotNull(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.IsNotNull()

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IS NOT NULL", e.SQL)
}

func TestTextExpr_As(t *testing.T) {
	expr := String(clause.Expr{SQL: "CONCAT(first, last)", Vars: nil})
	field := expr.As("full_name")
	assert.Equal(t, "full_name", field.Name())
}

func TestTextExpr_Cast(t *testing.T) {
	expr := String(clause.Expr{SQL: "price_str", Vars: nil})
	result := expr.Cast("SIGNED")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "CAST(? AS SIGNED)", e.SQL)
}

func TestTextExpr_CastSigned(t *testing.T) {
	expr := String(clause.Expr{SQL: "amount_str", Vars: nil})
	result := expr.CastSigned()

	// CastSigned 返回 IntExpr
	e := result.Gt(100)
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? > ?", exprResult.SQL)
}

func TestTextExpr_CastUnsigned(t *testing.T) {
	expr := String(clause.Expr{SQL: "count_str", Vars: nil})
	result := expr.CastUnsigned()

	// CastUnsigned 返回 IntExpr
	e := result.Gte(0)
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? >= ?", exprResult.SQL)
}

func TestTextExpr_CastDecimal(t *testing.T) {
	expr := String(clause.Expr{SQL: "price_str", Vars: nil})
	result := expr.CastDecimal(10, 2)

	// CastDecimal 返回 FloatExpr
	e := result.Lt(100.50)
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? < ?", exprResult.SQL)
}

func TestTextExpr_CastChar(t *testing.T) {
	expr := String(clause.Expr{SQL: "user_id", Vars: nil})
	result := expr.CastChar(10)

	// CastChar 返回 StringExpr[string]
	e := result.Like("123%")
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", exprResult.SQL)
}

func TestTextExpr_CastCharNoLength(t *testing.T) {
	expr := StringOf[int](clause.Expr{SQL: "user_id", Vars: nil})
	result := expr.CastChar()

	// 验证 SQL 生成
	e := result.Eq("test")
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", exprResult.SQL)
}

// 测试泛型类型安全性
func TestTextExpr_TypeSafety(t *testing.T) {
	// string 类型
	strExpr := String(clause.Expr{SQL: "name", Vars: nil})
	_ = strExpr.Eq("test")       // 只能传 string
	_ = strExpr.In("a", "b")     // 只能传 string
	_ = strExpr.Like("test%")    // 只能传 string
	_ = strExpr.Contains("abc")  // 只能传 string
	_ = strExpr.HasPrefix("pre") // 只能传 string

	// int 类型（用于数字字符串场景）
	intExpr := StringOf[int](clause.Expr{SQL: "code", Vars: nil})
	_ = intExpr.Eq(123)     // 只能传 int
	_ = intExpr.In(1, 2, 3) // 只能传 int
	_ = intExpr.Like(100)   // 只能传 int
}

// ==================== 字符串函数测试 ====================

func TestTextExpr_Upper(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Upper()

	// 链式调用测试
	e := result.Eq("JOHN")
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", exprResult.SQL)
}

func TestTextExpr_Lower(t *testing.T) {
	expr := String(clause.Expr{SQL: "email", Vars: nil})
	result := expr.Lower()

	e := result.Like("%@gmail.com")
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", exprResult.SQL)
}

func TestTextExpr_Trim(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Trim()

	e := result.Eq("test")
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", exprResult.SQL)
}

func TestTextExpr_LTrim(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.LTrim()

	e := result.Eq("test")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_RTrim(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.RTrim()

	e := result.Eq("test")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_Substring(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Substring(1, 3)

	e := result.Eq("JOH")
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", exprResult.SQL)
}

func TestTextExpr_Left(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Left(5)

	e := result.Like("J%")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_Right(t *testing.T) {
	expr := String(clause.Expr{SQL: "phone", Vars: nil})
	result := expr.Right(4)

	e := result.Eq("1234")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_Length(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Length()

	// Length 返回 IntExpr
	e := result.Gt(10)
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? > ?", exprResult.SQL)
}

func TestTextExpr_CharLength(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.CharLength()

	e := result.Lte(50)
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? <= ?", exprResult.SQL)
}

func TestTextExpr_Concat(t *testing.T) {
	expr := String(clause.Expr{SQL: "first_name", Vars: nil})
	lastName := clause.Expr{SQL: "last_name", Vars: nil}
	result := expr.Concat(lastName)

	e := result.Like("John%")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_Replace(t *testing.T) {
	expr := String(clause.Expr{SQL: "phone", Vars: nil})
	result := expr.Replace("-", "")

	e := result.Eq("1234567890")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_Locate(t *testing.T) {
	expr := String(clause.Expr{SQL: "email", Vars: nil})
	result := expr.Locate("@")

	// Locate 返回 IntExpr
	e := result.Gt(0)
	exprResult, ok := e.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? > ?", exprResult.SQL)
}

func TestTextExpr_Reverse(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Reverse()

	e := result.Eq("nhoJ")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_Repeat(t *testing.T) {
	expr := String(clause.Expr{SQL: "'*'", Vars: nil})
	result := expr.Repeat(5)

	e := result.Eq("*****")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_LPad(t *testing.T) {
	expr := String(clause.Expr{SQL: "id", Vars: nil})
	result := expr.LPad(5, "0")

	e := result.Eq("00001")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

func TestTextExpr_RPad(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})
	result := expr.RPad(20, " ")

	e := result.Like("John%")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)
}

// 测试链式调用
func TestTextExpr_Chaining(t *testing.T) {
	expr := String(clause.Expr{SQL: "name", Vars: nil})

	// 链式调用: UPPER(TRIM(name))
	result := expr.Trim().Upper()
	e := result.Eq("JOHN")
	_, ok := e.(clause.Expr)
	assert.True(t, ok)

	// 链式调用: LEFT(LOWER(email), 10)
	email := String(clause.Expr{SQL: "email", Vars: nil})
	result2 := email.Lower().Left(10)
	e2 := result2.Contains("@")
	_, ok = e2.(clause.Expr)
	assert.True(t, ok)
}
