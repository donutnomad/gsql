package gsql

import (
	"fmt"
	"strings"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
	"github.com/samber/lo"
)

// ==================== JSON 函数 ====================

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
func AsJson(expr Expression) fields.Json {
	return fields.AsJson(expr)
}

// JsonLit 创建一个 JSON 字面量表达式
// 用于将 JSON 字符串转换为 JsonExpr，以便传递给 JSON 函数
// 示例:
//
//	gsql.JsonLit(`{"name":"John","age":30}`)
//	gsql.JsonLit(`[1, 2, 3]`)
//
//goland:noinspection ALL
func JsonLit(jsonStr string) fields.Json {
	return fields.NewJson(clause.Expr{
		SQL:  "?",
		Vars: []any{jsonStr},
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
func JSON_OBJECT(pairs ...lo.Entry[string, Expression]) *jsonObjectBuilder {
	b := &jsonObjectBuilder{
		pairs: make([]lo.Entry[string, Expression], 0, len(pairs)),
	}
	b.pairs = append(b.pairs, pairs...)
	return b
}

type jsonObjectBuilder struct {
	pairs []lo.Entry[string, Expression]
}

func (j *jsonObjectBuilder) Add(key string, value Expression) *jsonObjectBuilder {
	j.pairs = append(j.pairs, lo.Entry[string, Expression]{Key: key, Value: value})
	return j
}

func (j *jsonObjectBuilder) toExpr() fields.Json {
	placeholders := make([]string, 0, len(j.pairs)*2)
	vars := make([]any, 0, len(j.pairs)*2)
	for _, pair := range j.pairs {
		placeholders = append(placeholders, "?", "?")
		vars = append(vars, pair.Key, pair.Value)
	}
	return fields.NewJson(clause.Expr{
		SQL:  fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(placeholders, ", ")),
		Vars: vars,
	})
}

func (j *jsonObjectBuilder) Build(builder clause.Builder) {
	j.toExpr().Build(builder)
}

func (j *jsonObjectBuilder) As(name string) field.IField {
	return fields.ScalarOf[any](j).As(name)
}

func (j *jsonObjectBuilder) ToExpr() Expression {
	return j.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (j *jsonObjectBuilder) ToJson() fields.Json {
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
func JSON_ARRAY(values ...Expression) *jsonArrayBuilder {
	return &jsonArrayBuilder{
		values: values,
	}
}

type jsonArrayBuilder struct {
	values []Expression
}

func (b *jsonArrayBuilder) Add(value Expression) *jsonArrayBuilder {
	b.values = append(b.values, value)
	return b
}

func (b *jsonArrayBuilder) toExpr() fields.Json {
	placeholders := make([]string, len(b.values))
	for i := range b.values {
		placeholders[i] = "?"
	}
	return fields.NewJson(clause.Expr{
		SQL:  fmt.Sprintf("JSON_ARRAY(%s)", strings.Join(placeholders, ", ")),
		Vars: lo.ToAnySlice(b.values),
	})
}

func (b *jsonArrayBuilder) Build(builder clause.Builder) {
	b.toExpr().Build(builder)
}

func (b *jsonArrayBuilder) As(name string) field.IField {
	return fields.ScalarOf[any](b).As(name)
}

func (b *jsonArrayBuilder) ToExpr() Expression {
	return b.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (b *jsonArrayBuilder) ToJson() fields.Json {
	return b.toExpr()
}

// ==================== JSON_UNQUOTE / JSON_QUOTE ====================

// JSON_QUOTE 为字符串添加引号，使其成为有效的 JSON 字符串值
// SELECT JSON_QUOTE('Hello World');
// SELECT JSON_QUOTE(users.name) FROM users;
//
//goland:noinspection ALL
func JSON_QUOTE(str Expression) fields.StringExpr[string] {
	return fields.StringOf[string](clause.Expr{
		SQL:  "JSON_QUOTE(?)",
		Vars: []any{str},
	})
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
func JSON_MERGE_PRESERVE(json1, json2 fields.JsonInput) *jsonMergeBuilder {
	return &jsonMergeBuilder{
		jsons:    []fields.JsonInput{json1, json2},
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
func JSON_MERGE_PATCH(json1, json2 fields.JsonInput) *jsonMergeBuilder {
	return &jsonMergeBuilder{
		jsons:    []fields.JsonInput{json1, json2},
		funcName: "JSON_MERGE_PATCH",
	}
}

type jsonMergeBuilder struct {
	jsons    []fields.JsonInput
	funcName string
}

func (b *jsonMergeBuilder) Merge(json fields.JsonInput) *jsonMergeBuilder {
	b.jsons = append(b.jsons, json)
	return b
}

func (b *jsonMergeBuilder) toExpr() fields.Json {
	placeholders := make([]string, len(b.jsons))
	for i := range b.jsons {
		placeholders[i] = "?"
	}
	return fields.NewJson(clause.Expr{
		SQL:  fmt.Sprintf("%s(%s)", b.funcName, strings.Join(placeholders, ", ")),
		Vars: lo.ToAnySlice(b.jsons),
	})
}

func (b *jsonMergeBuilder) Build(builder clause.Builder) {
	b.toExpr().Build(builder)
}

func (b *jsonMergeBuilder) As(name string) field.IField {
	return fields.ScalarOf[any](b).As(name)
}

func (b *jsonMergeBuilder) ToExpr() Expression {
	return b.toExpr()
}

// ToJson 返回 JsonExpr，用于类型安全的链式调用
func (b *jsonMergeBuilder) ToJson() fields.Json {
	return b.toExpr()
}
