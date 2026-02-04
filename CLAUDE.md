# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GSQL is a type-safe SQL query builder library for Go, built on top of GORM. It provides:
- Type-safe query building with Go generics
- Fluent chainable API
- CASE WHEN expressions, CTE (Common Table Expressions), UNION support
- JOINs, subqueries, and batch operations
- 100+ wrapped MySQL functions with type safety

## Build and Test Commands

```bash
# Run all tests (requires Docker for testcontainers)
go test ./...

# Run tutorial tests only (integration tests with MySQL container)
go test ./tutorial/...

# Run a specific test
go test -v -run TestBasic_Select ./tutorial/...

# Run unit tests (no database required)
go test ./internal/... ./clause/...

# Generate code (schema types from models)
go generate ./...

# Format and lint
goimports -w .
go vet ./...
```

## Architecture

### Core Query Builder Pattern

```
Select(fields) → From(table) → Where/Join/Order/Group → Execute(db)
```

The library uses a fluent builder pattern with two main types:
- `QueryBuilder` / `QueryBuilderG[T]` - Main query builder (generic version preserves model type)
- `baseQueryBuilder` / `baseQueryBuilderG[T]` - Intermediate builder before FROM clause

### Key Package Structure

```
gsql/
├── query.go, query_generic.go    # QueryBuilder and QueryBuilderG[T]
├── insert.go                     # InsertBuilder for INSERT operations
├── clause_*.go                   # SQL clause implementations (JOIN, CTE, UNION, etc.)
├── functions*.go                 # Wrapped MySQL functions (COUNT, SUM, IF, etc.)
├── typed_field.go               # Generated field types (IntField, StringField, etc.)
├── clause/                       # Expression and clause interfaces
│   └── expression.go            # Core Expression interface, builder helpers
├── field/                        # Public field type exports
│   └── base.go                  # Re-exports from internal/fields
├── internal/
│   └── fields/                  # Field type implementations
│       ├── int.go, string.go    # Type-specific fields with methods
│       └── *_generate.go        # Code generation for field types
└── tutorial/                     # Integration tests with testcontainers
    ├── models.go                # Model definitions with @Gsql annotation
    ├── generate.go              # Generated schema types (do not edit)
    └── *_test.go                # Test files by feature
```

### Schema Generation Pattern

Models are defined with `@Gsql` annotation and use `gorm` tags:

```go
// @Gsql
type Product struct {
    ID    uint64  `gorm:"column:id;primaryKey;autoIncrement"`
    Name  string  `gorm:"column:name;size:100;not null"`
    Price float64 `gorm:"column:price;not null"`
}
func (Product) TableName() string { return "products" }
```

Running `go generate ./...` creates a schema type:

```go
var ProductSchema = ProductSchemaType{
    tableName: "products",
    ID:        gsql.IntFieldOf[uint64]("products", "id", field.FlagPrimaryKey|field.FlagAutoIncrement),
    Name:      gsql.StringFieldOf[string]("products", "name"),
    Price:     gsql.FloatFieldOf[float64]("products", "price"),
}
```

### Field Type Hierarchy

Fields implement `field.IField` and provide type-safe operations:
- `IntField[T]` - numeric comparisons (Eq, Gt, Lt, Between, In)
- `StringField[T]` - string operations (Like, Contains, StartsWith)
- `DateTimeField[T]` - date/time operations
- `JsonField[T]` - JSON operations (Extract, Contains, Length)

Expression types (e.g., `IntExpr[T]`) are returned from functions and support the same operations.

### Query Execution Methods

```go
builder.Find(db, &results)     // Multiple results
builder.Take(db, &result)      // Single result, no ordering
builder.First(db, &result)     // Single result, ordered by primary key
builder.Count(db)              // Count rows
builder.Exist(db)              // Check existence
builder.Update(db, value)      // Update matching rows
builder.Delete(db, model)      // Delete matching rows
```

### Test Structure

Tests use testcontainers-go for MySQL integration:
- `TestMain` in `tutorial/testcontainer_test.go` starts MySQL container
- `setupTable[T]` helper creates/drops tables per test
- `getDB()` returns shared database connection

## Code Generation

The project uses `gogen` for generating schema types from model structs. Generated files:
- `tutorial/generate.go` - Schema types from models
- `typed_field.go` - Field type exports from internal/fields

To regenerate: `go generate ./...`

## Common Patterns

### Basic Query
```go
p := ProductSchema
gsql.Select(p.AllFields()...).
    From(&p).
    Where(p.Category.Eq("Electronics"), p.Price.Gte(500)).
    Order(p.Price, false).
    Limit(10).
    Find(db, &results)
```

### JOIN
```go
gsql.Select(o.AllFields()...).
    From(&o).
    Join(gsql.LeftJoin(&c).On(o.CustomerID.EqF(c.ID))).
    Find(db, &results)
```

### CTE
```go
gsql.With("cte_name", subquery).
    Select(fields...).
    From(gsql.TN("cte_name")).
    Find(db, &results)
```

### Subquery
```go
gsql.Select(p.Name, avgPrice.AsF("avg_price")).
    From(&p).
    Where(p.Price.GtF(
        gsql.Select(gsql.AVG(p2.Price)).From(&p2).AsF("avg"),
    ))
```
