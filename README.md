# GSQL - Type-Safe SQL Query Builder for GORM

A type-safe, fluent SQL query builder library built on top of GORM with support for complex queries, CASE expressions, CTE, batch operations, and 100+ MySQL functions.

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/)
[![GORM](https://img.shields.io/badge/GORM-v1.31.0-green)](https://gorm.io/)

## Features

- ✅ Type-safe query building with Go generics
- ✅ Fluent, chainable API
- ✅ CASE WHEN expressions builder
- ✅ CTE (Common Table Expressions) support
- ✅ BatchIn optimizer for large IN queries
- ✅ 100+ MySQL functions wrapped
- ✅ Subqueries and JOINs
- ✅ JSON field operations

## Installation

```bash
go get github.com/donutnomad/gsql
```

## Quick Start

### Basic Query

```go
package main

import (
    "github.com/donutnomad/gsql"
    "github.com/donutnomad/gsql/field"
)

func BasicQuery() string {
    // Define fields
    id := field.NewComparable[int64]("users", "id")
    name := field.NewPattern[string]("users", "name")
    age := field.NewComparable[int]("users", "age")

    // Build query
    sql := gsql.Select(id, name, age).
        From(gsql.TableName("users").Ptr()).
        Where(
            gsql.And(
                age.Gt(18),
                name.Contains("John"),
            ),
        ).
        Order(age, false).
        Limit(10).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT users.id, users.name, users.age
    // FROM users
    // WHERE users.age > 18 AND users.name LIKE '%John%'
    // ORDER BY users.age DESC
    // LIMIT 10
}
```

### CASE WHEN Expression

```go
func BuildComplexQuery() string {
    // Define fields
    amount := field.NewComparable[int64]("orders", "amount")
    userLevel := field.NewPattern[string]("orders", "user_level")
    status := field.NewPattern[string]("orders", "status")
    createdAt := field.NewComparable[int64]("orders", "created_at")

    // Build CASE expression for discount calculation
    discount := gsql.Case().
        When(
            gsql.And(
                userLevel.Eq("VIP"),
                amount.Gt(10000),
            ),
            gsql.Lit(0.7), // 30% off for VIP orders > 10000
        ).
        When(
            gsql.And(
                userLevel.Eq("Premium"),
                amount.Gt(5000),
            ),
            gsql.Lit(0.85), // 15% off for Premium orders > 5000
        ).
        When(
            gsql.Expr("first_order = ?", true),
            gsql.Lit(0.9), // 10% off for first orders
        ).
        Else(gsql.Lit(1.0)). // No discount
        End().AsF("discount_rate")

    // Build custom priority for sorting
    priority := gsql.Case().
        When(status.Eq("urgent"), gsql.Lit(1)).
        When(status.Eq("high"), gsql.Lit(2)).
        When(status.Eq("normal"), gsql.Lit(3)).
        Else(gsql.Lit(4)).
        End().AsF("priority")

    // Build complex query
    sql := gsql.Select(
        gsql.Field("id"),
        amount,
        userLevel,
        status,
        discount,
        priority,
        gsql.Mul(amount.ToExpr(), discount.ToExpr()).AsF("final_amount"),
    ).From(gsql.TableName("orders").Ptr()).
        Where(
            gsql.And(
                amount.Gt(100),
                status.In("pending", "processing", "urgent"),
                createdAt.Gte(1609459200), // Since 2021-01-01
            ),
        ).
        Order(priority, true).   // Order by priority ASC
        Order(amount, false).    // Then by amount DESC
        Limit(100).
        Offset(0).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT id, orders.amount, orders.user_level, orders.status,
    //   CASE
    //     WHEN orders.user_level = 'VIP' AND orders.amount > 10000 THEN 0.7
    //     WHEN orders.user_level = 'Premium' AND orders.amount > 5000 THEN 0.85
    //     WHEN first_order = true THEN 0.9
    //     ELSE 1.0
    //   END AS discount_rate,
    //   CASE
    //     WHEN orders.status = 'urgent' THEN 1
    //     WHEN orders.status = 'high' THEN 2
    //     WHEN orders.status = 'normal' THEN 3
    //     ELSE 4
    //   END AS priority,
    //   orders.amount * discount_rate AS final_amount
    // FROM orders
    // WHERE orders.amount > 100
    //   AND orders.status IN ('pending', 'processing', 'urgent')
    //   AND orders.created_at >= 1609459200
    // ORDER BY priority ASC, orders.amount DESC
    // LIMIT 100 OFFSET 0
}
```

### GROUP BY with Aggregation

```go
func GroupByWithAggregation() string {
    amount := field.NewComparable[int64]("orders", "amount")

    // Create amount range buckets
    amountRange := gsql.Case().
        When(amount.Lt(100), gsql.Lit("0-100")).
        When(amount.Lt(500), gsql.Lit("100-500")).
        When(amount.Lt(1000), gsql.Lit("500-1000")).
        Else(gsql.Lit("1000+")).
        End().AsF("amount_range")

    sql := gsql.Select(
        amountRange,
        gsql.COUNT().AsF("order_count"),
        gsql.SUM(amount.ToExpr()).AsF("total_amount"),
        gsql.AVG(amount.ToExpr()).AsF("avg_amount"),
    ).From(gsql.TableName("orders").Ptr()).
        GroupBy(amountRange).
        Having(gsql.Expr("COUNT(*) > ?", 10)).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT
    //   CASE
    //     WHEN orders.amount < 100 THEN '0-100'
    //     WHEN orders.amount < 500 THEN '100-500'
    //     WHEN orders.amount < 1000 THEN '500-1000'
    //     ELSE '1000+'
    //   END AS amount_range,
    //   COUNT(*) AS order_count,
    //   SUM(orders.amount) AS total_amount,
    //   AVG(orders.amount) AS avg_amount
    // FROM orders
    // GROUP BY amount_range
    // HAVING COUNT(*) > 10
}
```

### JOIN Operations

```go
func JoinQuery() string {
    // Users table
    userID := field.NewComparable[int64]("users", "id")
    userName := field.NewPattern[string]("users", "name")

    // Orders table
    orderID := field.NewComparable[int64]("orders", "id")
    orderUserID := field.NewComparable[int64]("orders", "user_id")
    orderAmount := field.NewComparable[int64]("orders", "amount")

    sql := gsql.Select(
        userName,
        gsql.COUNT(orderID).AsF("order_count"),
        gsql.SUM(orderAmount.ToExpr()).AsF("total_spent"),
    ).From(gsql.TableName("users").Ptr()).
        Join(gsql.InnerJoin(
            gsql.TableName("orders").Ptr(),
            userID.EqF(orderUserID),
        )).
        GroupBy(userID, userName).
        Having(gsql.Expr("COUNT(*) > ?", 5)).
        Order(gsql.Field("total_spent"), false).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT users.name, COUNT(orders.id) AS order_count, SUM(orders.amount) AS total_spent
    // FROM users
    // INNER JOIN orders ON users.id = orders.user_id
    // GROUP BY users.id, users.name
    // HAVING COUNT(*) > 5
    // ORDER BY total_spent DESC
}
```

### CTE (Common Table Expressions)

```go
func CTEQuery() string {
    id := field.NewComparable[int64]("users", "id")
    name := field.NewPattern[string]("users", "name")
    status := field.NewPattern[string]("users", "status")
    age := field.NewComparable[int]("users", "age")

    // Define CTE
    activeUsers := gsql.Select(id, name, age).
        From(gsql.TableName("users").Ptr()).
        Where(status.Eq("active"))

    // Use CTE in main query
    sql := gsql.Select(gsql.Star).
        From(gsql.TableName("active_users").Ptr()).
        With("active_users", activeUsers).
        Where(
            field.NewComparable[int]("active_users", "age").Gt(25),
        ).
        ToSQL()

    return sql
    // Output SQL:
    // WITH active_users AS (
    //   SELECT users.id, users.name, users.age
    //   FROM users
    //   WHERE users.status = 'active'
    // )
    // SELECT * FROM active_users WHERE active_users.age > 25
}
```

### Batch IN Optimization

```go
func BatchInQuery(db *gorm.DB, userIDs []int64) error {
    id := field.NewComparable[int64]("users", "id")

    // For large IN queries (10000+ items), use BatchIn with temp table strategy
    batchIn := gsql.BatchIn(id, userIDs)
    expr, cleanup, err := batchIn.Execute(db)
    if err != nil {
        return err
    }
    defer cleanup() // Cleanup temp table

    var users []User
    err = gsql.Select(gsql.Star).
        From(gsql.TableName("users").Ptr()).
        Where(expr).
        Find(db, &users)

    return err
    // Generated SQL:
    // CREATE TEMPORARY TABLE tmp_batch_in_xxx (val BIGINT, PRIMARY KEY (val))
    // INSERT INTO tmp_batch_in_xxx (val) VALUES (1), (2), ..., (1000)
    // ...
    // SELECT * FROM users WHERE id IN (SELECT val FROM tmp_batch_in_xxx)
}
```

### Date Functions

```go
func DateFunctions() string {
    createdAt := field.NewComparable[int64]("orders", "created_at")

    sql := gsql.Select(
        gsql.Field("id"),
        gsql.FROM_UNIXTIME(createdAt.ToExpr()).AsF("created_date"),
        gsql.DATE_FORMAT(
            gsql.FROM_UNIXTIME(createdAt.ToExpr()),
            "%Y-%m-%d",
        ).AsF("formatted_date"),
        gsql.YEAR(gsql.FROM_UNIXTIME(createdAt.ToExpr())).AsF("year"),
        gsql.MONTH(gsql.FROM_UNIXTIME(createdAt.ToExpr())).AsF("month"),
    ).From(gsql.TableName("orders").Ptr()).
        Where(
            createdAt.Gte(gsql.UNIX_TIMESTAMP(
                gsql.DATE_SUB(gsql.NOW(), "30 DAY"),
            ).ToExpr()),
        ).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT id,
    //   FROM_UNIXTIME(orders.created_at) AS created_date,
    //   DATE_FORMAT(FROM_UNIXTIME(orders.created_at), '%Y-%m-%d') AS formatted_date,
    //   YEAR(FROM_UNIXTIME(orders.created_at)) AS year,
    //   MONTH(FROM_UNIXTIME(orders.created_at)) AS month
    // FROM orders
    // WHERE orders.created_at >= UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 30 DAY))
}
```

### String Functions

```go
func StringFunctions() string {
    firstName := field.NewPattern[string]("users", "first_name")
    lastName := field.NewPattern[string]("users", "last_name")
    email := field.NewPattern[string]("users", "email")

    sql := gsql.Select(
        gsql.CONCAT(firstName.ToExpr(), gsql.Lit(" "), lastName.ToExpr()).AsF("full_name"),
        gsql.UPPER(email.ToExpr()).AsF("email_upper"),
        gsql.SUBSTRING(email.ToExpr(), 1, 5).AsF("email_prefix"),
        gsql.LENGTH(firstName.ToExpr()).AsF("name_length"),
    ).From(gsql.TableName("users").Ptr()).
        Where(
            email.HasSuffix("@example.com"),
        ).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT
    //   CONCAT(users.first_name, ' ', users.last_name) AS full_name,
    //   UPPER(users.email) AS email_upper,
    //   SUBSTRING(users.email, 1, 5) AS email_prefix,
    //   LENGTH(users.first_name) AS name_length
    // FROM users
    // WHERE users.email LIKE '%@example.com'
}
```

### Subquery

```go
func SubqueryExample() string {
    // Subquery: get average order amount
    orderAmount := field.NewComparable[int64]("orders", "amount")
    avgAmountSubquery := gsql.Select(
        gsql.AVG(orderAmount.ToExpr()).AsF("avg_amount"),
    ).From(gsql.TableName("orders").Ptr())

    // Main query: find orders above average
    sql := gsql.Select(gsql.Star).
        From(gsql.TableName("orders").Ptr()).
        Where(
            gsql.Expr("amount > (?)", avgAmountSubquery.ToExpr()),
        ).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT * FROM orders
    // WHERE amount > (SELECT AVG(orders.amount) AS avg_amount FROM orders)
}
```

### Conditional Aggregation

```go
func ConditionalAggregation() string {
    status := field.NewPattern[string]("orders", "status")
    amount := field.NewComparable[int64]("orders", "amount")
    userID := field.NewComparable[int64]("orders", "user_id")

    sql := gsql.Select(
        userID,
        gsql.COUNT().AsF("total_orders"),
        gsql.SUM(
            gsql.IF(
                status.Eq("completed"),
                amount.ToExpr(),
                gsql.Lit(0),
            ),
        ).AsF("completed_amount"),
        gsql.SUM(
            gsql.IF(
                status.Eq("pending"),
                amount.ToExpr(),
                gsql.Lit(0),
            ),
        ).AsF("pending_amount"),
    ).From(gsql.TableName("orders").Ptr()).
        GroupBy(userID).
        ToSQL()

    return sql
    // Output SQL:
    // SELECT
    //   orders.user_id,
    //   COUNT(*) AS total_orders,
    //   SUM(IF(orders.status = 'completed', orders.amount, 0)) AS completed_amount,
    //   SUM(IF(orders.status = 'pending', orders.amount, 0)) AS pending_amount
    // FROM orders
    // GROUP BY orders.user_id
}
```

## Field Types

### Comparable Fields

Supports comparison operations: `=`, `!=`, `>`, `>=`, `<`, `<=`, `IN`, `NOT IN`

```go
id := field.NewComparable[int64]("users", "id")
age := field.NewComparable[int]("users", "age")
salary := field.NewComparable[float64]("users", "salary")
createdAt := field.NewComparable[time.Time]("users", "created_at")

// Usage
id.Eq(100)              // id = 100
age.Gt(18)              // age > 18
age.Gte(18)             // age >= 18
age.Lt(60)              // age < 60
age.Lte(60)             // age <= 60
age.Not(0)              // age != 0
age.In(18, 25, 30)      // age IN (18, 25, 30)
age.NotIn(0, -1)        // age NOT IN (0, -1)
```

### Pattern Fields

Supports pattern matching: `LIKE`, `NOT LIKE`, prefix, suffix, contains

```go
name := field.NewPattern[string]("users", "name")
email := field.NewPattern[string]("users", "email")

// Usage
name.Eq("John")             // name = 'John'
name.Not("Admin")           // name != 'Admin'
name.Like("%admin%")        // name LIKE '%admin%'
name.NotLike("test%")       // name NOT LIKE 'test%'
name.Contains("John")       // name LIKE '%John%'
name.HasPrefix("admin")     // name LIKE 'admin%'
name.HasSuffix(".com")      // name LIKE '%.com'
```

## Executing Queries

```go
// Create database instance
db := gsql.NewDefaultGormDB(gormDB)

// Find multiple records
var users []User
err := query.Find(db, &users)

// Find first record
var user User
err := query.First(db, &user)

// Count records
count, err := query.Count(db)

// Check existence
exists, err := query.Exist(db)

// Update records
err := query.Where(id.Eq(100)).Update(db, map[string]any{
    "name": "New Name",
})

// Delete records
err := query.Where(id.Eq(100)).Delete(db, &User{})

// Debug mode (print SQL)
query.Debug().Find(db, &users)
```

## Advanced Features

### Index Hints

```go
query.UseIndex("idx_age")
query.ForceIndex("idx_name")
query.IgnoreIndex("idx_created_at")
query.UseIndexForJoin("idx_user_id")
query.UseIndexForOrderBy("idx_age")
```

### Row Locking

```go
query.ForUpdate()              // SELECT ... FOR UPDATE
query.ForShare()               // SELECT ... FOR SHARE
query.ForUpdate().Nowait()     // SELECT ... FOR UPDATE NOWAIT
query.ForUpdate().SkipLocked() // SELECT ... FOR UPDATE SKIP LOCKED
```

### JSON Operations

```go
jsonData := gsql.JSON_OBJECT().
    Add("id", userID.ToExpr()).
    Add("name", userName.ToExpr()).
    Add("age", userAge.ToExpr()).
    AsF("user_json")

sql := gsql.Select(jsonData).
    From(gsql.TableName("users").Ptr()).
    ToSQL()
// Output SQL:
// SELECT JSON_OBJECT('id', users.id, 'name', users.name, 'age', users.age) AS user_json
// FROM users
```

## MySQL Functions

GSQL wraps 100+ MySQL functions:

**Date/Time**: NOW, CURRENT_DATE, UNIX_TIMESTAMP, FROM_UNIXTIME, DATE_FORMAT, YEAR, MONTH, DAY, DATE_ADD, DATE_SUB, DATEDIFF, TIMESTAMPDIFF

**String**: CONCAT, CONCAT_WS, LENGTH, UPPER, LOWER, SUBSTRING, LEFT, RIGHT, TRIM, REPLACE, LOCATE

**Numeric**: ABS, CEIL, FLOOR, ROUND, MOD, POWER, SQRT, TRUNCATE

**Aggregate**: COUNT, SUM, AVG, MAX, MIN, GROUP_CONCAT, COUNT_DISTINCT

**Conditional**: IF, IFNULL, NULLIF, CASE

**Type Conversion**: CAST, CONVERT

See [functions.go](functions.go) for complete list.

## Documentation

- [CASE_WHEN_USAGE.md](docs/CASE_WHEN_USAGE.md) - CASE expression detailed usage
- [CTE_README.md](docs/CTE_README.md) - CTE feature documentation
- [BATCH_IN_README.md](docs/BATCH_IN_README.md) - BatchIn optimizer guide

## Requirements

- Go 1.25+
- GORM v1.31.0+
- MySQL 8.0+ (for CTE features)

## License

MIT License
