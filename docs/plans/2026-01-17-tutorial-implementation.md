# GSQL Tutorial 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 gsql 创建完整的用法大全，既作为用户学习文档，也作为集成测试套件。

**Architecture:** 使用 testcontainer 启动 MySQL 8.0 容器，在 tutorial/ 包中创建分层测试文件（basic、intermediate、advanced），每个文件使用共享的 gorm.DB 连接和 setupTable 辅助函数。

**Tech Stack:** Go 1.25+, testcontainers-go v0.35.0, MySQL 8.0, GORM v1.31.0, gsql

---

## Task 1: 创建 tutorial 目录和 testcontainer 基础设施

**Files:**
- Create: `tutorial/testcontainer_test.go`

**Step 1: 创建 tutorial 目录**

```bash
mkdir -p tutorial
```

**Step 2: 创建 testcontainer_test.go 文件**

```go
package tutorial

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 启动 MySQL 8.0 容器
	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("tutorial"),
		mysql.WithUsername("root"),
		mysql.WithPassword("password"),
	)
	if err != nil {
		fmt.Printf("Failed to start MySQL container: %v\n", err)
		os.Exit(1)
	}

	// 获取连接字符串
	connStr, err := mysqlContainer.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		fmt.Printf("Failed to get connection string: %v\n", err)
		_ = mysqlContainer.Terminate(ctx)
		os.Exit(1)
	}

	// 连接数据库
	sharedDB, err = gorm.Open(mysql.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		_ = mysqlContainer.Terminate(ctx)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()

	// 清理容器
	_ = mysqlContainer.Terminate(ctx)

	os.Exit(code)
}

// setupTable 创建表并在测试结束后清理
func setupTable[T any](t *testing.T, model *T) {
	t.Helper()
	err := sharedDB.AutoMigrate(model)
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}
	t.Cleanup(func() {
		_ = sharedDB.Migrator().DropTable(model)
	})
}

// getDB 返回共享的数据库连接
func getDB() *gorm.DB {
	return sharedDB
}
```

**Step 3: 验证文件可以编译**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go build ./tutorial/...`
Expected: No errors

**Step 4: Commit**

```bash
git add tutorial/testcontainer_test.go
git commit -m "feat(tutorial): add testcontainer infrastructure for MySQL 8.0"
```

---

## Task 2: 创建基础 Models

**Files:**
- Create: `tutorial/models.go`

**Step 1: 创建 models.go 文件**

```go
package tutorial

import (
	"database/sql"
	"time"
)

// ==================== Basic Models ====================

// Product 产品表 - 用于基础 CRUD 和函数测试
type Product struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string         `gorm:"column:name;size:100;not null"`
	Category    string         `gorm:"column:category;size:50"`
	Price       float64        `gorm:"column:price;not null"`
	Stock       int            `gorm:"column:stock;default:0"`
	Description sql.NullString `gorm:"column:description;type:text"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (Product) TableName() string { return "products" }

// Employee 员工表 - 用于基础查询和函数测试
type Employee struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Name       string    `gorm:"column:name;size:100;not null"`
	Email      string    `gorm:"column:email;size:100;uniqueIndex"`
	Department string    `gorm:"column:department;size:50"`
	Salary     float64   `gorm:"column:salary"`
	HireDate   time.Time `gorm:"column:hire_date"`
	BirthDate  time.Time `gorm:"column:birth_date"`
	IsActive   bool      `gorm:"column:is_active;default:true"`
}

func (Employee) TableName() string { return "employees" }

// ==================== Intermediate Models ====================

// Customer 客户表 - 用于 JOIN 测试
type Customer struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;size:100;not null"`
	Email     string    `gorm:"column:email;size:100"`
	Phone     string    `gorm:"column:phone;size:20"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Customer) TableName() string { return "customers" }

// Order 订单表 - 用于 JOIN 和聚合测试
type Order struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID uint64    `gorm:"column:customer_id;index"`
	OrderDate  time.Time `gorm:"column:order_date"`
	TotalPrice float64   `gorm:"column:total_price"`
	Status     string    `gorm:"column:status;size:20;default:'pending'"`
}

func (Order) TableName() string { return "orders" }

// OrderItem 订单项表 - 用于多表 JOIN
type OrderItem struct {
	ID        uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID   uint64  `gorm:"column:order_id;index"`
	ProductID uint64  `gorm:"column:product_id;index"`
	Quantity  int     `gorm:"column:quantity;not null"`
	UnitPrice float64 `gorm:"column:unit_price;not null"`
}

func (OrderItem) TableName() string { return "order_items" }

// ==================== Advanced Models ====================

// SalesRecord 销售记录 - 用于窗口函数测试
type SalesRecord struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Region     string    `gorm:"column:region;size:50;index"`
	Salesperson string   `gorm:"column:salesperson;size:100"`
	Amount     float64   `gorm:"column:amount"`
	SaleDate   time.Time `gorm:"column:sale_date;index"`
}

func (SalesRecord) TableName() string { return "sales_records" }

// OrgNode 组织节点 - 用于递归 CTE 测试
type OrgNode struct {
	ID       uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name     string  `gorm:"column:name;size:100;not null"`
	ParentID *uint64 `gorm:"column:parent_id;index"`
	Level    int     `gorm:"column:level;default:0"`
}

func (OrgNode) TableName() string { return "org_nodes" }

// UserProfile 用户配置 - 用于 JSON 测试
type UserProfile struct {
	ID       uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Username string `gorm:"column:username;size:100;not null"`
	Profile  string `gorm:"column:profile;type:json"` // JSON 字段
}

func (UserProfile) TableName() string { return "user_profiles" }

// Transaction 交易记录 - 用于锁测试
type Transaction struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	AccountID uint64    `gorm:"column:account_id;index"`
	Amount    float64   `gorm:"column:amount"`
	Type      string    `gorm:"column:type;size:20"` // credit/debit
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Transaction) TableName() string { return "transactions" }
```

**Step 2: 验证文件可以编译**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go build ./tutorial/...`
Expected: No errors

**Step 3: Commit**

```bash
git add tutorial/models.go
git commit -m "feat(tutorial): add model definitions for all test scenarios"
```

---

## Task 3: 生成 Schema 文件

**Files:**
- Create: `tutorial/generate.go`

**Step 1: 创建 generate.go 文件（手写 Schema，不依赖代码生成器）**

```go
package tutorial

import (
	"database/sql"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
)

// ==================== Product Schema ====================

type ProductSchemaType struct {
	ID          field.Comparable[uint64]
	Name        field.Pattern[string]
	Category    field.Pattern[string]
	Price       field.Comparable[float64]
	Stock       field.Comparable[int]
	Description field.Pattern[sql.NullString]
	CreatedAt   field.Comparable[time.Time]
	UpdatedAt   field.Comparable[time.Time]
	fieldType   Product
	alias       string
	tableName   string
}

func (t ProductSchemaType) TableName() string { return t.tableName }
func (t ProductSchemaType) Alias() string     { return t.alias }

func (t *ProductSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Name = t.Name.WithTable(&tn)
	t.Category = t.Category.WithTable(&tn)
	t.Price = t.Price.WithTable(&tn)
	t.Stock = t.Stock.WithTable(&tn)
	t.Description = t.Description.WithTable(&tn)
	t.CreatedAt = t.CreatedAt.WithTable(&tn)
	t.UpdatedAt = t.UpdatedAt.WithTable(&tn)
}

func (t ProductSchemaType) As(alias string) ProductSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t ProductSchemaType) ModelType() *Product { return &t.fieldType }

func (t ProductSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.Name, t.Category, t.Price, t.Stock, t.Description, t.CreatedAt, t.UpdatedAt}
}

