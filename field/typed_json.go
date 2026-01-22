package field

import (
	"strings"

	"github.com/donutnomad/gsql/clause"
)

// ==================== JSON 类型系统 ====================

// JsonInput JSON 输入接口，用于约束 JSON 函数的输入参数
// 实现此接口的类型可以作为 JSON 函数的输入
type JsonInput interface {
	Expression
	jsonInput() // 标记方法，用于类型约束
}

// JsonExpr JSON 类型表达式，用于 JSON 字段和 JSON 函数的返回值
// 支持作为 JSON 函数的输入参数，提供类型安全
// 支持链式调用 JSON 操作方法
type JsonExpr struct {
	baseComparableImpl[string] // 复用比较操作（JSON 值也可以比较）
}

// NewJsonExpr 创建一个新的 JsonExpr 实例
func NewJsonExpr(expr clause.Expression) JsonExpr {
	return JsonExpr{baseComparableImpl[string]{Expression: expr}}
}

// jsonInput 实现 JsonInput 接口
func (e JsonExpr) jsonInput() {}

// ==================== JsonExpr 方法 ====================

// Extract 从 JSON 文档中提取数据 (JSON_EXTRACT)
// SELECT JSON_EXTRACT('{"name":"John","age":30}', '$.name');
// SELECT JSON_EXTRACT(data, '$.user.email') FROM profiles;
// 示例: gsql.AsJson(u.Profile).Extract("$.name", "$.age")
// 独立函数: gsql.JSON_EXTRACT(gsql.AsJson(u.Profile), "$.name", "$.age")
func (e JsonExpr) Extract(paths ...string) JsonExpr {
	vars := make([]any, 0, len(paths)+1)
	placeholders := make([]string, 0, len(paths)+1)
	vars = append(vars, e.Expression)
	placeholders = append(placeholders, "?")
	for _, path := range paths {
		placeholders = append(placeholders, "?")
		vars = append(vars, path)
	}
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_EXTRACT(" + strings.Join(placeholders, ", ") + ")",
		Vars: vars,
	})
}

// Unquote 去除 JSON 值的引号 (JSON_UNQUOTE)
// SELECT JSON_UNQUOTE('"Hello World"');
// SELECT JSON_UNQUOTE(JSON_EXTRACT(data, '$.name')) FROM users;
// 示例: gsql.AsJson(u.Profile).Extract("$.name").Unquote()
// 独立函数: gsql.JSON_UNQUOTE(gsql.JSON_EXTRACT(gsql.AsJson(u.Profile), "$.name"))
func (e JsonExpr) Unquote() TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "JSON_UNQUOTE(?)",
		Vars: []any{e.Expression},
	})
}

// Keys 返回 JSON 对象的键 (JSON_KEYS)
// SELECT JSON_KEYS('{"a":1,"b":2}');
// SELECT JSON_KEYS('{"a":{"x":1,"y":2},"b":3}', '$.a');
// 示例: gsql.AsJson(u.Profile).Keys()
// 独立函数: gsql.JSON_KEYS(gsql.AsJson(u.Profile))
func (e JsonExpr) Keys(path ...string) JsonExpr {
	if len(path) > 0 {
		return NewJsonExpr(clause.Expr{
			SQL:  "JSON_KEYS(?, ?)",
			Vars: []any{e.Expression, path[0]},
		})
	}
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_KEYS(?)",
		Vars: []any{e.Expression},
	})
}

// Length 返回 JSON 文档的长度 (JSON_LENGTH)
// SELECT JSON_LENGTH('[1,2,3]');
// SELECT JSON_LENGTH('{"a":1,"b":2}');
// 示例: gsql.AsJson(u.Profile).Length("$.skills")
// 独立函数: gsql.JSON_LENGTH(gsql.AsJson(u.Profile), "$.skills")
func (e JsonExpr) Length(path ...string) IntExprT[int64] {
	if len(path) > 0 {
		return NewIntExprT[int64](clause.Expr{
			SQL:  "JSON_LENGTH(?, ?)",
			Vars: []any{e.Expression, path[0]},
		})
	}
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_LENGTH(?)",
		Vars: []any{e.Expression},
	})
}

// Contains 检查是否包含指定值 (JSON_CONTAINS)
// SELECT JSON_CONTAINS('{"a":1,"b":2}', '1', '$.a');
// SELECT JSON_CONTAINS('[1,2,3]', '2');
// 示例: gsql.AsJson(u.Profile).Contains(gsql.JsonLit(`"go"`), "$.skills")
// 独立函数: gsql.JSON_CONTAINS(gsql.AsJson(u.Profile), gsql.JsonLit(`"go"`), "$.skills")
func (e JsonExpr) Contains(candidate JsonInput, path ...string) IntExprT[int64] {
	if len(path) > 0 {
		return NewIntExprT[int64](clause.Expr{
			SQL:  "JSON_CONTAINS(?, ?, ?)",
			Vars: []any{e.Expression, candidate, path[0]},
		})
	}
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_CONTAINS(?, ?)",
		Vars: []any{e.Expression, candidate},
	})
}

