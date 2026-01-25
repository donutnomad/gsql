package gsql_test

import (
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/internal/fields"
)

// 演示 CASE WHEN 构建器的实际使用场景

func TestCaseExample_SearchedCase(t *testing.T) {
	// 场景：根据订单金额分级
	amount := gsql.NewIntField[int64]("", "amount")

	amountLevel := gsql.Cases.String().
		When(amount.Gt(10000), gsql.StringVal("VIP")).
		When(amount.Gt(5000), gsql.StringVal("Premium")).
		When(amount.Gt(1000), gsql.StringVal("Standard")).
		Else(gsql.StringVal("Basic"))

	sql := gsql.
		Select(
			gsql.Field("id"),
			gsql.Field("amount"),
			amountLevel.As("customer_level"),
		).
		From(gsql.TN("orders")).ToSQL()

	t.Logf("搜索式 CASE SQL:\n%s", sql)
}

func TestCaseExample_SimpleCaseValue(t *testing.T) {
	// 场景：将状态码转换为中文描述
	status := gsql.NewIntField[int]("", "status")

	statusDesc := gsql.CaseValue[int, string, gsql.StringExpr[string]](status.Expr()).
		When(gsql.IntVal(0), gsql.StringVal("待处理")).
		When(gsql.IntVal(1), gsql.StringVal("处理中")).
		When(gsql.IntVal(2), gsql.StringVal("已完成")).
		Else(gsql.StringVal("未知状态"))

	sql := gsql.Select(
		gsql.Field("id"),
		gsql.Field("status"),
		statusDesc.As("status_desc"),
	).From(gsql.TN("orders")).ToSQL()

	t.Logf("简单 CASE SQL:\n%s", sql)
}

func TestCaseExample_ComplexScenario(t *testing.T) {
	// 场景：根据多个条件计算折扣
	userLevel := gsql.NewStringField[string]("", "user_level")
	amount := gsql.NewIntField[int64]("", "amount")
	firstOrder := gsql.NewScalarField[bool]("", "first_order")

	discount := gsql.Cases.Float().
		When(
			gsql.And(
				userLevel.Eq("VIP"),
				amount.Gt(10000),
			),
			gsql.FloatVal(0.7), // 7折
		).
		When(
			gsql.And(
				userLevel.Eq("Premium"),
				amount.Gt(5000),
			),
			gsql.FloatVal(0.85), // 85折
		).
		When(firstOrder.Eq(true), gsql.FloatVal(0.9)). // 首单9折
		Else(gsql.FloatVal(1.0))                       // 原价

	// 计算最终金额（amount * discount_rate）

	sql := gsql.Select(
		gsql.Field("id"),
		gsql.Field("amount"),
		discount.As("discount_rate"),
		amount.Mul(discount).As("final_amount"),
	).From(gsql.TN("orders")).ToSQL()

	t.Logf("复杂条件 CASE SQL:\n%s", sql)
}

func TestCaseExample_InGroupBy(t *testing.T) {
	// 场景：按金额分段统计订单数
	amount := fields.NewIntField[int64]("", "amount")

	amountRange := gsql.Cases.String().
		When(amount.Lt(100), gsql.StringVal("0-100")).
		When(amount.Lt(500), gsql.StringVal("100-500")).
		When(amount.Lt(1000), gsql.StringVal("500-1000")).
		Else(gsql.StringVal("1000+"))

	sql := gsql.
		Select(
			amountRange.As("amount_range"),
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
	status := fields.NewStringField[string]("", "status")
	id := fields.NewIntField[uint]("", "id")

	priority := gsql.Cases.Int().
		When(status.Eq("urgent"), gsql.IntVal(1)).
		When(status.Eq("high"), gsql.IntVal(2)).
		When(status.Eq("normal"), gsql.IntVal(3)).
		Else(gsql.IntVal(4)).
		As("priority")

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
	userType := fields.NewStringField[string]("", "user_type")
	createdAt := fields.NewDateTimeField[time.Time]("", "created_at")
	monthCreatedAt := createdAt.Month()

	// 季节性折扣
	seasonDiscount := gsql.Cases.Float().
		When(monthCreatedAt.In(11, 12), gsql.FloatVal(0.8)). // 双11双12
		When(monthCreatedAt.Eq(6), gsql.FloatVal(0.9)).      // 618
		Else(gsql.FloatVal(1.0))

	seasonDiscount1 := fields.IntOf[float64](seasonDiscount)

	// VIP 在季节性折扣基础上再打 95 折
	finalDiscount := gsql.Cases.Float().
		When(
			userType.Eq("vip"),
			seasonDiscount1.Mul(0.95).AsFloat(), // VIP在活动基础上再打95折
		).
		Else(seasonDiscount1.AsFloat()).As("final_discount")

	sql := gsql.Select(
		gsql.Field("`id`"),
		gsql.Field("`user_type`"),
		finalDiscount,
	).From(gsql.TN("orders")).ToSQL()

	t.Logf("嵌套 CASE SQL:\n%s", sql)
}
