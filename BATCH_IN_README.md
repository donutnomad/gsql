# 大批量 IN/NOT IN 优化器 - 临时表方案

## 概述

`BatchInOptimizer` 是一个实验性功能，用于优化包含大量值的 IN/NOT IN 查询。当 IN 列表中的值数量超过阈值时，会自动创建临时表并建立索引，显著提升查询性能。

## 功能特性

- **真正的临时表**: 使用 `CREATE TEMPORARY TABLE` 创建临时表，支持索引
- **自动优化**: 根据值的数量自动选择最优策略
- **随机表名**: 自动生成唯一的临时表名，避免冲突
- **批量插入**: 支持配置批量插入大小，优化数据导入
- **类型推断**: 自动推断列类型并创建合适的表结构
- **自动清理**: 提供清理函数，确保资源正确释放

## 支持的策略

### 1. 临时表策略 (StrategyTempTable) ⭐ 推荐

创建临时表，插入数据，然后使用 IN (SELECT * FROM temp_table)。

**优点**:
- **性能最优**: 可以创建主键索引，查询速度快
- **SQL 简洁**: 查询表达式极短（~60 字符），不受数据量影响
- **适合超大数据量**: 10万+ 数据也能高效处理
- **可复用**: 同一个临时表可以被多次查询使用
- **自动清理**: 会话结束后自动删除，无需手动管理

**缺点**:
- 需要执行多条 SQL（创建表、插入数据、查询）
- 需要手动调用 Execute 方法

**生成的 SQL**:
```sql
-- 1. 创建临时表（带主键索引）
CREATE TEMPORARY TABLE tmp_batch_in_xxx (val BIGINT, PRIMARY KEY (val))

-- 2. 批量插入数据
INSERT INTO tmp_batch_in_xxx (val) VALUES (1), (2), ..., (1000)
INSERT INTO tmp_batch_in_xxx (val) VALUES (1001), (1002), ..., (2000)
...

-- 3. 使用临时表查询
SELECT * FROM users WHERE id IN (SELECT val FROM tmp_batch_in_xxx)

-- 4. 清理（可选，会话结束后自动删除）
DROP TEMPORARY TABLE IF EXISTS tmp_batch_in_xxx
```

### 2. 分片策略 (StrategyChunk)

将大 IN 列表拆分成多个小 IN，用 OR/AND 连接。

**优点**:
- 兼容性最好，适用于所有数据库
- 不需要额外的 SQL 步骤
- 可直接转为表达式使用

**缺点**:
- SQL 语句较长
- OR 连接可能影响执行计划

**生成 SQL 示例**:
```sql
SELECT * FROM users 
WHERE (id IN (1,2,...,1000) OR id IN (1001,1002,...,2000))
```

## 使用方法

### 基本用法（临时表策略）

```go
import "github.com/donutnomad/gsql"

// 假设有大量 ID
var userIDs []int64 // 10000 个 ID

// 步骤 1: 创建批量 IN 对象
batchIn := BatchIn(u.ID, userIDs)

// 步骤 2: 执行临时表流程（创建表、插入数据）
expr, cleanup, err := batchIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup() // 确保清理临时表

// 步骤 3: 使用返回的表达式查询
query := Select(u.ALL).From(u).Where(expr)
var users []User
err = query.Find(db, &users)
```

### 自动清理版本

```go
// 更简洁的写法（自动延迟清理）
batchIn := BatchIn(u.ID, userIDs)
expr, err := batchIn.ExecuteWithAutoCleanup(db)
if err != nil {
    log.Fatal(err)
}
// 注意：这个版本会立即 defer cleanup，所以确保在同一个函数内完成查询

query := Select(u.ALL).From(u).Where(expr)
var users []User
err = query.Find(db, &users)
```

### 使用分片策略（无需 Execute）

```go
// 配置分片策略
config := BatchInConfig{
    Strategy:  StrategyChunk,
    Threshold: 500,
    ChunkSize: 500, // 每个 IN 子句 500 个值
}

batchIn := BatchInWithConfig(u.ID, userIDs, config)

// 直接转为表达式使用（无需 Execute）
expr := batchIn.ToExpression()

query := Select(u.ALL).From(u).Where(expr)
var users []User
err := query.Find(db, &users)
```

### 自定义配置

