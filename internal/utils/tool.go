package utils

import (
	"reflect"
	"strings"
)

func Optional[T any](args []T, def T) T {
	if len(args) == 0 {
		return def
	}
	return args[0]
}

func IsWindowFunction(s string) bool {
	if len(s) < 2 {
		return false
	}
	if s[0] == '(' {
		s = s[1:]
	}
	return strings.HasPrefix(s, "ROW_NUMBER()") || strings.HasPrefix(s, "RANK()") || strings.HasPrefix(s, "DENSE_RANK()")
}

func IsLiteralFunctionName(s string) bool {
	// 检查是否以 ) 结尾
	if len(s) > 2 && s[len(s)-1] == ')' {
		// 找到第一个 ( 的位置
		openIndex := -1
		openCount := 0
		for i := 0; i < len(s); i++ {
			if s[i] == '(' {
				openCount++
				if openIndex == -1 {
					openIndex = i
				}
			}
		}

		// 必须有且只有一个 (，并且位置大于 0（前面有内容）
		if openCount == 1 && openIndex > 0 {
			// 取 ( 之前的部分
			s = s[:openIndex]
		} else {
			return false
		}
	}

	// 检查剩余字符是否都是大写字母或下划线
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || r == '_' {
		} else {
			return false
		}
	}
	return len(s) > 0
}

func IsNumber(s any) bool {
	typ := reflect.TypeOf(s)
	switch {
	case typ.Kind() == reflect.Int:
	case typ.Kind() == reflect.Float64:
	case typ.Kind() == reflect.Float32:
	case typ.Kind() == reflect.Int8:
	case typ.Kind() == reflect.Int16:
	case typ.Kind() == reflect.Int32:
	case typ.Kind() == reflect.Int64:
	case typ.Kind() == reflect.Uint:
	case typ.Kind() == reflect.Uint8:
	case typ.Kind() == reflect.Uint16:
	case typ.Kind() == reflect.Uint32:
	case typ.Kind() == reflect.Uint64:
	default:
		return false
	}
	return true
}

func IsString(s any) bool {
	typ := reflect.TypeOf(s)
	return typ.Kind() == reflect.String
}

// IsFunctionCall 检测字符串是否是一个函数调用表达式
// 例如: JSON_ARRAY((?,?)), CONCAT(?, ?), UPPER(name) 等
func IsFunctionCall(s string) bool {
	if len(s) < 3 {
		return false
	}

	// 必须以 ) 结尾
	if s[len(s)-1] != ')' {
		return false
	}

	// 找到第一个 ( 的位置
	openIndex := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			openIndex = i
			break
		}
	}

	// 必须有左括号，且左括号前面有内容（函数名）
	if openIndex <= 0 {
		return false
	}

	// 检查函数名部分（左括号之前）是否都是大写字母或下划线
	funcName := s[:openIndex]
	for _, r := range funcName {
		if !((r >= 'A' && r <= 'Z') || r == '_') {
			return false
		}
	}

	// 检查括号是否平衡
	openCount := 0
	closeCount := 0
	for i := openIndex; i < len(s); i++ {
		if s[i] == '(' {
			openCount++
		} else if s[i] == ')' {
			closeCount++
		}
		// 右括号不能超过左括号
		if closeCount > openCount {
			return false
		}
	}

	// 左右括号数量必须相等
	return openCount == closeCount && openCount > 0
}
