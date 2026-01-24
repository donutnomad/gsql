package fieldi

import (
	"fmt"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/types"
	"github.com/donutnomad/gsql/internal/utils"
)

type Base struct {
	tableName  string
	columnName string
	alias      string // 别名
	sql        Expression
	flags      types.FieldFlag // 字段标志
}

// ColumnName 返回列名
func (f Base) ColumnName() string {
	return f.columnName
}

// TableName 返回表名
func (f Base) TableName() string {
	return f.tableName
}

// SetAlias 设置别名并返回新的 Base
func (f Base) SetAlias(alias string) Base {
	f.alias = alias
	return f
}

func NewBase(tableName, name string, flags ...types.FieldFlag) *Base {
	var flag = types.FlagNone
	if len(flags) > 0 {
		flag = flags[0]
	}
	return &Base{
		tableName:  tableName,
		columnName: name,
		flags:      flag,
	}
}

// Flags 返回字段标志
func (f Base) Flags() types.FieldFlag {
	return f.flags
}

// HasFlag 判断是否有某个标志
func (f Base) HasFlag(flag types.FieldFlag) bool {
	return f.flags&flag != 0
}

// IsPrimaryKey 是否为主键
func (f Base) IsPrimaryKey() bool {
	return f.HasFlag(types.FlagPrimaryKey)
}

// IsUniqueIndex 是否为唯一索引
func (f Base) IsUniqueIndex() bool {
	return f.HasFlag(types.FlagUniqueIndex)
}

func NewBaseFromSql(expr Expression, alias string) *Base {
	return &Base{
		sql:   expr,
		alias: alias,
	}
}

// IsExpr 是否是一个表达式字段
func (f Base) IsExpr() bool {
	return f.sql != nil
}

// ToColumn 转换为clause.Column对象，只有函数和窗口函数才可以导出
// 使用在=,!=,>,<,等场景
func (f Base) ToColumn() clause.Column {
	if f.sql != nil {
		var mb = utils.NewMemoryBuilder()
		f.sql.Build(mb)
		var sql = mb.SQL.String()
		if utils.IsLiteralFunctionName(sql) { // like: FROM_UNIXTIME OR JSON_EXTRACT
			if len(mb.Vars) > 0 {
				sql = utils.Dialector.Explain(sql, mb.Vars...)
			}
			return clause.Column{
				Table: "",
				Name:  sql,
				Alias: "",
				Raw:   true,
			}
		}
		panic("expr field cannot to column")
	}
	//		if utils.IsLiteralFunctionName(sql) || utils.IsWindowFunction(sql) {
	//			if len(mb.Vars) > 0 {
	//				sql = utils.Dialector.Explain(sql, mb.Vars...)
	//			}
	//		}
	return NewColumnClause(f).Column
}

// ToExpr 转换为表达式
func (f Base) ToExpr() Expression {
	return NewColumnClause(f)
}

// Name 返回字段名称
// 对于expr，返回别名
// 对于普通字段，有别名的返回别名，否则返回真实名字
func (f Base) Name() string {
	if f.sql != nil {
		return f.alias
	}
	if len(f.alias) > 0 {
		return f.alias
	}
	return f.columnName
}

func (f Base) FullName() string {
	if f.tableName == "" {
		return fmt.Sprintf("`%s`", f.columnName)
	}
	return fmt.Sprintf("`%s`.`%s`", f.tableName, f.columnName)
}

func (f Base) Alias() string {
	return f.alias
}

// As 创建一个别名字段
func (f Base) As(alias string) IField {
	if f.sql != nil {
		return NewBaseFromSql(f.sql, alias)
	}
	b := NewBase(f.tableName, f.columnName)
	b.alias = alias
	return b
}

func (f Base) AsPrefix(prefix string) IField {
	return f.As(prefix + f.columnName)
}

func (f Base) AsSuffix(suffix string) IField {
	return f.As(f.columnName + suffix)
}

// Build Implements clause.Expression
func (f Base) Build(builder clause.Builder) {
	f.ToExpr().Build(builder)
}