func (t ProductSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var ProductSchema = ProductSchemaType{
	tableName:   "products",
	ID:          field.NewComparable[uint64]("products", "id", field.FlagPrimaryKey),
	Name:        field.NewPattern[string]("products", "name"),
	Category:    field.NewPattern[string]("products", "category"),
	Price:       field.NewComparable[float64]("products", "price"),
	Stock:       field.NewComparable[int]("products", "stock"),
	Description: field.NewPattern[sql.NullString]("products", "description"),
	CreatedAt:   field.NewComparable[time.Time]("products", "created_at"),
	UpdatedAt:   field.NewComparable[time.Time]("products", "updated_at"),
}

// ==================== Employee Schema ====================

type EmployeeSchemaType struct {
	ID         field.Comparable[uint64]
	Name       field.Pattern[string]
	Email      field.Pattern[string]
	Department field.Pattern[string]
	Salary     field.Comparable[float64]
	HireDate   field.Comparable[time.Time]
	BirthDate  field.Comparable[time.Time]
	IsActive   field.Comparable[bool]
	fieldType  Employee
	alias      string
	tableName  string
}

func (t EmployeeSchemaType) TableName() string { return t.tableName }
func (t EmployeeSchemaType) Alias() string     { return t.alias }

func (t *EmployeeSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Name = t.Name.WithTable(&tn)
	t.Email = t.Email.WithTable(&tn)
	t.Department = t.Department.WithTable(&tn)
	t.Salary = t.Salary.WithTable(&tn)
	t.HireDate = t.HireDate.WithTable(&tn)
	t.BirthDate = t.BirthDate.WithTable(&tn)
	t.IsActive = t.IsActive.WithTable(&tn)
}

func (t EmployeeSchemaType) As(alias string) EmployeeSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t EmployeeSchemaType) ModelType() *Employee { return &t.fieldType }

func (t EmployeeSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.Name, t.Email, t.Department, t.Salary, t.HireDate, t.BirthDate, t.IsActive}
}

func (t EmployeeSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var EmployeeSchema = EmployeeSchemaType{
	tableName:  "employees",
	ID:         field.NewComparable[uint64]("employees", "id", field.FlagPrimaryKey),
	Name:       field.NewPattern[string]("employees", "name"),
	Email:      field.NewPattern[string]("employees", "email"),
	Department: field.NewPattern[string]("employees", "department"),
	Salary:     field.NewComparable[float64]("employees", "salary"),
	HireDate:   field.NewComparable[time.Time]("employees", "hire_date"),
	BirthDate:  field.NewComparable[time.Time]("employees", "birth_date"),
	IsActive:   field.NewComparable[bool]("employees", "is_active"),
}

// ==================== Customer Schema ====================

type CustomerSchemaType struct {
	ID        field.Comparable[uint64]
	Name      field.Pattern[string]
	Email     field.Pattern[string]
	Phone     field.Pattern[string]
	CreatedAt field.Comparable[time.Time]
	fieldType Customer
	alias     string
	tableName string
}

func (t CustomerSchemaType) TableName() string { return t.tableName }
func (t CustomerSchemaType) Alias() string     { return t.alias }

func (t *CustomerSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Name = t.Name.WithTable(&tn)
	t.Email = t.Email.WithTable(&tn)
	t.Phone = t.Phone.WithTable(&tn)
	t.CreatedAt = t.CreatedAt.WithTable(&tn)
}

func (t CustomerSchemaType) As(alias string) CustomerSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t CustomerSchemaType) ModelType() *Customer { return &t.fieldType }

func (t CustomerSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.Name, t.Email, t.Phone, t.CreatedAt}
}

func (t CustomerSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var CustomerSchema = CustomerSchemaType{
	tableName: "customers",
	ID:        field.NewComparable[uint64]("customers", "id", field.FlagPrimaryKey),
	Name:      field.NewPattern[string]("customers", "name"),
	Email:     field.NewPattern[string]("customers", "email"),
	Phone:     field.NewPattern[string]("customers", "phone"),
	CreatedAt: field.NewComparable[time.Time]("customers", "created_at"),
}

// ==================== Order Schema ====================

type OrderSchemaType struct {
	ID         field.Comparable[uint64]
	CustomerID field.Comparable[uint64]
	OrderDate  field.Comparable[time.Time]
	TotalPrice field.Comparable[float64]
	Status     field.Pattern[string]
	fieldType  Order
	alias      string
	tableName  string
}

func (t OrderSchemaType) TableName() string { return t.tableName }
func (t OrderSchemaType) Alias() string     { return t.alias }

func (t *OrderSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.CustomerID = t.CustomerID.WithTable(&tn)
	t.OrderDate = t.OrderDate.WithTable(&tn)
	t.TotalPrice = t.TotalPrice.WithTable(&tn)
	t.Status = t.Status.WithTable(&tn)
}

func (t OrderSchemaType) As(alias string) OrderSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t OrderSchemaType) ModelType() *Order { return &t.fieldType }

func (t OrderSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.CustomerID, t.OrderDate, t.TotalPrice, t.Status}
}

func (t OrderSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var OrderSchema = OrderSchemaType{
	tableName:  "orders",
	ID:         field.NewComparable[uint64]("orders", "id", field.FlagPrimaryKey),
	CustomerID: field.NewComparable[uint64]("orders", "customer_id", field.FlagIndex),
	OrderDate:  field.NewComparable[time.Time]("orders", "order_date"),
	TotalPrice: field.NewComparable[float64]("orders", "total_price"),
	Status:     field.NewPattern[string]("orders", "status"),
}

// ==================== OrderItem Schema ====================

type OrderItemSchemaType struct {
	ID        field.Comparable[uint64]
	OrderID   field.Comparable[uint64]
	ProductID field.Comparable[uint64]
	Quantity  field.Comparable[int]
	UnitPrice field.Comparable[float64]
	fieldType OrderItem
	alias     string
	tableName string
}

func (t OrderItemSchemaType) TableName() string { return t.tableName }
func (t OrderItemSchemaType) Alias() string     { return t.alias }

func (t *OrderItemSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.OrderID = t.OrderID.WithTable(&tn)
	t.ProductID = t.ProductID.WithTable(&tn)
	t.Quantity = t.Quantity.WithTable(&tn)
	t.UnitPrice = t.UnitPrice.WithTable(&tn)
}

func (t OrderItemSchemaType) As(alias string) OrderItemSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t OrderItemSchemaType) ModelType() *OrderItem { return &t.fieldType }

func (t OrderItemSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.OrderID, t.ProductID, t.Quantity, t.UnitPrice}
}

func (t OrderItemSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var OrderItemSchema = OrderItemSchemaType{
	tableName: "order_items",
	ID:        field.NewComparable[uint64]("order_items", "id", field.FlagPrimaryKey),
	OrderID:   field.NewComparable[uint64]("order_items", "order_id", field.FlagIndex),
	ProductID: field.NewComparable[uint64]("order_items", "product_id", field.FlagIndex),
	Quantity:  field.NewComparable[int]("order_items", "quantity"),
	UnitPrice: field.NewComparable[float64]("order_items", "unit_price"),
}

// ==================== SalesRecord Schema ====================

type SalesRecordSchemaType struct {
	ID          field.Comparable[uint64]
	Region      field.Pattern[string]
	Salesperson field.Pattern[string]
	Amount      field.Comparable[float64]
	SaleDate    field.Comparable[time.Time]
	fieldType   SalesRecord
	alias       string
	tableName   string
}

func (t SalesRecordSchemaType) TableName() string { return t.tableName }
func (t SalesRecordSchemaType) Alias() string     { return t.alias }

func (t *SalesRecordSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Region = t.Region.WithTable(&tn)
	t.Salesperson = t.Salesperson.WithTable(&tn)
	t.Amount = t.Amount.WithTable(&tn)
	t.SaleDate = t.SaleDate.WithTable(&tn)
}

func (t SalesRecordSchemaType) As(alias string) SalesRecordSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t SalesRecordSchemaType) ModelType() *SalesRecord { return &t.fieldType }

func (t SalesRecordSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.Region, t.Salesperson, t.Amount, t.SaleDate}
}

