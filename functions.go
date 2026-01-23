package gsql

import (
	"fmt"
	"strings"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
)

// 白名单：允许的字符集
var allowedCharsets = map[string]bool{
	"utf8": true, "utf8mb4": true, "latin1": true,
	"gbk": true, "ascii": true, "binary": true,
	"ucs2": true, "utf16": true, "utf32": true,
}

var Star field.IField = field.NewBase("", "*")

// Deprecated: 使用 Lit 替代。
// Primitive 已被弃用，请使用更简洁的 Lit 函数。
func Primitive[T primitive](value T) field.ExpressionTo {
	return Lit(value)
}

func Lit[T primitive](value T) field.ExpressionTo {
	return Val(value)
}

func Val[T any](value T) field.ExpressionTo {
	return ExprTo{Expr("?", value)}
}

func Slice[T any](value ...T) field.ExpressionTo {
	return Val(value)
}

func Eq(val1 field.ExpressionTo, val2 field.ExpressionTo) field.Expression {
	return Expr("? = ?", val1, val2)
}

func Not(val1 field.ExpressionTo, val2 field.ExpressionTo) field.Expression {
	return Expr("? != ?", val1, val2)
}

func StarWith(tableName string) field.IField {
	return field.NewBaseFromSql(Expr("?.*", quoteClause{
		name: tableName,
	}), "")
}

type quoteClause struct {
	name string
}

func (q quoteClause) Build(builder clause.Builder) {
	builder.WriteQuoted(q.name)
}

// False 返回布尔值假
// SELECT FALSE;
// SELECT FALSE = 0;
// SELECT users.* FROM users WHERE users.is_active = FALSE;
var False field.ExpressionTo = ExprTo{clause.Expr{
	SQL: "FALSE",
}}

// Null 返回空值
// SELECT NULL;
// SELECT IFNULL(users.nickname, NULL) FROM users;
// UPDATE users SET deleted_at = NULL WHERE id = 1;
var Null field.ExpressionTo = ExprTo{clause.Expr{
	SQL: "NULL",
}}

// RAND 返回0到1之间的随机浮点数，可选种子参数
// SELECT RAND();
// SELECT RAND() * 100;
// SELECT RAND(123);
// SELECT * FROM users ORDER BY RAND() LIMIT 10;
func RAND() fields.FloatExpr[float64] {
	return fields.NewFloatExpr[float64](clause.Expr{SQL: "RAND()"})
}

// ==================== 聚合函数 ====================

// COUNT 计算行数或非NULL值的数量，不提供参数时统计所有行（包括NULL）
// 返回 IntExpr，支持 .Gt(), .Lt(), .Eq() 等比较操作
// 示例:
//
//	COUNT()           // COUNT(*)
//	COUNT(id)         // COUNT(id)
//	COUNT().Gt(5)     // COUNT(*) > 5
func COUNT(expr ...field.IField) fields.IntExpr[int64] {
	if len(expr) == 0 {
		return fields.NewIntExpr[int64](clause.Expr{SQL: "COUNT(*)"})
	}
	return fields.NewIntExpr[int64](clause.Expr{
		SQL:  "COUNT(?)",
		Vars: []any{expr[0].ToExpr()},
	})
}

// COUNT_DISTINCT 计算不重复的非NULL值的数量
// 返回 IntExpr，支持比较操作
// 示例:
//
//	COUNT_DISTINCT(city)       // COUNT(DISTINCT city)
//	COUNT_DISTINCT(id).Gt(10)  // COUNT(DISTINCT id) > 10
func COUNT_DISTINCT(expr field.IField) fields.IntExpr[int64] {
	return fields.NewIntExpr[int64](clause.Expr{
		SQL:  "COUNT(DISTINCT ?)",
		Vars: []any{expr.ToExpr()},
	})
}

// GROUP_CONCAT 将分组内的字符串连接起来，默认用逗号分隔，可指定分隔符
// SELECT GROUP_CONCAT(name) FROM users;
// SELECT GROUP_CONCAT(name SEPARATOR ';') FROM users;
// SELECT user_id, GROUP_CONCAT(product_name) FROM orders GROUP BY user_id;
// SELECT category, GROUP_CONCAT(DISTINCT tag ORDER BY tag) FROM products GROUP BY category;
func GROUP_CONCAT(expr field.Expression, separator ...string) fields.TextExpr[string] {
	if len(separator) > 0 {
		// 使用参数化查询代替字符串拼接
		return fields.NewTextExpr[string](clause.Expr{
			SQL:  "GROUP_CONCAT(? SEPARATOR ?)",
			Vars: []any{expr, separator[0]},
		})
	}
	return fields.NewTextExpr[string](clause.Expr{
		SQL:  "GROUP_CONCAT(?)",
		Vars: []any{expr},
	})
}

// ==================== 流程控制函数 ====================

// IF 条件判断函数，如果条件为真返回第一个值，否则返回第二个值
// SELECT IF(score >= 60, '及格', '不及格') FROM students;
// SELECT IF(stock > 0, 'In Stock', 'Out of Stock') FROM products;
// SELECT name, IF(age >= 18, '成年', '未成年') FROM users;
// SELECT SUM(IF(status = 'completed', amount, 0)) FROM orders;
func IF(condition, valueIfTrue, valueIfFalse field.Expression) field.ExpressionTo {
	return ExprTo{clause.Expr{
		SQL:  "IF(?, ?, ?)",
		Vars: []any{condition, valueIfTrue, valueIfFalse},
	}}
}

