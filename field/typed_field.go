package field

import (
	"github.com/donutnomad/gsql/clause"
)

// ==================== IntExprField 定义 ====================

// IntExprField 整数类型字段，同时实现 IField 接口和 IntExprT 的所有方法
// 使用场景：
//   - 替代 Comparable[int/int64/uint64] 等整数类型字段
//   - 支持字段操作（Order、GroupBy 等）的同时支持类型安全的数学运算
type IntExprField[T any] struct {
	Base
	IntExprT[T]
}

// NewIntExprField 创建一个新的 IntExprField 实例
func NewIntExprField[T any](tableName, name string, flags ...FieldFlag) IntExprField[T] {
	b := NewBase(tableName, name, flags...)
	return NewIntExprFieldFrom[T](b)
}

// NewIntExprFieldFrom 从 IField 创建 IntExprField
func NewIntExprFieldFrom[T any](field IField) IntExprField[T] {
	base := ifieldToBase(field)
	expr := base.ToExpr()
	return IntExprField[T]{
		Base:     base,
		IntExprT: NewIntExprT[T](expr),
	}
}

// Build 实现 clause.Expression 接口
func (f IntExprField[T]) Build(builder clause.Builder) {
	f.Base.ToExpr().Build(builder)
}

// ToExpr 转换为 Expression（覆盖 IntExprT 的方法，使用 Base 的实现）
func (f IntExprField[T]) ToExpr() Expression {
	return f.Base.ToExpr()
}

// As 创建一个别名字段（返回 IField）
func (f IntExprField[T]) As(alias string) IField {
	return f.Base.As(alias)
}

// WithTable 创建带新表名的字段
func (f IntExprField[T]) WithTable(tableName interface{ TableName() string }, fieldNames ...string) IntExprField[T] {
	name := f.Base.columnName
	if len(fieldNames) > 0 {
		name = fieldNames[0]
	}
	return NewIntExprField[T](tableName.TableName(), name)
}

// WithAlias 创建带别名的字段
func (f IntExprField[T]) WithAlias(alias string) IntExprField[T] {
	b := f.Base
	b.alias = alias
	return NewIntExprFieldFrom[T](b)
}

// FieldType 返回字段类型的零值
func (f IntExprField[T]) FieldType() T {
	var def T
	return def
}

// ==================== FloatExprField 定义 ====================

// FloatExprField 浮点类型字段，同时实现 IField 接口和 FloatExprT 的所有方法
// 使用场景：
//   - 替代 Comparable[float32/float64] 等浮点类型字段
//   - 支持字段操作的同时支持类型安全的数学运算和三角函数
type FloatExprField[T any] struct {
	Base
	FloatExprT[T]
}

// NewFloatExprField 创建一个新的 FloatExprField 实例
func NewFloatExprField[T any](tableName, name string, flags ...FieldFlag) FloatExprField[T] {
	b := NewBase(tableName, name, flags...)
	return NewFloatExprFieldFrom[T](b)
}

// NewFloatExprFieldFrom 从 IField 创建 FloatExprField
func NewFloatExprFieldFrom[T any](field IField) FloatExprField[T] {
	base := ifieldToBase(field)
	expr := base.ToExpr()
	return FloatExprField[T]{
		Base:       base,
		FloatExprT: NewFloatExprT[T](expr),
	}
}

// Build 实现 clause.Expression 接口
func (f FloatExprField[T]) Build(builder clause.Builder) {
	f.Base.ToExpr().Build(builder)
}

// ToExpr 转换为 Expression
func (f FloatExprField[T]) ToExpr() Expression {
	return f.Base.ToExpr()
}

// As 创建一个别名字段
func (f FloatExprField[T]) As(alias string) IField {
	return f.Base.As(alias)
}

// WithTable 创建带新表名的字段
func (f FloatExprField[T]) WithTable(tableName interface{ TableName() string }, fieldNames ...string) FloatExprField[T] {
	name := f.Base.columnName
	if len(fieldNames) > 0 {
		name = fieldNames[0]
	}
	return NewFloatExprField[T](tableName.TableName(), name)
}

// WithAlias 创建带别名的字段
func (f FloatExprField[T]) WithAlias(alias string) FloatExprField[T] {
	b := f.Base
	b.alias = alias
	return NewFloatExprFieldFrom[T](b)
}

