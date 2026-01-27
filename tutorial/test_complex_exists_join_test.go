package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
)

// ==================== Complex EXISTS + JOIN + Aggregation Tests ====================
// 这个测试文件模拟类似以下的复杂查询场景:
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

// ==================== 模型定义 ====================

// BalanceRecord 余额记录表
// @Gsql
type BalanceRecord struct {
	ID          uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	Contract    string  `gorm:"column:contract;size:42;not null;index"`
	Account     string  `gorm:"column:account;size:42;not null;index"`
	NFTID       uint64  `gorm:"column:nft_id;not null;index"`
	BlockNumber uint64  `gorm:"column:block_number;not null;index"`
	TotalHold   float64 `gorm:"column:total_hold;default:0"`
	Realized    float64 `gorm:"column:realized;default:0"`
	Committed   float64 `gorm:"column:committed;default:0"`
	Balance     float64 `gorm:"column:balance;default:0"`
}

func (BalanceRecord) TableName() string { return "balance_records" }

// RORRecord Rights of Record 记录表
// @Gsql
type RORRecord struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	NFTContract string `gorm:"column:nft_contract;size:42;not null;index"`
	NFTID       uint64 `gorm:"column:nft_id;not null;index"`
	Receiver    string `gorm:"column:receiver;size:42;not null;index"`
	NFTStatus   int    `gorm:"column:nft_status;default:0"`
}

func (RORRecord) TableName() string { return "ror_records" }

// NFTToken NFT 持有记录表
// @Gsql
type NFTToken struct {
	ID          uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	NFTContract string  `gorm:"column:nft_contract;size:42;not null;index"`
	Account     string  `gorm:"column:account;size:42;not null;index"`
	NFTID       uint64  `gorm:"column:nft_id;not null;index"`
	Balance     float64 `gorm:"column:balance;default:0"`
}

func (NFTToken) TableName() string { return "nft_tokens" }

// ==================== Schema 定义 ====================

type BalanceRecordSchemaType struct {
	ID          fields.IntField[uint64]
	Contract    fields.StringField[string]
	Account     fields.StringField[string]
	NFTID       fields.IntField[uint64]
	BlockNumber fields.IntField[uint64]
	TotalHold   fields.FloatField[float64]
	Realized    fields.FloatField[float64]
	Committed   fields.FloatField[float64]
	Balance     fields.FloatField[float64]
	fieldType   BalanceRecord
	alias       string
	tableName   string
}

func (t BalanceRecordSchemaType) TableName() string { return t.tableName }
func (t BalanceRecordSchemaType) Alias() string     { return t.alias }
func (t *BalanceRecordSchemaType) WithTable(tableName string) {
	tn := gsql.TN(tableName)
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
func (t BalanceRecordSchemaType) As(alias string) BalanceRecordSchemaType {
	var ret = t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}
func (t BalanceRecordSchemaType) ModelType() *BalanceRecord { return &t.fieldType }
func (t BalanceRecordSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.Contract, t.Account, t.NFTID, t.BlockNumber, t.TotalHold, t.Realized, t.Committed, t.Balance}
}

var BalanceRecordSchema = BalanceRecordSchemaType{
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
	fieldType:   BalanceRecord{},
}

type RORRecordSchemaType struct {
	ID          fields.IntField[uint64]
	NFTContract fields.StringField[string]
	NFTID       fields.IntField[uint64]
	Receiver    fields.StringField[string]
	NFTStatus   fields.IntField[int]
	fieldType   RORRecord
	alias       string
	tableName   string
}

func (t RORRecordSchemaType) TableName() string { return t.tableName }
func (t RORRecordSchemaType) Alias() string     { return t.alias }
func (t *RORRecordSchemaType) WithTable(tableName string) {
	tn := gsql.TN(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.NFTContract = t.NFTContract.WithTable(&tn)
	t.NFTID = t.NFTID.WithTable(&tn)
	t.Receiver = t.Receiver.WithTable(&tn)
	t.NFTStatus = t.NFTStatus.WithTable(&tn)
}
func (t RORRecordSchemaType) As(alias string) RORRecordSchemaType {
	var ret = t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}
func (t RORRecordSchemaType) ModelType() *RORRecord { return &t.fieldType }
func (t RORRecordSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.NFTContract, t.NFTID, t.Receiver, t.NFTStatus}
}

