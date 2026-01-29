package gsql

import (
	"strings"
	"testing"

	"github.com/donutnomad/gsql/internal/fields"
	"github.com/samber/mo"
)

// TestWhereEmptyConditions 测试空条件不会导致 "AND AND" 语法错误
// 这是一个回归测试，确保 mo.None 等空值不会产生无效 SQL
func TestWhereEmptyConditions(t *testing.T) {
	// 定义测试用的字段
	idField := fields.IntFieldOf[int64]("test_table", "id")
	nameField := fields.StringFieldOf[string]("test_table", "name")
	statusField := fields.IntFieldOf[int]("test_table", "status")
	clientIDField := fields.IntFieldOf[int64]("test_table", "client_id")
	priceField := fields.FloatFieldOf[float64]("test_table", "price")

	t.Run("EqOpt with None should not produce AND AND", func(t *testing.T) {
		sql := Select(idField, nameField, statusField, clientIDField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				nameField.EqOpt(mo.None[string]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		if strings.Contains(sql, "AND\n    AND") || strings.Contains(sql, "AND \n") {
			t.Errorf("SQL contains invalid AND pattern: %s", sql)
		}
		// 验证 SQL 包含正确的条件
		if !strings.Contains(sql, "`test_table`.`id` = 1") {
			t.Errorf("SQL should contain id = 1: %s", sql)
		}
		if !strings.Contains(sql, "`test_table`.`client_id` = 100") {
			t.Errorf("SQL should contain client_id = 100: %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("Multiple Opt methods with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.EqOpt(mo.None[int64]()),       // 空
				nameField.EqOpt(mo.None[string]()),    // 空
				statusField.EqOpt(mo.Some(1)),         // 有值
				clientIDField.EqOpt(mo.None[int64]()), // 空
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		// 应该只有 status = 1
		if !strings.Contains(sql, "`test_table`.`status` = 1") {
			t.Errorf("SQL should contain status = 1: %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("NotOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				nameField.NotOpt(mo.None[string]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("GtOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.GtOpt(mo.None[int]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("GteOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.GteOpt(mo.None[int]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("LtOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.LtOpt(mo.None[int]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("LteOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.LteOpt(mo.None[int]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("BetweenOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.BetweenOpt(mo.None[int](), mo.None[int]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("Between with nil", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.Between(nil, nil), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("LikeOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				nameField.LikeOpt(mo.None[string]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("ContainsOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				nameField.ContainsOpt(mo.None[string]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("HasPrefixOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				nameField.HasPrefixOpt(mo.None[string]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("HasSuffixOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				nameField.HasSuffixOpt(mo.None[string]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("All conditions are None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.EqOpt(mo.None[int64]()),
				nameField.EqOpt(mo.None[string]()),
				statusField.EqOpt(mo.None[int]()),
			).
			String()

		// 没有有效条件时，不应该有无效的 AND 模式
		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		// 检查是否有孤立的 AND
		if strings.Contains(sql, "WHERE AND") || strings.Contains(sql, "WHERE\n    AND") {
			t.Errorf("SQL contains 'WHERE AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("Empty In slice", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.In(), // 空切片
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("Empty NotIn slice", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				statusField.NotIn(), // 空切片
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("Mixed valid and empty conditions with OrderBy - real world case", func(t *testing.T) {
		// 这是报告的真实场景
		sql := Select(idField, nameField).
			From(TN("test_table")).
			Where(
				idField.EqOpt(mo.None[int64]()),   // 空
				clientIDField.Eq(1),               // 有效
				statusField.EqOpt(mo.None[int]()), // 空
			).
			OrderBy(idField.Desc()).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		if strings.Contains(sql, "AND\nORDER") || strings.Contains(sql, "AND ORDER") {
			t.Errorf("SQL contains 'AND ORDER' (missing condition before ORDER BY): %s", sql)
		}
		// 验证结构正确
		if !strings.Contains(sql, "`test_table`.`client_id` = 1") {
			t.Errorf("SQL should contain client_id = 1: %s", sql)
		}
		if !strings.Contains(sql, "ORDER BY") {
			t.Errorf("SQL should contain ORDER BY: %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("First condition is None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.EqOpt(mo.None[int64]()), // 空条件在最前
				clientIDField.Eq(1),
				statusField.Eq(2),
			).
			String()

		// 不应该以 AND 开头
		if strings.Contains(sql, "WHERE AND") {
			t.Errorf("SQL contains 'WHERE AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("Last condition is None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				clientIDField.Eq(100),
				statusField.EqOpt(mo.None[int]()), // 空条件在最后
			).
			String()

		// 不应该以 AND 结尾
		if strings.Contains(sql, "AND\n") && !strings.Contains(sql, "AND `") {
			t.Errorf("SQL ends with dangling AND: %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("NotBetween with nil", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				priceField.NotBetween(nil, nil), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})

	t.Run("NotBetweenOpt with None", func(t *testing.T) {
		sql := Select(idField).
			From(TN("test_table")).
			Where(
				idField.Eq(1),
				priceField.NotBetweenOpt(mo.None[float64](), mo.None[float64]()), // 空条件
				clientIDField.Eq(100),
			).
			String()

		if strings.Contains(sql, "AND AND") {
			t.Errorf("SQL contains 'AND AND': %s", sql)
		}
		t.Logf("Generated SQL: %s", sql)
	})
}

// TestConditionIsEmptySQL 测试 Condition.IsEmptySQL 方法
func TestConditionIsEmptySQL(t *testing.T) {
	idField := fields.IntFieldOf[int64]("test", "id")
	_ = fields.StringFieldOf[string]("test", "name") // nameField 暂时不用

	t.Run("EqOpt None returns empty condition", func(t *testing.T) {
		cond := idField.EqOpt(mo.None[int64]())
		if !cond.IsEmptySQL() {
			t.Error("EqOpt(None) should return empty condition")
		}
	})

	t.Run("EqOpt Some returns non-empty condition", func(t *testing.T) {
		cond := idField.EqOpt(mo.Some[int64](1))
		if cond.IsEmptySQL() {
			t.Error("EqOpt(Some) should return non-empty condition")
		}
	})

	t.Run("NotOpt None returns empty condition", func(t *testing.T) {
		cond := idField.NotOpt(mo.None[int64]())
		if !cond.IsEmptySQL() {
			t.Error("NotOpt(None) should return empty condition")
		}
	})

	t.Run("GtOpt None returns empty condition", func(t *testing.T) {
		cond := idField.GtOpt(mo.None[int64]())
		if !cond.IsEmptySQL() {
			t.Error("GtOpt(None) should return empty condition")
		}
	})

	t.Run("Empty In returns empty condition", func(t *testing.T) {
		cond := idField.In()
		if !cond.IsEmptySQL() {
			t.Error("In() with no args should return empty condition")
		}
	})

	t.Run("Empty NotIn returns empty condition", func(t *testing.T) {
		cond := idField.NotIn()
		if !cond.IsEmptySQL() {
			t.Error("NotIn() with no args should return empty condition")
		}
	})

	t.Run("Between nil nil returns empty condition", func(t *testing.T) {
		cond := idField.Between(nil, nil)
		if !cond.IsEmptySQL() {
			t.Error("Between(nil, nil) should return empty condition")
		}
	})

	t.Run("NotBetween nil nil returns empty condition", func(t *testing.T) {
		cond := idField.NotBetween(nil, nil)
		if !cond.IsEmptySQL() {
			t.Error("NotBetween(nil, nil) should return empty condition")
		}
	})
}