// ContainsPath 检查路径是否存在 (JSON_CONTAINS_PATH)
// SELECT JSON_CONTAINS_PATH('{"a":1,"b":2}', 'one', '$.a', '$.c');
// SELECT JSON_CONTAINS_PATH('{"a":1,"b":2}', 'all', '$.a', '$.b');
// mode: 'one' 或 'all'
// 示例: gsql.AsJson(u.Profile).ContainsPath("one", "$.name", "$.age")
// 独立函数: gsql.JSON_CONTAINS_PATH(gsql.AsJson(u.Profile), "one", "$.name", "$.age")
func (e JsonExpr) ContainsPath(mode string, paths ...string) IntExprT[int64] {
	vars := make([]any, 0, len(paths)+2)
	placeholders := make([]string, 0, len(paths)+2)
	vars = append(vars, e.Expression, mode)
	placeholders = append(placeholders, "?", "?")
	for _, path := range paths {
		placeholders = append(placeholders, "?")
		vars = append(vars, path)
	}
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_CONTAINS_PATH(" + strings.Join(placeholders, ", ") + ")",
		Vars: vars,
	})
}

// Type 返回 JSON 值的类型 (JSON_TYPE)
// SELECT JSON_TYPE('{"a":1}');
// SELECT JSON_TYPE('[1,2,3]');
// 返回: OBJECT, ARRAY, INTEGER, DOUBLE, STRING, BOOLEAN, NULL
// 示例: gsql.AsJson(u.Profile).Type()
// 独立函数: gsql.JSON_TYPE(gsql.AsJson(u.Profile))
func (e JsonExpr) Type() TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "JSON_TYPE(?)",
		Vars: []any{e.Expression},
	})
}

// Depth 返回 JSON 文档的最大深度 (JSON_DEPTH)
// SELECT JSON_DEPTH('{"a":{"b":{"c":1}}}');
// SELECT JSON_DEPTH('[1,[2,[3]]]');
// 示例: gsql.AsJson(u.Profile).Depth()
// 独立函数: gsql.JSON_DEPTH(gsql.AsJson(u.Profile))
func (e JsonExpr) Depth() IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_DEPTH(?)",
		Vars: []any{e.Expression},
	})
}

// Valid 检查是否为有效 JSON (JSON_VALID)
// SELECT JSON_VALID('{"a":1}');
// SELECT JSON_VALID('invalid json');
// 示例: gsql.AsJson(u.Profile).Valid()
// 独立函数: gsql.JSON_VALID(gsql.AsJson(u.Profile))
func (e JsonExpr) Valid() IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_VALID(?)",
		Vars: []any{e.Expression},
	})
}

// Pretty 格式化输出 JSON (JSON_PRETTY)
// SELECT JSON_PRETTY('{"a":1,"b":2}');
// 示例: gsql.AsJson(u.Profile).Pretty()
// 独立函数: gsql.JSON_PRETTY(gsql.AsJson(u.Profile))
func (e JsonExpr) Pretty() TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "JSON_PRETTY(?)",
		Vars: []any{e.Expression},
	})
}

// StorageSize 返回存储 JSON 所需的字节数 (JSON_STORAGE_SIZE)
// SELECT JSON_STORAGE_SIZE('{"a":1}');
// SELECT JSON_STORAGE_SIZE('[1,2,3,4,5]');
// 示例: gsql.AsJson(u.Profile).StorageSize()
// 独立函数: gsql.JSON_STORAGE_SIZE(gsql.AsJson(u.Profile))
func (e JsonExpr) StorageSize() IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_STORAGE_SIZE(?)",
		Vars: []any{e.Expression},
	})
}

// StorageFree 返回部分更新后释放的空间 (JSON_STORAGE_FREE)
// SELECT JSON_STORAGE_FREE(data) FROM users;
// 示例: gsql.AsJson(u.Profile).StorageFree()
// 独立函数: gsql.JSON_STORAGE_FREE(gsql.AsJson(u.Profile))
func (e JsonExpr) StorageFree() IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_STORAGE_FREE(?)",
		Vars: []any{e.Expression},
	})
}

