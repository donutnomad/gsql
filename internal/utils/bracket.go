package utils

import (
	"reflect"
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
	} else if IsLiteralFunctionName(sql) {
		return false
	}
	//else if strings.Contains(sql, "AS") || strings.Contains(sql, "as") {
	//	return false
	//}
	return true
}