// IFNULL 如果第一个表达式不为NULL则返回它，否则返回第二个表达式
// SELECT IFNULL(nickname, username) FROM users;
// SELECT IFNULL(discount, 0) FROM products;
// SELECT IFNULL(email, 'no-email') FROM contacts;
// SELECT name, IFNULL(phone, 'N/A') FROM users;
func IFNULL(expr1, expr2 any) field.ExpressionTo {
	return ExprTo{clause.Expr{
		SQL:  "IFNULL(?, ?)",
		Vars: []any{expr1, expr2},
	}}
}

// NULLIF 如果两个表达式相等则返回NULL，否则返回第一个表达式
// SELECT NULLIF(10, 10);
// SELECT NULLIF(10, 5);
// SELECT NULLIF(username, ”) FROM users;
// SELECT 100 / NULLIF(quantity, 0) FROM inventory;
func NULLIF(expr1, expr2 any) field.ExpressionTo {
	return ExprTo{clause.Expr{
		SQL:  "NULLIF(?, ?)",
		Vars: []any{expr1, expr2},
	}}
}

// ==================== 类型转换函数 ====================

// CAST 将表达式转换为指定的数据类型
// SELECT CAST('123' AS UNSIGNED);
// SELECT CAST('2023-10-26' AS DATE);
// SELECT CAST(price AS CHAR) FROM products;
// SELECT CAST(amount AS DECIMAL(10,2)) FROM orders;
// 使用示例：
//
//	CAST(field, CastTypeSigned)
//	CAST(field, CastTypeDate)
//	CAST(field, "DECIMAL(10,2)") // 对于需要指定精度的类型，可以直接传字符串
func CAST(expr field.Expression, dataType string) field.ExpressionTo {
	return ExprTo{clause.Expr{
		SQL:  fmt.Sprintf("CAST(? AS %s)", dataType),
		Vars: []any{expr},
	}}
}

// CONVERT 将表达式转换为指定的数据类型或字符集
// 语法1: CONVERT(expr, type) - 类型转换，与CAST类似
// SELECT CONVERT('123', UNSIGNED);
// SELECT CONVERT('2023-10-26', DATE);
// SELECT CONVERT(price, CHAR) FROM products;
// SELECT CONVERT(amount, DECIMAL(10,2)) FROM orders;
// 语法2: CONVERT(expr USING charset) - 字符集转换，使用CONVERT_CHARSET函数
// 使用示例：
//
//	CONVERT(field, CastTypeSigned)
//	CONVERT(field, CastTypeDate)
//	CONVERT(field, "DECIMAL(10,2)") // 对于需要指定精度的类型，可以直接传字符串
func CONVERT(expr field.Expression, dataType string) field.ExpressionTo {
	return ExprTo{clause.Expr{
		SQL:  fmt.Sprintf("CONVERT(?, %s)", dataType),
		Vars: []any{expr},
	}}
}

// CONVERT_CHARSET 将表达式转换为指定的字符集
// SELECT CONVERT(name USING utf8mb4) FROM users;
// SELECT CONVERT(content USING latin1) FROM articles;
// SELECT CONVERT(description USING gbk) FROM products;
// SELECT CONVERT(text USING utf8) FROM messages;
// 常用字符集: utf8, utf8mb4, latin1, gbk, ascii, binary
func CONVERT_CHARSET(expr field.Expression, charset string) field.ExpressionTo {
	// 验证字符集参数
	charset = strings.ToLower(strings.TrimSpace(charset))
	if !allowedCharsets[charset] {
		panic(fmt.Sprintf("CONVERT_CHARSET: invalid or unsupported charset: %s", charset))
	}

	return ExprTo{clause.Expr{
		SQL:  fmt.Sprintf("CONVERT(? USING %s)", charset),
		Vars: []any{expr},
	}}
}

// ==================== 其它常用函数 ====================

// DATABASE 返回当前使用的数据库名，如果未选择数据库则返回NULL
// SELECT DATABASE();
// INSERT INTO logs (db_name) VALUES (DATABASE());
// SELECT DATABASE() as current_db;
// SELECT * FROM information_schema.tables WHERE table_schema = DATABASE();
func DATABASE() fields.TextExpr[string] {
	return fields.NewTextExpr[string](clause.Expr{SQL: "DATABASE()"})
}

// USER 返回当前MySQL用户名和主机名，格式为 'user@host'
// SELECT USER();
// INSERT INTO audit_logs (user) VALUES (USER());
// SELECT USER() as current_user;
// SELECT * FROM connections WHERE user = USER();
func USER() fields.TextExpr[string] {
	return fields.NewTextExpr[string](clause.Expr{SQL: "USER()"})
}

// CURRENT_USER 返回当前MySQL用户名和主机名，与USER()相同
// SELECT CURRENT_USER();
// SELECT CURRENT_USER;
// INSERT INTO access_logs (accessed_by) VALUES (CURRENT_USER());
// SELECT CURRENT_USER() as authenticated_user;
func CURRENT_USER() fields.TextExpr[string] {
	return fields.NewTextExpr[string](clause.Expr{SQL: "CURRENT_USER()"})
}

// VERSION 返回MySQL服务器的版本号
// SELECT VERSION();
// SELECT VERSION() as mysql_version;
// INSERT INTO system_info (version) VALUES (VERSION());
// SELECT IF(VERSION() LIKE '8.%', 'MySQL 8', 'Older') as version_check;
func VERSION() fields.TextExpr[string] {
	return fields.NewTextExpr[string](clause.Expr{SQL: "VERSION()"})
}

// UUID 生成一个符合RFC 4122标准的通用唯一标识符（36字符的字符串）
// SELECT UUID();
// INSERT INTO records (id) VALUES (UUID());
// SELECT UUID() as unique_id;
// UPDATE sessions SET session_id = UUID() WHERE session_id IS NULL;
func UUID() fields.TextExpr[string] {
	return fields.NewTextExpr[string](clause.Expr{SQL: "UUID()"})
}