```go
// 配置临时表参数
config := BatchInConfig{
    Strategy:        StrategyTempTable,
    Threshold:       2000,  // 超过 2000 个值才使用临时表
    InsertBatchSize: 5000,  // 每批次插入 5000 条数据
}

batchIn := BatchInWithConfig(u.ID, userIDs, config)
expr, cleanup, err := batchIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// 使用表达式查询
query := Select(u.ALL).From(u).Where(expr)
var users []User
err = query.Find(db, &users)
```

### NOT IN 查询

```go
// 排除大量 ID
var excludeIDs []int64 // 5000 个 ID

batchNotIn := BatchNotIn(u.ID, excludeIDs)
expr, cleanup, err := batchNotIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

query := Select(u.ALL).From(u).Where(expr)
var users []User
err := query.Find(db, &users)
```

### 复用临时表

```go
// 创建临时表
batchIn := BatchIn(u.ID, userIDs)
expr, cleanup, err := batchIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// 查询 1
var activeUsers []User
err = Select(u.ALL).
    From(u).
    Where(expr, u.Status.Eq("active")).
    Find(db, &activeUsers)

// 查询 2: 复用同一个临时表
var inactiveUsers []User
err = Select(u.ALL).
    From(u).
    Where(expr, u.Status.Eq("inactive")).
    Find(db, &inactiveUsers)

// cleanup() 会在 defer 时执行，清理临时表
```

## 配置说明

### BatchInConfig 结构

```go
type BatchInConfig struct {
    // Strategy 优化策略
    // - StrategyChunk: 分片策略
    // - StrategyTempTable: 临时表策略（推荐）
    Strategy BatchInStrategy

    // Threshold 触发优化的阈值（默认 500）
    // 当 IN 列表长度 >= Threshold 时，会应用优化策略
    // 当 IN 列表长度 < Threshold 时，使用普通 IN
    // 推荐值：300-500（根据数据库类型调整）
    Threshold int

    // ChunkSize 分片大小（默认 500）
    // 仅在 StrategyChunk 策略下有效
    // 决定每个 IN 子句最多包含多少个值
    ChunkSize int

    // InsertBatchSize 插入临时表时的批次大小（默认 1000）
    // 仅在 StrategyTempTable 策略下有效
    // 决定每次 INSERT 语句插入多少条数据
    InsertBatchSize int
}
```

### 默认配置

```go
DefaultBatchInConfig() BatchInConfig {
    return BatchInConfig{
        Strategy:        StrategyTempTable,  // 使用临时表策略
        Threshold:       500,                // 500 个值以上才优化（业界推荐值）
        ChunkSize:       500,                // 分片大小 500
        InsertBatchSize: 1000,               // 每批次插入 1000 条
    }
}
```

## 性能对比

根据测试结果（5000 个值的 IN 查询）:

| 策略 | SQL 长度 | 索引支持 | 性能 | 适用场景 |
|------|----------|----------|------|---------|
| 普通 IN | 28,909 字符 | ❌ | ⭐⭐ | < 500 个值 |
| 分片策略 | 28,991 字符 | ❌ | ⭐⭐⭐ | 500-5000 个值 |
| 临时表策略 | 63 字符 | ✅ 主键索引 | ⭐⭐⭐⭐⭐ | > 1000 个值 |

**临时表优势**:
- 查询表达式长度固定（~60 字符），不受数据量影响
- 支持主键索引，查询速度极快
- 对于 10000 个值：普通 IN 需要 58,910 字符，临时表仅需 63 字符
- 执行计划稳定，不会因为值太多导致性能下降

## IN 列表大小最佳实践

根据业界经验和性能测试：

| 数量范围 | 性能影响 | 推荐方案 |
|---------|---------|---------|
| < 100 | 无 | 普通 IN，性能最优 |
| 100-300 | 轻微 | 普通 IN，可接受 |
| 300-500 | 明显 | 建议优化（临时表或分片）|
| > 500 | 严重 | 强烈建议优化 |

**不同数据库的建议**:
- **MySQL**: 建议 < 500（无硬性限制）
- **PostgreSQL**: 建议 < 1000
- **Oracle**: 硬性限制 1000
- **SQL Server**: 参数限制 2100

### 推荐配置

1. **小数据量 (< 500)**: 自动使用 **普通 IN**（默认）
   ```go
   // 阈值设置为 500，小数据量自动降级到普通 IN
   BatchInConfig{
       Strategy:  StrategyTempTable,
       Threshold: 500,  // 默认值
   }
   ```

