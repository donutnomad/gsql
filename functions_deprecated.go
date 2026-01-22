package gsql

//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Mul 方法替代，如 expr1.Mul(expr2)
//func Mul(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? * ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Div 方法替代，如 expr1.Div(expr2)
//func Div(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? / ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Add 方法替代，如 expr1.Add(expr2)
//func Add(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? + ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Sub 方法替代，如 expr1.Sub(expr2)
//func Sub(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? - ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Mod 方法替代，如 expr1.Mod(expr2)
//func Mod(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? % ?",
//		Vars: []any{expr1, expr2},
//	})
//}

// ==================== 字符串函数 ====================

//// Deprecated: 使用 TextExpr 的 Concat 方法替代，如 textExpr.Concat(args...)
//// CONCAT 拼接多个字符串，任意参数为NULL则返回NULL
//// SELECT CONCAT('Hello', ' ', 'World');
//// SELECT CONCAT(users.first_name, ' ', users.last_name) as full_name FROM users;
//// SELECT CONCAT('User:', users.id) FROM users;
//// SELECT CONCAT(YEAR(NOW()), '-', MONTH(NOW()));
//func CONCAT(args ...field.Expression) field.StringExpr {
//	placeholders := ""
//	for i := range args {
//		if i > 0 {
//			placeholders += ", "
//		}
//		placeholders += "?"
//	}
//	return field.NewStringExpr(clause.Expr{
//		SQL:  fmt.Sprintf("CONCAT(%s)", placeholders),
//		Vars: lo.ToAnySlice(args),
//	})
//}
//
//// CONCAT_WS 用指定分隔符拼接多个字符串，自动跳过NULL值，分隔符为NULL则返回NULL
//// SELECT CONCAT_WS(',', 'A', 'B', 'C');
//// SELECT CONCAT_WS('-', users.last_name, users.first_name) FROM users;
//// SELECT CONCAT_WS('/', YEAR(date), MONTH(date), DAY(date)) FROM logs;
//// SELECT CONCAT_WS(', ', city, state, country) FROM addresses;
//func CONCAT_WS(separator string, args ...any) field.StringExpr {
//	placeholders := "?"
//	allArgs := []any{separator}
//	for range args {
//		placeholders += ", ?"
//	}
//	allArgs = append(allArgs, args...)
//	return field.NewStringExpr(clause.Expr{
//		SQL:  fmt.Sprintf("CONCAT_WS(%s)", placeholders),
//		Vars: allArgs,
//	})
//}

//// Deprecated: 使用 TextExpr 的 Length 方法替代，如 textExpr.Length()
//// LENGTH 返回字符串的字节长度，UTF-8编码中一个中文字符通常占3个字节
//// SELECT LENGTH('Hello');
//// SELECT LENGTH('你好');
//// SELECT users.name, LENGTH(users.name) FROM users;
//// SELECT * FROM products WHERE LENGTH(product_code) = 8;
//func LENGTH(str field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "LENGTH(?)",
//		Vars: []any{str},
//	})
//}
//
//// Deprecated: 使用 TextExpr 的 CharLength 方法替代，如 textExpr.CharLength()
//// CHAR_LENGTH 返回字符串的字符长度，多字节字符按一个字符计算，是CHARACTER_LENGTH的同义词
//// SELECT CHAR_LENGTH('Hello');
//// SELECT CHAR_LENGTH('你好');
//// SELECT users.name, CHAR_LENGTH(users.name) FROM users;
//// SELECT * FROM articles WHERE CHAR_LENGTH(content) > 1000;
//func CHAR_LENGTH(str field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "CHAR_LENGTH(?)",
//		Vars: []any{str},
//	})
//}

