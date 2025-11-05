package utils

import (
	"fmt"
	"testing"
)

func TestExtractAS(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "(ROW_NUMBER() OVER(ORDER BY `products`.`price` DESC)) AS `rn`",
			expected: "ROW_NUMBER() OVER(ORDER BY `products`.`price` DESC)",
		},
		{
			input:    "(COUNT(*)) as count",
			expected: "COUNT(*)",
		},
		{
			input:    "name AS `user_name`",
			expected: "name",
		},
		{
			input:    "`users`.`id` AS id",
			expected: "`users`.`id`",
		},
		{
			input:    "(SUM(amount)) AS total",
			expected: "SUM(amount)",
		},
		{
			input:    "column_name as alias",
			expected: "column_name",
		},
		{
			input:    "simple_column",
			expected: "", // 没有 AS，应该匹配失败
		},
		{
			input:    "(COUNT(*))",
			expected: "", // 没有 AS，应该匹配失败
		},
		{
			input:    "`users`.`id`",
			expected: "", // 没有 AS，应该匹配失败
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			matches := aliasRegexp.FindStringSubmatch(tc.input)
			fmt.Printf("Input: %s\n", tc.input)
			fmt.Printf("Matches: %v\n", matches)

			var result string
			if len(matches) > 1 {
				// matches[1] 是括号内的内容，matches[2] 是非括号的内容
				if matches[1] != "" {
					result = matches[1]
				} else if matches[2] != "" {
					result = matches[2]
				}
			}

			fmt.Printf("Extracted: %s\n", result)
			fmt.Printf("Expected: %s\n\n", tc.expected)

			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestExtractAliasName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "column AS `user_name`",
			expected: "user_name",
		},
		{
			input:    "column AS user_name",
			expected: "user_name",
		},
		{
			input:    "(COUNT(*)) as count",
			expected: "count",
		},
		{
			input:    "(COUNT(*)) AS `count`",
			expected: "count",
		},
		{
			input:    "`users`.`id` AS id",
			expected: "id",
		},
		{
			input:    "(SUM(amount)) AS total",
			expected: "total",
		},
		{
			input:    "name as `rn`",
			expected: "rn",
		},
		{
			input:    "(ROW_NUMBER() OVER(ORDER BY price)) AS `row_num`",
			expected: "row_num",
		},
		{
			input:    "simple_column",
			expected: "", // 没有 AS，应该匹配失败
		},
		{
			input:    "column",
			expected: "", // 没有 AS，应该匹配失败
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := ExtractAliasName(tc.input)
			if result != tc.expected {
				t.Errorf("Input: %q\nExpected: %q\nGot: %q", tc.input, tc.expected, result)
			}
		})
	}
}
