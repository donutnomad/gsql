package utils

import "testing"

func TestIsFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// 正确的函数调用
		{
			name:     "JSON_ARRAY with nested parentheses",
			input:    "JSON_ARRAY((?,?))",
			expected: true,
		},
		{
			name:     "CONCAT with parameters",
			input:    "CONCAT(?, ?)",
			expected: true,
		},
		{
			name:     "UPPER with single parameter",
			input:    "UPPER(name)",
			expected: true,
		},
		{
			name:     "COUNT with star",
			input:    "COUNT(*)",
			expected: true,
		},
		{
			name:     "SUBSTRING with multiple params",
			input:    "SUBSTRING(?, 1, 10)",
			expected: true,
		},
		{
			name:     "Function with underscore",
			input:    "DATE_FORMAT(NOW(), '%Y-%m-%d')",
			expected: true,
		},
		{
			name:     "Nested function calls",
			input:    "UPPER(CONCAT(?, ?))",
			expected: true,
		},
		{
			name:     "Empty parameters",
			input:    "NOW()",
			expected: true,
		},

		// 错误的函数调用
		{
			name:     "No closing parenthesis",
			input:    "CONCAT(?, ?",
			expected: false,
		},
		{
			name:     "No opening parenthesis",
			input:    "CONCAT?, ?)",
			expected: false,
		},
		{
			name:     "Lowercase function name",
			input:    "concat(?, ?)",
			expected: false,
		},
		{
			name:     "Mixed case function name",
			input:    "Concat(?, ?)",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Only function name",
			input:    "CONCAT",
			expected: false,
		},
		{
			name:     "Unbalanced parentheses - more open",
			input:    "CONCAT((?, ?)",
			expected: false,
		},
		{
			name:     "Unbalanced parentheses - more close",
			input:    "CONCAT(?, ?))",
			expected: false,
		},
		{
			name:     "Function name with special chars",
			input:    "CON-CAT(?, ?)",
			expected: false,
		},
		{
			name:     "Function name with numbers",
			input:    "CONCAT2(?, ?)",
			expected: false,
		},
		{
			name:     "No function name",
			input:    "(?, ?)",
			expected: false,
		},
		{
			name:     "Wrong order parentheses",
			input:    "CONCAT)?, ?(",
			expected: false,
		},
		{
			name:     "Too short",
			input:    "A()",
			expected: true, // 技术上是有效的，虽然很短
		},
		{
			name:     "Only two chars",
			input:    "A(",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFunctionCall(tt.input)
			if result != tt.expected {
				t.Errorf("IsFunctionCall(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
