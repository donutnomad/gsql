package gsql

import (
	"strings"
	"testing"

	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/samber/lo"
)

// ==================== Complex EXISTS + JOIN + Aggregation SQL Generation Tests ====================
// 这些测试验证 gsql 库能否正确生成类似以下的复杂 SQL 查询:
//
// SELECT SUM(bb.total_hold) as hold, SUM(bb.realized) as realized, ...
// FROM balance_record AS bb
// LEFT JOIN ror_table AS rorTable ON rorTable.nft_id = bb.nft_id
// WHERE rorTable.nft_status = ?
//   AND bb.block_number = ?
//   AND bb.account = ?
//   AND bb.contract = ?
//   AND (
//     EXISTS (SELECT 1 FROM ror_table AS ror WHERE bb.contract = ror.nft_contract AND bb.account = ror.receiver AND bb.nft_id = ror.nft_id)
//     OR EXISTS (SELECT 1 FROM nt_table AS nt WHERE bb.contract = nt.nft_contract AND bb.account = nt.account AND nt.balance > 0 AND bb.nft_id = nt.nft_id)
//   )

// Schema 类型定义 (为测试目的简化)

type balanceRecordSchema struct {
	ID          fields.IntField[uint64]
	Contract    fields.StringField[string]
	Account     fields.StringField[string]
	NFTID       fields.IntField[uint64]
	BlockNumber fields.IntField[uint64]
	TotalHold   fields.FloatField[float64]
	Realized    fields.FloatField[float64]
	Committed   fields.FloatField[float64]
	Balance     fields.FloatField[float64]
	alias       string
	tableName   string
}