//// Deprecated: 使用 TextExpr 的 CharLength 方法替代，如 textExpr.CharLength()
//// CHARACTER_LENGTH 返回字符串的字符长度，多字节字符按一个字符计算，是CHAR_LENGTH的同义词
//// SELECT CHARACTER_LENGTH('Hello');
//// SELECT CHARACTER_LENGTH('你好世界');
//// SELECT CHARACTER_LENGTH(description) FROM products;
//// SELECT * FROM posts WHERE CHARACTER_LENGTH(title) < 50;
//func CHARACTER_LENGTH(str field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "CHARACTER_LENGTH(?)",
//		Vars: []any{str},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Upper 方法替代，如 textExpr.Upper()
//// UPPER 将字符串转换为大写，只对英文字母有效
//// SELECT UPPER('hello world');
//// SELECT UPPER(users.username) FROM users;
//// SELECT * FROM products WHERE UPPER(product_code) = 'ABC123';
//// UPDATE users SET username = UPPER(username) WHERE id = 1;
//func UPPER(str field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "UPPER(?)",
//		Vars: []any{str},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Upper 方法替代，如 textExpr.Upper()
//// UCASE 将字符串转换为大写，是UPPER的同义词
//// SELECT UCASE('hello world');
//// SELECT UCASE(email) FROM users;
//// SELECT * FROM codes WHERE UCASE(code) LIKE 'A%';
//// SELECT CONCAT(UCASE(LEFT(name, 1)), SUBSTRING(name, 2)) FROM users;
//func UCASE(str field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "UCASE(?)",
//		Vars: []any{str},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Lower 方法替代，如 textExpr.Lower()
//// LOWER 将字符串转换为小写，只对英文字母有效
//// SELECT LOWER('HELLO WORLD');
//// SELECT LOWER(users.email) FROM users;
//// SELECT * FROM users WHERE LOWER(username) = 'admin';
//// UPDATE users SET email = LOWER(email);
//func LOWER(str field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "LOWER(?)",
//		Vars: []any{str},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Lower 方法替代，如 textExpr.Lower()
//// LCASE 将字符串转换为小写，是LOWER的同义词
//// SELECT LCASE('HELLO WORLD');
//// SELECT LCASE(company_name) FROM companies;
//// SELECT * FROM domains WHERE LCASE(domain) = 'example.com';
//// SELECT LCASE(TRIM(email)) FROM users;
//func LCASE(str field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "LCASE(?)",
//		Vars: []any{str},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Substring 方法替代，如 textExpr.Substring(pos, length)
//// SUBSTRING 从字符串中提取子字符串，位置从1开始，是SUBSTR的同义词
//// SELECT SUBSTRING('Hello World', 1, 5);
//// SELECT SUBSTRING('Hello World', 7);
//// SELECT SUBSTRING(users.email, 1, LOCATE('@', users.email) - 1) FROM users;
//// SELECT SUBSTRING(product_code, 4, 3) FROM products;
//func SUBSTRING(str field.Expression, pos, length int) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "SUBSTRING(?, ?, ?)",
//		Vars: []any{str, pos, length},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Substring 方法替代，如 textExpr.Substring(pos, length)
//// SUBSTR 从字符串中提取子字符串，位置从1开始，是SUBSTRING的同义词
//// SELECT SUBSTR('Hello World', 1, 5);
//// SELECT SUBSTR('Hello World', 7);
//// SELECT SUBSTR(description, 1, 100) FROM articles;
//// SELECT SUBSTR(phone, -4) FROM users;
//func SUBSTR(str field.Expression, pos, length int) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "SUBSTR(?, ?, ?)",
//		Vars: []any{str, pos, length},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Left 方法替代，如 textExpr.Left(length)
//// LEFT 从字符串左侧提取指定长度的子字符串
//// SELECT LEFT('Hello World', 5);
//// SELECT LEFT(users.name, 1) as initial FROM users;
//// SELECT * FROM products WHERE LEFT(product_code, 2) = 'AB';
//// SELECT LEFT(email, LOCATE('@', email) - 1) FROM users;
//func LEFT(str field.Expression, length int) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "LEFT(?, ?)",
//		Vars: []any{str, length},
//	})
//}

