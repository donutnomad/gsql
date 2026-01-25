package fields

import (
	"fmt"

	"github.com/donutnomad/gsql/clause"
)

var _ clause.Expression = (*StringExpr[string])(nil)

type stringExpr[T any] = StringExpr[T]

// ==================== StringExpr 定义 ====================

// StringExpr 文本类型表达式，用于 VARCHAR 和 TEXT 类型字段
// @gentype default=[string]
// 支持比较操作和模式匹配操作
// 使用场景：
//   - CONCAT, SUBSTRING 等字符串函数的返回值
//   - UPPER, LOWER 等字符串转换函数的返回值
//   - 派生表中的文本列
type StringExpr[T any] struct {
	baseComparableImpl[T] // 只包含 Eq/Not/In/NotIn，字符串没有大小比较
	patternExprImpl[T]
	pointerExprImpl
	nullCondFuncSql
	baseExprSql
}

// String creates a StringExpr[string] from a clause expression.
func String(expr clause.Expression) StringExpr[string] {
	return StringOf[string](expr)
}

// StringE creates a StringExpr[string] from raw SQL with optional variables.
func StringE(sql string, vars ...any) StringExpr[string] {
	return String(clause.Expr{SQL: sql, Vars: vars})
}

// StringVal creates a StringExpr from a string literal value.
func StringVal[T ~string | any](val T) StringExpr[T] {
	return StringOf[T](anyToExpr(val))
}

func StringFrom[T any](field interface{ FieldType() T }) StringExpr[T] {
	return StringOf[T](anyToExpr(field))
}

// StringOf creates a generic StringExpr[T] from a clause expression.
func StringOf[T any](expr clause.Expression) StringExpr[T] {
	return StringExpr[T]{
		baseComparableImpl: baseComparableImpl[T]{Expression: expr},
		patternExprImpl:    patternExprImpl[T]{Expression: expr},
		pointerExprImpl:    pointerExprImpl{Expression: expr},
		nullCondFuncSql:    nullCondFuncSql{Expression: expr},
		baseExprSql:        baseExprSql{Expr: expr},
	}
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
// SELECT CAST(price AS SIGNED) FROM products;
// SELECT CAST(amount AS DECIMAL(10,2)) FROM orders;
// targetType: SIGNED, UNSIGNED, DECIMAL(m,n), CHAR, DATE, DATETIME, TIME, BINARY 等
func (e StringExpr[T]) Cast(targetType string) clause.Expression {
	return clause.Expr{
		SQL:  "CAST(? AS " + targetType + ")",
		Vars: []any{e.Unwrap()},
	}
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e StringExpr[T]) CastSigned() IntExpr[int64] {
	return IntOf[int64](clause.Expr{
		SQL:  "CAST(? AS SIGNED)",
		Vars: []any{e.Unwrap()},
	})
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e StringExpr[T]) CastUnsigned() IntExpr[uint64] {
	return IntOf[uint64](clause.Expr{
		SQL:  "CAST(? AS UNSIGNED)",
		Vars: []any{e.Unwrap()},
	})
}

// CastJson 转换为JSON (CAST AS JSON)
func (e StringExpr[T]) CastJson() JsonExpr {
	return JsonOf(clause.Expr{
		SQL:  "CAST(? AS JSON)",
		Vars: []any{e.Unwrap()},
	})
}

// CastDecimal 转换为小数 (CAST AS DECIMAL)
// precision: 总位数, scale: 小数位数
func (e StringExpr[T]) CastDecimal(precision, scale int) DecimalExpr[float64] {
	return DecimalOf[float64](clause.Expr{
		SQL:  fmt.Sprintf("CAST(? AS DECIMAL(%d, %d))", precision, scale),
		Vars: []any{e.Unwrap()},
	})
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e StringExpr[T]) CastChar(length ...int) StringExpr[string] {
	if len(length) > 0 {
		return StringOf[string](clause.Expr{
			SQL:  fmt.Sprintf("CAST(? AS CHAR(%d))", length[0]),
			Vars: []any{e.Unwrap()},
		})
	}
	return StringOf[string](clause.Expr{
		SQL:  "CAST(? AS CHAR)",
		Vars: []any{e.Unwrap()},
	})
}

// ==================== 字符串函数 ====================

// Upper 将字符串转换为大写 (UPPER)，只对英文字母有效
// SELECT UPPER('hello world');
// SELECT UPPER(users.username) FROM users;
// SELECT * FROM products WHERE UPPER(product_code) = 'ABC123';
// UPDATE users SET username = UPPER(username) WHERE id = 1;
func (e StringExpr[T]) Upper() StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "UPPER(?)",
		Vars: []any{e.Unwrap()},
	})
}

func init() {
	var t StringExpr[string]
	t.Lower().Like("123")
}

