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

var Star field.IField = fields.ScalarFieldOf[any]("", "*")

func Lit[T primitive](value T) *fields.LitExpr {
	return fields.NewLitExpr(value)
}

func Val[T any](value T) field.ExpressionTo {
	return ExprTo{Expression: Expr("?", value)}
}

func Slice[T any](value ...T) field.ExpressionTo {
	return Val(value)
}

func Eq(val1 field.ExpressionTo, val2 field.ExpressionTo) Expression {
	return Expr("? = ?", val1, val2)
}

func Not(val1 field.ExpressionTo, val2 field.ExpressionTo) Expression {
	return Expr("? != ?", val1, val2)
}

func StarWith(tableName string) field.IField {
	return fields.ScalarFieldOf[any](tableName, "*")
}

// False 返回布尔值假
// SELECT FALSE;
// SELECT FALSE = 0;
// SELECT users.* FROM users WHERE users.is_active = FALSE;
var False = fields.IntOf[bool](clause.Expr{
	SQL: "FALSE",
})

// Null 返回空值
// SELECT NULL;
// SELECT IFNULL(users.nickname, NULL) FROM users;
// UPDATE users SET deleted_at = NULL WHERE id = 1;
var Null = fields.ScalarOf[any](clause.Expr{
	SQL: "NULL",
})

// RAND 返回0到1之间的随机浮点数，可选种子参数
// 数据库支持: MySQL
// SELECT RAND();
// SELECT RAND() * 100;
// SELECT RAND(123);
// SELECT * FROM users ORDER BY RAND() LIMIT 10;
func RAND() fields.FloatExpr[float64] {
	return fields.FloatOf[float64](clause.Expr{SQL: "RAND()"})
}

// ==================== 聚合函数 ====================

// COUNT 计算行数或非NULL值的数量，不提供参数时统计所有行（包括NULL）
// 数据库支持: MySQL, PostgreSQL, SQLite
// 返回 IntExpr，支持 .Gt(), .Lt(), .Eq() 等比较操作
// 示例:
//
//	COUNT()           // COUNT(*)
//	COUNT(id)         // COUNT(id)
//	COUNT().Gt(5)     // COUNT(*) > 5
func COUNT(expr ...clause.Expression) fields.IntExpr[int64] {
	if len(expr) == 0 {
		return fields.IntOf[int64](clause.Expr{SQL: "COUNT(*)"})
	}
	return fields.IntOf[int64](clause.Expr{
		SQL:  "COUNT(?)",
		Vars: []any{expr[0]},
	})
}

// COUNT_DISTINCT 计算不重复的非NULL值的数量
// 数据库支持: MySQL, PostgreSQL, SQLite
// 返回 IntExpr，支持比较操作
// 示例:
//
//	COUNT_DISTINCT(city)       // COUNT(DISTINCT city)
//	COUNT_DISTINCT(id).Gt(10)  // COUNT(DISTINCT id) > 10
func COUNT_DISTINCT(expr field.IField) fields.IntExpr[int64] {
	return fields.IntOf[int64](clause.Expr{
		SQL:  "COUNT(DISTINCT ?)",
		Vars: []any{expr.ToExpr()},
	})
}

// GROUP_CONCAT 将分组内的字符串连接起来，默认用逗号分隔，可指定分隔符
// 数据库支持: MySQL (PostgreSQL 使用 STRING_AGG, SQLite 支持 GROUP_CONCAT 但语法略有不同)
// SELECT GROUP_CONCAT(name) FROM users;
// SELECT GROUP_CONCAT(name SEPARATOR ';') FROM users;
// SELECT user_id, GROUP_CONCAT(product_name) FROM orders GROUP BY user_id;
// SELECT category, GROUP_CONCAT(DISTINCT tag ORDER BY tag) FROM products GROUP BY category;
func GROUP_CONCAT(expr Expression, separator ...string) fields.StringExpr[string] {
	if len(separator) > 0 {
		// SEPARATOR 后面必须是字面值，不能使用参数占位符
		// 使用 strings.ReplaceAll 转义单引号防止 SQL 注入
		escaped := strings.ReplaceAll(separator[0], "'", "''")
		return fields.StringOf[string](clause.Expr{
			SQL:  "GROUP_CONCAT(? SEPARATOR '" + escaped + "')",
			Vars: []any{expr},
		})
	}
	return fields.StringOf[string](clause.Expr{
		SQL:  "GROUP_CONCAT(?)",
		Vars: []any{expr},
	})
}

// ==================== 流程控制函数 ====================

// IF 条件判断函数，如果条件为真返回第一个值，否则返回第二个值
// 数据库支持: MySQL (PostgreSQL/SQLite 使用 CASE WHEN 替代)
// SELECT IF(score >= 60, '及格', '不及格') FROM students;
// SELECT IF(stock > 0, 'In Stock', 'Out of Stock') FROM products;
// SELECT name, IF(age >= 18, '成年', '未成年') FROM users;
// SELECT SUM(IF(status = 'completed', amount, 0)) FROM orders;
func IF[Result interface{ ExprType() R }, R any](condition Condition, valueIfTrue, valueIfFalse Result) Result {
	return fields.CastExpr[Result](clause.Expr{
		SQL:  "IF(?, ?, ?)",
		Vars: []any{condition, valueIfTrue, valueIfFalse},
	})
}

func IFF[Result interface{ Expr() ResultExpr }, ResultExpr interface{ ExprType() R }, R any](condition Condition, valueIfTrue, valueIfFalse Result) ResultExpr {
	return fields.CastExpr[ResultExpr](clause.Expr{
		SQL:  "IF(?, ?, ?)",
		Vars: []any{condition, valueIfTrue, valueIfFalse},
	})
}

func COUNT_IF(condition Condition) IntExpr[int64] {
	return COUNT(
		IF(condition, IntVal(1), IntOf[int](nil)),
	)
}

// ==================== 类型转换函数 ====================

// CONVERT 将表达式转换为指定的数据类型或字符集
// 数据库支持: MySQL (PostgreSQL 使用 CAST 或 :: 操作符)
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
func CONVERT(expr Expression, dataType string) field.ExpressionTo {
	return ExprTo{Expression: clause.Expr{
		SQL:  fmt.Sprintf("CONVERT(?, %s)", dataType),
		Vars: []any{expr},
	}}
}

// CONVERT_CHARSET 将表达式转换为指定的字符集
// 数据库支持: MySQL
// SELECT CONVERT(name USING utf8mb4) FROM users;
// SELECT CONVERT(content USING latin1) FROM articles;
// SELECT CONVERT(description USING gbk) FROM products;
// SELECT CONVERT(text USING utf8) FROM messages;
// 常用字符集: utf8, utf8mb4, latin1, gbk, ascii, binary
func CONVERT_CHARSET(expr Expression, charset string) field.ExpressionTo {
	// 验证字符集参数
	charset = strings.ToLower(strings.TrimSpace(charset))
	if !allowedCharsets[charset] {
		panic(fmt.Sprintf("CONVERT_CHARSET: invalid or unsupported charset: %s", charset))
	}

	return ExprTo{Expression: clause.Expr{
		SQL:  fmt.Sprintf("CONVERT(? USING %s)", charset),
		Vars: []any{expr},
	}}
}

// JsonOf
//
//	gsql.JsonOf(u.Profile).Extract("$.name")
//	gsql.JsonOf(u.Profile).Length("$.skills")
func JsonOf(expr clause.Expression) JsonExpr {
	return fields.JsonOf(expr)
}