//// Deprecated: 使用 TextExpr 的 Right 方法替代，如 textExpr.Right(length)
//// RIGHT 从字符串右侧提取指定长度的子字符串
//// SELECT RIGHT('Hello World', 5);
//// SELECT RIGHT(phone, 4) as last_four FROM users;
//// SELECT * FROM files WHERE RIGHT(filename, 4) = '.pdf';
//// SELECT RIGHT(product_code, 3) FROM products;
//func RIGHT(str field.Expression, length int) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "RIGHT(?, ?)",
//		Vars: []any{str, length},
//	})
//}
//
//// Deprecated: 使用 TextExpr 的 Locate 方法替代，如 textExpr.Locate(substr)
//// LOCATE 返回子字符串在字符串中第一次出现的位置（从1开始），未找到返回0，可选起始位置
//// SELECT LOCATE('World', 'Hello World');
//// SELECT LOCATE('o', 'Hello World');
//// SELECT LOCATE('o', 'Hello World', 6);
//// SELECT * FROM users WHERE LOCATE('@', email) > 0;
//func LOCATE(substr, str field.Expression, pos ...int) field.IntExpr {
//	if len(pos) > 0 {
//		return field.NewIntExpr(clause.Expr{
//			SQL:  "LOCATE(?, ?, ?)",
//			Vars: []any{substr, str, pos[0]},
//		})
//	}
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "LOCATE(?, ?)",
//		Vars: []any{substr, str},
//	})
//}
//
//// INSTR 返回子字符串在字符串中第一次出现的位置（从1开始），未找到返回0
//// SELECT INSTR('Hello World', 'World');
//// SELECT INSTR('Hello World', 'o');
//// SELECT * FROM urls WHERE INSTR(url, 'https://') = 1;
//// SELECT INSTR(email, '@') as at_position FROM users;
//func INSTR(str, substr field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "INSTR(?, ?)",
//		Vars: []any{str, substr},
//	})
//}
//
//// Deprecated: 使用 TextExpr 的 Replace 方法替代，如 textExpr.Replace(from, to)
//// REPLACE 替换字符串中所有出现的子字符串
//// SELECT REPLACE('Hello World', 'World', 'MySQL');
//// SELECT REPLACE('www.example.com', 'www', 'mail');
//// SELECT REPLACE(phone, '-', ") FROM users;
//// UPDATE products SET description = REPLACE(description, 'old', 'new');
//func REPLACE(str, fromStr, toStr field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "REPLACE(?, ?, ?)",
//		Vars: []any{str, fromStr, toStr},
//	})
//}
//
//// Deprecated: 使用 TextExpr 的 Trim 方法替代，如 textExpr.Trim()
//// TRIM 去除字符串两端的空格，也可指定去除的字符
//// SELECT TRIM('  Hello World  ');
//// SELECT TRIM(BOTH 'x' FROM 'xxxHelloxxx');
//// SELECT TRIM(users.username) FROM users;
//// UPDATE users SET email = TRIM(email);
//func TRIM(str field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "TRIM(?)",
//		Vars: []any{str},
//	})
//}
//
//// Deprecated: 使用 TextExpr 的 LTrim 方法替代，如 textExpr.LTrim()
//// LTRIM 去除字符串左侧的空格
//// SELECT LTRIM('  Hello World  ');
//// SELECT LTRIM(users.name) FROM users;
//// SELECT * FROM products WHERE LTRIM(code) != code;
//// UPDATE users SET username = LTRIM(username);
//func LTRIM(str field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "LTRIM(?)",
//		Vars: []any{str},
//	})
//}
//
//// Deprecated: 使用 TextExpr 的 RTrim 方法替代，如 textExpr.RTrim()
//// RTRIM 去除字符串右侧的空格
//// SELECT RTRIM('  Hello World  ');
//// SELECT RTRIM(description) FROM products;
//// SELECT * FROM users WHERE RTRIM(email) != email;
//// UPDATE articles SET title = RTRIM(title);
//func RTRIM(str field.Expression) field.StringExpr {
//	return field.NewStringExpr(clause.Expr{
//		SQL:  "RTRIM(?)",
//		Vars: []any{str},
//	})
//}

