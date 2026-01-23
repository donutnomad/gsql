package fields

import (
	"fmt"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
)

var _ clause.Expression = (*TextExpr[string])(nil)

// ==================== TextExpr 定义 ====================

// TextExpr 文本类型表达式，用于 VARCHAR 和 TEXT 类型字段
// 支持比较操作和模式匹配操作
// 使用场景：
//   - CONCAT, SUBSTRING 等字符串函数的返回值
//   - UPPER, LOWER 等字符串转换函数的返回值
//   - 派生表中的文本列
type TextExpr[T any] struct {
	baseComparableImpl[T] // 只包含 Eq/Not/In/NotIn，字符串没有大小比较
	patternExprImpl[T]
	pointerExprImpl
	nullCondFuncSql
}

func NewTextExpr[T any](expr clause.Expression) TextExpr[T] {
	return TextExpr[T]{
		baseComparableImpl: baseComparableImpl[T]{Expression: expr},
		patternExprImpl:    patternExprImpl[T]{Expression: expr},
		pointerExprImpl:    pointerExprImpl{Expression: expr},
		nullCondFuncSql:    nullCondFuncSql{Expression: expr},
	}
}

// Build 实现 clause.Expression 接口
func (e TextExpr[T]) Build(builder clause.Builder) {
	e.baseComparableImpl.Expression.Build(builder)
}

func (e TextExpr[T]) ToExpr() clause.Expression {
	return e.baseComparableImpl.Expression
}

// As 创建一个别名字段
func (e TextExpr[T]) As(alias string) field.IField {
	return field.NewBaseFromSql(e.baseComparableImpl.Expression, alias)
}

// ==================== 类型转换 ====================

// Cast 类型转换 (CAST)
// SELECT CAST(price AS SIGNED) FROM products;
// SELECT CAST(amount AS DECIMAL(10,2)) FROM orders;
// targetType: SIGNED, UNSIGNED, DECIMAL(m,n), CHAR, DATE, DATETIME, TIME, BINARY 等
func (e TextExpr[T]) Cast(targetType string) clause.Expression {
	return clause.Expr{
		SQL:  "CAST(? AS " + targetType + ")",
		Vars: []any{e.baseComparableImpl.Expression},
	}
}

