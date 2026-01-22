package gsql

import (
	"fmt"
	"strings"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/samber/lo"
)

// ==================== JSON 函数 ====================

// 接口一致性检查 - 确保 builder 实现了所需接口
var (
	_ field.ExpressionTo = (*jsonObjectBuilder)(nil)
	_ field.ExpressionTo = (*jsonArrayBuilder)(nil)
	_ field.ExpressionTo = (*jsonSetBuilder)(nil)
	_ field.ExpressionTo = (*jsonRemoveBuilder)(nil)
	_ field.ExpressionTo = (*jsonArrayModifyBuilder)(nil)
	_ field.ExpressionTo = (*jsonMergeBuilder)(nil)
)

// PathValue 表示 JSON 路径和值的配对
type PathValue struct {
	Path  string
	Value any
}

// AsJson 将任意表达式包装为 JsonExpr
// 用于将字段或其他表达式转换为可以使用 JSON 方法的类型
// 示例:
//
//	gsql.AsJson(u.Profile).Extract("$.name")
//	gsql.AsJson(u.Profile).Length("$.skills")
//	gsql.AsJson(u.Profile).StorageSize()
//
//goland:noinspection ALL
func AsJson(expr field.Expression) field.JsonExpr {
	return field.AsJson(expr)
}

// JsonLit 创建一个 JSON 字面量表达式
// 用于将 JSON 字符串转换为 JsonExpr，以便传递给 JSON 函数
// 示例:
//
//	gsql.JsonLit(`{"name":"John","age":30}`)
//	gsql.JsonLit(`[1, 2, 3]`)
//
//goland:noinspection ALL
func JsonLit(jsonStr string) field.JsonExpr {
	return field.NewJsonExpr(clause.Expr{
		SQL:  "?",
		Vars: []any{jsonStr},
	})
}

// ==================== JSON_EXTRACT ====================

