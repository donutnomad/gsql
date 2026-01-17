package tutorial

import (
	"database/sql"
	"time"
)

// ==================== Basic Models ====================

// Product 产品表 - 用于基础 CRUD 和函数测试
// @Gsql
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
// @Gsql
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
// @Gsql
type Customer struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;size:100;not null"`
	Email     string    `gorm:"column:email;size:100"`
	Phone     string    `gorm:"column:phone;size:20"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Customer) TableName() string { return "customers" }

// Order 订单表 - 用于 JOIN 和聚合测试
// @Gsql
type Order struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	CustomerID uint64    `gorm:"column:customer_id;index"`
	OrderDate  time.Time `gorm:"column:order_date"`
	TotalPrice float64   `gorm:"column:total_price"`
	Status     string    `gorm:"column:status;size:20;default:'pending'"`
}

func (Order) TableName() string { return "orders" }

// OrderItem 订单项表 - 用于多表 JOIN
// @Gsql
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
// @Gsql
type SalesRecord struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Region      string    `gorm:"column:region;size:50;index"`
	Salesperson string    `gorm:"column:salesperson;size:100"`
	Amount      float64   `gorm:"column:amount"`
	SaleDate    time.Time `gorm:"column:sale_date;index"`
}

func (SalesRecord) TableName() string { return "sales_records" }

// OrgNode 组织节点 - 用于递归 CTE 测试
// @Gsql
type OrgNode struct {
	ID       uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name     string  `gorm:"column:name;size:100;not null"`
	ParentID *uint64 `gorm:"column:parent_id;index"`
	Level    int     `gorm:"column:level;default:0"`
}

func (OrgNode) TableName() string { return "org_nodes" }

// UserProfile 用户配置 - 用于 JSON 测试
// @Gsql
type UserProfile struct {
	ID       uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Username string `gorm:"column:username;size:100;not null"`
	Profile  string `gorm:"column:profile;type:json"` // JSON 字段
}

func (UserProfile) TableName() string { return "user_profiles" }

// Transaction 交易记录 - 用于锁测试
// @Gsql
type Transaction struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	AccountID uint64    `gorm:"column:account_id;index"`
	Amount    float64   `gorm:"column:amount"`
	Type      string    `gorm:"column:type;size:20"` // credit/debit
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Transaction) TableName() string { return "transactions" }
