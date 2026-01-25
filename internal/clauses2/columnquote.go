package clauses2

import (
	"fmt"

	"github.com/samber/lo"
	"gorm.io/gorm/clause"
)

type INoAS interface {
	NoAS()
}

type ColumnQuote struct {
	TableName  string // 表名
	ColumnName string // 字段名
	Alias      string // As Name
}

// NoAS remove AS `xxx`
func (c *ColumnQuote) NoAS() {
	c.Alias = ""
}

func (c *ColumnQuote) FullName() string {
	if c.TableName == "" {
		return fmt.Sprintf("`%s`", c.ColumnName)
	}
	return fmt.Sprintf("`%s`. `%s`", c.TableName, c.ColumnName)
}

// Name 返回字段名称
// 对于expr，返回别名
// 对于普通字段，有别名的返回别名，否则返回真实名字
func (c *ColumnQuote) Name() string {
	return lo.Ternary(len(c.Alias) > 0, c.Alias, c.ColumnName)
}

func (c *ColumnQuote) Build(builder clause.Builder) {
	if len(c.TableName) > 0 {
		builder.WriteQuoted(c.TableName)
		builder.WriteString(".")
	}
	if len(c.ColumnName) > 0 {
		if c.ColumnName == "*" {
			builder.WriteString(c.ColumnName)
		} else {
			builder.WriteQuoted(c.ColumnName)
		}
	} else {
		panic("[ColumnQuote] required column name")
	}
	if len(c.Alias) > 0 {
		builder.WriteString(" AS ")
		builder.WriteQuoted(c.Alias)
	}
}
