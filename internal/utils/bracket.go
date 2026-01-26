package utils

import (
	"reflect"

	"github.com/donutnomad/gsql/clause"
	gormclause "gorm.io/gorm/clause"
)

func AddVarAutoBracket(builder gormclause.Builder, values []any) {
	needBracket := IsNeedParentheses(values)
	if needBracket {
		_ = builder.WriteByte('(')
	}
	builder.AddVar(builder, values...)
	if needBracket {
		_ = builder.WriteByte(')')
	}
}

func WriteExpr(builder gormclause.Builder, expr gormclause.Expression) {
	parentheses := !clause.IsSimpleExpression(expr)
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