// Search 在 JSON 中搜索字符串值 (JSON_SEARCH)
// SELECT JSON_SEARCH('{"a":"abc","b":"def"}', 'one', 'abc');
// SELECT JSON_SEARCH('["abc","def","abc"]', 'all', 'abc');
// mode: 'one' 或 'all'
// 示例: gsql.AsJson(u.Profile).Search("one", "abc")
// 独立函数: gsql.JSON_SEARCH(gsql.AsJson(u.Profile), "one", "abc")
func (e JsonExpr) Search(mode string, searchStr any, escapePath ...any) TextExpr[string] {
	vars := []any{e.Expression, mode, searchStr}
	placeholders := []string{"?", "?", "?"}
	for _, ep := range escapePath {
		placeholders = append(placeholders, "?")
		vars = append(vars, ep)
	}
	return NewTextExpr[string](clause.Expr{
		SQL:  "JSON_SEARCH(" + strings.Join(placeholders, ", ") + ")",
		Vars: vars,
	})
}

// Set 设置 JSON 值 (JSON_SET) - 返回 JsonExpr，支持链式设置
// SELECT JSON_SET('{"a":1}', '$.b', 2);
// 示例: gsql.AsJson(u.Profile).Set("$.name", "John").Set("$.age", 18)
// 独立函数: gsql.JSON_SET(gsql.AsJson(u.Profile), "$.name", "John").Set("$.age", 18)
func (e JsonExpr) Set(path string, value any) JsonExpr {
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_SET(?, ?, ?)",
		Vars: []any{e.Expression, path, value},
	})
}

// Insert 插入 JSON 值，仅当路径不存在时 (JSON_INSERT)
// SELECT JSON_INSERT('{"a":1}', '$.b', 2);
// 示例: gsql.AsJson(u.Profile).Insert("$.created_at", now)
// 独立函数: gsql.JSON_INSERT(gsql.AsJson(u.Profile), "$.created_at", now)
func (e JsonExpr) Insert(path string, value any) JsonExpr {
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_INSERT(?, ?, ?)",
		Vars: []any{e.Expression, path, value},
	})
}

// Replace 替换 JSON 值，仅当路径存在时 (JSON_REPLACE)
// SELECT JSON_REPLACE('{"a":1}', '$.a', 2);
// 示例: gsql.AsJson(u.Profile).Replace("$.status", "inactive")
// 独立函数: gsql.JSON_REPLACE(gsql.AsJson(u.Profile), "$.status", "inactive")
func (e JsonExpr) Replace(path string, value any) JsonExpr {
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_REPLACE(?, ?, ?)",
		Vars: []any{e.Expression, path, value},
	})
}

// Remove 移除 JSON 元素 (JSON_REMOVE)
// SELECT JSON_REMOVE('{"a":1,"b":2}', '$.a');
// 示例: gsql.AsJson(u.Profile).Remove("$.temp", "$.old")
// 独立函数: gsql.JSON_REMOVE(gsql.AsJson(u.Profile), "$.temp").Remove("$.old")
func (e JsonExpr) Remove(paths ...string) JsonExpr {
	vars := make([]any, 0, len(paths)+1)
	placeholders := make([]string, 0, len(paths)+1)
	vars = append(vars, e.Expression)
	placeholders = append(placeholders, "?")
	for _, path := range paths {
		placeholders = append(placeholders, "?")
		vars = append(vars, path)
	}
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_REMOVE(" + strings.Join(placeholders, ", ") + ")",
		Vars: vars,
	})
}

// ArrayAppend 向 JSON 数组追加值 (JSON_ARRAY_APPEND)
// SELECT JSON_ARRAY_APPEND('[1,2]', '$', 3);
// 示例: gsql.AsJson(u.Tags).ArrayAppend("$", "new_tag")
// 独立函数: gsql.JSON_ARRAY_APPEND(gsql.AsJson(u.Tags), "$", "new_tag")
func (e JsonExpr) ArrayAppend(path string, value any) JsonExpr {
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_ARRAY_APPEND(?, ?, ?)",
		Vars: []any{e.Expression, path, value},
	})
}

// ArrayInsert 向 JSON 数组插入值 (JSON_ARRAY_INSERT)
// SELECT JSON_ARRAY_INSERT('[1,2]', '$[0]', 0);
// 示例: gsql.AsJson(u.Images).ArrayInsert("$[0]", "cover.jpg")
// 独立函数: gsql.JSON_ARRAY_INSERT(gsql.AsJson(u.Images), "$[0]", "cover.jpg")
func (e JsonExpr) ArrayInsert(path string, value any) JsonExpr {
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_ARRAY_INSERT(?, ?, ?)",
		Vars: []any{e.Expression, path, value},
	})
}