// ==================== 数值函数 ====================

//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Abs 方法替代，如 expr.Abs()
//// ABS 返回数值的绝对值
//// SELECT ABS(-10);
//// SELECT ABS(10);
//// SELECT ABS(users.balance) FROM users;
//// SELECT * FROM transactions WHERE ABS(amount) > 1000;
//func ABS(x field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "ABS(?)",
//		Vars: []any{x},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Ceil 方法替代，如 expr.Ceil()
//// CEIL 向上取整，返回大于或等于X的最小整数，是CEILING的同义词
//// SELECT CEIL(4.3);
//// SELECT CEIL(4.9);
//// SELECT CEIL(-4.3);
//// SELECT CEIL(price * 1.1) FROM products;
//func CEIL(x field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "CEIL(?)",
//		Vars: []any{x},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Ceil 方法替代，如 expr.Ceil()
//// CEILING 向上取整，返回大于或等于X的最小整数，是CEIL的同义词
//// SELECT CEILING(4.3);
//// SELECT CEILING(4.9);
//// SELECT CEILING(-4.3);
//// SELECT CEILING(total / 10) * 10 FROM orders;
//func CEILING(x field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "CEILING(?)",
//		Vars: []any{x},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Floor 方法替代，如 expr.Floor()
//// FLOOR 向下取整，返回小于或等于X的最大整数
//// SELECT FLOOR(4.3);
//// SELECT FLOOR(4.9);
//// SELECT FLOOR(-4.3);
//// SELECT FLOOR(price * 0.9) FROM products;
//func FLOOR(x field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "FLOOR(?)",
//		Vars: []any{x},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Round 方法替代，如 expr.Round(decimals...)
//// ROUND 四舍五入到指定小数位数，默认四舍五入到整数
//// SELECT ROUND(4.567);
//// SELECT ROUND(4.567, 2);
//// SELECT ROUND(4.567, 0);
//// SELECT ROUND(price, 2) FROM products;
//// SELECT ROUND(123.456, -1);
//func ROUND(x field.Expression, d ...int) field.FloatExpr {
//	if len(d) > 0 {
//		return field.NewFloatExpr(clause.Expr{
//			SQL:  "ROUND(?, ?)",
//			Vars: []any{x, d[0]},
//		})
//	}
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "ROUND(?)",
//		Vars: []any{x},
//	})
//}
//
//// MOD 返回N除以M的余数（模运算）
//// SELECT MOD(10, 3);
//// SELECT MOD(234, 10);
//// SELECT MOD(-10, 3);
//// SELECT * FROM users WHERE MOD(id, 2) = 0;
//func MOD(n, m field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "MOD(?, ?)",
//		Vars: []any{n, m},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Pow 方法替代，如 expr.Pow(exponent)
//// POWER 返回X的Y次幂，是POW的同义词
//// SELECT POWER(2, 3);
//// SELECT POWER(10, 2);
//// SELECT POWER(5, -1);
//// SELECT POWER(users.level, 2) FROM users;
//func POWER(x, y field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "POWER(?, ?)",
//		Vars: []any{x, y},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Pow 方法替代，如 expr.Pow(exponent)
//// POW 返回X的Y次幂，是POWER的同义词
//// SELECT POW(2, 3);
//// SELECT POW(10, 2);
//// SELECT POW(distance, 2) FROM locations;
//// SELECT SQRT(POW(x2 - x1, 2) + POW(y2 - y1, 2)) as distance FROM points;
//func POW(x, y field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "POW(?, ?)",
//		Vars: []any{x, y},
//	})
//}
//
//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Sqrt 方法替代，如 expr.Sqrt()
//// SQRT 返回X的平方根，X必须为非负数
//// SELECT SQRT(4);
//// SELECT SQRT(16);
//// SELECT SQRT(2);
//// SELECT SQRT(area) as side_length FROM squares;
//func SQRT(x field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "SQRT(?)",
//		Vars: []any{x},
//	})
//}