2. **中等数据量 (500-10000)**: 使用 **临时表策略**（默认）
   ```go
   BatchInConfig{
       Strategy:        StrategyTempTable,
       Threshold:       500,  // 默认值
       InsertBatchSize: 1000,
   }
   ```

3. **大数据量 (> 10000)**: 使用 **临时表策略** + 大批次插入
   ```go
   BatchInConfig{
       Strategy:        StrategyTempTable,
       Threshold:       500,
       InsertBatchSize: 5000,  // 增大批次提升插入速度
   }
   ```

4. **超保守配置 (降低阈值到 300)**:
   ```go
   BatchInConfig{
       Strategy:  StrategyTempTable,
       Threshold: 300,  // 更早触发优化
       ChunkSize: 300,
   }
   ```

4. **旧版本数据库/不支持临时表**: 使用 **分片策略**
   ```go
   BatchInConfig{
       Strategy:  StrategyChunk,
       Threshold: 500,
       ChunkSize: 500,
   }
   ```

## 实际应用场景

### 场景 1: 批量查询用户

```go
// 从其他服务获取了大量用户 ID
userIDs := []int64{1, 2, 3, ..., 10000} // 10000 个 ID

// 批量查询用户信息
batchIn := BatchIn(u.ID, userIDs)
expr, cleanup, err := batchIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

var users []User
err = Select(u.ALL).
    From(u).
    Where(expr).
    Find(db, &users)
```

### 场景 2: 排除已处理的订单

```go
// 已处理的订单 ID
processedOrderIDs := []int64{...} // 5000 个 ID

// 查询未处理的订单
batchNotIn := BatchNotIn(o.ID, processedOrderIDs)
expr, cleanup, err := batchNotIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

var pendingOrders []Order
err = Select(o.ALL).
    From(o).
    Where(
        o.Status.Eq("pending"),
        expr,
    ).
    Find(db, &pendingOrders)
```

### 场景 3: 多条件查询

```go
// 结合其他条件
batchIn := BatchIn(u.ID, targetUserIDs)
expr, cleanup, err := batchIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

var users []User
err = Select(u.ALL).
    From(u).
    Where(
        u.Status.Eq("active"),
        u.CreatedAt.Gte(time.Now().AddDate(0, -1, 0)),
        expr,
    ).
    Order(u.CreatedAt, false).
    Limit(100).
    Find(db, &users)
```

### 场景 4: 字符串类型批量查询

```go
// 批量查询用户名
usernames := []string{"user1", "user2", ..., "user5000"}

batchIn := BatchIn(u.Username, usernames)
expr, cleanup, err := batchIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

var users []User
err = Select(u.ALL).
    From(u).
    Where(expr).
    Find(db, &users)
```

### 场景 5: 统计分析

```go
// 大批量数据统计
batchIn := BatchIn(u.ID, hugeUserIDs) // 100000 个 ID
expr, cleanup, err := batchIn.Execute(db)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// 统计查询
var stats struct {
    TotalUsers  int
    ActiveUsers int
    TotalOrders int
}

// 复用临时表进行多次统计
db.Model(&User{}).Where(expr).Count(&stats.TotalUsers)
db.Model(&User{}).Where(expr, u.Status.Eq("active")).Count(&stats.ActiveUsers)
db.Model(&Order{}).Where("user_id IN (SELECT val FROM ?)", batchIn.tempTableName).Count(&stats.TotalOrders)
```

## 技术细节

### 临时表名生成

```go
// 使用加密随机数生成 16 位十六进制字符串
func generateTempTableName() string {
    b := make([]byte, 8)
    rand.Read(b)
    return "tmp_batch_in_" + hex.EncodeToString(b)
}

// 示例: tmp_batch_in_8548fb5e8b87df0a
```

### 列类型自动推断

根据第一个值的类型自动推断列类型：

| Go 类型 | SQL 类型 |
|---------|----------|
| int, int64 | BIGINT |
| uint, uint64 | BIGINT UNSIGNED |
| float32, float64 | DOUBLE |
| string | VARCHAR(255) |
| bool | BOOLEAN |

### 主键索引

临时表自动创建主键索引，确保查询性能：

```sql
CREATE TEMPORARY TABLE tmp_batch_in_xxx (
    val BIGINT, 
    PRIMARY KEY (val)  -- 自动创建主键索引
)
```

### 批量插入

支持配置批量插入大小，平衡性能和内存使用：