// MergePreserve 合并 JSON，保留重复键 (JSON_MERGE_PRESERVE)
// SELECT JSON_MERGE_PRESERVE('{"a":1}', '{"b":2}');
// 示例: gsql.AsJson(json1).MergePreserve(json2, json3)
// 独立函数: gsql.JSON_MERGE_PRESERVE(json1, json2).Merge(json3)
func (e JsonExpr) MergePreserve(others ...JsonInput) JsonExpr {
	vars := make([]any, 0, len(others)+1)
	placeholders := make([]string, 0, len(others)+1)
	vars = append(vars, e.Expression)
	placeholders = append(placeholders, "?")
	for _, other := range others {
		placeholders = append(placeholders, "?")
		vars = append(vars, other)
	}
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_MERGE_PRESERVE(" + strings.Join(placeholders, ", ") + ")",
		Vars: vars,
	})
}

// MergePatch 合并 JSON，覆盖重复键 (JSON_MERGE_PATCH)
// SELECT JSON_MERGE_PATCH('{"a":1}', '{"a":2}');
// 示例: gsql.AsJson(json1).MergePatch(json2, json3)
// 独立函数: gsql.JSON_MERGE_PATCH(json1, json2).Merge(json3)
func (e JsonExpr) MergePatch(others ...JsonInput) JsonExpr {
	vars := make([]any, 0, len(others)+1)
	placeholders := make([]string, 0, len(others)+1)
	vars = append(vars, e.Expression)
	placeholders = append(placeholders, "?")
	for _, other := range others {
		placeholders = append(placeholders, "?")
		vars = append(vars, other)
	}
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_MERGE_PATCH(" + strings.Join(placeholders, ", ") + ")",
		Vars: vars,
	})
}

// Overlaps 检查两个 JSON 是否有重叠元素 (JSON_OVERLAPS, MySQL 8.0.17+)
// SELECT JSON_OVERLAPS('[1,3,5]', '[2,3,4]');
// SELECT JSON_OVERLAPS('{"a":1}', '{"a":1,"b":2}');
// 示例: gsql.AsJson(u.Tags).Overlaps(gsql.JsonLit(`["go","python"]`))
func (e JsonExpr) Overlaps(other JsonInput) IntExprT[int64] {
	return NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_OVERLAPS(?, ?)",
		Vars: []any{e.Expression, other},
	})
}

// Value 从 JSON 文档中提取标量值 (JSON_VALUE, MySQL 8.0.21+)
// SELECT JSON_VALUE('{"name":"John"}', '$.name');
// SELECT JSON_VALUE('{"age":30}', '$.age' RETURNING SIGNED);
// 示例: gsql.AsJson(u.Profile).Value("$.name")
// 注意: 比 Extract + Unquote 更简洁，直接返回标量值
func (e JsonExpr) Value(path string) TextExpr[string] {
	return NewTextExpr[string](clause.Expr{
		SQL:  "JSON_VALUE(?, ?)",
		Vars: []any{e.Expression, path},
	})
}

func (e JsonExpr) As(alias string) IField {
	return NewBaseFromSql(e.Expression, alias)
}

// ==================== JSON 聚合函数 ====================

// JsonArrayAgg 将列值聚合为 JSON 数组 (JSON_ARRAYAGG, MySQL 8.0+)
// SELECT JSON_ARRAYAGG(name) FROM users;
// SELECT department, JSON_ARRAYAGG(name) FROM users GROUP BY department;
// 示例: field.JsonArrayAgg(u.Name)
func JsonArrayAgg(expr Expression) JsonExpr {
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_ARRAYAGG(?)",
		Vars: []any{expr},
	})
}

// JsonObjectAgg 将键值对聚合为 JSON 对象 (JSON_OBJECTAGG, MySQL 8.0+)
// SELECT JSON_OBJECTAGG(name, age) FROM users;
// SELECT department, JSON_OBJECTAGG(name, salary) FROM users GROUP BY department;
// 示例: field.JsonObjectAgg(u.Name, u.Age)
//  1. 键必须唯一 - 如果同一组内有重复的键，后面的值会覆盖前面的
//  2. 键必须是字符串 - MySQL 会自动将非字符串键转换为字符串
//  3. NULL 值 - 如果键为 NULL，该行会被忽略
func JsonObjectAgg(key, value Expression) JsonExpr {
	return NewJsonExpr(clause.Expr{
		SQL:  "JSON_OBJECTAGG(?, ?)",
		Vars: []any{key, value},
	})
}

// ==================== JSON 辅助函数 ====================

// AsJson 将任意表达式包装为 JsonExpr
// 用于将字段或其他表达式转换为可以使用 JSON 方法的类型
// 示例:
//
//	gsql.AsJson(u.Profile).Extract("$.name")
//	gsql.AsJson(u.Profile).Length("$.skills")
func AsJson(expr Expression) JsonExpr {
	if je, ok := expr.(JsonExpr); ok {
		return je
	}
	return NewJsonExpr(expr)
}