// Lower 将字符串转换为小写 (LOWER)，只对英文字母有效
// SELECT LOWER('HELLO WORLD');
// SELECT LOWER(users.email) FROM users;
// SELECT * FROM users WHERE LOWER(username) = 'admin';
// UPDATE users SET email = LOWER(email);
func (e StringExpr[T]) Lower() StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "LOWER(?)",
		Vars: []any{e.Unwrap()},
	})
}

// Trim 去除字符串两端的空格 (TRIM)
// SELECT TRIM('  Hello World  ');
// SELECT TRIM(users.username) FROM users;
// UPDATE users SET email = TRIM(email);
func (e StringExpr[T]) Trim() StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "TRIM(?)",
		Vars: []any{e.Unwrap()},
	})
}

// LTrim 去除字符串左侧的空格 (LTRIM)
// SELECT LTRIM('  Hello World  ');
// SELECT LTRIM(users.name) FROM users;
// SELECT * FROM products WHERE LTRIM(code) != code;
// UPDATE users SET username = LTRIM(username);
func (e StringExpr[T]) LTrim() StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "LTRIM(?)",
		Vars: []any{e.Unwrap()},
	})
}

// RTrim 去除字符串右侧的空格 (RTRIM)
// SELECT RTRIM('  Hello World  ');
// SELECT RTRIM(description) FROM products;
// SELECT * FROM users WHERE RTRIM(email) != email;
// UPDATE articles SET title = RTRIM(title);
func (e StringExpr[T]) RTrim() StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "RTRIM(?)",
		Vars: []any{e.Unwrap()},
	})
}

// Substring 从字符串中提取子字符串 (SUBSTRING)，位置从1开始
// SELECT SUBSTRING('Hello World', 1, 5);
// SELECT SUBSTRING(users.email, 1, LOCATE('@', users.email) - 1) FROM users;
// SELECT SUBSTRING(product_code, 4, 3) FROM products;
// pos: 起始位置（从1开始）, length: 长度
func (e StringExpr[T]) Substring(pos, length int) StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "SUBSTRING(?, ?, ?)",
		Vars: []any{e.Unwrap(), pos, length},
	})
}

// Left 从字符串左侧提取指定长度的子字符串 (LEFT)
// SELECT LEFT('Hello World', 5);
// SELECT LEFT(users.name, 1) as initial FROM users;
// SELECT * FROM products WHERE LEFT(product_code, 2) = 'AB';
// SELECT LEFT(email, LOCATE('@', email) - 1) FROM users;
func (e StringExpr[T]) Left(length int) StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "LEFT(?, ?)",
		Vars: []any{e.Unwrap(), length},
	})
}

// Right 从字符串右侧提取指定长度的子字符串 (RIGHT)
// SELECT RIGHT('Hello World', 5);
// SELECT RIGHT(phone, 4) as last_four FROM users;
// SELECT * FROM files WHERE RIGHT(filename, 4) = '.pdf';
// SELECT RIGHT(product_code, 3) FROM products;
func (e StringExpr[T]) Right(length int) StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "RIGHT(?, ?)",
		Vars: []any{e.Unwrap(), length},
	})
}

// Length 返回字符串的字节长度 (LENGTH)
// SELECT LENGTH('Hello');
// SELECT LENGTH('你好');
// SELECT users.name, LENGTH(users.name) FROM users;
// SELECT * FROM products WHERE LENGTH(product_code) = 8;
// 注意: UTF-8编码中一个中文字符通常占3个字节
func (e StringExpr[T]) Length() IntExpr[int64] {
	return IntOf[int64](clause.Expr{
		SQL:  "LENGTH(?)",
		Vars: []any{e.Unwrap()},
	})
}

// CharLength 返回字符串的字符长度 (CHAR_LENGTH)，多字节字符按一个字符计算
// SELECT CHAR_LENGTH('Hello');
// SELECT CHAR_LENGTH('你好');
// SELECT users.name, CHAR_LENGTH(users.name) FROM users;
// SELECT * FROM articles WHERE CHAR_LENGTH(content) > 1000;
func (e StringExpr[T]) CharLength() IntExpr[int64] {
	return IntOf[int64](clause.Expr{
		SQL:  "CHAR_LENGTH(?)",
		Vars: []any{e.Unwrap()},
	})
}

// Concat 拼接多个字符串 (CONCAT)，任意参数为NULL则返回NULL
// SELECT CONCAT('Hello', ' ', 'World');
// SELECT CONCAT(users.first_name, ' ', users.last_name) as full_name FROM users;
// SELECT CONCAT('User:', users.id) FROM users;
// SELECT CONCAT(YEAR(NOW()), '-', MONTH(NOW()));
func (e StringExpr[T]) Concat(args ...clause.Expression) StringExpr[T] {
	placeholders := "?"
	allArgs := []any{e.Unwrap()}
	for _, arg := range args {
		placeholders += ", ?"
		allArgs = append(allArgs, arg)
	}
	return StringOf[T](clause.Expr{
		SQL:  "CONCAT(" + placeholders + ")",
		Vars: allArgs,
	})
}