//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Sign 方法替代，如 expr.Sign()
//// SIGN 返回数值的符号：负数返回-1，零返回0，正数返回1
//// SELECT SIGN(-10);
//// SELECT SIGN(0);
//// SELECT SIGN(10);
//// SELECT SIGN(balance) FROM accounts;
//func SIGN(x field.Expression) field.IntExpr {
//	return field.NewIntExpr(clause.Expr{
//		SQL:  "SIGN(?)",
//		Vars: []any{x},
//	})
//}

//// Deprecated: 使用 IntExpr/FloatExpr/DecimalExpr 的 Truncate 方法替代，如 expr.Truncate(decimals)
//// TRUNCATE 截断数值到指定小数位数，不进行四舍五入
//// SELECT TRUNCATE(4.567, 2);
//// SELECT TRUNCATE(4.567, 0);
//// SELECT TRUNCATE(123.456, -1);
//// SELECT TRUNCATE(price, 2) FROM products;
//func TRUNCATE(x field.Expression, d int) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "TRUNCATE(?, ?)",
//		Vars: []any{x, d},
//	})
//}

// ==================== 以下函数已迁移到类型化方法，建议使用对应类型的方法 ====================
//// 例如: DateTimeExpr.Format(), DateTimeExpr.Year(), DateTimeExpr.AddInterval() 等
//
//// DATE_FORMAT 格式化日期/时间为指定字符串，支持各种格式化符号
//// Deprecated: 使用 DateTimeExpr.Format() 或 DateExpr.Format() 代替
//// SELECT DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s');
//// SELECT DATE_FORMAT('2023-10-26', '%Y年%m月%d日');
//// SELECT DATE_FORMAT(users.birthday, '%W %M %Y') FROM users;
//// SELECT DATE_FORMAT(NOW(), '%Y%m%d%H%i%s');
//func DATE_FORMAT(date field.Expression, format string) field.TextExpr[string] {
//	return field.NewTextExpr[string](clause.Expr{
//		SQL:  "DATE_FORMAT(?, ?)",
//		Vars: []any{date, format},
//	})
//}
//
//// YEAR 提取日期中的年份部分 (1000-9999)
//// Deprecated: 使用 DateTimeExpr.Year() 或 DateExpr.Year() 代替
//// SELECT YEAR(NOW());
//// SELECT YEAR('2023-10-26');
//// SELECT * FROM users WHERE YEAR(birthday) = 1990;
//// SELECT YEAR(users.created_at) as year FROM users GROUP BY year;
//func YEAR(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "YEAR(?)",
//		Vars: []any{date},
//	})
//}
//
//// MONTH 提取日期中的月份部分 (1-12)
//// Deprecated: 使用 DateTimeExpr.Month() 或 DateExpr.Month() 代替
//// SELECT MONTH(NOW());
//// SELECT MONTH('2023-10-26');
//// SELECT * FROM orders WHERE MONTH(order_date) = 10;
//// SELECT MONTH(users.birthday), COUNT(*) FROM users GROUP BY MONTH(users.birthday);
//func MONTH(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "MONTH(?)",
//		Vars: []any{date},
//	})
//}

//// DAY 提取日期中一个月中的天数 (1-31)，是DAYOFMONTH的同义词
//// Deprecated: 使用 DateTimeExpr.Day() 或 DateExpr.Day() 代替
//// SELECT DAY(NOW());
//// SELECT DAY('2023-10-26');
//// SELECT * FROM events WHERE DAY(event_date) = 15;
//// SELECT YEAR(date), MONTH(date), DAY(date) FROM logs;
//func DAY(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "DAY(?)",
//		Vars: []any{date},
//	})
//}