var RORRecordSchema = RORRecordSchemaType{
	tableName:   "ror_records",
	ID:          fields.IntFieldOf[uint64]("ror_records", "id", field.FlagPrimaryKey),
	NFTContract: fields.StringFieldOf[string]("ror_records", "nft_contract"),
	NFTID:       fields.IntFieldOf[uint64]("ror_records", "nft_id"),
	Receiver:    fields.StringFieldOf[string]("ror_records", "receiver"),
	NFTStatus:   fields.IntFieldOf[int]("ror_records", "nft_status"),
	fieldType:   RORRecord{},
}

type NFTTokenSchemaType struct {
	ID          fields.IntField[uint64]
	NFTContract fields.StringField[string]
	Account     fields.StringField[string]
	NFTID       fields.IntField[uint64]
	Balance     fields.FloatField[float64]
	fieldType   NFTToken
	alias       string
	tableName   string
}

func (t NFTTokenSchemaType) TableName() string { return t.tableName }
func (t NFTTokenSchemaType) Alias() string     { return t.alias }
func (t *NFTTokenSchemaType) WithTable(tableName string) {
	tn := gsql.TN(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.NFTContract = t.NFTContract.WithTable(&tn)
	t.Account = t.Account.WithTable(&tn)
	t.NFTID = t.NFTID.WithTable(&tn)
	t.Balance = t.Balance.WithTable(&tn)
}
func (t NFTTokenSchemaType) As(alias string) NFTTokenSchemaType {
	var ret = t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}
func (t NFTTokenSchemaType) ModelType() *NFTToken { return &t.fieldType }
func (t NFTTokenSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.NFTContract, t.Account, t.NFTID, t.Balance}
}

var NFTTokenSchema = NFTTokenSchemaType{
	tableName:   "nft_tokens",
	ID:          fields.IntFieldOf[uint64]("nft_tokens", "id", field.FlagPrimaryKey),
	NFTContract: fields.StringFieldOf[string]("nft_tokens", "nft_contract"),
	Account:     fields.StringFieldOf[string]("nft_tokens", "account"),
	NFTID:       fields.IntFieldOf[uint64]("nft_tokens", "nft_id"),
	Balance:     fields.FloatFieldOf[float64]("nft_tokens", "balance"),
	fieldType:   NFTToken{},
}

// ==================== 测试用例 ====================

