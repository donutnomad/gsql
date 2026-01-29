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

func TestBetweenBuild(t *testing.T) {
	tests := []struct {
		name        string
		between     Between
		expectedSQL string
	}{
		{
			name: "simple values",
			between: Between{
				Column: testExpr{sql: "`users`.`age`"},
				From:   10,
				To:     20,
			},
			expectedSQL: "`users`.`age` BETWEEN ? AND ?",
		},
		{
			name: "expression values - arithmetic",
			between: Between{
				Column: testExpr{sql: "`users`.`age`"},
				From:   testExpr{sql: "a + b"},
				To:     testExpr{sql: "c - d"},
			},
			expectedSQL: "`users`.`age` BETWEEN (a + b) AND (c - d)",
		},
		{
			name: "column expressions",
			between: Between{
				Column: testExpr{sql: "`orders`.`amount`"},
				From:   testExpr{sql: "`limits`.`min`"},
				To:     testExpr{sql: "`limits`.`max`"},
			},
			expectedSQL: "`orders`.`amount` BETWEEN `limits`.`min` AND `limits`.`max`",
		},
		{
			name: "function expressions",
			between: Between{
				Column: testExpr{sql: "`events`.`created_at`"},
				From:   testExpr{sql: "DATE_SUB(NOW())"},
				To:     testExpr{sql: "NOW()"},
			},
			expectedSQL: "`events`.`created_at` BETWEEN DATE_SUB(NOW()) AND NOW()",
		},
		{
			name: "mixed value and expression",
			between: Between{
				Column: testExpr{sql: "`products`.`price`"},
				From:   100,
				To:     testExpr{sql: "base_price * 2"},
			},
			expectedSQL: "`products`.`price` BETWEEN ? AND (base_price * 2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &testMemBuilder{}
			tt.between.Build(mb)
			got := mb.sql
			if got != tt.expectedSQL {
				t.Errorf("Between.Build() produced SQL = %q, want %q", got, tt.expectedSQL)
			}
		})
	}
}

func TestNotBetweenBuild(t *testing.T) {
	tests := []struct {
		name        string
		notBetween  NotBetween
		expectedSQL string
	}{
		{
			name: "simple values",
			notBetween: NotBetween{
				Column: testExpr{sql: "`users`.`age`"},
				From:   10,
				To:     20,
			},
			expectedSQL: "`users`.`age` NOT BETWEEN ? AND ?",
		},
		{
			name: "expression values - arithmetic",
			notBetween: NotBetween{
				Column: testExpr{sql: "`users`.`age`"},
				From:   testExpr{sql: "a + b"},
				To:     testExpr{sql: "c - d"},
			},
			expectedSQL: "`users`.`age` NOT BETWEEN (a + b) AND (c - d)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &testMemBuilder{}
			tt.notBetween.Build(mb)
			got := mb.sql
			if got != tt.expectedSQL {
				t.Errorf("NotBetween.Build() produced SQL = %q, want %q", got, tt.expectedSQL)
			}
		})
	}
}

func TestExprWithExpressionVars(t *testing.T) {
	// 测试 Expr 处理表达式变量时的行为
	// 验证算术表达式如 "? / ?" 中的表达式参数是否正确添加括号
	tests := []struct {
		name        string
		expr        Expr
		expectedSQL string
	}{
		{
			name: "division with simple values",
			expr: Expr{
				SQL:  "? / ?",
				Vars: []any{testExpr{sql: "`a`.`x`"}, 10},
			},
			expectedSQL: "`a`.`x` / ?",
		},
		{
			name: "division with arithmetic expression",
			expr: Expr{
				SQL:  "? / ?",
				Vars: []any{testExpr{sql: "`a`.`x`"}, testExpr{sql: "b + c"}},
			},
			expectedSQL: "`a`.`x` / (b + c)",
		},
		{
			name: "addition with both arithmetic expressions",
			expr: Expr{
				SQL:  "? + ?",
				Vars: []any{testExpr{sql: "a * b"}, testExpr{sql: "c - d"}},
			},
			expectedSQL: "(a * b) + (c - d)",
		},
		{
			name: "multiplication with column and function",
			expr: Expr{
				SQL:  "? * ?",
				Vars: []any{testExpr{sql: "`price`"}, testExpr{sql: "COUNT(*)"}},
			},
			expectedSQL: "`price` * COUNT(*)",
		},
		{
			name: "complex nested expression",
			expr: Expr{
				SQL:  "(? + ?) / ?",
				Vars: []any{testExpr{sql: "`a`"}, testExpr{sql: "`b`"}, testExpr{sql: "x + y"}},
			},
			// 第一个和第二个 ? 在括号内，不需要额外括号
			// 第三个 ? 不在括号内，需要括号
			expectedSQL: "(`a` + `b`) / (x + y)",
		},
		{
			name: "WithoutParentheses flag",
			expr: Expr{
				SQL:                "? + ?",
				Vars:               []any{testExpr{sql: "a * b"}, testExpr{sql: "c - d"}},
				WithoutParentheses: true,
			},
			// WithoutParentheses 标志禁用自动括号
			expectedSQL: "a * b + c - d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &testMemBuilder{}
			tt.expr.Build(mb)
			got := mb.sql
			if got != tt.expectedSQL {
				t.Errorf("Expr.Build() produced SQL = %q, want %q", got, tt.expectedSQL)
			}
		})
	}
}