// FieldType 返回字段类型的零值
func (f FloatExprField[T]) FieldType() T {
	var def T
	return def
}

// ==================== DecimalExprField 定义 ====================

// DecimalExprField 定点数类型字段，同时实现 IField 接口和 DecimalExprT 的所有方法
// 使用场景：
//   - 价格、金额等需要精确计算的字段
//   - 支持字段操作的同时支持类型安全的精确数学运算
type DecimalExprField[T any] struct {
	Base
	DecimalExprT[T]
}

// NewDecimalExprField 创建一个新的 DecimalExprField 实例
func NewDecimalExprField[T any](tableName, name string, flags ...FieldFlag) DecimalExprField[T] {
	b := NewBase(tableName, name, flags...)
	return NewDecimalExprFieldFrom[T](b)
}

// NewDecimalExprFieldFrom 从 IField 创建 DecimalExprField
func NewDecimalExprFieldFrom[T any](field IField) DecimalExprField[T] {
	base := ifieldToBase(field)
	expr := base.ToExpr()
	return DecimalExprField[T]{
		Base:         base,
		DecimalExprT: NewDecimalExprT[T](expr),
	}
}

// Build 实现 clause.Expression 接口
func (f DecimalExprField[T]) Build(builder clause.Builder) {
	f.Base.ToExpr().Build(builder)
}

// ToExpr 转换为 Expression
func (f DecimalExprField[T]) ToExpr() Expression {
	return f.Base.ToExpr()
}

// As 创建一个别名字段
func (f DecimalExprField[T]) As(alias string) IField {
	return f.Base.As(alias)
}

// WithTable 创建带新表名的字段
func (f DecimalExprField[T]) WithTable(tableName interface{ TableName() string }, fieldNames ...string) DecimalExprField[T] {
	name := f.Base.columnName
	if len(fieldNames) > 0 {
		name = fieldNames[0]
	}
	return NewDecimalExprField[T](tableName.TableName(), name)
}

// WithAlias 创建带别名的字段
func (f DecimalExprField[T]) WithAlias(alias string) DecimalExprField[T] {
	b := f.Base
	b.alias = alias
	return NewDecimalExprFieldFrom[T](b)
}

// FieldType 返回字段类型的零值
func (f DecimalExprField[T]) FieldType() T {
	var def T
	return def
}

// ==================== TextExprField 定义 ====================

// TextExprField 文本类型字段，同时实现 IField 接口和 TextExpr 的所有方法
// 使用场景：
//   - 替代 Pattern[string] 等字符串类型字段
//   - 支持字段操作的同时支持类型安全的字符串操作（Upper、Lower、Concat 等）
type TextExprField[T any] struct {
	Base
	TextExpr[T]
}

// NewTextExprField 创建一个新的 TextExprField 实例
func NewTextExprField[T any](tableName, name string, flags ...FieldFlag) TextExprField[T] {
	b := NewBase(tableName, name, flags...)
	return NewTextExprFieldFrom[T](b)
}

// NewTextExprFieldFrom 从 IField 创建 TextExprField
func NewTextExprFieldFrom[T any](field IField) TextExprField[T] {
	base := ifieldToBase(field)
	expr := base.ToExpr()
	return TextExprField[T]{
		Base:     base,
		TextExpr: NewTextExpr[T](expr),
	}
}

// Build 实现 clause.Expression 接口
func (f TextExprField[T]) Build(builder clause.Builder) {
	f.Base.ToExpr().Build(builder)
}

// ToExpr 转换为 Expression
func (f TextExprField[T]) ToExpr() Expression {
	return f.Base.ToExpr()
}

// As 创建一个别名字段
func (f TextExprField[T]) As(alias string) IField {
	return f.Base.As(alias)
}

// WithTable 创建带新表名的字段
func (f TextExprField[T]) WithTable(tableName interface{ TableName() string }, fieldNames ...string) TextExprField[T] {
	name := f.Base.columnName
	if len(fieldNames) > 0 {
		name = fieldNames[0]
	}
	return NewTextExprField[T](tableName.TableName(), name)
}

// WithAlias 创建带别名的字段
func (f TextExprField[T]) WithAlias(alias string) TextExprField[T] {
	b := f.Base
	b.alias = alias
	return NewTextExprFieldFrom[T](b)
}

// FieldType 返回字段类型的零值
func (f TextExprField[T]) FieldType() T {
	var def T
	return def
}