// TestComplexExistsWithJoinAndAggregation 测试复杂的 EXISTS + JOIN + 聚合场景
// 模拟原始查询:
//
//	SELECT SUM(bb.total_hold) as hold, SUM(bb.realized) as realized, ...
//	FROM balance_record AS bb
//	LEFT JOIN ror_table AS rorTable ON rorTable.nft_id = bb.nft_id
//	WHERE rorTable.nft_status = ?
//	  AND bb.block_number = ?
//	  AND bb.account = ?
//	  AND bb.contract = ?
//	  AND (
//	    EXISTS (SELECT 1 FROM ror_table AS ror WHERE ...)
//	    OR EXISTS (SELECT 1 FROM nt_table AS nt WHERE ...)
//	  )
func TestComplexExistsWithJoinAndAggregation(t *testing.T) {
	// 设置表
	setupTable(t, BalanceRecordSchema.ModelType())
	setupTable(t, RORRecordSchema.ModelType())
	setupTable(t, NFTTokenSchema.ModelType())
	db := getDB()

	// 插入测试数据
	contract := "0x1234567890abcdef1234567890abcdef12345678"
	account := "0xabcdef1234567890abcdef1234567890abcdef12"
	blockNumber := uint64(12345)

	// Balance records
	balanceRecords := []BalanceRecord{
		{Contract: contract, Account: account, NFTID: 1, BlockNumber: blockNumber, TotalHold: 100, Realized: 50, Committed: 25, Balance: 25},
		{Contract: contract, Account: account, NFTID: 2, BlockNumber: blockNumber, TotalHold: 200, Realized: 100, Committed: 50, Balance: 50},
		{Contract: contract, Account: account, NFTID: 3, BlockNumber: blockNumber, TotalHold: 300, Realized: 150, Committed: 75, Balance: 75}, // 不匹配 EXISTS
		{Contract: contract, Account: "other_account", NFTID: 4, BlockNumber: blockNumber, TotalHold: 400, Realized: 200, Committed: 100, Balance: 100},
	}
	if err := db.Create(&balanceRecords).Error; err != nil {
		t.Fatalf("Failed to create balance records: %v", err)
	}

	// ROR records - NFTID 1 匹配第一个 EXISTS
	rorRecords := []RORRecord{
		{NFTContract: contract, NFTID: 1, Receiver: account, NFTStatus: 1}, // 匹配
		{NFTContract: contract, NFTID: 2, Receiver: "other_receiver", NFTStatus: 1},
	}
	if err := db.Create(&rorRecords).Error; err != nil {
		t.Fatalf("Failed to create ROR records: %v", err)
	}

	// NFT tokens - NFTID 2 匹配第二个 EXISTS
	nftTokens := []NFTToken{
		{NFTContract: contract, Account: account, NFTID: 2, Balance: 10}, // 匹配
		{NFTContract: contract, Account: account, NFTID: 3, Balance: 0},  // balance = 0, 不匹配
	}
	if err := db.Create(&nftTokens).Error; err != nil {
		t.Fatalf("Failed to create NFT tokens: %v", err)
	}

	t.Run("Complex query with table aliases, JOIN, multiple EXISTS and aggregation", func(t *testing.T) {
		// 创建带别名的 schema
		bb := BalanceRecordSchema.As("bb")
		rorTable := RORRecordSchema.As("rorTable")
		ror := RORRecordSchema.As("ror")
		nt := NFTTokenSchema.As("nt")

		// 构建第一个 EXISTS 子查询
		// SELECT 1 FROM ror_records AS ror
		// WHERE bb.contract = ror.nft_contract AND bb.account = ror.receiver AND bb.nft_id = ror.nft_id
		existsSubquery1 := gsql.Select(gsql.Lit(1).As("_")).
			From(&ror).
			Where(
				bb.Contract.EqF(ror.NFTContract),
				bb.Account.EqF(ror.Receiver),
				bb.NFTID.EqF(ror.NFTID),
			)

		// 构建第二个 EXISTS 子查询
		// SELECT 1 FROM nft_tokens AS nt
		// WHERE bb.contract = nt.nft_contract AND bb.account = nt.account AND nt.balance > 0 AND bb.nft_id = nt.nft_id
		existsSubquery2 := gsql.Select(gsql.Lit(1).As("_")).
			From(&nt).
			Where(
				bb.Contract.EqF(nt.NFTContract),
				bb.Account.EqF(nt.Account),
				nt.Balance.Gt(0),
				bb.NFTID.EqF(nt.NFTID),
			)

		// 结果类型
		type AggregateResult struct {
			Hold      *float64 `gorm:"column:hold"`
			Realized  *float64 `gorm:"column:realized"`
			Committed *float64 `gorm:"column:committed"`
			Balance   *float64 `gorm:"column:balance"`
		}

		// 构建主查询
		// SELECT SUM(bb.total_hold) as hold, SUM(bb.realized) as realized, ...
		// FROM balance_records AS bb
		// LEFT JOIN ror_records AS rorTable ON rorTable.nft_id = bb.nft_id
		// WHERE rorTable.nft_status = 1
		//   AND bb.block_number = 12345
		//   AND bb.account = '0xabcdef...'
		//   AND bb.contract = '0x1234...'
		//   AND (EXISTS(...) OR EXISTS(...))
		var result AggregateResult
		err := gsql.Select(
			bb.TotalHold.Sum().As("hold"),
			bb.Realized.Sum().As("realized"),
			bb.Committed.Sum().As("committed"),
			bb.Balance.Sum().As("balance"),
		).From(&bb).
			Join(gsql.LeftJoin(&rorTable).On(rorTable.NFTID.EqF(bb.NFTID))).
			Where(
				rorTable.NFTStatus.Eq(1),
				bb.BlockNumber.Eq(blockNumber),
				bb.Account.Eq(account),
				bb.Contract.Eq(contract),
				gsql.Or(
					gsql.Exists(existsSubquery1),
					gsql.Exists(existsSubquery2),
				),
			).
			First(db, &result)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// 验证结果 - NFTID 1 和 2 应该匹配 (100+200=300, 50+100=150, 25+50=75, 25+50=75)
		if result.Hold == nil || *result.Hold != 300 {
			t.Errorf("Expected hold=300, got %v", result.Hold)
		}
		if result.Realized == nil || *result.Realized != 150 {
			t.Errorf("Expected realized=150, got %v", result.Realized)
		}
		if result.Committed == nil || *result.Committed != 75 {
			t.Errorf("Expected committed=75, got %v", result.Committed)
		}
		if result.Balance == nil || *result.Balance != 75 {
			t.Errorf("Expected balance=75, got %v", result.Balance)
		}
	})

	t.Run("Print SQL for verification", func(t *testing.T) {
		bb := BalanceRecordSchema.As("bb")
		rorTable := RORRecordSchema.As("rorTable")
		ror := RORRecordSchema.As("ror")
		nt := NFTTokenSchema.As("nt")

		existsSubquery1 := gsql.Select(gsql.Lit(1).As("_")).
			From(&ror).
			Where(
				bb.Contract.EqF(ror.NFTContract),
				bb.Account.EqF(ror.Receiver),
				bb.NFTID.EqF(ror.NFTID),
			)

		existsSubquery2 := gsql.Select(gsql.Lit(1).As("_")).
			From(&nt).
			Where(
				bb.Contract.EqF(nt.NFTContract),
				bb.Account.EqF(nt.Account),
				nt.Balance.Gt(0),
				bb.NFTID.EqF(nt.NFTID),
			)

		sql := gsql.Select(
			bb.TotalHold.Sum().As("hold"),
			bb.Realized.Sum().As("realized"),
			bb.Committed.Sum().As("committed"),
			bb.Balance.Sum().As("balance"),
		).From(&bb).
			Join(gsql.LeftJoin(&rorTable).On(rorTable.NFTID.EqF(bb.NFTID))).
			Where(
				rorTable.NFTStatus.Eq(1),
				bb.BlockNumber.Eq(blockNumber),
				bb.Account.Eq(account),
				bb.Contract.Eq(contract),
				gsql.Or(
					gsql.Exists(existsSubquery1),
					gsql.Exists(existsSubquery2),
				),
			).
			ToSQL()

		t.Logf("Generated SQL:\n%s", sql)
	})
}