func TestExprModOperator(t *testing.T) {
	// 测试 MOD 操作符处理表达式变量时的行为
	tests := []struct {
		name        string
		expr        Expr
		expectedSQL string
	}{
		{
			name: "MOD with simple values",
			expr: Expr{
				SQL:  "? MOD ?",
				Vars: []any{testExpr{sql: "`id`"}, 2},
			},
			expectedSQL: "`id` MOD ?",
		},
		{
			name: "MOD with arithmetic expression",
			expr: Expr{
				SQL:  "? MOD ?",
				Vars: []any{testExpr{sql: "`a` + `b`"}, testExpr{sql: "c - d"}},
			},
			// 复杂表达式需要括号
			expectedSQL: "(`a` + `b`) MOD (c - d)",
		},
		{
			name: "MOD with column expressions",
			expr: Expr{
				SQL:  "? MOD ?",
				Vars: []any{testExpr{sql: "`users`.`id`"}, testExpr{sql: "`config`.`value`"}},
			},
			// 简单列名不需要括号
			expectedSQL: "`users`.`id` MOD `config`.`value`",
		},
		{
			name: "MOD with function",
			expr: Expr{
				SQL:  "? MOD ?",
				Vars: []any{testExpr{sql: "ABS(`value`)"}, 10},
			},
			expectedSQL: "ABS(`value`) MOD ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &testMemBuilder{}
			tt.expr.Build(mb)
			got := mb.sql
			if got != tt.expectedSQL {
				t.Errorf("Expr.Build() produced SQL = %q, want %q", got, tt.expectedSQL)
			}
		})
	}
}

func TestExprBetweenWithExpressions(t *testing.T) {
	// 测试原来的 BETWEEN 模板是否能正确处理表达式
	tests := []struct {
		name        string
		expr        Expr
		expectedSQL string
	}{
		{
			name: "BETWEEN with simple values",
			expr: Expr{
				SQL:  "? BETWEEN ? AND ?",
				Vars: []any{testExpr{sql: "`age`"}, 10, 20},
			},
			expectedSQL: "`age` BETWEEN ? AND ?",
		},
		{
			name: "BETWEEN with arithmetic expressions",
			expr: Expr{
				SQL:  "? BETWEEN ? AND ?",
				Vars: []any{testExpr{sql: "`age`"}, testExpr{sql: "a + b"}, testExpr{sql: "c - d"}},
			},
			expectedSQL: "`age` BETWEEN (a + b) AND (c - d)",
		},
		{
			name: "BETWEEN with IF function - real world case",
			expr: Expr{
				SQL: "? BETWEEN ? AND ?",
				Vars: []any{
					testExpr{sql: "FROM_UNIXTIME(`order`.`block_time`)"},
					testExpr{sql: "`log`.`created_at`"},
					testExpr{sql: "IF(`log`.`bind` = TRUE, '2026-01-29', `log`.`unbind_at`)"},
				},
			},
			// IF() 是函数，不加括号
			expectedSQL: "FROM_UNIXTIME(`order`.`block_time`) BETWEEN `log`.`created_at` AND IF(`log`.`bind` = TRUE, '2026-01-29', `log`.`unbind_at`)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &testMemBuilder{}
			tt.expr.Build(mb)
			got := mb.sql
			if got != tt.expectedSQL {
				t.Errorf("Expr.Build() produced SQL = %q, want %q", got, tt.expectedSQL)
			}
		})
	}
}