```go
// 例如：10000 个值，InsertBatchSize = 1000
// 会执行 10 次 INSERT，每次插入 1000 个值
INSERT INTO tmp_batch_in_xxx (val) VALUES (1), (2), ..., (1000)
INSERT INTO tmp_batch_in_xxx (val) VALUES (1001), (1002), ..., (2000)
...
```

## 注意事项

### 1. 临时表生命周期

- 临时表在**会话（连接）级别**创建，不同连接看不到
- 会话结束后**自动删除**，无需担心残留
- 建议手动调用 `cleanup()` 以尽早释放资源

### 2. 阈值设置

- **太低**: 会导致小数据量也创建临时表，增加开销
- **太高**: 大数据量时无法触发优化
- **推荐**: 300-500 之间（默认 500）
- **保守**: 300（更早触发优化，适合性能敏感场景）
- **激进**: 800-1000（容忍更大的 IN 列表）

### 3. 数据库兼容性

- **临时表策略**: 
  - MySQL: 5.0+
  - PostgreSQL: 8.0+
  - SQLite: 3.0+
  - MariaDB: 5.1+
  
- **分片策略**: 所有数据库都支持

### 4. 性能考虑

- 临时表创建和插入数据有一定开销，但查询性能显著提升
- 对于 < 1000 个值，普通 IN 可能更快（无需创建表）
- 对于 > 10000 个值，临时表优势明显（避免 SQL 解析瓶颈）
- 临时表可被多次查询复用，摊薄创建开销

### 5. 内存和磁盘

- 临时表默认存储在内存中（具体取决于数据库配置）
- 如果数据量超过 `tmp_table_size`，会写入磁盘
- MySQL 可通过 `SET SESSION tmp_table_size = xxx` 调整

### 6. 事务处理

- 临时表在事务内创建，事务回滚时自动删除
- 多个查询可在同一事务内共享临时表

## API 参考

### 便捷函数

```go
// BatchIn 创建批量优化的 IN 对象（默认配置）
func BatchIn[T any](column field.Comparable[T], values []T) *TempTableBatchIn

// BatchNotIn 创建批量优化的 NOT IN 对象（默认配置）
func BatchNotIn[T any](column field.Comparable[T], values []T) *TempTableBatchIn

// BatchInWithConfig 创建批量优化的 IN 对象（自定义配置）
func BatchInWithConfig[T any](column field.Comparable[T], values []T, config BatchInConfig) *TempTableBatchIn

// BatchNotInWithConfig 创建批量优化的 NOT IN 对象（自定义配置）
func BatchNotInWithConfig[T any](column field.Comparable[T], values []T, config BatchInConfig) *TempTableBatchIn
```

### TempTableBatchIn 对象方法

```go
// Execute 执行完整的临时表流程（创建表、插入数据）
// 返回：查询表达式、清理函数、错误
func (t *TempTableBatchIn) Execute(db *gorm.DB) (field.Expression, func() error, error)

// ExecuteWithAutoCleanup 执行并自动注册清理（使用 defer）
// 返回：查询表达式、错误
func (t *TempTableBatchIn) ExecuteWithAutoCleanup(db *gorm.DB) (field.Expression, error)

// ToExpression 转换为普通表达式（仅用于简单 IN 或分片策略）
// 临时表策略会 panic，必须使用 Execute
func (t *TempTableBatchIn) ToExpression() field.Expression
```

## 测试

运行测试:

```bash
# 运行所有测试
go test -v ./lib/gsql -run TestBatchInOptimizer

# 运行 benchmark
go test -bench=BenchmarkBatchIn ./lib/gsql -benchmem
```

## 未来改进

- [ ] 支持组合列（多列 IN）
- [ ] 支持自动检测数据库类型并选择最优策略
- [ ] 提供统计信息（优化次数、节省的时间等）
- [ ] 支持分布式缓存临时表（Redis 等）
- [ ] 集成到 query.go 和 query_g.go（根据用户反馈决定）

## 反馈

这是一个实验性功能，欢迎提供反馈和建议！

**为什么选择临时表而不是 CTE + VALUES？**

1. **性能更优**: 临时表支持索引，CTE + VALUES 不支持
2. **可复用**: 同一个临时表可以被多次查询使用
3. **适用范围更广**: 适合超大数据量（10万+）
4. **执行计划稳定**: 数据库优化器可以更好地处理临时表
