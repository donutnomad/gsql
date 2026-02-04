package gsql_test

import (
	"fmt"
	"testing"

	"github.com/donutnomad/gsql"
)

// TestCTEBasic 测试基本的 CTE 用法
func TestCTEBasic(t *testing.T) {
	// WITH user_summary AS (
	//   SELECT id, name FROM users WHERE age > 18
	// )
	// SELECT * FROM users
	sql := gsql.With("user_summary",
		gsql.Select(gsql.Field("id"), gsql.Field("name")).
			From(gsql.TN("users")).
			Where(gsql.Expr("age > ?", 18)),
	).Select(gsql.Star).
		From(gsql.TN("users")).
		ToSQL()

	t.Logf("CTE Basic SQL:\n%s", sql)

	if !contains(sql, "WITH") {
		t.Errorf("Expected SQL to contain WITH")
	}
	if !contains(sql, "user_summary") {
		t.Errorf("Expected SQL to contain user_summary")
	}
}

// TestCTEMultiple 测试多个 CTE
func TestCTEMultiple(t *testing.T) {
	// WITH
	//   young AS (SELECT * FROM users WHERE age < 30),
	//   old AS (SELECT * FROM users WHERE age >= 30)
	// SELECT * FROM combined
	sql := gsql.With("young",
		gsql.Select(gsql.Star).
			From(gsql.TN("users")).
			Where(gsql.Expr("age < ?", 30)),
	).And("old",
		gsql.Select(gsql.Star).
			From(gsql.TN("users")).
			Where(gsql.Expr("age >= ?", 30)),
	).Select(gsql.Star).
		From(gsql.TN("combined")).
		ToSQL()

	t.Logf("Multiple CTE SQL:\n%s", sql)

	if !contains(sql, "WITH") {
		t.Errorf("Expected SQL to contain WITH")
	}
	if !contains(sql, "young") {
		t.Errorf("Expected SQL to contain young")
	}
	if !contains(sql, "old") {
		t.Errorf("Expected SQL to contain old")
	}
}

// TestCTEWithColumns 测试带列名的 CTE
func TestCTEWithColumns(t *testing.T) {
	// WITH user_info (user_id, user_name) AS (
	//   SELECT id, name FROM users
	// )
	// SELECT * FROM users
	sql := gsql.With("user_info",
		gsql.Select(gsql.Field("id"), gsql.Field("name")).
			From(gsql.TN("users")),
		"user_id", "user_name", // 指定列名
	).Select(gsql.Star).
		From(gsql.TN("users")).
		ToSQL()

	t.Logf("CTE with columns SQL:\n%s", sql)

	if !contains(sql, "user_info") {
		t.Errorf("Expected SQL to contain user_info")
	}
	if !contains(sql, "user_id") {
		t.Errorf("Expected SQL to contain user_id")
	}
}

// TestCTERecursive 测试递归 CTE
func TestCTERecursive(t *testing.T) {
	// WITH RECURSIVE numbers AS (
	//   SELECT 1 as n
	// )
	// SELECT * FROM numbers
	sql := gsql.WithRecursive("numbers",
		gsql.Select(gsql.Lit(1).As("n")).
			From(gsql.TN("dual")),
	).Select(gsql.Star).
		From(gsql.TN("numbers")).
		ToSQL()

	t.Logf("Recursive CTE SQL:\n%s", sql)

	if !contains(sql, "WITH RECURSIVE") {
		t.Errorf("Expected SQL to contain WITH RECURSIVE")
	}
	if !contains(sql, "numbers") {
		t.Errorf("Expected SQL to contain numbers")
	}
}

// TestCTEWithJoin 测试 CTE 与 JOIN
func TestCTEWithJoin(t *testing.T) {
	// WITH active_users AS (
	//   SELECT id FROM users WHERE status = 'active'
	// )
	// SELECT * FROM orders o
	// INNER JOIN active_users au ON o.user_id = au.id
	sql := gsql.With("active_users",
		gsql.Select(gsql.Field("id")).
			From(gsql.TN("users")).
			Where(gsql.Expr("status = ?", "active")),
	).Select(gsql.Star).
		From(gsql.TN("orders")).
		Join(gsql.InnerJoin(gsql.TN("active_users")).
			On(gsql.Expr("orders.user_id = active_users.id"))).
		ToSQL()

	t.Logf("CTE with JOIN SQL:\n%s", sql)

	if !contains(sql, "WITH") {
		t.Errorf("Expected SQL to contain WITH")
	}
	if !contains(sql, "active_users") {
		t.Errorf("Expected SQL to contain active_users")
	}
	if !contains(sql, "INNER JOIN") {
		t.Errorf("Expected SQL to contain INNER JOIN")
	}
}