func (t SalesRecordSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var SalesRecordSchema = SalesRecordSchemaType{
	tableName:   "sales_records",
	ID:          field.NewComparable[uint64]("sales_records", "id", field.FlagPrimaryKey),
	Region:      field.NewPattern[string]("sales_records", "region", field.FlagIndex),
	Salesperson: field.NewPattern[string]("sales_records", "salesperson"),
	Amount:      field.NewComparable[float64]("sales_records", "amount"),
	SaleDate:    field.NewComparable[time.Time]("sales_records", "sale_date", field.FlagIndex),
}

// ==================== OrgNode Schema ====================

type OrgNodeSchemaType struct {
	ID        field.Comparable[uint64]
	Name      field.Pattern[string]
	ParentID  field.Comparable[*uint64]
	Level     field.Comparable[int]
	fieldType OrgNode
	alias     string
	tableName string
}

func (t OrgNodeSchemaType) TableName() string { return t.tableName }
func (t OrgNodeSchemaType) Alias() string     { return t.alias }

func (t *OrgNodeSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Name = t.Name.WithTable(&tn)
	t.ParentID = t.ParentID.WithTable(&tn)
	t.Level = t.Level.WithTable(&tn)
}

func (t OrgNodeSchemaType) As(alias string) OrgNodeSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t OrgNodeSchemaType) ModelType() *OrgNode { return &t.fieldType }

func (t OrgNodeSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.Name, t.ParentID, t.Level}
}

func (t OrgNodeSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var OrgNodeSchema = OrgNodeSchemaType{
	tableName: "org_nodes",
	ID:        field.NewComparable[uint64]("org_nodes", "id", field.FlagPrimaryKey),
	Name:      field.NewPattern[string]("org_nodes", "name"),
	ParentID:  field.NewComparable[*uint64]("org_nodes", "parent_id", field.FlagIndex),
	Level:     field.NewComparable[int]("org_nodes", "level"),
}

// ==================== UserProfile Schema ====================

type UserProfileSchemaType struct {
	ID        field.Comparable[uint64]
	Username  field.Pattern[string]
	Profile   field.Pattern[string]
	fieldType UserProfile
	alias     string
	tableName string
}

func (t UserProfileSchemaType) TableName() string { return t.tableName }
func (t UserProfileSchemaType) Alias() string     { return t.alias }

func (t *UserProfileSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Username = t.Username.WithTable(&tn)
	t.Profile = t.Profile.WithTable(&tn)
}

func (t UserProfileSchemaType) As(alias string) UserProfileSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t UserProfileSchemaType) ModelType() *UserProfile { return &t.fieldType }

func (t UserProfileSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.Username, t.Profile}
}

