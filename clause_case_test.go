package gsql_test

import (
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
)

// 演示 CASE WHEN 构建器的实际使用场景

func TestCaseExample_SearchedCase(t *testing.T) {
	// 场景：根据订单金额分级
	amount := field.NewComparable[int64]("", "amount")

	amountLevel := gsql.Case().
		When(amount.Gt(10000), gsql.Lit("VIP")).
		When(amount.Gt(5000), gsql.Lit("Premium")).
		When(amount.Gt(1000), gsql.Lit("Standard")).
		Else(gsql.Lit("Basic")).
		End().AsF("customer_level")

	sql := gsql.Select(
		gsql.Field("id"),
		gsql.Field("amount"),
		amountLevel,
	).From(gsql.TN("orders")).ToSQL()

	t.Logf("搜索式 CASE SQL:\n%s", sql)
}

func TestCaseExample_SimpleCaseValue(t *testing.T) {
	// 场景：将状态码转换为中文描述
	status := field.NewComparable[int]("", "status")

	statusDesc := gsql.CaseValue(status.ToExpr()).
		When(gsql.Lit(0), gsql.Lit("待处理")).
		When(gsql.Lit(1), gsql.Lit("处理中")).
		When(gsql.Lit(2), gsql.Lit("已完成")).
		Else(gsql.Lit("未知状态")).
		End().AsF("status_desc")

	sql := gsql.Select(
		gsql.Field("id"),
		gsql.Field("status"),
		statusDesc,
	).From(gsql.TN("orders")).ToSQL()

	t.Logf("简单 CASE SQL:\n%s", sql)
}

func TestCaseExample_ComplexScenario(t *testing.T) {
	// 场景：根据多个条件计算折扣
	userLevel := field.NewPattern[string]("", "user_level")
	amount := field.NewComparable[int64]("", "amount")
	firstOrder := field.NewComparable[bool]("", "first_order")

	discount := gsql.Case().
		When(
			gsql.And(
				userLevel.Eq("VIP"),
				amount.Gt(10000),
			),
			gsql.Lit(0.7), // 7折
		).
		When(
			gsql.And(
				userLevel.Eq("Premium"),
				amount.Gt(5000),
			),
			gsql.Lit(0.85), // 85折
		).
		When(firstOrder.Eq(true), gsql.Lit(0.9)). // 首单9折
		Else(gsql.Lit(1.0)).                      // 原价
		End().AsF("discount_rate")

	// 计算最终金额（amount * discount_rate）
	finalAmount := gsql.Field("amount * (" + discount.Name() + ")").As("final_amount")

	sql := gsql.Select(
		gsql.Field("id"),
		gsql.Field("amount"),
		discount,
		finalAmount,
	).From(gsql.TN("orders")).ToSQL()

	t.Logf("复杂条件 CASE SQL:\n%s", sql)
}

func TestCaseExample_InGroupBy(t *testing.T) {
	// 场景：按金额分段统计订单数
	amount := fields.NewIntExprField[int64]("", "amount")

	amountRange := gsql.Case().
		When(amount.Lt(100), gsql.Lit("0-100")).
		When(amount.Lt(500), gsql.Lit("100-500")).
		When(amount.Lt(1000), gsql.Lit("500-1000")).
		Else(gsql.Lit("1000+")).
		End().AsF("amount_range")

	sql := gsql.
		Select(
			amountRange,
			gsql.COUNT().As("order_count"),
			amount.Sum().As("total_amount"),
		).
		From(gsql.TN("orders")).
		GroupBy(amountRange).
		ToSQL()

	t.Logf("GROUP BY with CASE SQL:\n%s", sql)
}

func TestCaseExample_InOrderBy(t *testing.T) {
	// 场景：自定义排序优先级
	status := field.NewPattern[string]("", "status")
	id := field.NewPattern[uint]("", "id")

	priority := gsql.Case().
		When(status.Eq("urgent"), gsql.Lit(1)).
		When(status.Eq("high"), gsql.Lit(2)).
		When(status.Eq("normal"), gsql.Lit(3)).
		Else(gsql.Lit(4)).
		End().AsF("priority")

	sql := gsql.
		Select(
			id,
			status,
			priority,
		).
		From(gsql.TN("tasks")).
		Order(priority, true). // 按优先级升序
		ToSQL()

	t.Logf("ORDER BY with CASE SQL:\n%s", sql)
}

func TestCaseExample_NestedCase(t *testing.T) {
	// 场景：嵌套 CASE 表达式
	userType := field.NewPattern[string]("", "user_type")
	createdAt := fields.NewDateTimeExprField[time.Time]("", "created_at")
	monthCreatedAt := createdAt.Month()

	// 季节性折扣
	seasonDiscount := gsql.Case().
		When(monthCreatedAt.In(11, 12), gsql.Lit(0.8)). // 双11双12
		When(monthCreatedAt.Eq(6), gsql.Lit(0.9)).      // 618
		Else(gsql.Lit(1.0)).
		End()

	seasonDiscount1 := fields.NewIntExpr[float64](seasonDiscount)

	// VIP 在季节性折扣基础上再打 95 折
	finalDiscount := gsql.Case().
		When(
			userType.Eq("vip"),
			seasonDiscount1.Mul(0.95), // VIP在活动基础上再打95折
		).
		Else(seasonDiscount1).
		End().AsF("final_discount")

	sql := gsql.Select(
		gsql.Field("`id`"),
		gsql.Field("`user_type`"),
		finalDiscount,
	).From(gsql.TN("orders")).ToSQL()

	t.Logf("嵌套 CASE SQL:\n%s", sql)
}
