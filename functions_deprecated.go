package gsql

//
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Mul 方法替代，如 expr1.Mul(expr2)
//func Mul(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? * ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Div 方法替代，如 expr1.Div(expr2)
//func Div(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? / ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Add 方法替代，如 expr1.Add(expr2)
//func Add(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? + ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Sub 方法替代，如 expr1.Sub(expr2)
//func Sub(expr1, expr2 field.Expression) field.FloatExpr {
//	return field.NewFloatExpr(clause.Expr{
//		SQL:  "? - ?",
//		Vars: []any{expr1, expr2},
//	})
//}
//
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Mod 方法替代，如 expr1.Mod(expr2)
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

//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Abs 方法替代，如 expr.Abs()
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
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Ceil 方法替代，如 expr.Ceil()
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
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Ceil 方法替代，如 expr.Ceil()
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
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Floor 方法替代，如 expr.Floor()
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
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Round 方法替代，如 expr.Round(decimals...)
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
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Pow 方法替代，如 expr.Pow(exponent)
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
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Pow 方法替代，如 expr.Pow(exponent)
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
//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Sqrt 方法替代，如 expr.Sqrt()
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

//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Sign 方法替代，如 expr.Sign()
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

//// Deprecated: 使用 IntExprT/FloatExprT/DecimalExprT 的 Truncate 方法替代，如 expr.Truncate(decimals)
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
