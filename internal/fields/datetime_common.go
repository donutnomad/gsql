package fields

import (
	"fmt"
	"strconv"
	"strings"
)

// 允许的时间间隔单位
var allowedIntervalUnits = map[string]bool{
	"MICROSECOND": true,
	"SECOND":      true,
	"MINUTE":      true,
	"HOUR":        true,
	"DAY":         true,
	"WEEK":        true,
	"MONTH":       true,
	"QUARTER":     true,
	"YEAR":        true,
}

// parseInterval 解析并验证时间间隔格式
func parseInterval(interval string, funcName string) string {
	parts := strings.Fields(interval)
	if len(parts) != 2 {
		panic(fmt.Sprintf("%s: invalid interval format, expected '<number> <unit>' (e.g., '1 DAY')", funcName))
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(fmt.Sprintf("%s: interval value must be a number, got: %s", funcName, parts[0]))
	}

	unit := strings.ToUpper(parts[1])
	if !allowedIntervalUnits[unit] {
		panic(fmt.Sprintf("%s: invalid interval unit: %s", funcName, unit))
	}

	return fmt.Sprintf("%d %s", num, unit)
}