//// DAYOFMONTH 提取日期中一个月中的天数 (1-31)，是DAY的同义词
//// Deprecated: 使用 DateTimeExpr.DayOfMonth() 或 DateExpr.DayOfMonth() 代替
//// SELECT DAYOFMONTH(NOW());
//// SELECT DAYOFMONTH('2023-10-26');
//// SELECT * FROM users WHERE DAYOFMONTH(birthday) = 1;
//// SELECT DAYOFMONTH(created_at) FROM orders;
//func DAYOFMONTH(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "DAYOFMONTH(?)",
//		Vars: []any{date},
//	})
//}

//// WEEK 提取日期在一年中的周数 (0-53)，可选第二个参数指定周开始于周日还是周一
//// Deprecated: 使用 DateTimeExpr.Week() 或 DateExpr.Week() 代替
//// SELECT WEEK(NOW());
//// SELECT WEEK('2023-10-26');
//// SELECT WEEK(NOW(), 1);
//// SELECT * FROM orders WHERE WEEK(order_date) = WEEK(NOW());
//func WEEK(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "WEEK(?)",
//		Vars: []any{date},
//	})
//}

//// WEEKOFYEAR 提取日期在一年中的周数 (1-53)，相当于WEEK(date, 3)
//// Deprecated: 使用 DateTimeExpr.WeekOfYear() 或 DateExpr.WeekOfYear() 代替
//// SELECT WEEKOFYEAR(NOW());
//// SELECT WEEKOFYEAR('2023-10-26');
//// SELECT * FROM events WHERE WEEKOFYEAR(event_date) = 43;
//// SELECT WEEKOFYEAR(created_at), COUNT(*) FROM orders GROUP BY WEEKOFYEAR(created_at);
//func WEEKOFYEAR(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "WEEKOFYEAR(?)",
//		Vars: []any{date},
//	})
//}

//// HOUR 提取时间中的小时部分 (0-23)
//// Deprecated: 使用 DateTimeExpr.Hour() 或 TimeExpr.Hour() 代替
//// SELECT HOUR(NOW());
//// SELECT HOUR('2023-10-26 14:30:45');
//// SELECT * FROM logs WHERE HOUR(log_time) BETWEEN 9 AND 17;
//// SELECT HOUR(users.last_login) FROM users;
//func HOUR(time field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "HOUR(?)",
//		Vars: []any{time},
//	})
//}

//// MINUTE 提取时间中的分钟部分 (0-59)
//// Deprecated: 使用 DateTimeExpr.Minute() 或 TimeExpr.Minute() 代替
//// SELECT MINUTE(NOW());
//// SELECT MINUTE('2023-10-26 14:30:45');
//// SELECT * FROM schedules WHERE MINUTE(start_time) = 0;
//// SELECT HOUR(time), MINUTE(time) FROM appointments;
//func MINUTE(time field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "MINUTE(?)",
//		Vars: []any{time},
//	})
//}

//// SECOND 提取时间中的秒数部分 (0-59)
//// Deprecated: 使用 DateTimeExpr.Second() 或 TimeExpr.Second() 代替
//// SELECT SECOND(NOW());
//// SELECT SECOND('2023-10-26 14:30:45');
//// SELECT * FROM events WHERE SECOND(event_time) = 0;
//// SELECT HOUR(time), MINUTE(time), SECOND(time) FROM logs;
//func SECOND(time field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "SECOND(?)",
//		Vars: []any{time},
//	})
//}

//// DAYOFWEEK 返回日期在一周中的索引 (1=周日, 2=周一, ..., 7=周六)
//// Deprecated: 使用 DateTimeExpr.DayOfWeek() 或 DateExpr.DayOfWeek() 代替
//// SELECT DAYOFWEEK(NOW());
//// SELECT DAYOFWEEK('2023-10-26');
//// SELECT * FROM events WHERE DAYOFWEEK(event_date) IN (1, 7);
//// SELECT CASE DAYOFWEEK(date) WHEN 1 THEN '周日' WHEN 2 THEN '周一' END FROM logs;
//func DAYOFWEEK(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "DAYOFWEEK(?)",
//		Vars: []any{date},
//	})
//}

