package field

import (
	"testing"

	"github.com/donutnomad/gsql/clause"
	"github.com/stretchr/testify/assert"
)

func TestTextExpr_Like(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "UPPER(name)", Vars: nil})
	result := expr.Like("JOHN%")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Len(t, e.Vars, 2)
	assert.Equal(t, "JOHN%", e.Vars[1])
}

func TestTextExpr_LikeWithEscape(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Like("100%", '\\')

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ? ESCAPE ?", e.SQL)
	assert.Len(t, e.Vars, 3)
	assert.Equal(t, "\\", e.Vars[2])
}

func TestTextExpr_NotLike(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "LOWER(email)", Vars: nil})
	result := expr.NotLike("%test%")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? NOT LIKE ?", e.SQL)
	assert.Len(t, e.Vars, 2)
}

func TestTextExpr_Contains(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "description", Vars: nil})
	result := expr.Contains("keyword")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Equal(t, "%keyword%", e.Vars[1])
}

func TestTextExpr_HasPrefix(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "code", Vars: nil})
	result := expr.HasPrefix("PRE_")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Equal(t, "PRE_%", e.Vars[1])
}

func TestTextExpr_HasSuffix(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "filename", Vars: nil})
	result := expr.HasSuffix(".txt")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? LIKE ?", e.SQL)
	assert.Equal(t, "%.txt", e.Vars[1])
}

func TestTextExpr_Eq(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "UPPER(name)", Vars: nil})
	result := expr.Eq("JOHN")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? = ?", e.SQL)
	assert.Equal(t, "JOHN", e.Vars[1])
}

func TestTextExpr_Gt(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "name", Vars: nil})
	result := expr.Gt("A")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? > ?", e.SQL)
	assert.Equal(t, "A", e.Vars[1])
}

func TestTextExpr_In(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "status", Vars: nil})
	result := expr.In("active", "pending")

	e, ok := result.(clause.Expr)
	assert.True(t, ok)
	assert.Equal(t, "? IN ?", e.SQL)
}

func TestTextExpr_AsF(t *testing.T) {
	expr := NewTextExpr(clause.Expr{SQL: "CONCAT(first, last)", Vars: nil})
	field := expr.AsF("full_name")

	assert.Equal(t, "full_name", field.Name())
}

func TestTextExpr_ImplementsNumericExpr(t *testing.T) {
	var _ NumericExpr = TextExpr{}
}
