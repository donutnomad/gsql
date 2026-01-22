package gsql_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/donutnomad/gsql/internal/utils"
)

// aliasRegexp 匹配 (xxx) AS `xxx` 或 xxx AS xxx 等情况，并提取前面的 xxx
// 第一个捕获组: 括号内的内容 (如果有括号)
// 第二个捕获组: 非括号的内容 (如果没有括号)
var aliasRegexp = regexp.MustCompile(`(?i)\((.+)\)\s+AS\s+` + "`" + `?\w+` + "`" + `?|(\S+)\s+AS\s+` + "`" + `?\w+` + "`" + `?`)

func TestGG(t *testing.T) {
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

func TestDD(t *testing.T) {
	b := utils.NewMemoryBuilder()
	bookF := gsql.AsJson(gsql.Expr("book"))

	bookF.Extract("$.name").Build(b)
	fmt.Println(b.SQL.String())
	fmt.Println(b.Vars)
	f := field.NewComparableFrom[string](
		gsql.AsJson(bookF).Extract("$.name").As(""),
	)
	_ = f
	f1 := field.NewComparableFrom[string](
		gsql.Field("JSON_EXTRACT(book, '$.name')").As(""),
	)
	_ = f1
	s := gsql.Select().From(gsql.TN("books")).Where(
		gsql.AsJson(bookF).Extract("$.name").Eq("2"),
	).ToSQL()
	fmt.Println(s)

	// 创建字段
	//price := field.NewComparable[float64]("products", "price")
	//
	//// 示例 1: 简单的 ROW_NUMBER，按价格降序排列
	//
	//rn := field.NewComparableWithField[int](
	//	gsql.RowNumber().
	//		OrderBy(price, true). // 按价格降序
	//		AsF("rn"),
	//)
	//
	//s2 := gsql.Select(rn).From(gsql.TableName2("books")).Where(rn.Eq(1)).ToSQL()

	//fmt.Println(s2)
}

// 演示 ROW_NUMBER() OVER() 窗口函数的使用
func ExampleRowNumber() {
	// 创建字段
	category := fields.NewTextExprField[string]("products", "category")
	price := fields.NewFloatExprField[float64]("products", "price")
	name := fields.NewTextExprField[string]("products", "name")

	products := &gsql.Table{Name: "products"}

	// 示例 1: 简单的 ROW_NUMBER，按价格降序排列
	rn1 := gsql.RowNumber().
		OrderBy(price.Desc()). // 按价格降序
		AsF("row_num")

	query1 := gsql.Select(name, price, rn1).
		From(products)

	fmt.Println("示例 1 - 简单排序:")
	fmt.Println(query1.ToSQL())

	// 示例 2: 带 PARTITION BY 的 ROW_NUMBER，在每个分类内按价格排序
	rn2 := gsql.RowNumber().
		PartitionBy(category). // 按类别分组
		OrderBy(price.Desc()). // 每组内按价格降序
		AsF("row_num")

	query2 := gsql.Select(category, name, price, rn2).
		From(products)

	fmt.Println("\n示例 2 - 分组排序:")
	fmt.Println(query2.ToSQL())

	// 示例 3: 多个 PARTITION BY 字段
	createdAt := field.NewComparable[string]("products", "created_at")
	rn3 := gsql.RowNumber().
		PartitionBy(category, createdAt). // 按多个字段分组
		OrderBy(price.Asc()).             // 升序排序
		AsF("row_num")

	query3 := gsql.Select(category, createdAt, name, price, rn3).
		From(products)

	fmt.Println("\n示例 3 - 多字段分组:")
	fmt.Println(query3.ToSQL())

	// Output:
	// 示例 1 - 简单排序:
	// SELECT `products`.`name`, `products`.`price`, ROW_NUMBER() OVER(ORDER BY `products`.`price` DESC) AS `row_num` FROM `products`
	//
	// 示例 2 - 分组排序:
	// SELECT `products`.`category`, `products`.`name`, `products`.`price`, ROW_NUMBER() OVER(PARTITION BY `products`.`category` ORDER BY `products`.`price` DESC) AS `row_num` FROM `products`
	//
	// 示例 3 - 多字段分组:
	// SELECT `products`.`category`, `products`.`created_at`, `products`.`name`, `products`.`price`, ROW_NUMBER() OVER(PARTITION BY `products`.`category`, `products`.`created_at` ORDER BY `products`.`price` ASC) AS `row_num` FROM `products`
}

// 演示 RANK() 窗口函数的使用
func ExampleRank() {
	// 创建字段
	category := fields.NewTextExprField[string]("products", "category")
	price := fields.NewFloatExprField[float64]("products", "price")
	name := fields.NewTextExprField[string]("products", "name")

	products := &gsql.Table{Name: "products"}

	// RANK() 会为相同的值分配相同的排名，下一个排名会跳过
	rank := gsql.Rank().
		PartitionBy(category).
		OrderBy(price.Desc()).
		AsF("price_rank")

	query := gsql.Select(category, name, price, rank).
		From(products)

	fmt.Println("RANK() 示例:")
	fmt.Println(query.ToSQL())

	// Output:
	// RANK() 示例:
	// SELECT `products`.`category`, `products`.`name`, `products`.`price`, RANK() OVER(PARTITION BY `products`.`category` ORDER BY `products`.`price` DESC) AS `price_rank` FROM `products`
}

// 演示 DENSE_RANK() 窗口函数的使用
func ExampleDenseRank() {
	// 创建字段
	score := fields.NewIntExprField[int]("students", "score")
	name := fields.NewTextExprField[string]("students", "name")

	students := &gsql.Table{Name: "students"}

	// DENSE_RANK() 会为相同的值分配相同的排名，下一个排名连续
	denseRank := gsql.DenseRank().
		OrderBy(score.Desc()).
		AsF("score_rank")

	query := gsql.Select(name, score, denseRank).
		From(students)

	fmt.Println("DENSE_RANK() 示例:")
	fmt.Println(query.ToSQL())

	// Output:
	// DENSE_RANK() 示例:
	// SELECT `students`.`name`, `students`.`score`, DENSE_RANK() OVER(ORDER BY `students`.`score` DESC) AS `score_rank` FROM `students`
}