// JSON_EXTRACT 从 JSON 文档中提取数据，使用 JSON 路径表达式定位元素
// SELECT JSON_EXTRACT('{"name":"John","age":30}', '$.name');
// SELECT JSON_EXTRACT(data, '$.user.email') FROM profiles;
// 路径语法: $ 表示根, .key 访问对象键, [n] 访问数组索引, [*] 访问所有数组元素
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Extract("$.name", "$.age")
//
//goland:noinspection ALL
func JSON_EXTRACT(json field.JsonInput, paths ...string) field.JsonExpr {
	var vars = make([]any, 0, len(paths)+1)
	var placeholders = make([]string, 0, len(paths)+1)

	vars = append(vars, json)
	placeholders = append(placeholders, "?")
	for _, path := range paths {
		placeholders = append(placeholders, "?")
		vars = append(vars, path)
	}

	return field.NewJsonExpr(clause.Expr{
		SQL:  fmt.Sprintf("JSON_EXTRACT(%s)", strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

// ==================== JSON_OBJECT (Builder) ====================

// JSON_OBJECT 创建 JSON 对象
// 使用方式:
//
//	gsql.JSON_OBJECT().Add("name", name).Add("age", age)
//	gsql.JSON_OBJECT(lo.T2("name", name), lo.T2("age", age))
//
//goland:noinspection ALL
func JSON_OBJECT(pairs ...lo.Entry[string, field.Expression]) *jsonObjectBuilder {
	b := &jsonObjectBuilder{
		pairs: make([]lo.Entry[string, field.Expression], 0, len(pairs)),
	}
	b.pairs = append(b.pairs, pairs...)
	return b
}

type jsonObjectBuilder struct {
	pairs []lo.Entry[string, field.Expression]
}

func (j *jsonObjectBuilder) Add(key string, value field.Expression) *jsonObjectBuilder {
	j.pairs = append(j.pairs, lo.Entry[string, field.Expression]{Key: key, Value: value})
	return j
}

func (j *jsonObjectBuilder) toExpr() field.JsonExpr {
	placeholders := make([]string, 0, len(j.pairs)*2)
	vars := make([]any, 0, len(j.pairs)*2)
	for _, pair := range j.pairs {
		placeholders = append(placeholders, "?", "?")
		vars = append(vars, pair.Key, pair.Value)
	}
	return field.NewJsonExpr(clause.Expr{
		SQL:  fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

func (j *jsonObjectBuilder) Build(builder clause.Builder) {
	j.toExpr().Build(builder)
}

func (j *jsonObjectBuilder) AsF(name ...string) field.IField {
	var alias = ""
	if len(name) > 0 {
		alias = name[0]
	}
	return field.NewBaseFromSql(j.toExpr(), alias)
}

func (j *jsonObjectBuilder) ToExpr() field.Expression {
	return j.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (j *jsonObjectBuilder) ToJson() field.JsonExpr {
	return j.toExpr()
}

// ==================== JSON_ARRAY (Builder) ====================

// JSON_ARRAY 创建 JSON 数组，接受多个值作为数组元素
// 使用方式:
//
//	gsql.JSON_ARRAY().Add(val1).Add(val2)
//	gsql.JSON_ARRAY(val1, val2).Add(val3)
//
//goland:noinspection ALL
func JSON_ARRAY(values ...field.Expression) *jsonArrayBuilder {
	return &jsonArrayBuilder{
		values: values,
	}
}

type jsonArrayBuilder struct {
	values []field.Expression
}

func (b *jsonArrayBuilder) Add(value field.Expression) *jsonArrayBuilder {
	b.values = append(b.values, value)
	return b
}

func (b *jsonArrayBuilder) toExpr() field.JsonExpr {
	placeholders := make([]string, len(b.values))
	for i := range b.values {
		placeholders[i] = "?"
	}
	return field.NewJsonExpr(clause.Expr{
		SQL:  fmt.Sprintf("JSON_ARRAY(%s)", strings.Join(placeholders, ", ")),
		Vars: lo.ToAnySlice(b.values),
	})
}

func (b *jsonArrayBuilder) Build(builder clause.Builder) {
	b.toExpr().Build(builder)
}

func (b *jsonArrayBuilder) AsF(name ...string) field.IField {
	var alias = ""
	if len(name) > 0 {
		alias = name[0]
	}
	return field.NewBaseFromSql(b.toExpr(), alias)
}

func (b *jsonArrayBuilder) ToExpr() field.Expression {
	return b.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (b *jsonArrayBuilder) ToJson() field.JsonExpr {
	return b.toExpr()
}

// ==================== JSON_UNQUOTE / JSON_QUOTE ====================

// JSON_UNQUOTE 去除 JSON 值的引号，通常与 JSON_EXTRACT 配合使用
// SELECT JSON_UNQUOTE('"Hello World"');
// SELECT JSON_UNQUOTE(JSON_EXTRACT(data, '$.name')) FROM users;
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Extract("$.name").Unquote()
//
//goland:noinspection ALL
func JSON_UNQUOTE(json field.JsonInput) field.TextExpr[string] {
	return field.NewTextExpr[string](clause.Expr{
		SQL:  "JSON_UNQUOTE(?)",
		Vars: []any{json},
	})
}

// JSON_QUOTE 为字符串添加引号，使其成为有效的 JSON 字符串值
// SELECT JSON_QUOTE('Hello World');
// SELECT JSON_QUOTE(users.name) FROM users;
//
//goland:noinspection ALL
func JSON_QUOTE(str field.Expression) field.TextExpr[string] {
	return field.NewTextExpr[string](clause.Expr{
		SQL:  "JSON_QUOTE(?)",
		Vars: []any{str},
	})
}

// ==================== JSON_CONTAINS / JSON_CONTAINS_PATH ====================

// JSON_CONTAINS 检查 JSON 文档是否在指定路径包含候选值，可选路径参数
// SELECT JSON_CONTAINS('{"a":1,"b":2}', '1', '$.a');
// SELECT JSON_CONTAINS('[1,2,3]', '2');
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Contains(gsql.JsonLit(`"go"`), "$.skills")
//
//goland:noinspection ALL
func JSON_CONTAINS(target, candidate field.JsonInput, path ...string) field.IntExprT[int64] {
	if len(path) > 0 {
		return field.NewIntExprT[int64](clause.Expr{
			SQL:  "JSON_CONTAINS(?, ?, ?)",
			Vars: []any{target, candidate, path[0]},
		})
	}
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_CONTAINS(?, ?)",
		Vars: []any{target, candidate},
	})
}

// JSON_CONTAINS_PATH 检查 JSON 文档中是否存在指定路径，mode 可以是 'one' 或 'all'
// SELECT JSON_CONTAINS_PATH('{"a":1,"b":2}', 'one', '$.a', '$.c');
// SELECT JSON_CONTAINS_PATH('{"a":1,"b":2}', 'all', '$.a', '$.b');
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).ContainsPath("one", "$.name", "$.age")
//
//goland:noinspection ALL
func JSON_CONTAINS_PATH(json field.JsonInput, mode string, paths ...string) field.IntExprT[int64] {
	placeholders := make([]string, 0, len(paths)+2)
	vars := make([]any, 0, len(paths)+2)
	placeholders = append(placeholders, "?", "?")
	vars = append(vars, json, mode)
	for _, path := range paths {
		placeholders = append(placeholders, "?")
		vars = append(vars, path)
	}
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  fmt.Sprintf("JSON_CONTAINS_PATH(%s)", strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

// ==================== JSON_KEYS / JSON_LENGTH ====================

// JSON_KEYS 返回 JSON 对象的顶级键或指定路径的键，结果为 JSON 数组
// SELECT JSON_KEYS('{"a":1,"b":2}');
// SELECT JSON_KEYS('{"a":{"x":1,"y":2},"b":3}', '$.a');
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Keys()
//
//goland:noinspection ALL
func JSON_KEYS(json field.JsonInput, path ...string) field.JsonExpr {
	if len(path) > 0 {
		return field.NewJsonExpr(clause.Expr{
			SQL:  "JSON_KEYS(?, ?)",
			Vars: []any{json, path[0]},
		})
	}
	return field.NewJsonExpr(clause.Expr{
		SQL:  "JSON_KEYS(?)",
		Vars: []any{json},
	})
}

// JSON_LENGTH 返回 JSON 文档的长度（对象的键数量或数组的元素数量），可指定路径
// SELECT JSON_LENGTH('[1,2,3]');
// SELECT JSON_LENGTH('{"a":1,"b":2}');
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Length("$.skills")
//
//goland:noinspection ALL
func JSON_LENGTH(json field.JsonInput, path ...string) field.IntExprT[int64] {
	if len(path) > 0 {
		return field.NewIntExprT[int64](clause.Expr{
			SQL:  "JSON_LENGTH(?, ?)",
			Vars: []any{json, path[0]},
		})
	}
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_LENGTH(?)",
		Vars: []any{json},
	})
}

// ==================== JSON_SET (Builder) ====================

// JSON_SET 在 JSON 文档中设置值，路径存在则替换，不存在则插入
// 使用方式:
//
//	gsql.JSON_SET(data, "$.name", "John").Set("$.age", 18)
//
// 也可以使用方法调用（支持链式设置）:
//
//	gsql.AsJson(u.Profile).Set("$.name", "John").Set("$.age", 18)
//
//goland:noinspection ALL
func JSON_SET(json field.JsonInput, path string, value any) *jsonSetBuilder {
	return &jsonSetBuilder{
		json:     json,
		pairs:    []PathValue{{Path: path, Value: value}},
		funcName: "JSON_SET",
	}
}

type jsonSetBuilder struct {
	json     field.JsonInput
	pairs    []PathValue
	funcName string
}

func (b *jsonSetBuilder) Set(path string, value any) *jsonSetBuilder {
	b.pairs = append(b.pairs, PathValue{Path: path, Value: value})
	return b
}

func (b *jsonSetBuilder) toExpr() field.JsonExpr {
	placeholders := make([]string, 0, len(b.pairs)*2+1)
	vars := make([]any, 0, len(b.pairs)*2+1)
	placeholders = append(placeholders, "?")
	vars = append(vars, b.json)
	for _, pv := range b.pairs {
		placeholders = append(placeholders, "?", "?")
		vars = append(vars, pv.Path, pv.Value)
	}
	return field.NewJsonExpr(clause.Expr{
		SQL:  fmt.Sprintf("%s(%s)", b.funcName, strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

func (b *jsonSetBuilder) Build(builder clause.Builder) {
	b.toExpr().Build(builder)
}

func (b *jsonSetBuilder) AsF(name ...string) field.IField {
	var alias = ""
	if len(name) > 0 {
		alias = name[0]
	}
	return field.NewBaseFromSql(b.toExpr(), alias)
}

func (b *jsonSetBuilder) ToExpr() field.Expression {
	return b.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (b *jsonSetBuilder) ToJson() field.JsonExpr {
	return b.toExpr()
}

// ==================== JSON_INSERT (Builder) ====================

// JSON_INSERT 在 JSON 文档中插入值，仅当路径不存在时插入
// 使用方式:
//
//	gsql.JSON_INSERT(data, "$.created_at", now).Set("$.views", 0)
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Insert("$.created_at", now)
//
//goland:noinspection ALL
func JSON_INSERT(json field.JsonInput, path string, value any) *jsonSetBuilder {
	return &jsonSetBuilder{
		json:     json,
		pairs:    []PathValue{{Path: path, Value: value}},
		funcName: "JSON_INSERT",
	}
}

// ==================== JSON_REPLACE (Builder) ====================

// JSON_REPLACE 在 JSON 文档中替换值，仅当路径存在时替换
// 使用方式:
//
//	gsql.JSON_REPLACE(data, "$.status", "inactive").Set("$.updated_at", now)
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Replace("$.status", "inactive")
//
//goland:noinspection ALL
func JSON_REPLACE(json field.JsonInput, path string, value any) *jsonSetBuilder {
	return &jsonSetBuilder{
		json:     json,
		pairs:    []PathValue{{Path: path, Value: value}},
		funcName: "JSON_REPLACE",
	}
}

// ==================== JSON_REMOVE (Builder) ====================

// JSON_REMOVE 从 JSON 文档中移除指定路径的元素
// 使用方式:
//
//	gsql.JSON_REMOVE(data, "$.temp").Remove("$.old").Remove("$.deprecated")
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Remove("$.temp", "$.old")
//
//goland:noinspection ALL
func JSON_REMOVE(json field.JsonInput, path string) *jsonRemoveBuilder {
	return &jsonRemoveBuilder{
		json:  json,
		paths: []string{path},
	}
}

type jsonRemoveBuilder struct {
	json  field.JsonInput
	paths []string
}

func (b *jsonRemoveBuilder) Remove(path string) *jsonRemoveBuilder {
	b.paths = append(b.paths, path)
	return b
}

func (b *jsonRemoveBuilder) toExpr() field.JsonExpr {
	placeholders := make([]string, 0, len(b.paths)+1)
	vars := make([]any, 0, len(b.paths)+1)
	placeholders = append(placeholders, "?")
	vars = append(vars, b.json)
	for _, p := range b.paths {
		placeholders = append(placeholders, "?")
		vars = append(vars, p)
	}
	return field.NewJsonExpr(clause.Expr{
		SQL:  fmt.Sprintf("JSON_REMOVE(%s)", strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

func (b *jsonRemoveBuilder) Build(builder clause.Builder) {
	b.toExpr().Build(builder)
}

func (b *jsonRemoveBuilder) AsF(name ...string) field.IField {
	var alias = ""
	if len(name) > 0 {
		alias = name[0]
	}
	return field.NewBaseFromSql(b.toExpr(), alias)
}

func (b *jsonRemoveBuilder) ToExpr() field.Expression {
	return b.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (b *jsonRemoveBuilder) ToJson() field.JsonExpr {
	return b.toExpr()
}

// ==================== JSON_ARRAY_APPEND (Builder) ====================

// JSON_ARRAY_APPEND 向 JSON 数组的指定路径追加值
// 使用方式:
//
//	gsql.JSON_ARRAY_APPEND(tags, "$", "new_tag").Append("$.items", item)
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Tags).ArrayAppend("$", "new_tag")
//
//goland:noinspection ALL
func JSON_ARRAY_APPEND(json field.JsonInput, path string, value any) *jsonArrayModifyBuilder {
	return &jsonArrayModifyBuilder{
		json:     json,
		pairs:    []PathValue{{Path: path, Value: value}},
		funcName: "JSON_ARRAY_APPEND",
	}
}

type jsonArrayModifyBuilder struct {
	json     field.JsonInput
	pairs    []PathValue
	funcName string
}

func (b *jsonArrayModifyBuilder) Append(path string, value any) *jsonArrayModifyBuilder {
	b.pairs = append(b.pairs, PathValue{Path: path, Value: value})
	return b
}

// Insert 为 JSON_ARRAY_INSERT 添加额外的 path-value 对
func (b *jsonArrayModifyBuilder) Insert(path string, value any) *jsonArrayModifyBuilder {
	b.pairs = append(b.pairs, PathValue{Path: path, Value: value})
	return b
}

func (b *jsonArrayModifyBuilder) toExpr() field.JsonExpr {
	placeholders := make([]string, 0, len(b.pairs)*2+1)
	vars := make([]any, 0, len(b.pairs)*2+1)
	placeholders = append(placeholders, "?")
	vars = append(vars, b.json)
	for _, pv := range b.pairs {
		placeholders = append(placeholders, "?", "?")
		vars = append(vars, pv.Path, pv.Value)
	}
	return field.NewJsonExpr(clause.Expr{
		SQL:  fmt.Sprintf("%s(%s)", b.funcName, strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

func (b *jsonArrayModifyBuilder) Build(builder clause.Builder) {
	b.toExpr().Build(builder)
}

func (b *jsonArrayModifyBuilder) AsF(name ...string) field.IField {
	var alias = ""
	if len(name) > 0 {
		alias = name[0]
	}
	return field.NewBaseFromSql(b.toExpr(), alias)
}

func (b *jsonArrayModifyBuilder) ToExpr() field.Expression {
	return b.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (b *jsonArrayModifyBuilder) ToJson() field.JsonExpr {
	return b.toExpr()
}

// ==================== JSON_ARRAY_INSERT (Builder) ====================

// JSON_ARRAY_INSERT 向 JSON 数组的指定位置插入值
// 使用方式:
//
//	gsql.JSON_ARRAY_INSERT(images, "$[0]", "cover.jpg").Insert("$[1]", "image.jpg")
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Images).ArrayInsert("$[0]", "cover.jpg")
//
//goland:noinspection ALL
func JSON_ARRAY_INSERT(json field.JsonInput, path string, value any) *jsonArrayModifyBuilder {
	return &jsonArrayModifyBuilder{
		json:     json,
		pairs:    []PathValue{{Path: path, Value: value}},
		funcName: "JSON_ARRAY_INSERT",
	}
}

// ==================== JSON_MERGE_PRESERVE / JSON_MERGE_PATCH (Builder) ====================

// JSON_MERGE_PRESERVE 合并多个 JSON 文档，保留所有重复的键
// 使用方式:
//
//	gsql.JSON_MERGE_PRESERVE(json1, json2).Merge(json3).Merge(json4)
//
// 也可以使用方法调用:
//
//	gsql.AsJson(json1).MergePreserve(json2, json3)
//
//goland:noinspection ALL
func JSON_MERGE_PRESERVE(json1, json2 field.JsonInput) *jsonMergeBuilder {
	return &jsonMergeBuilder{
		jsons:    []field.JsonInput{json1, json2},
		funcName: "JSON_MERGE_PRESERVE",
	}
}

// JSON_MERGE_PATCH 使用 RFC 7396 语义合并 JSON 文档，后面的文档会覆盖前面的键
// 使用方式:
//
//	gsql.JSON_MERGE_PATCH(json1, json2).Merge(json3)
//
// 也可以使用方法调用:
//
//	gsql.AsJson(json1).MergePatch(json2, json3)
//
//goland:noinspection ALL
func JSON_MERGE_PATCH(json1, json2 field.JsonInput) *jsonMergeBuilder {
	return &jsonMergeBuilder{
		jsons:    []field.JsonInput{json1, json2},
		funcName: "JSON_MERGE_PATCH",
	}
}

type jsonMergeBuilder struct {
	jsons    []field.JsonInput
	funcName string
}

func (b *jsonMergeBuilder) Merge(json field.JsonInput) *jsonMergeBuilder {
	b.jsons = append(b.jsons, json)
	return b
}

func (b *jsonMergeBuilder) toExpr() field.JsonExpr {
	placeholders := make([]string, len(b.jsons))
	for i := range b.jsons {
		placeholders[i] = "?"
	}
	return field.NewJsonExpr(clause.Expr{
		SQL:  fmt.Sprintf("%s(%s)", b.funcName, strings.Join(placeholders, ", ")),
		Vars: lo.ToAnySlice(b.jsons),
	})
}

func (b *jsonMergeBuilder) Build(builder clause.Builder) {
	b.toExpr().Build(builder)
}

func (b *jsonMergeBuilder) AsF(name ...string) field.IField {
	var alias = ""
	if len(name) > 0 {
		alias = name[0]
	}
	return field.NewBaseFromSql(b.toExpr(), alias)
}

func (b *jsonMergeBuilder) ToExpr() field.Expression {
	return b.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (b *jsonMergeBuilder) ToJson() field.JsonExpr {
	return b.toExpr()
}

// ==================== JSON_VALID / JSON_TYPE / JSON_DEPTH ====================

// JSON_VALID 检查值是否为有效的 JSON 文档，返回 1 表示有效，0 表示无效
// SELECT JSON_VALID('{"a":1}');
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Valid()
//
//goland:noinspection ALL
func JSON_VALID(json field.JsonInput) field.IntExprT[int64] {
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_VALID(?)",
		Vars: []any{json},
	})
}

// JSON_TYPE 返回 JSON 值的类型（OBJECT, ARRAY, INTEGER, DOUBLE, STRING, BOOLEAN, NULL）
// SELECT JSON_TYPE('{"a":1}');
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Type()
//
//goland:noinspection ALL
func JSON_TYPE(json field.JsonInput) field.TextExpr[string] {
	return field.NewTextExpr[string](clause.Expr{
		SQL:  "JSON_TYPE(?)",
		Vars: []any{json},
	})
}

// JSON_DEPTH 返回 JSON 文档的最大深度，空数组/对象或标量值的深度为 1
// SELECT JSON_DEPTH('{"a":{"b":{"c":1}}}');
//
// 也可以使用方法调用:
//
//	gsql.AsJson(u.Profile).Depth()
//
//goland:noinspection ALL
func JSON_DEPTH(json field.JsonInput) field.IntExprT[int64] {
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_DEPTH(?)",
		Vars: []any{json},
	})
}

// ==================== JSON_PRETTY / JSON_SEARCH ====================

// JSON_PRETTY 以易读的格式打印 JSON 文档（带缩进和换行）
// SELECT JSON_PRETTY('{"a":1,"b":2}');
// SELECT JSON_PRETTY(data) FROM users LIMIT 1;
//
//goland:noinspection ALL
func JSON_PRETTY(json field.JsonInput) field.TextExpr[string] {
	return field.NewTextExpr[string](clause.Expr{
		SQL:  "JSON_PRETTY(?)",
		Vars: []any{json},
	})
}

// JSON_SEARCH 在 JSON 文档中搜索字符串值，返回匹配路径
// mode: 'one' 返回第一个匹配，'all' 返回所有匹配
// SELECT JSON_SEARCH('{"a":"abc","b":"def"}', 'one', 'abc');
// SELECT JSON_SEARCH('["abc","def","abc"]', 'all', 'abc');
//
//goland:noinspection ALL
func JSON_SEARCH(json field.JsonInput, mode string, searchStr any, escapePath ...any) field.TextExpr[string] {
	placeholders := []string{"?", "?", "?"}
	vars := []any{json, mode, searchStr}
	for _, ep := range escapePath {
		placeholders = append(placeholders, "?")
		vars = append(vars, ep)
	}
	return field.NewTextExpr[string](clause.Expr{
		SQL:  fmt.Sprintf("JSON_SEARCH(%s)", strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

// ==================== JSON_STORAGE_SIZE / JSON_STORAGE_FREE ====================

// JSON_STORAGE_SIZE 返回存储 JSON 文档所需的字节数（MySQL 5.7.22+）
// SELECT JSON_STORAGE_SIZE('{"a":1}');
// SELECT JSON_STORAGE_SIZE('[1,2,3,4,5]');
//
//goland:noinspection ALL
func JSON_STORAGE_SIZE(json field.JsonInput) field.IntExprT[int64] {
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_STORAGE_SIZE(?)",
		Vars: []any{json},
	})
}

// JSON_STORAGE_FREE 返回 JSON 列值的二进制表示中部分更新后释放的空间（MySQL 8.0.13+）
// SELECT JSON_STORAGE_FREE(data) FROM users;
// SELECT id, JSON_STORAGE_FREE(metadata) as free_space FROM products;
//
//goland:noinspection ALL
func JSON_STORAGE_FREE(json field.JsonInput) field.IntExprT[int64] {
	return field.NewIntExprT[int64](clause.Expr{
		SQL:  "JSON_STORAGE_FREE(?)",
		Vars: []any{json},
	})
}
