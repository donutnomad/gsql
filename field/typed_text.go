package field

import (
	"github.com/donutnomad/gsql/clause"
)

// TextExpr 文本类型表达式，用于 VARCHAR 和 TEXT 类型字段
// 支持比较操作（继承自 numericExprBase）和模式匹配操作
// 使用场景：
//   - CONCAT, SUBSTRING 等字符串函数的返回值
//   - UPPER, LOWER 等字符串转换函数的返回值
//   - 派生表中的文本列
type TextExpr struct {
	numericExprBase // 复用比较操作
}

// NewTextExpr 创建一个新的 TextExpr 实例
func NewTextExpr(expr clause.Expression) TextExpr {
	return TextExpr{numericExprBase{Expression: expr}}
}

// Like 模式匹配 (LIKE)
// SELECT * FROM users WHERE UPPER(name) LIKE 'JOHN%';
func (e TextExpr) Like(pattern string, escape ...byte) Expression {
	if len(escape) > 0 {
		return clause.Expr{
			SQL:  "? LIKE ? ESCAPE ?",
			Vars: []any{e.Expression, pattern, string(escape[0])},
		}
	}
	return clause.Expr{
		SQL:  "? LIKE ?",
		Vars: []any{e.Expression, pattern},
	}
}

// NotLike 不匹配模式 (NOT LIKE)
// SELECT * FROM users WHERE CONCAT(first_name, last_name) NOT LIKE '%test%';
func (e TextExpr) NotLike(pattern string, escape ...byte) Expression {
	if len(escape) > 0 {
		return clause.Expr{
			SQL:  "? NOT LIKE ? ESCAPE ?",
			Vars: []any{e.Expression, pattern, string(escape[0])},
		}
	}
	return clause.Expr{
		SQL:  "? NOT LIKE ?",
		Vars: []any{e.Expression, pattern},
	}
}

// Contains 包含 (LIKE '%value%')
// SELECT * FROM users WHERE LOWER(email) LIKE '%@gmail.com%';
func (e TextExpr) Contains(value string) Expression {
	return e.Like("%" + value + "%")
}

// HasPrefix 前缀匹配 (LIKE 'value%')
// SELECT * FROM users WHERE UPPER(name) LIKE 'JOHN%';
func (e TextExpr) HasPrefix(value string) Expression {
	return e.Like(value + "%")
}

// HasSuffix 后缀匹配 (LIKE '%value')
// SELECT * FROM users WHERE LOWER(email) LIKE '%@example.com';
func (e TextExpr) HasSuffix(value string) Expression {
	return e.Like("%" + value)
}