//// DAYOFYEAR 返回日期在一年中的天数 (1-366)
//// Deprecated: 使用 DateTimeExpr.DayOfYear() 或 DateExpr.DayOfYear() 代替
//// SELECT DAYOFYEAR(NOW());
//// SELECT DAYOFYEAR('2023-10-26');
//// SELECT * FROM logs WHERE DAYOFYEAR(log_date) = 1;
//// SELECT DAYOFYEAR(created_at), COUNT(*) FROM orders GROUP BY DAYOFYEAR(created_at);
//func DAYOFYEAR(date field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "DAYOFYEAR(?)",
//		Vars: []any{date},
//	})
//}

//// DATE_ADD 在日期上增加一个时间间隔，支持多种时间单位
//// Deprecated: 使用 DateTimeExpr.AddInterval() 或 DateExpr.AddInterval() 代替
//// SELECT DATE_ADD(NOW(), INTERVAL 1 DAY);
//// SELECT DATE_ADD('2023-10-26', INTERVAL 2 HOUR);
//// SELECT DATE_ADD(users.created_at, INTERVAL 7 DAY) FROM users;
//// SELECT * FROM orders WHERE DATE_ADD(order_date, INTERVAL 30 DAY) > NOW();
//// 支持单位: MICROSECOND, SECOND, MINUTE, HOUR, DAY, WEEK, MONTH, QUARTER, YEAR
//func DATE_ADD(date field.Expression, interval string) field.ExpressionTo {
//	// 解析并验证 interval 格式 (例如: "1 DAY")
//	parts := strings.Fields(interval)
//	if len(parts) != 2 {
//		panic("DATE_ADD: invalid interval format, expected '<number> <unit>' (e.g., '1 DAY')")
//	}
//
//	// 验证数字部分
//	num, err := strconv.Atoi(parts[0])
//	if err != nil {
//		panic(fmt.Sprintf("DATE_ADD: interval value must be a number, got: %s", parts[0]))
//	}
//
//	// 验证单位部分
//	unit := strings.ToUpper(parts[1])
//	if !allowedIntervalUnits[unit] {
//		panic(fmt.Sprintf("DATE_ADD: invalid interval unit: %s", unit))
//	}
//
//	// 使用验证后的值重新构建 interval，而不是使用原始输入
//	// 这样可以防止恶意输入绕过验证
//	safeInterval := fmt.Sprintf("%d %s", num, unit)
//
//	return ExprTo{clause.Expr{
//		SQL:  fmt.Sprintf("DATE_ADD(?, INTERVAL %s)", safeInterval),
//		Vars: []any{date},
//	}}
//}

//// DATE_SUB 从日期中减去一个时间间隔，支持多种时间单位
//// Deprecated: 使用 DateTimeExpr.SubInterval() 或 DateExpr.SubInterval() 代替
//// SELECT DATE_SUB(NOW(), INTERVAL 1 DAY);
//// SELECT DATE_SUB('2023-10-26', INTERVAL 2 HOUR);
//// SELECT DATE_SUB(users.expire_date, INTERVAL 1 MONTH) FROM users;
//// SELECT * FROM logs WHERE log_date >= DATE_SUB(NOW(), INTERVAL 7 DAY);
//// 支持单位: MICROSECOND, SECOND, MINUTE, HOUR, DAY, WEEK, MONTH, QUARTER, YEAR
//func DATE_SUB(date field.Expression, interval string) field.ExpressionTo {
//	// 解析并验证 interval 格式 (例如: "1 DAY")
//	parts := strings.Fields(interval)
//	if len(parts) != 2 {
//		panic("DATE_SUB: invalid interval format, expected '<number> <unit>' (e.g., '1 DAY')")
//	}
//
//	// 验证数字部分
//	num, err := strconv.Atoi(parts[0])
//	if err != nil {
//		panic(fmt.Sprintf("DATE_SUB: interval value must be a number, got: %s", parts[0]))
//	}
//
//	// 验证单位部分
//	unit := strings.ToUpper(parts[1])
//	if !allowedIntervalUnits[unit] {
//		panic(fmt.Sprintf("DATE_SUB: invalid interval unit: %s", unit))
//	}
//
//	// 使用验证后的值重新构建 interval，而不是使用原始输入
//	// 这样可以防止恶意输入绕过验证
//	safeInterval := fmt.Sprintf("%d %s", num, unit)
//
//	return ExprTo{clause.Expr{
//		SQL:  fmt.Sprintf("DATE_SUB(?, INTERVAL %s)", safeInterval),
//		Vars: []any{date},
//	}}
//}

