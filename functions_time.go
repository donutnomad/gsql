package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
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
func UNIX_TIMESTAMP(date ...field.Expression) fields.IntExpr[int64] {
	if len(date) == 0 {
		return fields.NewIntExpr[int64](clause.Expr{SQL: "UNIX_TIMESTAMP()"})
	}
	return fields.NewIntExpr[int64](clause.Expr{
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
