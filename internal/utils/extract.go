package utils

import (
	"regexp"
	"strings"
)

// aliasRegexp 匹配 (xxx) AS `xxx` 或 xxx AS xxx 等情况，并提取前面的 xxx
// 第一个捕获组: 括号内的内容 (如果有括号)
// 第二个捕获组: 非括号的内容 (如果没有括号)
var aliasRegexp = regexp.MustCompile(`(?i)\((.+)\)\s+AS\s+` + "`" + `?\w+` + "`" + `?|(\S+)\s+AS\s+` + "`" + `?\w+` + "`" + `?`)

// aliasNameRegexp 匹配 AS 后面的别名，支持 AS `xxx` 或 AS xxx 格式
// 捕获组: AS 后面的别名名称（不包含反引号）
var aliasNameRegexp = regexp.MustCompile(`(?i)AS\s+` + "`" + `?(\w+)` + "`" + `?`)

func ExtractSQLFromAS(sql string) string {
	matches := aliasRegexp.FindStringSubmatch(sql)
	if len(matches) > 0 {
		return matches[1]
	}
	return ""
}

// ExtractAliasName 从 SQL 语句中提取 AS 后面的别名
// 例如: "column AS `user_name`" -> "user_name"
//
//	"column AS total" -> "total"
func ExtractAliasName(sql string) string {
	matches := aliasNameRegexp.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func AutoParentheses(sql string) string {
	if !strings.HasPrefix(sql, "(") {
		return "(" + sql + ")"
	}
	return sql
}