func (t balanceRecordSchema) TableName() string { return t.tableName }
func (t balanceRecordSchema) Alias() string     { return t.alias }
func (t *balanceRecordSchema) WithTable(tableName string) {
	tn := TN(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Contract = t.Contract.WithTable(&tn)
	t.Account = t.Account.WithTable(&tn)
	t.NFTID = t.NFTID.WithTable(&tn)
	t.BlockNumber = t.BlockNumber.WithTable(&tn)
	t.TotalHold = t.TotalHold.WithTable(&tn)
	t.Realized = t.Realized.WithTable(&tn)
	t.Committed = t.Committed.WithTable(&tn)
	t.Balance = t.Balance.WithTable(&tn)
}
func (t balanceRecordSchema) As(alias string) balanceRecordSchema {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

type rorRecordSchema struct {
	ID          fields.IntField[uint64]
	NFTContract fields.StringField[string]
	NFTID       fields.IntField[uint64]
	Receiver    fields.StringField[string]
	NFTStatus   fields.IntField[int]
	alias       string
	tableName   string
}

func (t rorRecordSchema) TableName() string { return t.tableName }
func (t rorRecordSchema) Alias() string     { return t.alias }
func (t *rorRecordSchema) WithTable(tableName string) {
	tn := TN(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.NFTContract = t.NFTContract.WithTable(&tn)
	t.NFTID = t.NFTID.WithTable(&tn)
	t.Receiver = t.Receiver.WithTable(&tn)
	t.NFTStatus = t.NFTStatus.WithTable(&tn)
}
func (t rorRecordSchema) As(alias string) rorRecordSchema {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

type nftTokenSchema struct {
	ID          fields.IntField[uint64]
	NFTContract fields.StringField[string]
	Account     fields.StringField[string]
	NFTID       fields.IntField[uint64]
	Balance     fields.FloatField[float64]
	alias       string
	tableName   string
}

func (t nftTokenSchema) TableName() string { return t.tableName }
func (t nftTokenSchema) Alias() string     { return t.alias }
func (t *nftTokenSchema) WithTable(tableName string) {
	tn := TN(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.NFTContract = t.NFTContract.WithTable(&tn)
	t.Account = t.Account.WithTable(&tn)
	t.NFTID = t.NFTID.WithTable(&tn)
	t.Balance = t.Balance.WithTable(&tn)
}
func (t nftTokenSchema) As(alias string) nftTokenSchema {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

// 定义基础 Schema
var balanceRecordSchemaBase = balanceRecordSchema{
	tableName:   "balance_records",
	ID:          fields.IntFieldOf[uint64]("balance_records", "id", field.FlagPrimaryKey),
	Contract:    fields.StringFieldOf[string]("balance_records", "contract"),
	Account:     fields.StringFieldOf[string]("balance_records", "account"),
	NFTID:       fields.IntFieldOf[uint64]("balance_records", "nft_id"),
	BlockNumber: fields.IntFieldOf[uint64]("balance_records", "block_number"),
	TotalHold:   fields.FloatFieldOf[float64]("balance_records", "total_hold"),
	Realized:    fields.FloatFieldOf[float64]("balance_records", "realized"),
	Committed:   fields.FloatFieldOf[float64]("balance_records", "committed"),
	Balance:     fields.FloatFieldOf[float64]("balance_records", "balance"),
}

var rorRecordSchemaBase = rorRecordSchema{
	tableName:   "ror_records",
	ID:          fields.IntFieldOf[uint64]("ror_records", "id", field.FlagPrimaryKey),
	NFTContract: fields.StringFieldOf[string]("ror_records", "nft_contract"),
	NFTID:       fields.IntFieldOf[uint64]("ror_records", "nft_id"),
	Receiver:    fields.StringFieldOf[string]("ror_records", "receiver"),
	NFTStatus:   fields.IntFieldOf[int]("ror_records", "nft_status"),
}

var nftTokenSchemaBase = nftTokenSchema{
	tableName:   "nft_tokens",
	ID:          fields.IntFieldOf[uint64]("nft_tokens", "id", field.FlagPrimaryKey),
	NFTContract: fields.StringFieldOf[string]("nft_tokens", "nft_contract"),
	Account:     fields.StringFieldOf[string]("nft_tokens", "account"),
	NFTID:       fields.IntFieldOf[uint64]("nft_tokens", "nft_id"),
	Balance:     fields.FloatFieldOf[float64]("nft_tokens", "balance"),
}

// TestComplexExistsJoinAggregationSQL 测试复杂的 EXISTS + JOIN + 聚合 SQL 生成
func TestComplexExistsJoinAggregationSQL(t *testing.T) {
	// 创建带别名的 schema
	bb := balanceRecordSchemaBase.As("bb")
	rorTable := rorRecordSchemaBase.As("rorTable")
	ror := rorRecordSchemaBase.As("ror")
	nt := nftTokenSchemaBase.As("nt")

	// 构建第一个 EXISTS 子查询
	existsSubquery1 := Select(Lit(1).As("_")).
		From(&ror).
		Where(
			bb.Contract.EqF(ror.NFTContract),
			bb.Account.EqF(ror.Receiver),
			bb.NFTID.EqF(ror.NFTID),
		)

	// 构建第二个 EXISTS 子查询
	existsSubquery2 := Select(Lit(1).As("_")).
		From(&nt).
		Where(
			bb.Contract.EqF(nt.NFTContract),
			bb.Account.EqF(nt.Account),
			nt.Balance.Gt(0),
			bb.NFTID.EqF(nt.NFTID),
		)

	// 构建主查询
	sql := Select(
		bb.TotalHold.Sum().As("hold"),
		bb.Realized.Sum().As("realized"),
		bb.Committed.Sum().As("committed"),
		bb.Balance.Sum().As("balance"),
	).From(&bb).
		Join(LeftJoin(&rorTable).On(rorTable.NFTID.EqF(bb.NFTID))).
		Where(
			rorTable.NFTStatus.Eq(1),
			bb.BlockNumber.Eq(uint64(12345)),
			bb.Account.Eq("0xabcdef"),
			bb.Contract.Eq("0x123456"),
			Or(
				Exists(existsSubquery1),
				Exists(existsSubquery2),
			),
		).
		ToSQL()

	t.Logf("Generated SQL:\n%s", sql)

	// 验证 SQL 包含关键部分
	checks := []string{
		"SUM(`bb`.`total_hold`) AS `hold`",
		"SUM(`bb`.`realized`) AS `realized`",
		"SUM(`bb`.`committed`) AS `committed`",
		"SUM(`bb`.`balance`) AS `balance`",
		"FROM balance_records AS bb",
		"LEFT JOIN `ror_records` AS `rorTable` ON `rorTable`.`nft_id` = `bb`.`nft_id`",
		"`rorTable`.`nft_status` = 1",
		"`bb`.`block_number` = 12345",
		"`bb`.`account` = '0xabcdef'",
		"`bb`.`contract` = '0x123456'",
		"EXISTS",
		"FROM ror_records AS ror",
		"`bb`.`contract` = `ror`.`nft_contract`",
		"`bb`.`account` = `ror`.`receiver`",
		"`bb`.`nft_id` = `ror`.`nft_id`",
		"FROM nft_tokens AS nt",
		"`bb`.`contract` = `nt`.`nft_contract`",
		"`bb`.`account` = `nt`.`account`",
		"`nt`.`balance` > 0",
		"`bb`.`nft_id` = `nt`.`nft_id`",
		" OR ",
	}

	for _, check := range checks {
		if !strings.Contains(sql, check) {
			t.Errorf("SQL should contain '%s'", check)
		}
	}
}

// TestMultipleExistsWithOR 测试多个 EXISTS 与 OR 条件
func TestMultipleExistsWithOR(t *testing.T) {
	// 简单的 Schema 用于测试
	userID := fields.IntFieldOf[uint64]("users", "id")
	userName := fields.StringFieldOf[string]("users", "name")
	orderUserID := fields.IntFieldOf[uint64]("orders", "user_id")
	orderTotal := fields.FloatFieldOf[float64]("orders", "total")
	reviewUserID := fields.IntFieldOf[uint64]("reviews", "user_id")
	reviewRating := fields.IntFieldOf[int]("reviews", "rating")

	// EXISTS 1: 有高价值订单
	exists1 := Select(Lit(1).As("_")).
		From(TN("orders")).
		Where(
			orderUserID.EqF(userID),
			orderTotal.Gt(1000),
		)

	// EXISTS 2: 有好评
	exists2 := Select(Lit(1).As("_")).
		From(TN("reviews")).
		Where(
			reviewUserID.EqF(userID),
			reviewRating.Gte(4),
		)

	sql := Select(userID, userName).
		From(TN("users")).
		Where(
			Or(
				Exists(exists1),
				Exists(exists2),
			),
		).
		ToSQL()

	t.Logf("Multiple EXISTS with OR:\n%s", sql)

	// 验证 SQL 结构
	checks := []string{
		"SELECT `users`.`id`, `users`.`name` FROM `users`",
		"WHERE (",
		"EXISTS",
		" OR ",
		"`orders`.`user_id` = `users`.`id`",
		"`orders`.`total` > 1000",
		"`reviews`.`user_id` = `users`.`id`",
		"`reviews`.`rating` >= 4",
	}

	for _, check := range checks {
		if !strings.Contains(sql, check) {
			t.Errorf("SQL should contain '%s'", check)
		}
	}
}

// TestMultipleExistsWithAND 测试多个 EXISTS 与 AND 条件
func TestMultipleExistsWithAND(t *testing.T) {
	userID := fields.IntFieldOf[uint64]("users", "id")
	userName := fields.StringFieldOf[string]("users", "name")
	orderUserID := fields.IntFieldOf[uint64]("orders", "user_id")
	orderTotal := fields.FloatFieldOf[float64]("orders", "total")
	reviewUserID := fields.IntFieldOf[uint64]("reviews", "user_id")
	reviewRating := fields.IntFieldOf[int]("reviews", "rating")

	exists1 := Select(Lit(1).As("_")).
		From(TN("orders")).
		Where(
			orderUserID.EqF(userID),
			orderTotal.Gt(1000),
		)

	exists2 := Select(Lit(1).As("_")).
		From(TN("reviews")).
		Where(
			reviewUserID.EqF(userID),
			reviewRating.Gte(4),
		)

	sql := Select(userID, userName).
		From(TN("users")).
		Where(
			And(
				Exists(exists1),
				Exists(exists2),
			),
		).
		ToSQL()

	t.Logf("Multiple EXISTS with AND:\n%s", sql)

	if !strings.Contains(sql, " AND ") || !strings.Contains(sql, "EXISTS") {
		t.Error("SQL should contain 'AND' and 'EXISTS'")
	}
}

// TestNestedORAndConditions 测试嵌套的 OR 和 AND 条件
func TestNestedORAndConditions(t *testing.T) {
	userID := fields.IntFieldOf[uint64]("users", "id")
	userName := fields.StringFieldOf[string]("users", "name")
	userStatus := fields.StringFieldOf[string]("users", "status")
	orderUserID := fields.IntFieldOf[uint64]("orders", "user_id")
	orderTotal := fields.FloatFieldOf[float64]("orders", "total")
	reviewUserID := fields.IntFieldOf[uint64]("reviews", "user_id")

	// 查询: (status = 'vip' AND EXISTS(高价值订单)) OR NOT EXISTS(任何评论)
	existsHighValue := Select(Lit(1).As("_")).
		From(TN("orders")).
		Where(
			orderUserID.EqF(userID),
			orderTotal.Gt(1000),
		)

	existsAnyReview := Select(Lit(1).As("_")).
		From(TN("reviews")).
		Where(reviewUserID.EqF(userID))

	sql := Select(userID, userName).
		From(TN("users")).
		Where(
			Or(
				And(
					userStatus.Eq("vip"),
					Exists(existsHighValue),
				),
				NotExists(existsAnyReview),
			),
		).
		ToSQL()

	t.Logf("Nested OR/AND with EXISTS:\n%s", sql)

	checks := []string{
		"`users`.`status` = 'vip'",
		"EXISTS",
		"NOT EXISTS",
		"`orders`.`total` > 1000",
	}

	for _, check := range checks {
		if !strings.Contains(sql, check) {
			t.Errorf("SQL should contain '%s'", check)
		}
	}
}

// TestExistsInJoinedSubquery 测试 EXISTS 子查询中包含 JOIN
func TestExistsInJoinedSubquery(t *testing.T) {
	userID := fields.IntFieldOf[uint64]("users", "id")
	userName := fields.StringFieldOf[string]("users", "name")

	// EXISTS 子查询包含 JOIN
	// SELECT 1 FROM orders o
	// JOIN order_items oi ON o.id = oi.order_id
	// JOIN products p ON oi.product_id = p.id
	// WHERE o.user_id = users.id AND p.category = 'Electronics'
	orderID := fields.IntFieldOf[uint64]("o", "id")
	orderUserID := fields.IntFieldOf[uint64]("o", "user_id")
	oiOrderID := fields.IntFieldOf[uint64]("oi", "order_id")
	oiProductID := fields.IntFieldOf[uint64]("oi", "product_id")
	productID := fields.IntFieldOf[uint64]("p", "id")
	productCategory := fields.StringFieldOf[string]("p", "category")

	// 构建带 JOIN 的 EXISTS 子查询
	existsSubquery := Select(Lit(1).As("_")).
		From(TN("orders AS o")).
		Join(
			InnerJoin(TN("order_items AS oi")).On(orderID.EqF(oiOrderID)),
			InnerJoin(TN("products AS p")).On(oiProductID.EqF(productID)),
		).
		Where(
			orderUserID.EqF(userID),
			productCategory.Eq("Electronics"),
		)

	sql := Select(userID, userName).
		From(TN("users")).
		Where(Exists(existsSubquery)).
		ToSQL()

	t.Logf("EXISTS with JOIN subquery:\n%s", sql)

	checks := []string{
		"EXISTS",
		"FROM orders AS o",
		"INNER JOIN",
		"`o`.`user_id` = `users`.`id`",
		"`p`.`category` = 'Electronics'",
	}

	for _, check := range checks {
		if !strings.Contains(sql, check) {
			t.Errorf("SQL should contain '%s'", check)
		}
	}
}

// TestComplexWhereWithMultipleConditionTypes 测试包含多种条件类型的复杂 WHERE
func TestComplexWhereWithMultipleConditionTypes(t *testing.T) {
	bb := balanceRecordSchemaBase.As("bb")
	ror := rorRecordSchemaBase.As("ror")
	nt := nftTokenSchemaBase.As("nt")

	existsROR := Select(Lit(1).As("_")).
		From(&ror).
		Where(
			bb.Contract.EqF(ror.NFTContract),
			bb.Account.EqF(ror.Receiver),
			bb.NFTID.EqF(ror.NFTID),
		)

	existsNFT := Select(Lit(1).As("_")).
		From(&nt).
		Where(
			bb.Contract.EqF(nt.NFTContract),
			bb.Account.EqF(nt.Account),
			nt.Balance.Gt(0),
			bb.NFTID.EqF(nt.NFTID),
		)

	// 复杂条件: 基本条件 + IN + BETWEEN + (EXISTS OR EXISTS)
	sql := Select(bb.TotalHold.Sum().As("total")).
		From(&bb).
		Where(
			bb.BlockNumber.Eq(uint64(12345)),
			bb.Account.In("0xabc", "0xdef"),
			bb.TotalHold.Between(lo.ToPtr[float64](100), lo.ToPtr[float64](1000)),
			Or(
				Exists(existsROR),
				Exists(existsNFT),
			),
		).
		ToSQL()

	t.Logf("Complex WHERE with multiple condition types:\n%s", sql)

	checks := []string{
		"`bb`.`block_number` = 12345",
		"`bb`.`account` IN ('0xabc','0xdef')",
		"`bb`.`total_hold` >= 100 AND `bb`.`total_hold` < 1000",
		"EXISTS",
		" OR ",
	}

	for _, check := range checks {
		if !strings.Contains(sql, check) {
			t.Errorf("SQL should contain '%s'", check)
		}
	}
}

// TestTableAliasConsistency 测试表别名在整个查询中的一致性
func TestTableAliasConsistency(t *testing.T) {
	bb := balanceRecordSchemaBase.As("bb")

	// 确保同一个 schema 的别名在 SELECT, FROM, WHERE 中一致
	sql := Select(
		bb.ID,
		bb.Contract,
		bb.TotalHold,
	).From(&bb).
		Where(
			bb.BlockNumber.Eq(uint64(100)),
			bb.TotalHold.Gt(0),
		).
		GroupBy(bb.Contract).
		Having(bb.TotalHold.Sum().Gt(1000)).
		OrderBy(bb.TotalHold.Desc()).
		ToSQL()

	t.Logf("Table alias consistency:\n%s", sql)

	// 所有引用都应该使用 'bb' 别名
	if strings.Contains(sql, "`balance_records`.") {
		t.Error("SQL should not contain original table name when alias is used")
	}

	// 验证别名被正确使用
	bbCount := strings.Count(sql, "`bb`.")
	if bbCount < 5 { // SELECT 中 3 个 + WHERE 中 2 个
		t.Errorf("Expected at least 5 occurrences of 'bb.' alias, got %d", bbCount)
	}
}
