# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

GSQL 是一个基于 GORM 的类型安全 SQL 查询构建器库，使用 Go 泛型提供流畅的链式 API。主要特性包括：
- 类型安全的查询构建
- CASE WHEN 表达式构建器
- CTE (Common Table Expressions) 支持
- BatchIn 优化器用于大型 IN 查询
- 100+ MySQL 函数封装
- 子查询和 JOIN 支持
- JSON 字段操作

## 开发命令

### 运行测试
```bash
# 运行所有测试
go test -v ./...

# 运行特定包的测试
go test -v ./field/...
go test -v .

# 运行特定测试文件
go test -v -run TestName

# 运行测试并显示覆盖率
go test -v -cover ./...
```

### 代码检查和格式化
```bash
# 格式化代码
go fmt ./...
gofmt -w .

# 使用 goimports 格式化并管理导入
goimports -w ./...

# 运行 go vet 检查
go vet ./...

# 使用 staticcheck (如果安装)
staticcheck ./...
```

### 构建
```bash
# 构建项目
go build ./...

# 检查依赖
go mod tidy
go mod verify
```

## 核心架构

### 1. 查询构建器层次结构

项目使用了分层的查询构建器设计:

- **baseQueryBuilder / baseQueryBuilderG[T]**: 初始构建器，只处理 SELECT 和 CTE
  - 使用 `Select()` 添加字段
  - 使用 `From()` 转换为完整的 QueryBuilder

- **QueryBuilder / QueryBuilderG[T]**: 完整的查询构建器
  - `QueryBuilder` 是 `QueryBuilderG[any]` 的类型别名，用于非泛型场景
  - `QueryBuilderG[T]` 是泛型版本，提供类型安全的模型操作
  - 包含所有查询子句：WHERE, JOIN, ORDER BY, GROUP BY, HAVING, LIMIT, OFFSET
  - 支持索引提示、分区、锁定等高级特性

### 2. Field 系统

field 包提供了类型安全的字段表示:

- **IField**: 所有字段的基础接口
- **Comparable[T]**: 支持比较操作 (=, !=, >, >=, <, <=, IN, NOT IN)
  - 使用 `field.NewComparable[T](table, column)` 创建
  - 示例: `id := field.NewComparable[int64]("users", "id")`

- **Pattern[T]**: 支持模式匹配 (LIKE, NOT LIKE, Contains, HasPrefix, HasSuffix)
  - 使用 `field.NewPattern[T](table, column)` 创建
  - 示例: `name := field.NewPattern[string]("users", "name")`

- **BaseFields**: 特殊类型，表示一组字段（如 Star 表示 SELECT *）

### 3. Clause 系统

每个 SQL 子句都有对应的实现:

- **clause_case.go**: CASE WHEN 表达式
  - 使用 `gsql.Case().When().Else().End()` 构建

- **clause_cte.go**: CTE (WITH 子句)
  - 使用 `query.With(name, subquery)` 添加

- **clause_batch_in.go**: 大型 IN 查询优化
  - 使用临时表策略处理 10000+ 条目的 IN 查询

- **clause_join.go**: JOIN 操作
  - 支持 INNER JOIN, LEFT JOIN, RIGHT JOIN, FULL JOIN

- **clause_json.go**: JSON 操作
  - JSON_OBJECT, JSON_ARRAY 等函数

### 4. 函数系统 (functions.go)

提供 100+ MySQL 函数的 Go 封装:
- 日期/时间函数: NOW, FROM_UNIXTIME, DATE_FORMAT 等
- 字符串函数: CONCAT, UPPER, SUBSTRING 等
- 数值函数: ABS, CEIL, FLOOR, ROUND 等
- 聚合函数: COUNT, SUM, AVG, MAX, MIN 等
- 条件函数: IF, IFNULL, CASE 等

所有函数返回 `ExprTo` 类型，可以使用 `.AsF(alias)` 添加别名

### 5. 查询执行

查询通过 `IDB` 接口执行，主要实现是 `DefaultGormDB`:

执行方法:
- `Find(db, &dest)`: 查询多条记录
- `First(db, &dest)`: 查询第一条记录
- `Count(db)`: 计数
- `Exist(db)`: 检查是否存在
- `Update(db, values)`: 更新
- `Delete(db, &dest)`: 删除

## 代码约定

### 命名规范
- 公共 API 使用 PascalCase (如 `SelectG`, `QueryBuilderG`)
- 带 `G` 后缀的是泛型版本 (如 `SelectG[T]`, `QueryBuilderG[T]`)
- 字段构造函数使用 `New` 前缀 (如 `NewComparable`, `NewPattern`)
- 聚合函数和 MySQL 函数使用大写 (如 `COUNT`, `SUM`, `CONCAT`)

### 类型安全设计
- 使用泛型确保字段类型与操作匹配
- `Comparable[T]` 只允许可比较类型
- `Pattern[T]` 主要用于字符串类型
- 函数返回 `ExprTo` 提供统一的表达式接口

### 方法链式调用
所有查询构建方法返回 `*QueryBuilder` 或 `*QueryBuilderG[T]`，支持链式调用:
```go
query.From(table).
    Where(condition).
    Order(field, true).
    Limit(10)
```

### Expression vs IField
- `IField`: 代表数据库字段（有表名和列名）
- `Expression`: 代表任意 SQL 表达式（可能是计算、函数调用等）
- 使用 `.ToExpr()` 将 IField 转换为 Expression
- 使用 `.AsF(alias)` 将 Expression 转换为带别名的 IField

## 重要文件说明

- **query.go**: 非泛型 QueryBuilder 定义和基础方法
- **query_generic.go**: 泛型 QueryBuilderG[T] 的完整实现
- **public.go**: 公共辅助函数和便捷方法
- **interfaces.go**: 核心接口定义 (IDB, IField, ITableName 等)
- **utils.go**: 内部工具函数
- **field/**: 字段类型系统的完整实现

## 测试结构

- 单元测试文件使用 `*_test.go` 命名
- 示例测试在 `example_*_test.go` 中
- 主要测试目录: `./`, `./field/`, `./scopes/`
- 测试使用 MySQL 驱动，需要实际数据库连接（见各测试文件的 setup）

## 扩展点

添加新功能时考虑以下扩展点:
1. 新的字段类型: 在 `field/` 包中添加
2. 新的 MySQL 函数: 在 `functions.go` 中添加
3. 新的 SQL 子句: 创建 `clause_*.go` 文件
4. 新的查询方法: 在 `query_generic.go` 中添加到 `QueryBuilderG[T]`

## 注意事项

1. **GORM 版本依赖**: 项目依赖 GORM v1.31.0+，某些内部 API 可能在不同版本中变化
2. **MySQL 特定功能**: 某些功能（如 CTE）需要 MySQL 8.0+
3. **性能考虑**: 大型 IN 查询应使用 `BatchIn` 优化器
4. **类型约束**: 泛型参数 T 必须是可以被 GORM 处理的类型
5. **沙盒模式命令限制**: 在沙盒模式中执行 go 命令时，必须直接使用 `go` 开头，组合命令会被阻止
   - ✅ `go test -v -run TestName ./tutorial/...`
   - ❌ `cd tutorial && go test -v -run TestName ./...`
