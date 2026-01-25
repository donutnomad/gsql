package fieldi

import (
	"github.com/donutnomad/gsql/clause"
)

type ExpressionTo interface {
	clause.Expression
	// AsF as field
	AsF(name ...string) IField
	//ToField(name string) IField
}

type IToExpr interface {
	ToExpr() clause.Expression
}

type IField interface {
	clause.Expression
	// ToExpr 转换为表达式
	ToExpr() clause.Expression

	// FullName 返回table.name，如果没有table，和Name()值相同
	FullName() string
	// Name 返回字段名称
	// 对于expr，返回别名
	// 对于普通字段，有别名的返回别名，否则返回真实名字
	Name() string
	// As 创建一个别名字段
	As(alias string) IField
	Alias() string
}