//// DATEDIFF 返回两个日期之间相差的天数 (date1 - date2)，只计算日期部分，忽略时间
//// Deprecated: 使用 DateTimeExpr.DateDiff() 或 DateExpr.DateDiff() 代替
//// SELECT DATEDIFF(NOW(), '2023-01-01');
//// SELECT DATEDIFF('2023-10-26', '2023-10-20');
//// SELECT users.name, DATEDIFF(NOW(), users.birthday) / 365 as age FROM users;
//// SELECT * FROM orders WHERE DATEDIFF(NOW(), order_date) > 30;
//func DATEDIFF(expr1, expr2 field.Expression) field.IntExpr[int64] {
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  "DATEDIFF(?, ?)",
//		Vars: []any{expr1, expr2},
//	})
//}

//// TIMEDIFF 返回两个时间/日期时间之间的差值，结果为时间格式 (HH:MM:SS)
//// Deprecated: 使用 DateTimeExpr.TimeDiff() 或 TimeExpr.TimeDiff() 代替
//// SELECT TIMEDIFF(NOW(), '2023-10-26 10:00:00');
//// SELECT TIMEDIFF('14:30:00', '10:15:00');
//// SELECT TIMEDIFF(end_time, start_time) FROM events;
//// SELECT * FROM logs WHERE TIMEDIFF(NOW(), log_time) > '01:00:00';
//func TIMEDIFF(expr1, expr2 field.Expression) field.ExpressionTo {
//	return ExprTo{clause.Expr{
//		SQL:  "TIMEDIFF(?, ?)",
//		Vars: []any{expr1, expr2},
//	}}
//}

//// TIMESTAMPDIFF 返回两个日期时间表达式之间的差值，以指定单位表示 (expr2 - expr1)
//// Deprecated: 使用 DateTimeExpr.TimestampDiff() 代替
//// SELECT TIMESTAMPDIFF(SECOND, '2023-10-26 10:00:00', '2023-10-26 10:05:00');
//// SELECT TIMESTAMPDIFF(HOUR, start_time, end_time) FROM events;
//// SELECT TIMESTAMPDIFF(YEAR, users.birthday, NOW()) as age FROM users;
//// SELECT * FROM orders WHERE TIMESTAMPDIFF(DAY, order_date, NOW()) > 30;
//// 支持单位: MICROSECOND, SECOND, MINUTE, HOUR, DAY, WEEK, MONTH, QUARTER, YEAR
//func TIMESTAMPDIFF(unit string, expr1, expr2 field.Expression) field.IntExpr[int64] {
//	// 验证单位参数
//	unit = strings.ToUpper(strings.TrimSpace(unit))
//	if !allowedIntervalUnits[unit] {
//		panic(fmt.Sprintf("TIMESTAMPDIFF: invalid unit: %s", unit))
//	}
//
//	return field.NewIntExprT[int64](clause.Expr{
//		SQL:  fmt.Sprintf("TIMESTAMPDIFF(%s, ?, ?)", unit),
//		Vars: []any{expr1, expr2},
//	})
//}
