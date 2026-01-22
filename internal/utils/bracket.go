package utils

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/donutnomad/gsql/clause"
)

func AddVarAutoBracket(builder clause.Builder, values []any) {
	needBracket := IsNeedParentheses(values)
	if needBracket {
		_ = builder.WriteByte('(')
	}
	builder.AddVar(builder, values...)
	if needBracket {
		_ = builder.WriteByte(')')
	}
}

func WriteExpr(builder clause.Builder, expr clause.Expression) {
	parentheses := IsNeedParenthesesExpr(expr)
	if parentheses {
		builder.WriteByte('(')
	}
	expr.Build(builder)
	if parentheses {
		builder.WriteByte(')')
	}
}

func IsNeedParentheses(vars []any) bool {
	for _, v := range vars {
		switch v := v.(type) {
		case []any:
			if len(v) > 0 {
				return false
			}
		default:
			switch rv := reflect.ValueOf(v); rv.Kind() {
			case reflect.Slice, reflect.Array:
				if rv.Len() == 0 {
				} else if rv.Type().Elem() == reflect.TypeOf(uint8(0)) {
				} else {
					return false
				}
			default:
			}
		}
	}
	return true
}

func IsNeedParenthesesExpr(expr clause.Expression) bool {
	var mb = NewMemoryBuilder()
	expr.Build(mb)
	var sql = mb.SQL.String()
	var vars = mb.Vars

	if sql == "(?)" {
		return false
	} else if strings.HasSuffix(sql, ".*") {
		return false
	} else if len(vars) == 1 && sql == "?" {
		arg := vars[0]
		if IsNumber(arg) || IsString(arg) {
			return false
		}
	} else if IsFunctionCall(sql) {
		return false
	} else if isQuotedColumn(sql) {
		// 简单的反引号包裹的列名不需要括号，如 `id` 或 `table`.`column`
		return false
	}
	// SELECT 1  这种1不需要加括号
	if _, e := strconv.ParseInt(sql, 10, 64); e == nil {
		return false
	}
	//else if strings.Contains(sql, "AS") || strings.Contains(sql, "as") {
	//	return false
	//}
	return true
}

// isQuotedColumn 检查是否是简单的反引号包裹的列名
// 匹配: `column` 或 `table`.`column`
func isQuotedColumn(sql string) bool {
	if len(sql) < 3 {
		return false
	}
	// 必须以反引号开始和结束
	if sql[0] != '`' || sql[len(sql)-1] != '`' {
		return false
	}
	// 计算反引号数量，必须是偶数（成对出现）
	backtickCount := 0
	for _, c := range sql {
		if c == '`' {
			backtickCount++
		}
	}
	// 简单列名: `column` (2个反引号)
	// 带表名: `table`.`column` (4个反引号)
	if backtickCount != 2 && backtickCount != 4 {
		return false
	}
	// 确保只包含有效字符：反引号、字母、数字、下划线、点
	for _, c := range sql {
		if c != '`' && c != '.' && !isValidColumnChar(c) {
			return false
		}
	}
	return true
}

// isValidColumnChar 检查是否是有效的列名字符
func isValidColumnChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
