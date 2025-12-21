package gsql

import (
	"strings"
	"testing"

	"github.com/donutnomad/gsql/field"
)

// TestDateAddSQLInjectionPrevention 测试 DATE_ADD 函数的 SQL 注入防护
func TestDateAddSQLInjectionPrevention(t *testing.T) {
	dateField := field.NewComparable[string]("users", "created_at")

	tests := []struct {
		name        string
		interval    string
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "正常输入",
			interval:    "1 DAY",
			shouldPanic: false,
		},
		{
			name:        "正常输入 - 多个空格",
			interval:    "1   DAY",
			shouldPanic: false, // strings.Fields 会自动处理多个空格
		},
		{
			name:        "SQL注入尝试 - 额外的SQL",
			interval:    "1 DAY) OR 1=1 --",
			shouldPanic: true,
			panicMsg:    "invalid interval format",
		},
		{
			name:        "SQL注入尝试 - 非法单位",
			interval:    "1 DAYS",
			shouldPanic: true,
			panicMsg:    "invalid interval unit",
		},
		{
			name:        "SQL注入尝试 - 非法数字",
			interval:    "abc DAY",
			shouldPanic: true,
			panicMsg:    "interval value must be a number",
		},
		{
			name:        "空字符串",
			interval:    "",
			shouldPanic: true,
			panicMsg:    "invalid interval format",
		},
		{
			name:        "只有数字",
			interval:    "1",
			shouldPanic: true,
			panicMsg:    "invalid interval format",
		},
		{
			name:        "只有单位",
			interval:    "DAY",
			shouldPanic: true,
			panicMsg:    "invalid interval format",
		},
		{
			name:        "负数",
			interval:    "-1 DAY",
			shouldPanic: false, // 负数是合法的
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						panicMsg := r.(string)
						if !strings.Contains(panicMsg, tt.panicMsg) {
							t.Errorf("期望 panic 消息包含 %q, 但得到 %q", tt.panicMsg, panicMsg)
						}
					} else {
						t.Error("期望 panic 但没有发生")
					}
				}()
			}
			result := DATE_ADD(dateField.ToExpr(), tt.interval)
			if !tt.shouldPanic {
				// 验证函数执行成功
				_ = result
			}
		})
	}
}

// TestDateSubSQLInjectionPrevention 测试 DATE_SUB 函数的 SQL 注入防护
func TestDateSubSQLInjectionPrevention(t *testing.T) {
	dateField := field.NewComparable[string]("users", "created_at")

	tests := []struct {
		name        string
		interval    string
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "正常输入",
			interval:    "7 DAY",
			shouldPanic: false,
		},
		{
			name:        "SQL注入尝试",
			interval:    "1 DAY); DROP TABLE users; --",
			shouldPanic: true,
			panicMsg:    "invalid interval format",
		},
		{
			name:        "非法单位",
			interval:    "1 UNKNOWN",
			shouldPanic: true,
			panicMsg:    "invalid interval unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						panicMsg := r.(string)
						if !strings.Contains(panicMsg, tt.panicMsg) {
							t.Errorf("期望 panic 消息包含 %q, 但得到 %q", tt.panicMsg, panicMsg)
						}
					} else {
						t.Error("期望 panic 但没有发生")
					}
				}()
			}
			_ = DATE_SUB(dateField.ToExpr(), tt.interval)
		})
	}
}

// TestTimestampDiffSQLInjectionPrevention 测试 TIMESTAMPDIFF 函数的 SQL 注入防护
func TestTimestampDiffSQLInjectionPrevention(t *testing.T) {
	expr1 := field.NewComparable[string]("users", "created_at")
	expr2 := field.NewComparable[string]("users", "updated_at")

	tests := []struct {
		name        string
		unit        string
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "正常输入 - 小写",
			unit:        "day",
			shouldPanic: false,
		},
		{
			name:        "正常输入 - 大写",
			unit:        "DAY",
			shouldPanic: false,
		},
		{
			name:        "SQL注入尝试",
			unit:        "DAY); DROP TABLE users; --",
			shouldPanic: true,
			panicMsg:    "invalid unit",
		},
		{
			name:        "非法单位",
			unit:        "DAYS",
			shouldPanic: true,
			panicMsg:    "invalid unit",
		},
		{
			name:        "前后有空格",
			unit:        "  DAY  ",
			shouldPanic: false, // TrimSpace 会处理
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						panicMsg := r.(string)
						if !strings.Contains(panicMsg, tt.panicMsg) {
							t.Errorf("期望 panic 消息包含 %q, 但得到 %q", tt.panicMsg, panicMsg)
						}
					} else {
						t.Error("期望 panic 但没有发生")
					}
				}()
			}
			_ = TIMESTAMPDIFF(tt.unit, expr1.ToExpr(), expr2.ToExpr())
		})
	}
}

// TestConvertCharsetSQLInjectionPrevention 测试 CONVERT_CHARSET 函数的 SQL 注入防护
func TestConvertCharsetSQLInjectionPrevention(t *testing.T) {
	expr := field.NewComparable[string]("users", "name")

	tests := []struct {
		name        string
		charset     string
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "正常输入 - 小写",
			charset:     "utf8",
			shouldPanic: false,
		},
		{
			name:        "正常输入 - 大写",
			charset:     "UTF8",
			shouldPanic: false, // ToLower 会处理
		},
		{
			name:        "SQL注入尝试",
			charset:     "utf8); DROP TABLE users; --",
			shouldPanic: true,
			panicMsg:    "invalid or unsupported charset",
		},
		{
			name:        "非法字符集",
			charset:     "unknown",
			shouldPanic: true,
			panicMsg:    "invalid or unsupported charset",
		},
		{
			name:        "前后有空格",
			charset:     "  utf8  ",
			shouldPanic: false, // TrimSpace 会处理
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						panicMsg := r.(string)
						if !strings.Contains(panicMsg, tt.panicMsg) {
							t.Errorf("期望 panic 消息包含 %q, 但得到 %q", tt.panicMsg, panicMsg)
						}
					} else {
						t.Error("期望 panic 但没有发生")
					}
				}()
			}
			_ = CONVERT_CHARSET(expr.ToExpr(), tt.charset)
		})
	}
}

// TestAllowedIntervalUnits 测试时间间隔单位白名单的完整性
func TestAllowedIntervalUnits(t *testing.T) {
	expectedUnits := []string{
		"MICROSECOND", "SECOND", "MINUTE", "HOUR", "DAY", "WEEK", "MONTH", "QUARTER", "YEAR",
		"SECOND_MICROSECOND", "MINUTE_MICROSECOND", "MINUTE_SECOND",
		"HOUR_MICROSECOND", "HOUR_SECOND", "HOUR_MINUTE",
		"DAY_MICROSECOND", "DAY_SECOND", "DAY_MINUTE", "DAY_HOUR",
		"YEAR_MONTH",
	}

	for _, unit := range expectedUnits {
		if !allowedIntervalUnits[unit] {
			t.Errorf("预期的时间单位 %q 不在白名单中", unit)
		}
	}
}

// TestAllowedCharsets 测试字符集白名单的完整性
func TestAllowedCharsets(t *testing.T) {
	expectedCharsets := []string{
		"utf8", "utf8mb4", "latin1", "gbk", "ascii", "binary", "ucs2", "utf16", "utf32",
	}

	for _, charset := range expectedCharsets {
		if !allowedCharsets[charset] {
			t.Errorf("预期的字符集 %q 不在白名单中", charset)
		}
	}
}
