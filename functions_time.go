package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
)

// CURRENT_TIMESTAMP 返回当前日期和时间 (YYYY-MM-DD HH:MM:SS)，是NOW()的同义词
// SELECT CURRENT_TIMESTAMP;
// SELECT CURRENT_TIMESTAMP();
// INSERT INTO logs (created_at) VALUES (CURRENT_TIMESTAMP);
// UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = 1;
func CURRENT_TIMESTAMP() field.ExpressionTo {
	return ExprTo{clause.Expr{
		SQL: "CURRENT_TIMESTAMP",
	}}
}

// ==================== 日期和时间函数 ====================

// NOW 返回当前日期和时间 (YYYY-MM-DD HH:MM:SS)
// SELECT NOW();
// SELECT NOW() + 0;
// INSERT INTO logs (created_at) VALUES (NOW());
// SELECT * FROM orders WHERE order_time > NOW() - INTERVAL 1 DAY;
func NOW() field.ExpressionTo {
	return ExprTo{clause.Expr{SQL: "NOW()"}}
}

// CURRENT_DATE 返回当前日期 (YYYY-MM-DD)，不包含时间部分
// SELECT CURRENT_DATE;
// SELECT CURRENT_DATE();
// SELECT * FROM users WHERE DATE(created_at) = CURRENT_DATE;
// SELECT DATEDIFF(CURRENT_DATE, users.birthday) FROM users;
func CURRENT_DATE() field.ExpressionTo {
	return ExprTo{clause.Expr{SQL: "CURRENT_DATE()"}}
}

// CURDATE 返回当前日期 (YYYY-MM-DD)，是CURRENT_DATE的同义词
// SELECT CURDATE();
// SELECT CURDATE() + 0;
// SELECT * FROM events WHERE event_date >= CURDATE();
// SELECT YEAR(CURDATE()), MONTH(CURDATE()), DAY(CURDATE());
func CURDATE() field.ExpressionTo {
	return ExprTo{clause.Expr{SQL: "CURDATE()"}}
}

// CURRENT_TIME 返回当前时间 (HH:MM:SS)，不包含日期部分
// SELECT CURRENT_TIME;
// SELECT CURRENT_TIME();
// SELECT CURRENT_TIME + 0;
// SELECT * FROM schedules WHERE start_time <= CURRENT_TIME AND end_time >= CURRENT_TIME;
func CURRENT_TIME() field.ExpressionTo {
	return ExprTo{clause.Expr{SQL: "CURRENT_TIME()"}}
}

// CURTIME 返回当前时间 (HH:MM:SS)，是CURRENT_TIME的同义词
// SELECT CURTIME();
// SELECT CURTIME() + 0;
// SELECT HOUR(CURTIME()), MINUTE(CURTIME()), SECOND(CURTIME());
// SELECT * FROM shifts WHERE shift_start <= CURTIME() AND shift_end >= CURTIME();
func CURTIME() field.ExpressionTo {
	return ExprTo{clause.Expr{SQL: "CURTIME()"}}
}

// UTC_TIMESTAMP 返回当前的UTC日期和时间 (YYYY-MM-DD HH:MM:SS)
// SELECT UTC_TIMESTAMP;
// SELECT UTC_TIMESTAMP();
// SELECT UTC_TIMESTAMP(), NOW();
// INSERT INTO global_logs (created_at_utc) VALUES (UTC_TIMESTAMP());
func UTC_TIMESTAMP() field.ExpressionTo {
	return ExprTo{clause.Expr{SQL: "UTC_TIMESTAMP()"}}
}

// UNIX_TIMESTAMP 返回Unix时间戳（秒），如果不提供参数则返回当前时间戳，提供参数则转换指定时间为时间戳
// SELECT UNIX_TIMESTAMP();
// SELECT UNIX_TIMESTAMP('2023-10-26 10:30:00');
// SELECT UNIX_TIMESTAMP(NOW());
// SELECT UNIX_TIMESTAMP(users.created_at) FROM users;
// SELECT * FROM orders WHERE UNIX_TIMESTAMP(order_time) > 1698306600;
func UNIX_TIMESTAMP(date ...field.Expression) field.IntExprT[int64] {
	if len(date) == 0 {
		return field.NewIntExprT[int64](clause.Expr{SQL: "UNIX_TIMESTAMP()"})
	}
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "UNIX_TIMESTAMP(?)",
		Vars: []any{date[0]},
	})
}