// CastSigned 转换为有符号整数 (CAST AS SIGNED)
func (e TextExpr[T]) CastSigned() IntExpr[int64] {
	return NewIntExpr[int64](clause.Expr{
		SQL:  "CAST(? AS SIGNED)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// CastUnsigned 转换为无符号整数 (CAST AS UNSIGNED)
func (e TextExpr[T]) CastUnsigned() IntExpr[uint64] {
	return NewIntExpr[uint64](clause.Expr{
		SQL:  "CAST(? AS UNSIGNED)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// CastDecimal 转换为小数 (CAST AS DECIMAL)
// precision: 总位数, scale: 小数位数
func (e TextExpr[T]) CastDecimal(precision, scale int) DecimalExpr[float64] {
	return NewDecimalExpr[float64](clause.Expr{
		SQL:  fmt.Sprintf("CAST(? AS DECIMAL(%d, %d))", precision, scale),
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// CastChar 转换为字符串 (CAST AS CHAR)
func (e TextExpr[T]) CastChar(length ...int) TextExpr[string] {
	if len(length) > 0 {
		return NewTextExpr[string](clause.Expr{
			SQL:  fmt.Sprintf("CAST(? AS CHAR(%d))", length[0]),
			Vars: []any{e.baseComparableImpl.Expression},
		})
	}
	return NewTextExpr[string](clause.Expr{
		SQL:  "CAST(? AS CHAR)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// ==================== 字符串函数 ====================

// Upper 将字符串转换为大写 (UPPER)，只对英文字母有效
// SELECT UPPER('hello world');
// SELECT UPPER(users.username) FROM users;
// SELECT * FROM products WHERE UPPER(product_code) = 'ABC123';
// UPDATE users SET username = UPPER(username) WHERE id = 1;
func (e TextExpr[T]) Upper() TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "UPPER(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

func init() {
	var t TextExpr[string]
	t.Lower().Like("123")
}

// Lower 将字符串转换为小写 (LOWER)，只对英文字母有效
// SELECT LOWER('HELLO WORLD');
// SELECT LOWER(users.email) FROM users;
// SELECT * FROM users WHERE LOWER(username) = 'admin';
// UPDATE users SET email = LOWER(email);
func (e TextExpr[T]) Lower() TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "LOWER(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Trim 去除字符串两端的空格 (TRIM)
// SELECT TRIM('  Hello World  ');
// SELECT TRIM(users.username) FROM users;
// UPDATE users SET email = TRIM(email);
func (e TextExpr[T]) Trim() TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "TRIM(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// LTrim 去除字符串左侧的空格 (LTRIM)
// SELECT LTRIM('  Hello World  ');
// SELECT LTRIM(users.name) FROM users;
// SELECT * FROM products WHERE LTRIM(code) != code;
// UPDATE users SET username = LTRIM(username);
func (e TextExpr[T]) LTrim() TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "LTRIM(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// RTrim 去除字符串右侧的空格 (RTRIM)
// SELECT RTRIM('  Hello World  ');
// SELECT RTRIM(description) FROM products;
// SELECT * FROM users WHERE RTRIM(email) != email;
// UPDATE articles SET title = RTRIM(title);
func (e TextExpr[T]) RTrim() TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "RTRIM(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Substring 从字符串中提取子字符串 (SUBSTRING)，位置从1开始
// SELECT SUBSTRING('Hello World', 1, 5);
// SELECT SUBSTRING(users.email, 1, LOCATE('@', users.email) - 1) FROM users;
// SELECT SUBSTRING(product_code, 4, 3) FROM products;
// pos: 起始位置（从1开始）, length: 长度
func (e TextExpr[T]) Substring(pos, length int) TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "SUBSTRING(?, ?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, pos, length},
	})
}

// Left 从字符串左侧提取指定长度的子字符串 (LEFT)
// SELECT LEFT('Hello World', 5);
// SELECT LEFT(users.name, 1) as initial FROM users;
// SELECT * FROM products WHERE LEFT(product_code, 2) = 'AB';
// SELECT LEFT(email, LOCATE('@', email) - 1) FROM users;
func (e TextExpr[T]) Left(length int) TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "LEFT(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, length},
	})
}

// Right 从字符串右侧提取指定长度的子字符串 (RIGHT)
// SELECT RIGHT('Hello World', 5);
// SELECT RIGHT(phone, 4) as last_four FROM users;
// SELECT * FROM files WHERE RIGHT(filename, 4) = '.pdf';
// SELECT RIGHT(product_code, 3) FROM products;
func (e TextExpr[T]) Right(length int) TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "RIGHT(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, length},
	})
}

// Length 返回字符串的字节长度 (LENGTH)
// SELECT LENGTH('Hello');
// SELECT LENGTH('你好');
// SELECT users.name, LENGTH(users.name) FROM users;
// SELECT * FROM products WHERE LENGTH(product_code) = 8;
// 注意: UTF-8编码中一个中文字符通常占3个字节
func (e TextExpr[T]) Length() IntExpr[int64] {
	return NewIntExpr[int64](clause.Expr{
		SQL:  "LENGTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// CharLength 返回字符串的字符长度 (CHAR_LENGTH)，多字节字符按一个字符计算
// SELECT CHAR_LENGTH('Hello');
// SELECT CHAR_LENGTH('你好');
// SELECT users.name, CHAR_LENGTH(users.name) FROM users;
// SELECT * FROM articles WHERE CHAR_LENGTH(content) > 1000;
func (e TextExpr[T]) CharLength() IntExpr[int64] {
	return NewIntExpr[int64](clause.Expr{
		SQL:  "CHAR_LENGTH(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Concat 拼接多个字符串 (CONCAT)，任意参数为NULL则返回NULL
// SELECT CONCAT('Hello', ' ', 'World');
// SELECT CONCAT(users.first_name, ' ', users.last_name) as full_name FROM users;
// SELECT CONCAT('User:', users.id) FROM users;
// SELECT CONCAT(YEAR(NOW()), '-', MONTH(NOW()));
func (e TextExpr[T]) Concat(args ...clause.Expression) TextExpr[T] {
	placeholders := "?"
	allArgs := []any{e.baseComparableImpl.Expression}
	for _, arg := range args {
		placeholders += ", ?"
		allArgs = append(allArgs, arg)
	}
	return NewTextExpr[T](clause.Expr{
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
func (e TextExpr[T]) ConcatWS(separator string, args ...clause.Expression) TextExpr[T] {
	placeholders := "?, ?"
	allArgs := []any{separator, e.baseComparableImpl.Expression}
	for _, arg := range args {
		placeholders += ", ?"
		allArgs = append(allArgs, arg)
	}
	return NewTextExpr[T](clause.Expr{
		SQL:  "CONCAT_WS(" + placeholders + ")",
		Vars: allArgs,
	})
}

// Replace 替换字符串中所有出现的子字符串 (REPLACE)
// SELECT REPLACE('Hello World', 'World', 'MySQL');
// SELECT REPLACE('www.example.com', 'www', 'mail');
// SELECT REPLACE(phone, '-', ”) FROM users;
// UPDATE products SET description = REPLACE(description, 'old', 'new');
func (e TextExpr[T]) Replace(from, to string) TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "REPLACE(?, ?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, from, to},
	})
}

// Locate 查找子字符串位置 (LOCATE)
// SELECT LOCATE('@', email) FROM users;
// 返回子字符串第一次出现的位置（从1开始），未找到返回0
func (e TextExpr[T]) Locate(substr string) IntExpr[int64] {
	return NewIntExpr[int64](clause.Expr{
		SQL:  "LOCATE(?, ?)",
		Vars: []any{substr, e.baseComparableImpl.Expression},
	})
}

// Reverse 反转字符串 (REVERSE)
// SELECT REVERSE(name) FROM users;
func (e TextExpr[T]) Reverse() TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "REVERSE(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}

// Repeat 重复字符串 (REPEAT)
// SELECT REPEAT('*', 10);
func (e TextExpr[T]) Repeat(count int) TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "REPEAT(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, count},
	})
}

// LPad 左侧填充 (LPAD)
// SELECT LPAD(id, 5, '0') FROM users; -- 结果如 "00001"
func (e TextExpr[T]) LPad(length int, padStr string) TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "LPAD(?, ?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, length, padStr},
	})
}

// RPad 右侧填充 (RPAD)
// SELECT RPAD(name, 20, ' ') FROM users;
func (e TextExpr[T]) RPad(length int, padStr string) TextExpr[T] {
	return NewTextExpr[T](clause.Expr{
		SQL:  "RPAD(?, ?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, length, padStr},
	})
}

// ==================== 日期时间转换 ====================

// ToDate 将字符串按照指定格式转换为日期/时间 (STR_TO_DATE)
// SELECT STR_TO_DATE('2023-10-26', '%Y-%m-%d');
// SELECT STR_TO_DATE('2023年10月26日', '%Y年%m月%d日');
// SELECT STR_TO_DATE('10/26/2023 10:30:45', '%m/%d/%Y %H:%i:%s');
// SELECT * FROM orders WHERE order_date = STR_TO_DATE('20231026', '%Y%m%d');
func (e TextExpr[T]) ToDate(format string) DateTimeExpr[string] {
	return NewDateTimeExpr[string](clause.Expr{
		SQL:  "STR_TO_DATE(?, ?)",
		Vars: []any{e.baseComparableImpl.Expression, format},
	})
}

// ==================== 网络函数 ====================

// InetAton 将点分十进制的IPv4地址转换为整数形式 (INET_ATON)
// SELECT INET_ATON('192.168.1.1'); -- 结果为 3232235777
// INSERT INTO ip_logs (ip_num) VALUES (INET_ATON('192.168.1.100'));
func (e TextExpr[T]) InetAton() IntExpr[uint32] {
	return NewIntExpr[uint32](clause.Expr{
		SQL:  "INET_ATON(?)",
		Vars: []any{e.baseComparableImpl.Expression},
	})
}
