package clause

import (
	"testing"
)

// 用于测试的简单表达式构建器
type testExpr struct {
	sql  string
	vars []any
}

func (e testExpr) Build(builder Builder) {
	for _, c := range e.sql {
		if c == '?' && len(e.vars) > 0 {
			builder.AddVar(builder, e.vars[0])
			e.vars = e.vars[1:]
		} else {
			builder.WriteByte(byte(c))
		}
	}
}

func TestIsWrappedInParens(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		// 被完整括号包裹的情况
		{"simple parens", "(abc)", true},
		{"subquery", "(SELECT id FROM users)", true},
		{"nested parens", "((a + b))", true},
		{"placeholder in parens", "(?)", true},

		// 不是完整括号包裹的情况
		{"no parens", "abc", false},
		{"parens in middle", "(a) + (b)", false},
		{"starts with paren only", "(a + b", false},
		{"ends with paren only", "a + b)", false},
		{"empty string", "", false},
		{"single char", "a", false},
		{"just open paren", "(", false},
		{"just close paren", ")", false},
		{"multiple groups", "(a)(b)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWrappedInParens(tt.sql)
			if got != tt.expected {
				t.Errorf("isWrappedInParens(%q) = %v, want %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestIsSQLKeyword(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		// SQL 关键字
		{"NULL uppercase", "NULL", true},
		{"NULL lowercase", "null", true},
		{"NULL mixed", "Null", true},
		{"TRUE", "TRUE", true},
		{"FALSE", "FALSE", true},
		{"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP", true},
		{"CURRENT_DATE", "CURRENT_DATE", true},
		{"CURRENT_TIME", "CURRENT_TIME", true},
		{"CURRENT_USER", "CURRENT_USER", true},
		{"NOW", "NOW", true},

		// 非 SQL 关键字
		{"column name", "id", false},
		{"function call", "COUNT(*)", false},
		{"random word", "hello", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSQLKeyword(tt.sql)
			if got != tt.expected {
				t.Errorf("isSQLKeyword(%q) = %v, want %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestIsNumericLiteral(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		// 有效的数字字面量
		{"integer", "100", true},
		{"zero", "0", true},
		{"negative integer", "-50", true},
		{"float", "3.14", true},
		{"negative float", "-0.5", true},
		{"large number", "1234567890", true},
		{"decimal starting with dot", ".5", true}, // MySQL 允许 .5 作为数字

		// 无效的数字字面量
		{"empty", "", false},
		{"just minus", "-", false},
		{"letter", "abc", false},
		{"mixed", "12ab", false},
		{"multiple dots", "1.2.3", false},
		{"column name", "`id`", false},
		{"expression", "1 + 2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNumericLiteral(tt.sql)
			if got != tt.expected {
				t.Errorf("isNumericLiteral(%q) = %v, want %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestIsStringLiteral(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		// 有效的字符串字面量
		{"simple string", "'hello'", true},
		{"empty string", "''", true},
		{"string with space", "'hello world'", true},

		// 无效的字符串字面量
		{"no quotes", "hello", false},
		{"double quotes", "\"hello\"", false},
		{"backticks", "`hello`", false},
		{"only open quote", "'hello", false},
		{"only close quote", "hello'", false},
		{"empty", "", false},
		{"single char", "'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStringLiteral(tt.sql)
			if got != tt.expected {
				t.Errorf("isStringLiteral(%q) = %v, want %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestIsFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		// 有效的函数调用
		{"COUNT", "COUNT(*)", true},
		{"SUM", "SUM(amount)", true},
		{"UPPER", "UPPER(?)", true},
		{"nested function", "CONCAT(UPPER(?),?)", true},
		{"JSON_OBJECT", "JSON_OBJECT(?,?)", true},
		{"function with underscore", "DATE_FORMAT(?)", true},

		// 无效的函数调用
		{"lowercase function", "count(*)", false},
		{"no parens", "COUNT", false},
		{"empty parens name", "()", false},
		{"column name", "`id`", false},
		{"expression", "a + b", false},
		{"unbalanced parens", "COUNT(()", false},
		{"short string", "A(", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFunctionCall(tt.sql)
			if got != tt.expected {
				t.Errorf("isFunctionCall(%q) = %v, want %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestIsQuotedColumn(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		// 有效的反引号列名
		{"simple column", "`id`", true},
		{"column with underscore", "`user_id`", true},
		{"table.column", "`users`.`id`", true},

		// 无效的情况
		{"no backticks", "id", false},
		{"single quotes", "'id'", false},
		{"double quotes", "\"id\"", false},
		{"expression", "`a` + `b`", false},
		{"three parts", "`a`.`b`.`c`", false},
		{"empty", "", false},
		{"just backticks", "``", false},
		{"special char", "`id-name`", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isQuotedColumn(tt.sql)
			if got != tt.expected {
				t.Errorf("isQuotedColumn(%q) = %v, want %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestIsSimpleExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     Expression
		expected bool
	}{
		// 简单表达式（不需要括号）
		{
			name:     "empty expression",
			expr:     testExpr{sql: ""},
			expected: true,
		},
		{
			name:     "column reference with table",
			expr:     testExpr{sql: "`users`.`id`"},
			expected: true,
		},
		{
			name:     "simple column",
			expr:     testExpr{sql: "`id`"},
			expected: true,
		},
		{
			name:     "function call",
			expr:     testExpr{sql: "COUNT(*)"},
			expected: true,
		},
		{
			name:     "placeholder with value",
			expr:     testExpr{sql: "?", vars: []any{100}},
			expected: true,
		},
		{
			name:     "wrapped in parens",
			expr:     testExpr{sql: "(SELECT id FROM users)"},
			expected: true,
		},
		{
			name:     "NULL keyword",
			expr:     testExpr{sql: "NULL"},
			expected: true,
		},
		{
			name:     "numeric literal",
			expr:     testExpr{sql: "100"},
			expected: true,
		},
		{
			name:     "string literal",
			expr:     testExpr{sql: "'hello'"},
			expected: true,
		},
		{
			name:     "table star",
			expr:     testExpr{sql: "`users`.*"},
			expected: true,
		},

		// 复杂表达式（需要括号）
		{
			name:     "arithmetic expression",
			expr:     testExpr{sql: "a + b"},
			expected: false,
		},
		{
			name:     "subquery without parens",
			expr:     testExpr{sql: "SELECT id FROM users"},
			expected: false,
		},
		{
			name:     "comparison",
			expr:     testExpr{sql: "a > b"},
			expected: false,
		},
		{
			name:     "logical expression",
			expr:     testExpr{sql: "a AND b"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSimpleExpression(tt.expr)
			if got != tt.expected {
				t.Errorf("IsSimpleExpression() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAddVarWithParens(t *testing.T) {
	tests := []struct {
		name        string
		value       any
		expectedSQL string
	}{
		{
			name:        "simple integer",
			value:       100,
			expectedSQL: "?",
		},
		{
			name:        "column expression",
			value:       testExpr{sql: "`users`.`id`"},
			expectedSQL: "`users`.`id`",
		},
		{
			name:        "function expression",
			value:       testExpr{sql: "COUNT(*)"},
			expectedSQL: "COUNT(*)",
		},
		{
			name:        "subquery needs parens",
			value:       testExpr{sql: "SELECT id FROM users"},
			expectedSQL: "(SELECT id FROM users)",
		},
		{
			name:        "arithmetic needs parens",
			value:       testExpr{sql: "a + b"},
			expectedSQL: "(a + b)",
		},
		{
			name:        "already wrapped subquery",
			value:       testExpr{sql: "(SELECT id FROM users)"},
			expectedSQL: "(SELECT id FROM users)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &testMemBuilder{}
			AddVarWithParens(mb, tt.value)
			got := mb.sql
			if got != tt.expectedSQL {
				t.Errorf("AddVarWithParens() produced SQL = %q, want %q", got, tt.expectedSQL)
			}
		})
	}
}

// testMemBuilder 用于测试的简单 Builder 实现
type testMemBuilder struct {
	sql  string
	vars []any
}

func (m *testMemBuilder) WriteByte(b byte) error {
	m.sql += string(b)
	return nil
}

func (m *testMemBuilder) WriteString(s string) (int, error) {
	m.sql += s
	return len(s), nil
}

func (m *testMemBuilder) WriteQuoted(field any) {
	switch v := field.(type) {
	case Column:
		if v.Table != "" {
			m.sql += "`" + v.Table + "`.`" + v.Name + "`"
		} else {
			m.sql += "`" + v.Name + "`"
		}
	case string:
		m.sql += "`" + v + "`"
	}
}

func (m *testMemBuilder) AddVar(writer Writer, vars ...any) {
	for i, v := range vars {
		if i > 0 {
			m.sql += ","
		}
		if expr, ok := v.(Expression); ok {
			expr.Build(m)
		} else {
			m.vars = append(m.vars, v)
			m.sql += "?"
		}
	}
}

func (m *testMemBuilder) AddError(err error) error {
	return err
}