// FROM_UNIXTIME 将Unix时间戳（秒）转换为DATETIME类型，如果提供了format，将转换为VARCHAR类型
// SELECT FROM_UNIXTIME(1698306600, '%Y年%m月%d日 %H时%i分%s秒');
// SELECT FROM_UNIXTIME(1698306600);
// SELECT FROM_UNIXTIME(users.time);
// SELECT FROM_UNIXTIME(users.time + 3600);
func FROM_UNIXTIME(date field.Expression, format ...string) field.ExpressionTo {
	if len(format) > 0 {
		return ExprTo{clause.Expr{
			SQL:  "FROM_UNIXTIME(?, ?)",
			Vars: []any{date, format[0]},
		}}
	}
	return ExprTo{clause.Expr{
		SQL:  "FROM_UNIXTIME(?)",
		Vars: []any{date},
	}}
}

// STR_TO_DATE 将字符串按照指定格式转换为日期/时间，格式需要与字符串匹配
// SELECT STR_TO_DATE('2023-10-26', '%Y-%m-%d');
// SELECT STR_TO_DATE('2023年10月26日', '%Y年%m月%d日');
// SELECT STR_TO_DATE('10/26/2023 10:30:45', '%m/%d/%Y %H:%i:%s');
// SELECT * FROM orders WHERE order_date = STR_TO_DATE('20231026', '%Y%m%d');
func STR_TO_DATE(str string, format string) field.ExpressionTo {
	return ExprTo{clause.Expr{
		SQL:  "STR_TO_DATE(?, ?)",
		Vars: []any{str, format},
	}}
}

// ==================== 以下函数已迁移到类型化方法，建议使用对应类型的方法 ====================
// 例如: DateTimeExpr.Format(), DateTimeExpr.Year(), DateTimeExpr.AddInterval() 等

// DATE_FORMAT 格式化日期/时间为指定字符串，支持各种格式化符号
// Deprecated: 使用 DateTimeExpr.Format() 或 DateExpr.Format() 代替
// SELECT DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s');
// SELECT DATE_FORMAT('2023-10-26', '%Y年%m月%d日');
// SELECT DATE_FORMAT(users.birthday, '%W %M %Y') FROM users;
// SELECT DATE_FORMAT(NOW(), '%Y%m%d%H%i%s');
func DATE_FORMAT(date field.Expression, format string) field.TextExpr[string] {
	return field.NewTextExpr[string](clause.Expr{
		SQL:  "DATE_FORMAT(?, ?)",
		Vars: []any{date, format},
	})
}

// YEAR 提取日期中的年份部分 (1000-9999)
// Deprecated: 使用 DateTimeExpr.Year() 或 DateExpr.Year() 代替
// SELECT YEAR(NOW());
// SELECT YEAR('2023-10-26');
// SELECT * FROM users WHERE YEAR(birthday) = 1990;
// SELECT YEAR(users.created_at) as year FROM users GROUP BY year;
func YEAR(date field.Expression) field.IntExprT[int64] {
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "YEAR(?)",
		Vars: []any{date},
	})
}

// MONTH 提取日期中的月份部分 (1-12)
// Deprecated: 使用 DateTimeExpr.Month() 或 DateExpr.Month() 代替
// SELECT MONTH(NOW());
// SELECT MONTH('2023-10-26');
// SELECT * FROM orders WHERE MONTH(order_date) = 10;
// SELECT MONTH(users.birthday), COUNT(*) FROM users GROUP BY MONTH(users.birthday);
func MONTH(date field.Expression) field.IntExprT[int64] {
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "MONTH(?)",
		Vars: []any{date},
	})
}

//// DAY 提取日期中一个月中的天数 (1-31)，是DAYOFMONTH的同义词
//// Deprecated: 使用 DateTimeExpr.Day() 或 DateExpr.Day() 代替
//// SELECT DAY(NOW());
//// SELECT DAY('2023-10-26');
//// SELECT * FROM events WHERE DAY(event_date) = 15;
//// SELECT YEAR(date), MONTH(date), DAY(date) FROM logs;
//func DAY(date field.Expression) field.IntExprT[int64] {
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
//func DAYOFMONTH(date field.Expression) field.IntExprT[int64] {
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
//func WEEK(date field.Expression) field.IntExprT[int64] {
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
//func WEEKOFYEAR(date field.Expression) field.IntExprT[int64] {
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
//func HOUR(time field.Expression) field.IntExprT[int64] {
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
//func MINUTE(time field.Expression) field.IntExprT[int64] {
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
//func SECOND(time field.Expression) field.IntExprT[int64] {
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
//func DAYOFWEEK(date field.Expression) field.IntExprT[int64] {
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
//func DAYOFYEAR(date field.Expression) field.IntExprT[int64] {
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
//func DATEDIFF(expr1, expr2 field.Expression) field.IntExprT[int64] {
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
//func TIMESTAMPDIFF(unit string, expr1, expr2 field.Expression) field.IntExprT[int64] {
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