func (t UserProfileSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var UserProfileSchema = UserProfileSchemaType{
	tableName: "user_profiles",
	ID:        field.NewComparable[uint64]("user_profiles", "id", field.FlagPrimaryKey),
	Username:  field.NewPattern[string]("user_profiles", "username"),
	Profile:   field.NewPattern[string]("user_profiles", "profile"),
}

// ==================== Transaction Schema ====================

type TransactionSchemaType struct {
	ID        field.Comparable[uint64]
	AccountID field.Comparable[uint64]
	Amount    field.Comparable[float64]
	Type      field.Pattern[string]
	CreatedAt field.Comparable[time.Time]
	fieldType Transaction
	alias     string
	tableName string
}

func (t TransactionSchemaType) TableName() string { return t.tableName }
func (t TransactionSchemaType) Alias() string     { return t.alias }

func (t *TransactionSchemaType) WithTable(tableName string) {
	tn := gsql.TableName(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.AccountID = t.AccountID.WithTable(&tn)
	t.Amount = t.Amount.WithTable(&tn)
	t.Type = t.Type.WithTable(&tn)
	t.CreatedAt = t.CreatedAt.WithTable(&tn)
}

func (t TransactionSchemaType) As(alias string) TransactionSchemaType {
	ret := t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t TransactionSchemaType) ModelType() *Transaction { return &t.fieldType }

func (t TransactionSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{t.ID, t.AccountID, t.Amount, t.Type, t.CreatedAt}
}

func (t TransactionSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var TransactionSchema = TransactionSchemaType{
	tableName: "transactions",
	ID:        field.NewComparable[uint64]("transactions", "id", field.FlagPrimaryKey),
	AccountID: field.NewComparable[uint64]("transactions", "account_id", field.FlagIndex),
	Amount:    field.NewComparable[float64]("transactions", "amount"),
	Type:      field.NewPattern[string]("transactions", "type"),
	CreatedAt: field.NewComparable[time.Time]("transactions", "created_at"),
}
```

**Step 2: 验证文件可以编译**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go build ./tutorial/...`
Expected: No errors

**Step 3: Commit**

```bash
git add tutorial/generate.go
git commit -m "feat(tutorial): add schema definitions for all models"
```

---

## Task 4: 创建 basic_test.go - 基础 CRUD 测试

**Files:**
- Create: `tutorial/basic_test.go`

**Step 1: 创建 basic_test.go 文件（第一部分：CRUD 测试）**

```go
package tutorial

import (
	"testing"
	"time"

	"github.com/donutnomad/gsql"
)

// ==================== 基础 CRUD 测试 ====================

func TestBasic_Select(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	products := []Product{
		{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100},
		{Name: "MacBook", Category: "Electronics", Price: 1999.99, Stock: 50},
		{Name: "iPad", Category: "Electronics", Price: 799.99, Stock: 200},
		{Name: "AirPods", Category: "Accessories", Price: 199.99, Stock: 500},
	}
	db.Create(&products)

	// 测试 1: WHERE + ORDER BY
	var result []Product
	err := gsql.Select(p.AllFields()...).
		From(&p).
		Where(p.Category.Eq("Electronics")).
		Order(p.Price, true). // DESC
		Find(db, &result)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 products, got %d", len(result))
	}
	if result[0].Name != "MacBook" {
		t.Errorf("Expected MacBook first, got %s", result[0].Name)
	}

	// 测试 2: LIMIT + OFFSET
	var paged []Product
	err = gsql.Select(p.AllFields()...).
		From(&p).
		Order(p.Price, false). // ASC
		Limit(2).
		Offset(1).
		Find(db, &paged)

	if err != nil {
		t.Fatalf("Paged query failed: %v", err)
	}
	if len(paged) != 2 {
		t.Errorf("Expected 2 products, got %d", len(paged))
	}
}

func TestBasic_Insert(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 单条插入
	product := Product{Name: "Test Product", Category: "Test", Price: 9.99, Stock: 10}
	err := db.Create(&product).Error
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	if product.ID == 0 {
		t.Error("Expected auto-generated ID")
	}

	// 批量插入
	batch := []Product{
		{Name: "Batch1", Category: "Batch", Price: 1.00, Stock: 1},
		{Name: "Batch2", Category: "Batch", Price: 2.00, Stock: 2},
		{Name: "Batch3", Category: "Batch", Price: 3.00, Stock: 3},
	}
	err = db.Create(&batch).Error
	if err != nil {
		t.Fatalf("Batch insert failed: %v", err)
	}

	// 验证数量
	count, err := gsql.Select(p.ID).From(&p).Where(p.Category.Eq("Batch")).Count(db)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 batch products, got %d", count)
	}
}

func TestBasic_Update(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	product := Product{Name: "Update Test", Category: "Test", Price: 10.00, Stock: 10}
	db.Create(&product)

	// 条件更新
	result := db.Model(&Product{}).
		Where("id = ?", product.ID).
		Updates(map[string]any{"price": 20.00, "stock": 20})

	if result.Error != nil {
		t.Fatalf("Update failed: %v", result.Error)
	}
	if result.RowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	// 验证更新
	var updated Product
	db.First(&updated, product.ID)
	if updated.Price != 20.00 {
		t.Errorf("Expected price 20.00, got %f", updated.Price)
	}
}

func TestBasic_Delete(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	products := []Product{
		{Name: "Delete1", Category: "Delete", Price: 1.00, Stock: 1},
		{Name: "Delete2", Category: "Delete", Price: 2.00, Stock: 2},
		{Name: "Keep", Category: "Keep", Price: 3.00, Stock: 3},
	}
	db.Create(&products)

	// 条件删除
	result := db.Where("category = ?", "Delete").Delete(&Product{})
	if result.Error != nil {
		t.Fatalf("Delete failed: %v", result.Error)
	}
	if result.RowsAffected != 2 {
		t.Errorf("Expected 2 rows deleted, got %d", result.RowsAffected)
	}

	// 验证删除
	count, _ := gsql.Select(p.ID).From(&p).Count(db)
	if count != 1 {
		t.Errorf("Expected 1 product remaining, got %d", count)
	}
}

// ==================== 日期时间函数测试 ====================

func TestFunc_DateTime(t *testing.T) {
	e := EmployeeSchema
	setupTable(t, e.ModelType())
	db := getDB()

	// 插入测试数据
	hireDate, _ := time.Parse("2006-01-02", "2020-06-15")
	birthDate, _ := time.Parse("2006-01-02", "1990-03-20")
	employee := Employee{
		Name:       "John Doe",
		Email:      "john@example.com",
		Department: "Engineering",
		Salary:     50000,
		HireDate:   hireDate,
		BirthDate:  birthDate,
		IsActive:   true,
	}
	db.Create(&employee)

	// NOW()
	var now time.Time
	db.Raw("SELECT NOW()").Scan(&now)
	if now.IsZero() {
		t.Error("NOW() returned zero time")
	}

	// YEAR(), MONTH(), DAY()
	var year, month, day int
	db.Raw("SELECT YEAR(?), MONTH(?), DAY(?)", hireDate, hireDate, hireDate).Row().Scan(&year, &month, &day)
	if year != 2020 || month != 6 || day != 15 {
		t.Errorf("Date parts mismatch: %d-%d-%d", year, month, day)
	}

	// DATE_FORMAT()
	var formatted string
	db.Raw("SELECT DATE_FORMAT(?, '%Y年%m月%d日')", hireDate).Scan(&formatted)
	if formatted != "2020年06月15日" {
		t.Errorf("Expected '2020年06月15日', got '%s'", formatted)
	}

	// DATEDIFF()
	var diff int
	db.Raw("SELECT DATEDIFF(?, ?)", hireDate, birthDate).Scan(&diff)
	if diff <= 0 {
		t.Error("DATEDIFF should return positive days")
	}

	// DATE_ADD()
	var futureDate time.Time
	db.Raw("SELECT DATE_ADD(?, INTERVAL 1 YEAR)", hireDate).Scan(&futureDate)
	if futureDate.Year() != 2021 {
		t.Errorf("Expected 2021, got %d", futureDate.Year())
	}

	t.Log("DateTime functions test passed")
}

// ==================== 字符串函数测试 ====================

func TestFunc_String(t *testing.T) {
	db := getDB()

	// CONCAT()
	var concat string
	db.Raw("SELECT CONCAT('Hello', ' ', 'World')").Scan(&concat)
	if concat != "Hello World" {
		t.Errorf("CONCAT: expected 'Hello World', got '%s'", concat)
	}

	// UPPER() / LOWER()
	var upper, lower string
	db.Raw("SELECT UPPER('hello'), LOWER('WORLD')").Row().Scan(&upper, &lower)
	if upper != "HELLO" || lower != "world" {
		t.Errorf("UPPER/LOWER: got '%s'/'%s'", upper, lower)
	}

	// SUBSTRING()
	var substr string
	db.Raw("SELECT SUBSTRING('Hello World', 7, 5)").Scan(&substr)
	if substr != "World" {
		t.Errorf("SUBSTRING: expected 'World', got '%s'", substr)
	}

	// LENGTH() / CHAR_LENGTH()
	var byteLen, charLen int
	db.Raw("SELECT LENGTH('中文'), CHAR_LENGTH('中文')").Row().Scan(&byteLen, &charLen)
	if charLen != 2 {
		t.Errorf("CHAR_LENGTH: expected 2, got %d", charLen)
	}

	// TRIM()
	var trimmed string
	db.Raw("SELECT TRIM('  hello  ')").Scan(&trimmed)
	if trimmed != "hello" {
		t.Errorf("TRIM: expected 'hello', got '%s'", trimmed)
	}

	// REPLACE()
	var replaced string
	db.Raw("SELECT REPLACE('Hello World', 'World', 'Go')").Scan(&replaced)
	if replaced != "Hello Go" {
		t.Errorf("REPLACE: expected 'Hello Go', got '%s'", replaced)
	}

	t.Log("String functions test passed")
}

// ==================== 数值函数测试 ====================

func TestFunc_Numeric(t *testing.T) {
	db := getDB()

	// ABS()
	var abs float64
	db.Raw("SELECT ABS(-42.5)").Scan(&abs)
	if abs != 42.5 {
		t.Errorf("ABS: expected 42.5, got %f", abs)
	}

	// CEIL() / FLOOR()
	var ceil, floor int
	db.Raw("SELECT CEIL(4.2), FLOOR(4.8)").Row().Scan(&ceil, &floor)
	if ceil != 5 || floor != 4 {
		t.Errorf("CEIL/FLOOR: got %d/%d", ceil, floor)
	}

	// ROUND()
	var rounded float64
	db.Raw("SELECT ROUND(3.14159, 2)").Scan(&rounded)
	if rounded != 3.14 {
		t.Errorf("ROUND: expected 3.14, got %f", rounded)
	}

	// MOD()
	var mod int
	db.Raw("SELECT MOD(10, 3)").Scan(&mod)
	if mod != 1 {
		t.Errorf("MOD: expected 1, got %d", mod)
	}

	// POWER()
	var power float64
	db.Raw("SELECT POWER(2, 10)").Scan(&power)
	if power != 1024 {
		t.Errorf("POWER: expected 1024, got %f", power)
	}

	// SQRT()
	var sqrt float64
	db.Raw("SELECT SQRT(16)").Scan(&sqrt)
	if sqrt != 4 {
		t.Errorf("SQRT: expected 4, got %f", sqrt)
	}

	t.Log("Numeric functions test passed")
}

// ==================== 聚合函数测试 ====================

func TestFunc_Aggregate(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	products := []Product{
		{Name: "P1", Category: "A", Price: 100, Stock: 10},
		{Name: "P2", Category: "A", Price: 200, Stock: 20},
		{Name: "P3", Category: "B", Price: 300, Stock: 30},
		{Name: "P4", Category: "B", Price: 400, Stock: 40},
	}
	db.Create(&products)

	// COUNT()
	count, err := gsql.Select(p.ID).From(&p).Count(db)
	if err != nil || count != 4 {
		t.Errorf("COUNT: expected 4, got %d", count)
	}

	// SUM()
	var sum float64
	db.Model(&Product{}).Select("SUM(price)").Scan(&sum)
	if sum != 1000 {
		t.Errorf("SUM: expected 1000, got %f", sum)
	}

	// AVG()
	var avg float64
	db.Model(&Product{}).Select("AVG(price)").Scan(&avg)
	if avg != 250 {
		t.Errorf("AVG: expected 250, got %f", avg)
	}

	// MAX() / MIN()
	var max, min float64
	db.Model(&Product{}).Select("MAX(price), MIN(price)").Row().Scan(&max, &min)
	if max != 400 || min != 100 {
		t.Errorf("MAX/MIN: got %f/%f", max, min)
	}

	t.Log("Aggregate functions test passed")
}

// ==================== 流程控制函数测试 ====================

func TestFunc_FlowControl(t *testing.T) {
	db := getDB()

	// IF()
	var ifResult string
	db.Raw("SELECT IF(1 > 0, 'yes', 'no')").Scan(&ifResult)
	if ifResult != "yes" {
		t.Errorf("IF: expected 'yes', got '%s'", ifResult)
	}

	// IFNULL()
	var ifnullResult string
	db.Raw("SELECT IFNULL(NULL, 'default')").Scan(&ifnullResult)
	if ifnullResult != "default" {
		t.Errorf("IFNULL: expected 'default', got '%s'", ifnullResult)
	}

	// NULLIF()
	var nullifResult *string
	db.Raw("SELECT NULLIF('a', 'a')").Scan(&nullifResult)
	if nullifResult != nil {
		t.Errorf("NULLIF: expected NULL, got '%v'", nullifResult)
	}

	t.Log("Flow control functions test passed")
}

// ==================== 类型转换函数测试 ====================

func TestFunc_TypeConvert(t *testing.T) {
	db := getDB()

	// CAST()
	var castInt int
	db.Raw("SELECT CAST('123' AS SIGNED)").Scan(&castInt)
	if castInt != 123 {
		t.Errorf("CAST: expected 123, got %d", castInt)
	}

	// CONVERT()
	var convertStr string
	db.Raw("SELECT CONVERT(123, CHAR)").Scan(&convertStr)
	if convertStr != "123" {
		t.Errorf("CONVERT: expected '123', got '%s'", convertStr)
	}

	t.Log("Type convert functions test passed")
}

// ==================== 算术运算测试 ====================

func TestFunc_Arithmetic(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	product := Product{Name: "Test", Category: "Test", Price: 100, Stock: 50}
	db.Create(&product)

	// Add
	addExpr := gsql.Add(p.Price.ToExpr(), gsql.Lit(50).ToExpr())
	var addResult float64
	gsql.Select(addExpr.AsF("result")).From(&p).Where(p.ID.Eq(product.ID)).First(db, &addResult)
	if addResult != 150 {
		t.Errorf("Add: expected 150, got %f", addResult)
	}

	// Sub
	subExpr := gsql.Sub(p.Price.ToExpr(), gsql.Lit(30).ToExpr())
	var subResult float64
	gsql.Select(subExpr.AsF("result")).From(&p).Where(p.ID.Eq(product.ID)).First(db, &subResult)
	if subResult != 70 {
		t.Errorf("Sub: expected 70, got %f", subResult)
	}

	// Mul
	mulExpr := gsql.Mul(p.Price.ToExpr(), gsql.Lit(2).ToExpr())
	var mulResult float64
	gsql.Select(mulExpr.AsF("result")).From(&p).Where(p.ID.Eq(product.ID)).First(db, &mulResult)
	if mulResult != 200 {
		t.Errorf("Mul: expected 200, got %f", mulResult)
	}

	// Div
	divExpr := gsql.Div(p.Price.ToExpr(), gsql.Lit(4).ToExpr())
	var divResult float64
	gsql.Select(divExpr.AsF("result")).From(&p).Where(p.ID.Eq(product.ID)).First(db, &divResult)
	if divResult != 25 {
		t.Errorf("Div: expected 25, got %f", divResult)
	}

	// Mod
	modExpr := gsql.Mod(p.Stock.ToExpr(), gsql.Lit(7).ToExpr())
	var modResult int
	gsql.Select(modExpr.AsF("result")).From(&p).Where(p.ID.Eq(product.ID)).First(db, &modResult)
	if modResult != 1 {
		t.Errorf("Mod: expected 1, got %d", modResult)
	}

	t.Log("Arithmetic functions test passed")
}
```

**Step 2: 验证测试文件可以编译**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go build ./tutorial/...`
Expected: No errors

**Step 3: Commit**

```bash
git add tutorial/basic_test.go
git commit -m "feat(tutorial): add basic CRUD and function tests"
```

---

## Task 5: 创建 intermediate_test.go - JOIN/子查询/GROUP BY 测试

**Files:**
- Create: `tutorial/intermediate_test.go`

**Step 1: 创建 intermediate_test.go 文件**

```go
package tutorial

import (
	"testing"
	"time"

	"github.com/donutnomad/gsql"
)

// ==================== JOIN 测试 ====================

func TestInter_InnerJoin(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// 插入测试数据
	customers := []Customer{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}
	db.Create(&customers)

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 200.00, Status: "pending"},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 150.00, Status: "completed"},
	}
	db.Create(&orders)

	// INNER JOIN
	type Result struct {
		CustomerName string  `gorm:"column:customer_name"`
		TotalPrice   float64 `gorm:"column:total_price"`
	}

	var results []Result
	err := gsql.Select(c.Name.AsF("customer_name"), o.TotalPrice).
		From(&o).
		Join(gsql.InnerJoin(&c).On(o.CustomerID.EqF(c.ID))).
		Where(o.Status.Eq("completed")).
		Find(db, &results)

	if err != nil {
		t.Fatalf("INNER JOIN failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	t.Log("INNER JOIN test passed")
}

func TestInter_LeftJoin(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// 插入测试数据：Charlie 没有订单
	customers := []Customer{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Charlie", Email: "charlie@example.com"},
	}
	db.Create(&customers)

	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.00, Status: "completed"},
	}
	db.Create(&orders)

	// LEFT JOIN - 包含没有订单的客户
	type Result struct {
		CustomerName string   `gorm:"column:customer_name"`
		TotalPrice   *float64 `gorm:"column:total_price"` // 可能为 NULL
	}

	var results []Result
	err := gsql.Select(c.Name.AsF("customer_name"), o.TotalPrice).
		From(&c).
		Join(gsql.LeftJoin(&o).On(c.ID.EqF(o.CustomerID))).
		Find(db, &results)

	if err != nil {
		t.Fatalf("LEFT JOIN failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results (including customer without order), got %d", len(results))
	}

	// 验证有一个客户的 TotalPrice 为 NULL
	var hasNull bool
	for _, r := range results {
		if r.TotalPrice == nil {
			hasNull = true
			break
		}
	}
	if !hasNull {
		t.Error("Expected at least one NULL total_price")
	}

	t.Log("LEFT JOIN test passed")
}