func TestExprWithComplexExpressions(t *testing.T) {
	// 测试复杂表达式如 IF、CASE WHEN 作为变量时的行为
	tests := []struct {
		name        string
		expr        Expr
		between     Between
		useBetween  bool
		expectedSQL string
	}{
		{
			name: "IF expression in arithmetic",
			expr: Expr{
				SQL:  "? + ?",
				Vars: []any{testExpr{sql: "`base`"}, testExpr{sql: "IF(a > 0, a, 0)"}},
			},
			// IF() 是函数调用形式，不需要括号
			expectedSQL: "`base` + IF(a > 0, a, 0)",
		},
		{
			name: "CASE WHEN in arithmetic",
			expr: Expr{
				SQL:  "? * ?",
				Vars: []any{testExpr{sql: "`price`"}, testExpr{sql: "CASE WHEN type = 1 THEN 0.9 ELSE 1 END"}},
			},
			// CASE WHEN 不是函数形式，需要括号
			expectedSQL: "`price` * (CASE WHEN type = 1 THEN 0.9 ELSE 1 END)",
		},
		{
			name: "CASE WHEN in BETWEEN",
			between: Between{
				Column: testExpr{sql: "`amount`"},
				From:   testExpr{sql: "CASE WHEN level = 1 THEN 0 ELSE 100 END"},
				To:     testExpr{sql: "CASE WHEN level = 1 THEN 1000 ELSE 5000 END"},
			},
			expectedSQL: "`amount` BETWEEN (CASE WHEN level = 1 THEN 0 ELSE 100 END) AND (CASE WHEN level = 1 THEN 1000 ELSE 5000 END)",
			useBetween:  true,
		},
		{
			name: "subquery in arithmetic",
			expr: Expr{
				SQL:  "? / ?",
				Vars: []any{testExpr{sql: "`total`"}, testExpr{sql: "SELECT COUNT(*) FROM users"}},
			},
			// 子查询需要括号
			expectedSQL: "`total` / (SELECT COUNT(*) FROM users)",
		},
		{
			name: "already wrapped subquery",
			expr: Expr{
				SQL:  "? / ?",
				Vars: []any{testExpr{sql: "`total`"}, testExpr{sql: "(SELECT COUNT(*) FROM users)"}},
			},
			// 已经有括号的子查询不重复添加
			expectedSQL: "`total` / (SELECT COUNT(*) FROM users)",
		},
		{
			name: "COALESCE function",
			expr: Expr{
				SQL:  "? + ?",
				Vars: []any{testExpr{sql: "COALESCE(`a`, 0)"}, testExpr{sql: "COALESCE(`b`, 0)"}},
			},
			// 函数调用不需要括号
			expectedSQL: "COALESCE(`a`, 0) + COALESCE(`b`, 0)",
		},
		{
			name: "NULLIF function",
			expr: Expr{
				SQL:  "? / ?",
				Vars: []any{testExpr{sql: "`total`"}, testExpr{sql: "NULLIF(`count`, 0)"}},
			},
			expectedSQL: "`total` / NULLIF(`count`, 0)",
		},
		{
			name: "nested CASE in IF",
			expr: Expr{
				SQL:  "? + ?",
				Vars: []any{testExpr{sql: "`base`"}, testExpr{sql: "IF(x > 0, CASE WHEN y = 1 THEN 10 ELSE 20 END, 0)"}},
			},
			// IF() 外层是函数形式
			expectedSQL: "`base` + IF(x > 0, CASE WHEN y = 1 THEN 10 ELSE 20 END, 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &testMemBuilder{}
			if tt.useBetween {
				tt.between.Build(mb)
			} else {
				tt.expr.Build(mb)
			}
			got := mb.sql
			if got != tt.expectedSQL {
				t.Errorf("Build() produced SQL = %q, want %q", got, tt.expectedSQL)
			}
		})
	}
}
