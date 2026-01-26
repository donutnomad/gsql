package clause

import (
	"database/sql/driver"
	"reflect"
	"strings"

	"gorm.io/gorm/clause"
)

// NeedsParentheses 用于判断表达式是否需要括号包裹
// 子查询、复杂表达式实现此接口返回 true
type NeedsParentheses interface {
	NeedsParentheses() bool
}

// AddVarWithParens 智能添加变量，如果是需要括号的表达式则自动加括号
func AddVarWithParens(builder Builder, value any) {
	// 1. 先检查是否实现了 NeedsParentheses 接口（显式控制）
	if np, ok := value.(NeedsParentheses); ok {
		if np.NeedsParentheses() {
			builder.WriteByte('(')
			if expr, ok := value.(Expression); ok {
				expr.Build(builder)
			} else {
				builder.AddVar(builder, value)
			}
			builder.WriteByte(')')
			return
		}
		// 明确不需要括号的表达式，直接构建
		if expr, ok := value.(Expression); ok {
			expr.Build(builder)
			return
		}
	}

	// 2. 检查是否是 Expression 类型，使用 SQL 字符串分析判断是否需要括号
	if expr, ok := value.(Expression); ok {
		if isSimpleExpression(expr) {
			// 简单表达式直接构建，不加括号
			expr.Build(builder)
		} else {
			// 复杂表达式（如子查询）需要括号
			builder.WriteByte('(')
			expr.Build(builder)
			builder.WriteByte(')')
		}
		return
	}

	// 3. 普通值直接添加
	builder.AddVar(builder, value)
}

// IsSimpleExpression 通过构建 SQL 字符串来判断表达式是否是简单表达式
// 简单表达式不需要括号，包括：单个列名、函数调用、字面量等
// 导出此函数供其他包使用，避免代码重复
func IsSimpleExpression(expr Expression) bool {
	mb := &memBuilder{sql: &strings.Builder{}}
	expr.Build(mb)
	sql := mb.sql.String()
	vars := mb.vars

	// 1. 空字符串
	if sql == "" {
		return true
	}

	// 2. 已被完整括号包裹（如子查询已经有括号）
	if isWrappedInParens(sql) {
		return true
	}

	// 3. 表名.* 形式
	if strings.HasSuffix(sql, ".*") {
		return true
	}

	// 4. 单个占位符且是简单值
	if sql == "?" && len(vars) <= 1 {
		return true
	}

	// 5. 函数调用（大写函数名 + 平衡括号）
	if isFunctionCall(sql) {
		return true
	}

	// 6. 反引号列名
	if isQuotedColumn(sql) {
		return true
	}

	// 7. SQL 关键字字面量
	if isSQLKeyword(sql) {
		return true
	}

	// 8. 数字字面量（整数或浮点）
	if isNumericLiteral(sql) {
		return true
	}

	// 9. 字符串字面量 'text'
	if isStringLiteral(sql) {
		return true
	}

	// 其他情况需要括号
	return false
}

// isSimpleExpression 是 IsSimpleExpression 的内部别名，保持向后兼容
func isSimpleExpression(expr Expression) bool {
	return IsSimpleExpression(expr)
}

// isWrappedInParens 检查字符串是否被一对完整的括号包裹
// "(SELECT ...)" -> true
// "(a) + (b)" -> false（括号在中间闭合）
func isWrappedInParens(sql string) bool {
	if len(sql) < 2 || sql[0] != '(' || sql[len(sql)-1] != ')' {
		return false
	}
	depth := 0
	for i := range len(sql) {
		if sql[i] == '(' {
			depth++
		} else if sql[i] == ')' {
			depth--
		}
		// 如果在最后一个字符之前 depth 变成 0，说明括号在中间闭合
		if depth == 0 && i < len(sql)-1 {
			return false
		}
	}
	return depth == 0
}

// isSQLKeyword 检查是否是 SQL 关键字字面量
func isSQLKeyword(sql string) bool {
	upper := strings.ToUpper(sql)
	switch upper {
	case "NULL", "TRUE", "FALSE",
		"CURRENT_TIMESTAMP", "CURRENT_DATE", "CURRENT_TIME",
		"CURRENT_USER", "NOW":
		return true
	}
	return false
}

// isNumericLiteral 检查是否是数字字面量
func isNumericLiteral(sql string) bool {
	if len(sql) == 0 {
		return false
	}
	// 允许负号开头
	start := 0
	if sql[0] == '-' {
		start = 1
	}
	if start >= len(sql) {
		return false
	}
	// 检查是否是有效数字
	hasDigit := false
	hasDot := false
	for i := start; i < len(sql); i++ {
		c := sql[i]
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c == '.' && !hasDot {
			hasDot = true
		} else {
			return false
		}
	}
	return hasDigit
}

// isStringLiteral 检查是否是字符串字面量 'text'
func isStringLiteral(sql string) bool {
	return len(sql) >= 2 && sql[0] == '\'' && sql[len(sql)-1] == '\''
}

// memBuilder 简化的内存构建器，用于构建 SQL 字符串
type memBuilder struct {
	sql  *strings.Builder
	vars []any
}

func (m *memBuilder) WriteByte(b byte) error {
	return m.sql.WriteByte(b)
}

func (m *memBuilder) WriteString(s string) (int, error) {
	return m.sql.WriteString(s)
}

func (m *memBuilder) WriteQuoted(field any) {
	switch v := field.(type) {
	case clause.Column:
		if v.Table != "" {
			m.sql.WriteByte('`')
			m.sql.WriteString(v.Table)
			m.sql.WriteString("`.`")
			m.sql.WriteString(v.Name)
			m.sql.WriteByte('`')
		} else {
			m.sql.WriteByte('`')
			m.sql.WriteString(v.Name)
			m.sql.WriteByte('`')
		}
	case string:
		m.sql.WriteByte('`')
		m.sql.WriteString(v)
		m.sql.WriteByte('`')
	default:
		m.sql.WriteString("?")
	}
}

