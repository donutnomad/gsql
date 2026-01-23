package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
)

// RowValue 返回 INSERT 语句中要插入的行值，用于 ON DUPLICATE KEY UPDATE 子句
// 根据 SetMySQLVersion 设置的全局配置，生成不同的 SQL 语法：
//   - MySQLVersionDefault: 生成 VALUES(column) 语法
//   - MySQLVersion8020: 生成 _new.column 语法
//
// 示例:
//
//	// 简单更新：使用插入行的值更新
//	gsql.Set(t.Count, gsql.RowValue(t.Count))
//
//	// 条件更新：只有当新版本号更大时才更新
//	gsql.Set(t.Value, gsql.IF(
//	    gsql.Expr("? >= ?", gsql.RowValue(t.Version), t.Version),
//	    gsql.RowValue(t.Value),
//	    t.Value,
//	))
func RowValue(f field.IField) field.ExpressionTo {
	if GetMySQLVersion() >= MySQLVersion8020 {
		// MySQL 8.0.20+ 新语法: _new.column
		return ExprTo{clause.Expr{
			SQL:  "?",
			Vars: []any{rowValueExpr{field: f}},
		}}
	}
	// 旧语法: VALUES(column)
	return ExprTo{clause.Expr{
		SQL:  "VALUES(?)",
		Vars: []any{f.ToExpr()},
	}}
}

// rowValueExpr 用于生成 MySQL 8.0.20+ 的行别名引用语法
type rowValueExpr struct {
	field field.IField
}

func (e rowValueExpr) Build(builder clause.Builder) {
	builder.WriteQuoted(insertRowAlias)
	builder.WriteByte('.')
	builder.WriteQuoted(e.field.Name())
}

// VALUES 返回 INSERT 语句中指定列的值，用于 ON DUPLICATE KEY UPDATE 子句
// Deprecated: 请使用 RowValue 替代，它会根据 MySQL 版本自动选择正确的语法
func VALUES(f field.IField) field.ExpressionTo {
	return RowValue(f)
}

// InsertValue 返回 INSERT 语句中要插入的行值
// Deprecated: 请使用 RowValue 替代，命名更清晰
func InsertValue(f field.IField) field.ExpressionTo {
	return RowValue(f)
}
