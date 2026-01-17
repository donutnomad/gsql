package field

// TypedColumn 类型化列工厂，用于从派生表/子查询中导出列
// 使用示例:
//   subquery := gsql.Select(gsql.COUNT().AsF("cnt")).From(orders).GroupBy(orders.UserID)
//   cnt := field.IntColumn("cnt").From(subquery)
//   mainQuery := gsql.Select(cnt).From(subquery.As("sub")).Where(cnt.Gt(10))

// IntColumn 创建整数类型列引用
func IntColumn(name string) IntColumnBuilder {
	return IntColumnBuilder{name: name}
}

// IntColumnBuilder 整数列构建器
type IntColumnBuilder struct {
	name string
}

// From 指定列来源的子查询/派生表
func (b IntColumnBuilder) From(source interface{ TableName() string }) Comparable[int64] {
	return NewComparable[int64](source.TableName(), b.name)
}

// FloatColumn 创建浮点类型列引用
func FloatColumn(name string) FloatColumnBuilder {
	return FloatColumnBuilder{name: name}
}

// FloatColumnBuilder 浮点列构建器
type FloatColumnBuilder struct {
	name string
}

// From 指定列来源的子查询/派生表
func (b FloatColumnBuilder) From(source interface{ TableName() string }) Comparable[float64] {
	return NewComparable[float64](source.TableName(), b.name)
}

// StringColumn 创建字符串类型列引用
func StringColumn(name string) StringColumnBuilder {
	return StringColumnBuilder{name: name}
}

// StringColumnBuilder 字符串列构建器
type StringColumnBuilder struct {
	name string
}

// From 指定列来源的子查询/派生表
func (b StringColumnBuilder) From(source interface{ TableName() string }) Pattern[string] {
	return NewPattern[string](source.TableName(), b.name)
}

// BoolColumn 创建布尔类型列引用
func BoolColumn(name string) BoolColumnBuilder {
	return BoolColumnBuilder{name: name}
}

// BoolColumnBuilder 布尔列构建器
type BoolColumnBuilder struct {
	name string
}

// From 指定列来源的子查询/派生表
func (b BoolColumnBuilder) From(source interface{ TableName() string }) Comparable[bool] {
	return NewComparable[bool](source.TableName(), b.name)
}

// TimeColumn 创建时间类型列引用 (用于 time.Time 类型)
func TimeColumn(name string) TimeColumnBuilder {
	return TimeColumnBuilder{name: name}
}

// TimeColumnBuilder 时间列构建器
type TimeColumnBuilder struct {
	name string
}

// From 指定列来源的子查询/派生表
// 返回 Comparable[any] 以支持各种时间类型
func (b TimeColumnBuilder) From(source interface{ TableName() string }) Comparable[any] {
	return NewComparable[any](source.TableName(), b.name)
}

// Column 创建泛型类型列引用
func Column[T any](name string) ColumnBuilder[T] {
	return ColumnBuilder[T]{name: name}
}

// ColumnBuilder 泛型列构建器
type ColumnBuilder[T any] struct {
	name string
}

// From 指定列来源的子查询/派生表
func (b ColumnBuilder[T]) From(source interface{ TableName() string }) Comparable[T] {
	return NewComparable[T](source.TableName(), b.name)
}