func (m *memBuilder) AddVar(writer clause.Writer, vars ...any) {
	for i, v := range vars {
		if i > 0 {
			writer.WriteByte(',')
		}
		if expr, ok := v.(Expression); ok {
			expr.Build(m)
		} else {
			m.vars = append(m.vars, v)
			writer.WriteByte('?')
		}
	}
}

func (m *memBuilder) AddError(err error) error {
	return err
}

// isFunctionCall 检测字符串是否是一个函数调用表达式
func isFunctionCall(s string) bool {
	if len(s) < 3 {
		return false
	}
	if s[len(s)-1] != ')' {
		return false
	}
	openIndex := strings.IndexByte(s, '(')
	if openIndex <= 0 {
		return false
	}
	funcName := s[:openIndex]
	for _, r := range funcName {
		if !((r >= 'A' && r <= 'Z') || r == '_') {
			return false
		}
	}
	// 检查括号平衡
	openCount, closeCount := 0, 0
	for i := openIndex; i < len(s); i++ {
		if s[i] == '(' {
			openCount++
		} else if s[i] == ')' {
			closeCount++
		}
		if closeCount > openCount {
			return false
		}
	}
	return openCount == closeCount && openCount > 0
}

// isQuotedColumn 检查是否是简单的反引号包裹的列名
func isQuotedColumn(sql string) bool {
	if len(sql) < 3 {
		return false
	}
	if sql[0] != '`' || sql[len(sql)-1] != '`' {
		return false
	}
	backtickCount := strings.Count(sql, "`")
	if backtickCount != 2 && backtickCount != 4 {
		return false
	}
	for _, c := range sql {
		if c != '`' && c != '.' && !isValidColumnChar(c) {
			return false
		}
	}
	return true
}

func isValidColumnChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// Eq equal to for where
type Eq struct {
	Column Expression
	Value  any
}

func (eq Eq) Build(builder Builder) {
	eq.Column.Build(builder)

	switch eq.Value.(type) {
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []any:
		rv := reflect.ValueOf(eq.Value)
		if rv.Len() == 0 {
			builder.WriteString(" IN (NULL)")
		} else {
			builder.WriteString(" IN (")
			for i := 0; i < rv.Len(); i++ {
				if i > 0 {
					builder.WriteByte(',')
				}
				builder.AddVar(builder, rv.Index(i).Interface())
			}
			builder.WriteByte(')')
		}
	default:
		if eqNil(eq.Value) {
			builder.WriteString(" IS NULL")
		} else {
			builder.WriteString(" = ")
			AddVarWithParens(builder, eq.Value)
		}
	}
}

func (eq Eq) NegationBuild(builder Builder) {
	Neq(eq).Build(builder)
}

// Neq not equal to for where
type Neq Eq

func (neq Neq) Build(builder Builder) {
	neq.Column.Build(builder)

	switch neq.Value.(type) {
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []any:
		builder.WriteString(" NOT IN (")
		rv := reflect.ValueOf(neq.Value)
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.AddVar(builder, rv.Index(i).Interface())
		}
		builder.WriteByte(')')
	default:
		if eqNil(neq.Value) {
			builder.WriteString(" IS NOT NULL")
		} else {
			builder.WriteString(" <> ")
			AddVarWithParens(builder, neq.Value)
		}
	}
}

func (neq Neq) NegationBuild(builder Builder) {
	Eq(neq).Build(builder)
}

// Like whether string matches regular expression
type Like Eq

func (like Like) Build(builder Builder) {
	like.Column.Build(builder)
	builder.WriteString(" LIKE ")
	builder.AddVar(builder, like.Value)
}

func (like Like) NegationBuild(builder Builder) {
	builder.WriteQuoted(like.Column)
	builder.WriteString(" NOT LIKE ")
	builder.AddVar(builder, like.Value)
}

// Gt greater than for where
type Gt Eq

func (gt Gt) Build(builder Builder) {
	gt.Column.Build(builder)
	builder.WriteString(" > ")
	AddVarWithParens(builder, gt.Value)
}

func (gt Gt) NegationBuild(builder Builder) {
	Lte(gt).Build(builder)
}

// Gte greater than or equal to for where
type Gte Eq

func (gte Gte) Build(builder Builder) {
	gte.Column.Build(builder)
	builder.WriteString(" >= ")
	AddVarWithParens(builder, gte.Value)
}

func (gte Gte) NegationBuild(builder Builder) {
	Lt(gte).Build(builder)
}

// Lt less than for where
type Lt Eq

func (lt Lt) Build(builder Builder) {
	lt.Column.Build(builder)
	builder.WriteString(" < ")
	AddVarWithParens(builder, lt.Value)
}

func (lt Lt) NegationBuild(builder Builder) {
	Gte(lt).Build(builder)
}

// Lte less than or equal to for where
type Lte Eq

func (lte Lte) Build(builder Builder) {
	lte.Column.Build(builder)
	builder.WriteString(" <= ")
	AddVarWithParens(builder, lte.Value)
}

func (lte Lte) NegationBuild(builder Builder) {
	Gt(lte).Build(builder)
}

func eqNil(value any) bool {
	if valuer, ok := value.(driver.Valuer); ok && !eqNilReflect(valuer) {
		value, _ = valuer.Value()
	}

	return value == nil || eqNilReflect(value)
}

func eqNilReflect(value any) bool {
	reflectValue := reflect.ValueOf(value)
	return reflectValue.Kind() == reflect.Ptr && reflectValue.IsNil()
}