// TestMultipleExistsWithDifferentConditions 测试多个 EXISTS 子查询与不同条件组合
func TestMultipleExistsWithDifferentConditions(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	oi := OrderItemSchema
	p := ProductSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	setupTable(t, oi.ModelType())
	setupTable(t, p.ModelType())
	db := getDB()

	// 创建测试数据
	customers := []Customer{
		{Name: "Alice", Email: "alice@test.com", Phone: "111"},
		{Name: "Bob", Email: "bob@test.com", Phone: "222"},
		{Name: "Charlie", Email: "charlie@test.com", Phone: "333"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	products := []Product{
		{Name: "Laptop", Category: "Electronics", Price: 1000, Stock: 50},
		{Name: "Mouse", Category: "Electronics", Price: 50, Stock: 200},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// Alice 有一个高价值订单
	order1 := Order{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 1500, Status: "completed"}
	if err := db.Create(&order1).Error; err != nil {
		t.Fatalf("Failed to create order1: %v", err)
	}
	orderItem1 := OrderItem{OrderID: order1.ID, ProductID: products[0].ID, Quantity: 1, UnitPrice: 1000}
	if err := db.Create(&orderItem1).Error; err != nil {
		t.Fatalf("Failed to create order item1: %v", err)
	}

	// Bob 有一个低价值订单
	order2 := Order{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 50, Status: "completed"}
	if err := db.Create(&order2).Error; err != nil {
		t.Fatalf("Failed to create order2: %v", err)
	}
	orderItem2 := OrderItem{OrderID: order2.ID, ProductID: products[1].ID, Quantity: 1, UnitPrice: 50}
	if err := db.Create(&orderItem2).Error; err != nil {
		t.Fatalf("Failed to create order item2: %v", err)
	}

	t.Run("Find customers with high-value orders OR who bought electronics", func(t *testing.T) {
		// 第一个 EXISTS: 有高价值订单 (>500)
		// SELECT 1 FROM orders WHERE orders.customer_id = customers.id AND orders.total_price > 500
		existsHighValue := gsql.Select(gsql.Lit(1).As("_")).
			From(&o).
			Where(
				o.CustomerID.EqF(c.ID),
				o.TotalPrice.Gt(500),
			)

		// 第二个 EXISTS: 购买了电子产品
		// SELECT 1 FROM orders o2
		// JOIN order_items oi ON o2.id = oi.order_id
		// JOIN products p ON oi.product_id = p.id
		// WHERE o2.customer_id = customers.id AND p.category = 'Electronics'
		o2 := OrderSchema.As("o2")
		oi2 := OrderItemSchema.As("oi2")
		p2 := ProductSchema.As("p2")

		existsBoughtElectronics := gsql.Select(gsql.Lit(1).As("_")).
			From(&o2).
			Join(
				gsql.InnerJoin(&oi2).On(o2.ID.EqF(oi2.OrderID)),
				gsql.InnerJoin(&p2).On(oi2.ProductID.EqF(p2.ID)),
			).
			Where(
				o2.CustomerID.EqF(c.ID),
				p2.Category.Eq("Electronics"),
			)

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(
				gsql.Or(
					gsql.Exists(existsHighValue),
					gsql.Exists(existsBoughtElectronics),
				),
			).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// Alice 和 Bob 都应该匹配
		// Alice: 高价值订单
		// Bob: 购买了电子产品 (Mouse)
		if len(results) != 2 {
			t.Errorf("Expected 2 customers, got %d", len(results))
		}
	})

	t.Run("Find customers with high-value orders AND who bought electronics", func(t *testing.T) {
		existsHighValue := gsql.Select(gsql.Lit(1).As("_")).
			From(&o).
			Where(
				o.CustomerID.EqF(c.ID),
				o.TotalPrice.Gt(500),
			)

		o2 := OrderSchema.As("o2")
		oi2 := OrderItemSchema.As("oi2")
		p2 := ProductSchema.As("p2")

		existsBoughtElectronics := gsql.Select(gsql.Lit(1).As("_")).
			From(&o2).
			Join(
				gsql.InnerJoin(&oi2).On(o2.ID.EqF(oi2.OrderID)),
				gsql.InnerJoin(&p2).On(oi2.ProductID.EqF(p2.ID)),
			).
			Where(
				o2.CustomerID.EqF(c.ID),
				p2.Category.Eq("Electronics"),
			)

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(
				gsql.And(
					gsql.Exists(existsHighValue),
					gsql.Exists(existsBoughtElectronics),
				),
			).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// 只有 Alice 同时满足两个条件
		if len(results) != 1 {
			t.Errorf("Expected 1 customer, got %d", len(results))
		}
		if len(results) > 0 && results[0].Name != "Alice" {
			t.Errorf("Expected Alice, got %s", results[0].Name)
		}
	})
}

// TestExistsWithTableAliasesAndJoin 测试 EXISTS 子查询中使用表别名和 JOIN
func TestExistsWithTableAliasesAndJoin(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	customers := []Customer{
		{Name: "VIP", Email: "vip@test.com", Phone: "111"},
		{Name: "Regular", Email: "regular@test.com", Phone: "222"},
		{Name: "NoOrders", Email: "no@test.com", Phone: "333"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 1000, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 2000, Status: "completed"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 100, Status: "completed"},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("Failed to create orders: %v", err)
	}

	t.Run("Aggregation with EXISTS filter using table aliases", func(t *testing.T) {
		// 使用别名
		cMain := CustomerSchema.As("c")
		oSub := OrderSchema.As("o_sub")

		// EXISTS 子查询: 检查客户是否有高价值订单
		existsHighValue := gsql.Select(gsql.Lit(1).As("_")).
			From(&oSub).
			Where(
				oSub.CustomerID.EqF(cMain.ID),
				oSub.TotalPrice.Gt(500),
			)

		type Result struct {
			Name       string   `gorm:"column:name"`
			TotalSpent *float64 `gorm:"column:total_spent"`
		}

		// 主查询: 汇总 VIP 客户的订单
		oJoin := OrderSchema.As("o_join")

		var results []Result
		err := gsql.Select(
			cMain.Name,
			oJoin.TotalPrice.Sum().As("total_spent"),
		).From(&cMain).
			Join(gsql.LeftJoin(&oJoin).On(cMain.ID.EqF(oJoin.CustomerID))).
			Where(gsql.Exists(existsHighValue)).
			GroupBy(cMain.ID, cMain.Name).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// 只有 VIP 客户有高价值订单
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if len(results) > 0 {
			if results[0].Name != "VIP" {
				t.Errorf("Expected VIP, got %s", results[0].Name)
			}
			if results[0].TotalSpent == nil || *results[0].TotalSpent != 3000 {
				t.Errorf("Expected total_spent=3000, got %v", results[0].TotalSpent)
			}
		}
	})

	t.Run("Print SQL with aliases", func(t *testing.T) {
		cMain := CustomerSchema.As("c")
		oSub := OrderSchema.As("o_sub")
		oJoin := OrderSchema.As("o_join")

		existsHighValue := gsql.Select(gsql.Lit(1).As("_")).
			From(&oSub).
			Where(
				oSub.CustomerID.EqF(cMain.ID),
				oSub.TotalPrice.Gt(500),
			)

		sql := gsql.Select(
			cMain.Name,
			oJoin.TotalPrice.Sum().As("total_spent"),
		).From(&cMain).
			Join(gsql.LeftJoin(&oJoin).On(cMain.ID.EqF(oJoin.CustomerID))).
			Where(gsql.Exists(existsHighValue)).
			GroupBy(cMain.ID, cMain.Name).
			ToSQL()

		t.Logf("Generated SQL with aliases:\n%s", sql)
	})
}

// TestNestedExistsConditions 测试嵌套的 EXISTS 条件
func TestNestedExistsConditions(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	oi := OrderItemSchema
	p := ProductSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	setupTable(t, oi.ModelType())
	setupTable(t, p.ModelType())
	db := getDB()

	customers := []Customer{
		{Name: "NestedTest_ElectronicsBuyer", Email: "nested_elec@test.com", Phone: "n111"},
		{Name: "NestedTest_ClothingBuyer", Email: "nested_cloth@test.com", Phone: "n222"},
		{Name: "NestedTest_MixedBuyer", Email: "nested_mixed@test.com", Phone: "n333"},
		{Name: "NestedTest_NoBuyer", Email: "nested_no@test.com", Phone: "n444"},
	}
	if err := db.Create(&customers).Error; err != nil {
		t.Fatalf("Failed to create customers: %v", err)
	}

	products := []Product{
		{Name: "NestedTest_Laptop", Category: "Electronics", Price: 1000, Stock: 50},
		{Name: "NestedTest_TShirt", Category: "Clothing", Price: 30, Stock: 200},
	}
	if err := db.Create(&products).Error; err != nil {
		t.Fatalf("Failed to create products: %v", err)
	}

	// ElectronicsBuyer 只买电子产品
	order1 := Order{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 1000, Status: "completed"}
	db.Create(&order1)
	db.Create(&OrderItem{OrderID: order1.ID, ProductID: products[0].ID, Quantity: 1, UnitPrice: 1000})

	// ClothingBuyer 只买服装
	order2 := Order{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 30, Status: "completed"}
	db.Create(&order2)
	db.Create(&OrderItem{OrderID: order2.ID, ProductID: products[1].ID, Quantity: 1, UnitPrice: 30})

	// MixedBuyer 两种都买
	order3 := Order{CustomerID: customers[2].ID, OrderDate: time.Now(), TotalPrice: 1030, Status: "completed"}
	db.Create(&order3)
	db.Create(&OrderItem{OrderID: order3.ID, ProductID: products[0].ID, Quantity: 1, UnitPrice: 1000})
	db.Create(&OrderItem{OrderID: order3.ID, ProductID: products[1].ID, Quantity: 1, UnitPrice: 30})

	t.Run("Complex nested conditions: (EXISTS A AND EXISTS B) OR (NOT EXISTS C)", func(t *testing.T) {
		// 查找: (买过电子产品 AND 买过服装) OR 没有任何订单

		// EXISTS: 买过电子产品
		o1 := OrderSchema.As("o1")
		oi1 := OrderItemSchema.As("oi1")
		p1 := ProductSchema.As("p1")

		existsElectronics := gsql.Select(gsql.Lit(1).As("_")).
			From(&o1).
			Join(
				gsql.InnerJoin(&oi1).On(o1.ID.EqF(oi1.OrderID)),
				gsql.InnerJoin(&p1).On(oi1.ProductID.EqF(p1.ID)),
			).
			Where(
				o1.CustomerID.EqF(c.ID),
				p1.Category.Eq("Electronics"),
			)

		// EXISTS: 买过服装
		o2 := OrderSchema.As("o2")
		oi2 := OrderItemSchema.As("oi2")
		p2 := ProductSchema.As("p2")

		existsClothing := gsql.Select(gsql.Lit(1).As("_")).
			From(&o2).
			Join(
				gsql.InnerJoin(&oi2).On(o2.ID.EqF(oi2.OrderID)),
				gsql.InnerJoin(&p2).On(oi2.ProductID.EqF(p2.ID)),
			).
			Where(
				o2.CustomerID.EqF(c.ID),
				p2.Category.Eq("Clothing"),
			)

		// NOT EXISTS: 没有订单
		notExistsOrders := gsql.Select(gsql.Lit(1).As("_")).
			From(&o).
			Where(o.CustomerID.EqF(c.ID))

		var results []Customer
		err := gsql.Select(c.AllFields()...).
			From(&c).
			Where(
				c.Name.HasPrefix("NestedTest_"), // 只查询本测试的数据
				gsql.Or(
					gsql.And(
						gsql.Exists(existsElectronics),
						gsql.Exists(existsClothing),
					),
					gsql.NotExists(notExistsOrders),
				),
			).
			OrderBy(c.Name.Asc()).
			Find(db, &results)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// MixedBuyer (两种都买) 和 NoBuyer (没有订单)
		if len(results) != 2 {
			t.Errorf("Expected 2 customers, got %d", len(results))
			for _, r := range results {
				t.Logf("  - %s", r.Name)
			}
		}

		expectedNames := map[string]bool{"NestedTest_MixedBuyer": true, "NestedTest_NoBuyer": true}
		for _, r := range results {
			if !expectedNames[r.Name] {
				t.Errorf("Unexpected customer: %s", r.Name)
			}
		}
	})
}