func TestInter_MultiJoin(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	oi := OrderItemSchema
	p := ProductSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	setupTable(t, oi.ModelType())
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	customer := Customer{Name: "Alice", Email: "alice@example.com"}
	db.Create(&customer)

	product := Product{Name: "iPhone", Category: "Electronics", Price: 999.99, Stock: 100}
	db.Create(&product)

	order := Order{CustomerID: customer.ID, OrderDate: time.Now(), TotalPrice: 999.99, Status: "completed"}
	db.Create(&order)

	orderItem := OrderItem{OrderID: order.ID, ProductID: product.ID, Quantity: 1, UnitPrice: 999.99}
	db.Create(&orderItem)

	// 三表 JOIN
	type Result struct {
		CustomerName string  `gorm:"column:customer_name"`
		ProductName  string  `gorm:"column:product_name"`
		Quantity     int     `gorm:"column:quantity"`
		UnitPrice    float64 `gorm:"column:unit_price"`
	}

	var results []Result
	err := gsql.Select(
		c.Name.AsF("customer_name"),
		p.Name.AsF("product_name"),
		oi.Quantity,
		oi.UnitPrice,
	).
		From(&o).
		Join(
			gsql.InnerJoin(&c).On(o.CustomerID.EqF(c.ID)),
			gsql.InnerJoin(&oi).On(o.ID.EqF(oi.OrderID)),
			gsql.InnerJoin(&p).On(oi.ProductID.EqF(p.ID)),
		).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Multi JOIN failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].CustomerName != "Alice" || results[0].ProductName != "iPhone" {
		t.Errorf("Unexpected result: %+v", results[0])
	}

	t.Log("Multi JOIN test passed")
}

// ==================== 子查询测试 ====================