// ConcatWS 用指定分隔符拼接多个字符串 (CONCAT_WS)，自动跳过NULL值
// SELECT CONCAT_WS(',', 'A', 'B', 'C'); -- 结果为 'A,B,C'
// SELECT CONCAT_WS('-', users.last_name, users.first_name) FROM users;
// SELECT CONCAT_WS('/', YEAR(date), MONTH(date), DAY(date)) FROM logs;
// SELECT CONCAT_WS(', ', city, state, country) FROM addresses;
// 注意：分隔符为NULL则返回NULL，但参数中的NULL会被跳过
func (e StringExpr[T]) ConcatWS(separator string, args ...clause.Expression) StringExpr[T] {
	placeholders := "?, ?"
	allArgs := []any{separator, e.Unwrap()}
	for _, arg := range args {
		placeholders += ", ?"
		allArgs = append(allArgs, arg)
	}
	return StringOf[T](clause.Expr{
		SQL:  "CONCAT_WS(" + placeholders + ")",
		Vars: allArgs,
	})
}

// Replace 替换字符串中所有出现的子字符串 (REPLACE)
// SELECT REPLACE('Hello World', 'World', 'MySQL');
// SELECT REPLACE('www.example.com', 'www', 'mail');
// SELECT REPLACE(phone, '-', ”) FROM users;
// UPDATE products SET description = REPLACE(description, 'old', 'new');
func (e StringExpr[T]) Replace(from, to string) StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "REPLACE(?, ?, ?)",
		Vars: []any{e.Unwrap(), from, to},
	})
}

// Locate 查找子字符串位置 (LOCATE)
// SELECT LOCATE('@', email) FROM users;
// 返回子字符串第一次出现的位置（从1开始），未找到返回0
func (e StringExpr[T]) Locate(substr string) IntExpr[int64] {
	return IntOf[int64](clause.Expr{
		SQL:  "LOCATE(?, ?)",
		Vars: []any{substr, e.Unwrap()},
	})
}

// Reverse 反转字符串 (REVERSE)
// SELECT REVERSE(name) FROM users;
func (e StringExpr[T]) Reverse() StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "REVERSE(?)",
		Vars: []any{e.Unwrap()},
	})
}

// Repeat 重复字符串 (REPEAT)
// SELECT REPEAT('*', 10);
func (e StringExpr[T]) Repeat(count int) StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "REPEAT(?, ?)",
		Vars: []any{e.Unwrap(), count},
	})
}

// LPad 左侧填充 (LPAD)
// SELECT LPAD(id, 5, '0') FROM users; -- 结果如 "00001"
func (e StringExpr[T]) LPad(length int, padStr string) StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "LPAD(?, ?, ?)",
		Vars: []any{e.Unwrap(), length, padStr},
	})
}

// RPad 右侧填充 (RPAD)
// SELECT RPAD(name, 20, ' ') FROM users;
func (e StringExpr[T]) RPad(length int, padStr string) StringExpr[T] {
	return StringOf[T](clause.Expr{
		SQL:  "RPAD(?, ?, ?)",
		Vars: []any{e.Unwrap(), length, padStr},
	})
}

// ==================== 日期时间转换 ====================

// ToDate 将字符串按照指定格式转换为日期/时间 (STR_TO_DATE)
// SELECT STR_TO_DATE('2023-10-26', '%Y-%m-%d');
// SELECT STR_TO_DATE('2023年10月26日', '%Y年%m月%d日');
// SELECT STR_TO_DATE('10/26/2023 10:30:45', '%m/%d/%Y %H:%i:%s');
// SELECT * FROM orders WHERE order_date = STR_TO_DATE('20231026', '%Y%m%d');
func (e StringExpr[T]) ToDate(format string) DateTimeExpr[string] {
	return DateTimeOf[string](clause.Expr{
		SQL:  "STR_TO_DATE(?, ?)",
		Vars: []any{e.Unwrap(), format},
	})
}

// ==================== 网络函数 ====================

// InetAton 将点分十进制的IPv4地址转换为整数形式 (INET_ATON)
// SELECT INET_ATON('192.168.1.1'); -- 结果为 3232235777
// INSERT INTO ip_logs (ip_num) VALUES (INET_ATON('192.168.1.100'));
func (e StringExpr[T]) InetAton() IntExpr[uint32] {
	return IntOf[uint32](clause.Expr{
		SQL:  "INET_ATON(?)",
		Vars: []any{e.Unwrap()},
	})
}

func (e StringExpr[T]) Unwrap() clause.Expression {
	return e.baseComparableImpl.Expression
}