// TestCTEWithWhere 测试 CTE 与 WHERE 子句
func TestCTEWithWhere(t *testing.T) {
	sql := gsql.With("summary",
		gsql.Select(gsql.Field("id"), gsql.Field("total")).
			From(gsql.TN("orders")),
	).Select(gsql.Star).
		From(gsql.TN("summary")).
		Where(gsql.Expr("total > ?", 1000)).
		ToSQL()

	t.Logf("CTE with WHERE SQL:\n%s", sql)

	if !contains(sql, "WITH") || !contains(sql, "WHERE") {
		t.Errorf("Expected SQL to contain WITH and WHERE")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestCTEExample_Basic 演示基本的 CTE 用法
func TestCTEExample_Basic(t *testing.T) {
	// 创建一个 CTE，然后在主查询中使用它
	sql :=
		gsql.With("user_summary",
			gsql.Select(gsql.Field("id"), gsql.Field("name"), gsql.Field("age")).
				From(gsql.TN("users")).
				Where(gsql.Expr("age > ?", 18)),
		).
			Select(gsql.Star).
			From(gsql.TN("user_summary")).
			Where(gsql.Expr("age < ?", 30)).
			ToSQL()

	fmt.Printf("基本 CTE 用法:\n%s\n\n", sql)
}

// TestCTEExample_Multiple 演示多个 CTE
func TestCTEExample_Multiple(t *testing.T) {
	// 使用多个 CTE，分别统计年轻用户和老用户
	sql := gsql.With("young_users",
		gsql.Select(gsql.Star).
			From(gsql.TN("users")).
			Where(gsql.Expr("age < ?", 30)),
	).
		And("old_users",
			gsql.Select(gsql.Star).
				From(gsql.TN("users")).
				Where(gsql.Expr("age >= ?", 30)),
		).
		Select(gsql.Star).
		From(gsql.TN("young_users")).
		ToSQL()

	fmt.Printf("多个 CTE:\n%s\n\n", sql)
}

// TestCTEExample_WithColumns 演示带列名的 CTE
func TestCTEExample_WithColumns(t *testing.T) {
	// 显式指定 CTE 的列名
	sql := gsql.With("user_info",
		gsql.Select(gsql.Field("id"), gsql.Field("name"), gsql.Field("email")).
			From(gsql.TN("users")),
		"user_id", "user_name", "user_email", // 指定列名
	).
		Select(gsql.Field("user_id"), gsql.Field("user_name")).
		From(gsql.TN("user_info")).
		ToSQL()

	fmt.Printf("带列名的 CTE:\n%s\n\n", sql)
}

// TestCTEExample_Recursive 演示递归 CTE
func TestCTEExample_Recursive(t *testing.T) {
	// 递归 CTE 示例：生成数字序列
	sql := gsql.WithRecursive("numbers",
		gsql.Select(gsql.Field("n")).
			From(gsql.TN("dual")).
			Where(gsql.Expr("n = ?", 1)),
	).
		Select(gsql.Star).
		From(gsql.TN("numbers")).
		Where(gsql.Expr("n <= ?", 10)).
		ToSQL()

	fmt.Printf("递归 CTE (数字序列基础):\n%s\n\n", sql)
}

// TestCTEExample_RecursiveTree 演示递归 CTE 查询树形结构
func TestCTEExample_RecursiveTree(t *testing.T) {
	// 递归查询组织架构树的锚点部分
	sql := gsql.WithRecursive("org_tree",
		gsql.Select(
			gsql.Field("id"),
			gsql.Field("name"),
			gsql.Field("parent_id"),
			gsql.Field("level"),
		).
			From(gsql.TN("departments")).
			Where(gsql.Expr("parent_id IS NULL")),
	).
		Select(gsql.Star).
		From(gsql.TN("org_tree")).
		Order(gsql.Field("level"), true).
		ToSQL()

	fmt.Printf("递归 CTE (组织架构树基础):\n%s\n\n", sql)
}

// TestCTEExample_WithJoin 演示 CTE 与 JOIN
func TestCTEExample_WithJoin(t *testing.T) {
	sql := gsql.With("active_users",
		gsql.Select(gsql.Field("id"), gsql.Field("name")).
			From(gsql.TN("users")).
			Where(gsql.Expr("status = ?", "active")),
	).Select(
		gsql.Field("orders.id"),
		gsql.Field("orders.total"),
		gsql.Field("active_users.name"),
	).
		From(gsql.TN("orders")).
		Join(gsql.InnerJoin(gsql.TN("active_users")).
			On(gsql.Expr("orders.user_id = active_users.id"))).
		ToSQL()

	fmt.Printf("CTE 与 JOIN:\n%s\n\n", sql)
}

// TestCTEExample_Aggregation 演示 CTE 与聚合
func TestCTEExample_Aggregation(t *testing.T) {
	sql := gsql.With("monthly_sales",
		gsql.Select(
			gsql.Field("month"),
			gsql.Field("total"),
		).
			From(gsql.TN("orders")).
			GroupBy(gsql.Expr("month")),
	).
		Select(
			gsql.Field("month"),
			gsql.Field("total"),
		).
		From(gsql.TN("monthly_sales")).
		Order(gsql.Field("month"), true).
		ToSQL()

	fmt.Printf("CTE 与聚合:\n%s\n\n", sql)
}