func TestInter_SubqueryInWhere(t *testing.T) {
	c := CustomerSchema
	o := OrderSchema
	setupTable(t, c.ModelType())
	setupTable(t, o.ModelType())
	db := getDB()

	// 插入测试数据
	customers := []Customer{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
		{Name: "Charlie", Email: "charlie@example.com"},
	}
	db.Create(&customers)

	// 只有 Alice 和 Bob 有订单
	orders := []Order{
		{CustomerID: customers[0].ID, OrderDate: time.Now(), TotalPrice: 100.00},
		{CustomerID: customers[1].ID, OrderDate: time.Now(), TotalPrice: 200.00},
	}
	db.Create(&orders)

	// 子查询：查找有订单的客户
	subquery := gsql.Select(o.CustomerID).From(&o)

	var results []Customer
	err := gsql.Select(c.AllFields()...).
		From(&c).
		Where(c.ID.InSubquery(subquery)).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Subquery in WHERE failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 customers with orders, got %d", len(results))
	}

	t.Log("Subquery in WHERE test passed")
}

// ==================== GROUP BY / HAVING 测试 ====================

func TestInter_GroupBy(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	products := []Product{
		{Name: "P1", Category: "Electronics", Price: 100, Stock: 10},
		{Name: "P2", Category: "Electronics", Price: 200, Stock: 20},
		{Name: "P3", Category: "Clothing", Price: 50, Stock: 100},
		{Name: "P4", Category: "Clothing", Price: 80, Stock: 50},
		{Name: "P5", Category: "Books", Price: 20, Stock: 200},
	}
	db.Create(&products)

	// GROUP BY
	type Result struct {
		Category   string  `gorm:"column:category"`
		TotalStock int     `gorm:"column:total_stock"`
		AvgPrice   float64 `gorm:"column:avg_price"`
	}

	var results []Result
	err := gsql.Select(
		p.Category,
		gsql.SUM(p.Stock.ToExpr()).AsF("total_stock"),
		gsql.AVG(p.Price.ToExpr()).AsF("avg_price"),
	).
		From(&p).
		GroupBy(p.Category).
		Find(db, &results)

	if err != nil {
		t.Fatalf("GROUP BY failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(results))
	}

	t.Log("GROUP BY test passed")
}

func TestInter_Having(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	products := []Product{
		{Name: "P1", Category: "Electronics", Price: 100, Stock: 10},
		{Name: "P2", Category: "Electronics", Price: 200, Stock: 20},
		{Name: "P3", Category: "Clothing", Price: 50, Stock: 100},
		{Name: "P4", Category: "Books", Price: 20, Stock: 200},
	}
	db.Create(&products)

	// GROUP BY + HAVING
	type Result struct {
		Category   string `gorm:"column:category"`
		TotalStock int    `gorm:"column:total_stock"`
	}

	var results []Result
	err := gsql.Select(
		p.Category,
		gsql.SUM(p.Stock.ToExpr()).AsF("total_stock"),
	).
		From(&p).
		GroupBy(p.Category).
		Having(gsql.Expr("SUM(stock) > ?", 50)).
		Find(db, &results)

	if err != nil {
		t.Fatalf("HAVING failed: %v", err)
	}
	// 只有 Clothing (150) 和 Books (200) 的 stock > 50
	if len(results) != 2 {
		t.Errorf("Expected 2 categories with stock > 50, got %d", len(results))
	}

	t.Log("HAVING test passed")
}

// ==================== UNION 测试 ====================

func TestInter_Union(t *testing.T) {
	e := EmployeeSchema
	setupTable(t, e.ModelType())
	db := getDB()

	// 插入测试数据
	employees := []Employee{
		{Name: "Alice", Email: "alice@example.com", Department: "Engineering", Salary: 80000},
		{Name: "Bob", Email: "bob@example.com", Department: "Marketing", Salary: 60000},
		{Name: "Charlie", Email: "charlie@example.com", Department: "Engineering", Salary: 90000},
	}
	db.Create(&employees)

	// UNION: 工程部员工 UNION 高薪员工
	query1 := gsql.Select(e.Name, e.Department, e.Salary).
		From(&e).
		Where(e.Department.Eq("Engineering"))

	query2 := gsql.Select(e.Name, e.Department, e.Salary).
		From(&e).
		Where(e.Salary.Gt(70000))

	type Result struct {
		Name       string  `gorm:"column:name"`
		Department string  `gorm:"column:department"`
		Salary     float64 `gorm:"column:salary"`
	}

	var results []Result
	err := query1.Union(query2).Find(db, &results)

	if err != nil {
		t.Fatalf("UNION failed: %v", err)
	}
	// Alice 和 Charlie 在工程部，Alice 和 Charlie 高薪（去重后应该是 2 人）
	if len(results) != 2 {
		t.Errorf("Expected 2 results after UNION, got %d", len(results))
	}

	t.Log("UNION test passed")
}

// ==================== CASE WHEN 测试 ====================

func TestInter_CaseWhen(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	products := []Product{
		{Name: "P1", Category: "Electronics", Price: 1000, Stock: 10},
		{Name: "P2", Category: "Electronics", Price: 500, Stock: 20},
		{Name: "P3", Category: "Clothing", Price: 100, Stock: 100},
	}
	db.Create(&products)

	// CASE WHEN
	priceLevel := gsql.Case().
		When(p.Price.Gte(800), gsql.Lit("High")).
		When(p.Price.Gte(300), gsql.Lit("Medium")).
		Else(gsql.Lit("Low")).
		End().AsF("price_level")

	type Result struct {
		Name       string `gorm:"column:name"`
		Price      float64
		PriceLevel string `gorm:"column:price_level"`
	}

	var results []Result
	err := gsql.Select(p.Name, p.Price, priceLevel).
		From(&p).
		Order(p.Price, true).
		Find(db, &results)

	if err != nil {
		t.Fatalf("CASE WHEN failed: %v", err)
	}

	// 验证结果
	expectedLevels := map[string]string{
		"P1": "High",
		"P2": "Medium",
		"P3": "Low",
	}
	for _, r := range results {
		if r.PriceLevel != expectedLevels[r.Name] {
			t.Errorf("Product %s: expected %s, got %s", r.Name, expectedLevels[r.Name], r.PriceLevel)
		}
	}

	t.Log("CASE WHEN test passed")
}

// ==================== 索引提示测试 ====================

func TestInter_IndexHint(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 创建索引
	db.Exec("CREATE INDEX idx_products_category ON products(category)")
	db.Exec("CREATE INDEX idx_products_price ON products(price)")

	// 插入测试数据
	products := []Product{
		{Name: "P1", Category: "Electronics", Price: 100},
		{Name: "P2", Category: "Electronics", Price: 200},
	}
	db.Create(&products)

	// USE INDEX
	sql1 := gsql.Select(p.AllFields()...).
		From(&p).
		UseIndex("idx_products_category").
		Where(p.Category.Eq("Electronics")).
		ToSQL()

	if sql1 == "" {
		t.Error("USE INDEX SQL should not be empty")
	}
	t.Logf("USE INDEX SQL: %s", sql1)

	// FORCE INDEX
	sql2 := gsql.Select(p.AllFields()...).
		From(&p).
		ForceIndex("idx_products_price").
		Where(p.Price.Gt(50)).
		ToSQL()

	if sql2 == "" {
		t.Error("FORCE INDEX SQL should not be empty")
	}
	t.Logf("FORCE INDEX SQL: %s", sql2)

	t.Log("Index hint test passed")
}
```

**Step 2: 验证测试文件可以编译**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go build ./tutorial/...`
Expected: No errors

**Step 3: Commit**

```bash
git add tutorial/intermediate_test.go
git commit -m "feat(tutorial): add intermediate tests for JOIN, subquery, GROUP BY, UNION, CASE WHEN"
```

---

## Task 6: 创建 advanced_test.go - CTE/窗口函数/JSON/锁 测试

**Files:**
- Create: `tutorial/advanced_test.go`

**Step 1: 创建 advanced_test.go 文件**

```go
package tutorial

import (
	"testing"
	"time"

	"github.com/donutnomad/gsql"
)

// ==================== CTE 测试 ====================

func TestAdv_BasicCTE(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// 插入测试数据
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 1500, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Charlie", Amount: 2000, SaleDate: time.Now()},
	}
	db.Create(&records)

	// 基础 CTE
	cte := gsql.With("high_sales",
		gsql.Select(s.AllFields()...).
			From(&s).
			Where(s.Amount.Gt(1200)),
	)

	type Result struct {
		Salesperson string  `gorm:"column:salesperson"`
		Amount      float64 `gorm:"column:amount"`
	}

	var results []Result
	err := cte.Select(gsql.Field("salesperson"), gsql.Field("amount")).
		From(gsql.TN("high_sales")).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Basic CTE failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 high sales records, got %d", len(results))
	}

	t.Log("Basic CTE test passed")
}

func TestAdv_RecursiveCTE(t *testing.T) {
	o := OrgNodeSchema
	setupTable(t, o.ModelType())
	db := getDB()

	// 插入组织树数据
	// CEO (id=1) -> VP1 (id=2), VP2 (id=3) -> Manager1 (id=4)
	nodes := []OrgNode{
		{ID: 1, Name: "CEO", ParentID: nil, Level: 0},
		{ID: 2, Name: "VP Engineering", ParentID: ptr(uint64(1)), Level: 1},
		{ID: 3, Name: "VP Sales", ParentID: ptr(uint64(1)), Level: 1},
		{ID: 4, Name: "Manager", ParentID: ptr(uint64(2)), Level: 2},
	}
	for _, node := range nodes {
		db.Create(&node)
	}

	// 递归 CTE - 获取组织树
	sql := gsql.WithRecursive("org_tree",
		gsql.Select(o.ID, o.Name, o.ParentID, o.Level).
			From(&o).
			Where(o.ParentID.IsNull()),
	).Select(gsql.Star).
		From(gsql.TN("org_tree")).
		ToSQL()

	if sql == "" {
		t.Error("Recursive CTE SQL should not be empty")
	}
	t.Logf("Recursive CTE SQL: %s", sql)

	t.Log("Recursive CTE test passed")
}

func TestAdv_MultipleCTE(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// 插入测试数据
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Bob", Amount: 2000, SaleDate: time.Now()},
	}
	db.Create(&records)

	// 多个 CTE
	sql := gsql.With("north_sales",
		gsql.Select(s.AllFields()...).From(&s).Where(s.Region.Eq("North")),
	).And("south_sales",
		gsql.Select(s.AllFields()...).From(&s).Where(s.Region.Eq("South")),
	).Select(gsql.Star).
		From(gsql.TN("north_sales")).
		ToSQL()

	if sql == "" {
		t.Error("Multiple CTE SQL should not be empty")
	}
	t.Logf("Multiple CTE SQL: %s", sql)

	t.Log("Multiple CTE test passed")
}

// ==================== 窗口函数测试 ====================

func TestAdv_RowNumber(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// 插入测试数据
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 1500, SaleDate: time.Now()},
		{Region: "South", Salesperson: "Charlie", Amount: 2000, SaleDate: time.Now()},
		{Region: "South", Salesperson: "David", Amount: 1800, SaleDate: time.Now()},
	}
	db.Create(&records)

	// ROW_NUMBER() OVER (PARTITION BY region ORDER BY amount DESC)
	rn := gsql.RowNumber().
		PartitionBy(s.Region).
		OrderBy(s.Amount, true).
		AsF("row_num")

	type Result struct {
		Region      string  `gorm:"column:region"`
		Salesperson string  `gorm:"column:salesperson"`
		Amount      float64 `gorm:"column:amount"`
		RowNum      int     `gorm:"column:row_num"`
	}

	var results []Result
	err := gsql.Select(s.Region, s.Salesperson, s.Amount, rn).
		From(&s).
		Find(db, &results)

	if err != nil {
		t.Fatalf("ROW_NUMBER failed: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(results))
	}

	// 验证每个区域的第一名
	for _, r := range results {
		if r.RowNum == 1 {
			if r.Region == "North" && r.Salesperson != "Bob" {
				t.Errorf("North #1 should be Bob, got %s", r.Salesperson)
			}
			if r.Region == "South" && r.Salesperson != "Charlie" {
				t.Errorf("South #1 should be Charlie, got %s", r.Salesperson)
			}
		}
	}

	t.Log("ROW_NUMBER test passed")
}

func TestAdv_RankDenseRank(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// 插入测试数据 (包含相同金额)
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 1000, SaleDate: time.Now()},
		{Region: "North", Salesperson: "Bob", Amount: 1000, SaleDate: time.Now()}, // 同金额
		{Region: "North", Salesperson: "Charlie", Amount: 500, SaleDate: time.Now()},
	}
	db.Create(&records)

	// RANK()
	rank := gsql.Rank().
		OrderBy(s.Amount, true).
		AsF("rank_num")

	// DENSE_RANK()
	denseRank := gsql.DenseRank().
		OrderBy(s.Amount, true).
		AsF("dense_rank_num")

	type Result struct {
		Salesperson  string  `gorm:"column:salesperson"`
		Amount       float64 `gorm:"column:amount"`
		RankNum      int     `gorm:"column:rank_num"`
		DenseRankNum int     `gorm:"column:dense_rank_num"`
	}

	var results []Result
	err := gsql.Select(s.Salesperson, s.Amount, rank, denseRank).
		From(&s).
		Find(db, &results)

	if err != nil {
		t.Fatalf("RANK/DENSE_RANK failed: %v", err)
	}

	t.Logf("RANK/DENSE_RANK results: %+v", results)
	t.Log("RANK/DENSE_RANK test passed")
}

func TestAdv_LagLead(t *testing.T) {
	s := SalesRecordSchema
	setupTable(t, s.ModelType())
	db := getDB()

	// 插入测试数据
	baseTime := time.Now()
	records := []SalesRecord{
		{Region: "North", Salesperson: "Alice", Amount: 100, SaleDate: baseTime},
		{Region: "North", Salesperson: "Alice", Amount: 150, SaleDate: baseTime.Add(24 * time.Hour)},
		{Region: "North", Salesperson: "Alice", Amount: 200, SaleDate: baseTime.Add(48 * time.Hour)},
	}
	db.Create(&records)

	// LAG() - 获取前一条记录的值
	lag := gsql.Lag(s.Amount, 1, gsql.Lit(0)).
		OrderBy(s.SaleDate, false).
		AsF("prev_amount")

	// LEAD() - 获取后一条记录的值
	lead := gsql.Lead(s.Amount, 1, gsql.Lit(0)).
		OrderBy(s.SaleDate, false).
		AsF("next_amount")

	type Result struct {
		Amount     float64 `gorm:"column:amount"`
		PrevAmount float64 `gorm:"column:prev_amount"`
		NextAmount float64 `gorm:"column:next_amount"`
	}

	var results []Result
	err := gsql.Select(s.Amount, lag, lead).
		From(&s).
		Where(s.Salesperson.Eq("Alice")).
		Find(db, &results)

	if err != nil {
		t.Fatalf("LAG/LEAD failed: %v", err)
	}

	t.Logf("LAG/LEAD results: %+v", results)
	t.Log("LAG/LEAD test passed")
}

// ==================== JSON 测试 ====================

func TestAdv_JsonExtract(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// 插入 JSON 数据
	profiles := []UserProfile{
		{Username: "alice", Profile: `{"age": 25, "city": "New York", "tags": ["developer", "golang"]}`},
		{Username: "bob", Profile: `{"age": 30, "city": "San Francisco", "tags": ["manager"]}`},
	}
	db.Create(&profiles)

	// JSON_EXTRACT
	type Result struct {
		Username string `gorm:"column:username"`
		Age      int    `gorm:"column:age"`
		City     string `gorm:"column:city"`
	}

	var results []Result
	err := db.Raw(`
		SELECT username,
		       JSON_EXTRACT(profile, '$.age') as age,
		       JSON_UNQUOTE(JSON_EXTRACT(profile, '$.city')) as city
		FROM user_profiles
		WHERE JSON_EXTRACT(profile, '$.age') > 20
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("JSON_EXTRACT failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	t.Log("JSON_EXTRACT test passed")
}

func TestAdv_JsonModify(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// 插入 JSON 数据
	profile := UserProfile{Username: "alice", Profile: `{"age": 25, "city": "New York"}`}
	db.Create(&profile)

	// JSON_SET - 添加/更新字段
	err := db.Exec(`
		UPDATE user_profiles
		SET profile = JSON_SET(profile, '$.country', 'USA', '$.age', 26)
		WHERE username = 'alice'
	`).Error

	if err != nil {
		t.Fatalf("JSON_SET failed: %v", err)
	}

	// 验证更新
	var updated UserProfile
	db.First(&updated, profile.ID)

	var country string
	db.Raw(`SELECT JSON_UNQUOTE(JSON_EXTRACT(profile, '$.country')) FROM user_profiles WHERE id = ?`, profile.ID).Scan(&country)
	if country != "USA" {
		t.Errorf("Expected country 'USA', got '%s'", country)
	}

	t.Log("JSON_MODIFY test passed")
}

func TestAdv_JsonQuery(t *testing.T) {
	u := UserProfileSchema
	setupTable(t, u.ModelType())
	db := getDB()

	// 插入 JSON 数据
	profiles := []UserProfile{
		{Username: "alice", Profile: `{"skills": ["go", "python", "javascript"]}`},
		{Username: "bob", Profile: `{"skills": ["java", "python"]}`},
	}
	db.Create(&profiles)

	// JSON_CONTAINS - 查找包含特定技能的用户
	var results []UserProfile
	err := db.Raw(`
		SELECT * FROM user_profiles
		WHERE JSON_CONTAINS(profile, '"go"', '$.skills')
	`).Scan(&results).Error

	if err != nil {
		t.Fatalf("JSON_CONTAINS failed: %v", err)
	}
	if len(results) != 1 || results[0].Username != "alice" {
		t.Errorf("Expected alice with go skill, got %+v", results)
	}

	t.Log("JSON_QUERY test passed")
}

// ==================== 锁测试 ====================

func TestAdv_ForUpdate(t *testing.T) {
	tx := TransactionSchema
	setupTable(t, tx.ModelType())
	db := getDB()

	// 插入测试数据
	trans := Transaction{AccountID: 1, Amount: 100.00, Type: "credit"}
	db.Create(&trans)

	// FOR UPDATE
	sql := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForUpdate().
		ToSQL()

	if sql == "" {
		t.Error("FOR UPDATE SQL should not be empty")
	}
	t.Logf("FOR UPDATE SQL: %s", sql)

	// NOWAIT
	sqlNowait := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForUpdate().
		Nowait().
		ToSQL()

	t.Logf("FOR UPDATE NOWAIT SQL: %s", sqlNowait)

	t.Log("FOR UPDATE test passed")
}

func TestAdv_LockInShare(t *testing.T) {
	tx := TransactionSchema
	setupTable(t, tx.ModelType())
	db := getDB()

	// 插入测试数据
	trans := Transaction{AccountID: 1, Amount: 100.00, Type: "credit"}
	db.Create(&trans)

	// LOCK IN SHARE MODE
	sql := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForShare().
		ToSQL()

	if sql == "" {
		t.Error("LOCK IN SHARE MODE SQL should not be empty")
	}
	t.Logf("LOCK IN SHARE MODE SQL: %s", sql)

	// SKIP LOCKED
	sqlSkip := gsql.Select(tx.AllFields()...).
		From(&tx).
		Where(tx.AccountID.Eq(1)).
		ForShare().
		SkipLocked().
		ToSQL()

	t.Logf("FOR SHARE SKIP LOCKED SQL: %s", sqlSkip)

	t.Log("LOCK IN SHARE MODE test passed")
}

// ==================== BatchIn 测试 ====================

func TestAdv_BatchIn(t *testing.T) {
	p := ProductSchema
	setupTable(t, p.ModelType())
	db := getDB()

	// 插入测试数据
	var products []Product
	for i := 0; i < 100; i++ {
		products = append(products, Product{
			Name:     "Product" + string(rune('A'+i%26)),
			Category: "Category" + string(rune('0'+i%10)),
			Price:    float64(i * 10),
			Stock:    i,
		})
	}
	db.Create(&products)

	// 收集 ID
	var ids []uint64
	for _, p := range products {
		ids = append(ids, p.ID)
	}

	// 使用 BatchIn（这里数据量小，会降级为普通 IN）
	batchIn := gsql.BatchIn(p.ID, ids)

	// 直接使用表达式（小数据量不需要临时表）
	expr := batchIn.ToExpression()

	var results []Product
	err := gsql.Select(p.AllFields()...).
		From(&p).
		Where(expr).
		Find(db, &results)

	if err != nil {
		t.Fatalf("BatchIn failed: %v", err)
	}
	if len(results) != 100 {
		t.Errorf("Expected 100 products, got %d", len(results))
	}

	t.Log("BatchIn test passed")
}

// ==================== 辅助函数 ====================

func ptr[T any](v T) *T {
	return &v
}
```

**Step 2: 验证测试文件可以编译**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go build ./tutorial/...`
Expected: No errors

**Step 3: Commit**

```bash
git add tutorial/advanced_test.go
git commit -m "feat(tutorial): add advanced tests for CTE, window functions, JSON, locks, BatchIn"
```

---

## Task 7: 运行完整测试套件

**Step 1: 运行所有 tutorial 测试**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go test -v ./tutorial/... -timeout 5m`
Expected: All tests pass (需要 Docker 运行 testcontainer)

**Step 2: 如果有测试失败，修复问题并重新运行**

**Step 3: Commit 任何修复**

```bash
git add -A
git commit -m "fix(tutorial): fix test issues discovered during integration testing"
```

---

## Task 8: 最终验证和文档更新

**Step 1: 运行完整项目测试**

Run: `cd /Users/ubuntu/Projects/go/cuti/gsql/.worktrees/tutorial && go test ./...`
Expected: No regression in existing tests

**Step 2: 更新设计文档状态**

将 `docs/plans/2026-01-17-tutorial-design.md` 中添加完成状态标记。

**Step 3: Final Commit**

```bash
git add -A
git commit -m "docs(tutorial): mark tutorial implementation as complete"
```

---

## 执行摘要

| Task | 描述 | 文件 |
|------|------|------|
| 1 | Testcontainer 基础设施 | `tutorial/testcontainer_test.go` |
| 2 | 基础 Models | `tutorial/models.go` |
| 3 | Schema 定义 | `tutorial/generate.go` |
| 4 | 基础 CRUD + 函数测试 | `tutorial/basic_test.go` |
| 5 | JOIN/子查询/GROUP BY | `tutorial/intermediate_test.go` |
| 6 | CTE/窗口函数/JSON/锁 | `tutorial/advanced_test.go` |
| 7 | 集成测试运行 | - |
| 8 | 最终验证 | - |

**预计测试覆盖：**
- 基础 CRUD：4 个测试
- 函数测试：8 组（DateTime、String、Numeric、Aggregate、FlowControl、TypeConvert、Arithmetic）
- JOIN 测试：3 个（Inner、Left、Multi）
- 子查询测试：1 个
- GROUP BY/HAVING：2 个
- UNION：1 个
- CASE WHEN：1 个
- 索引提示：1 个
- CTE 测试：3 个（Basic、Recursive、Multiple）
- 窗口函数：3 个（RowNumber、Rank/DenseRank、Lag/Lead）
- JSON 测试：3 个（Extract、Modify、Query）
- 锁测试：2 个（ForUpdate、ForShare）
- BatchIn：1 个

**总计：约 30+ 个测试函数**
