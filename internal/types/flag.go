package types

// FieldFlag 字段标志位
type FieldFlag uint32

const (
	FlagNone          FieldFlag = 0
	FlagPrimaryKey    FieldFlag = 1 << 0 // 主键
	FlagUniqueIndex   FieldFlag = 1 << 1 // 唯一索引
	FlagIndex         FieldFlag = 1 << 2 // 普通索引
	FlagAutoIncrement FieldFlag = 1 << 3 // 自增
)
