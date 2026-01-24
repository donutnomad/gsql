package gsql

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"regexp"
	"strings"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/samber/lo"
)

// Deprecated: 使用 TN 替代。
func TableName(name string) Table {
	return Table{Name: name}
}

// Deprecated: 使用 TN 替代。
func TableName2(name string) Table2 {
	return TN(name)
}

// TN TableName
func TN(tableName string) Table2 {
	return Table2{Name: tableName}
}

func MapPattern[OUT any, IN any](input field.Pattern[IN]) field.Pattern[OUT] {
	return field.NewPatternFrom[OUT](input.Base)
}

func MapComparable[OUT any, IN any](input field.Comparable[IN]) field.Comparable[OUT] {
	return field.NewComparableFrom[OUT](input.Base)
}

func Field(sql string, args ...any) field.IField {
	return field.NewBaseFromSql(Expr(sql, args...), "")
}

func FieldExpr(expr field.Expression, alias string) field.IField {
	return field.NewBaseFromSql(expr, alias)
}

func Expr(sql string, args ...any) clause.Expression {
	return clause.Expr{SQL: sql, Vars: args}
}

func DefineTempTable[Model any, ModelT any](types ModelT, builder *QueryBuilder) templateTable[ModelT, Model] {
	return DefineTable[Model, ModelT](fmt.Sprintf("%s%d", "temp_", rand.N(32)), types, builder)
}

func DefineTempTableAny[ModelT any](types ModelT, builder *QueryBuilder) templateTable[ModelT, any] {
	return DefineTable[any, ModelT](fmt.Sprintf("%s%d", "temp_", rand.N(32)), types, builder)
}

func DefineTable[Model any, ModelT any](tableName string, types ModelT, builder field.IToExpr) templateTable[ModelT, Model] {
	if v, ok := builder.(*QueryBuilder); ok {
		if len(v.selects) == 0 {
			b := v.Clone()
			b.selects = append(b.selects, Star)
			builder = b
		}
	}

	var newTable = reflect.ValueOf(tableNameFn(tableName))
	var ty = &types

	rv := reflect.ValueOf(ty)
	if rv.Kind() != reflect.Ptr {
		panic("input must be a pointer")
	}
	if rv.IsNil() {
		panic("input pointer is nil")
	}

	// 解引用获取指针指向的结构体
	rv = rv.Elem()

	if rv.Kind() != reflect.Struct {
		panic("input must be pointer to struct")
	}

	for i := 0; i < rv.NumField(); i++ {
		fieldValue := rv.Field(i)
		fieldType := rv.Type().Field(i)

		if !fieldType.IsExported() || !fieldValue.CanSet() {
			continue
		}

		fieldV2 := fieldValue.Addr().Interface()
		if v, ok := fieldV2.(interface{ WithTable(tableName string) }); ok {
			v.WithTable(tableName)
		} else {
			withTableMethod := fieldValue.MethodByName("WithTable")
			if !withTableMethod.IsValid() {
				continue
			}
			results := withTableMethod.Call([]reflect.Value{newTable})
			if len(results) > 0 {
				fieldValue.Set(results[0])
			}
		}
	}

	return templateTable[ModelT, Model]{
		Fields:    *ty,
		tableName: tableName,
		expr:      builder.ToExpr(),
	}
}

var (
	createClauses = []string{"INSERT", "VALUES", "ON CONFLICT"}
	queryClauses  = []string{"CTE", "SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT", "FOR"}
	updateClauses = []string{"UPDATE", "SET", "WHERE"}
	deleteClauses = []string{"DELETE", "FROM", "WHERE"}
)

var tableRegexp = regexp.MustCompile(`(?i)(?:.+? AS (\w+)\s*(?:$|,)|^\w+\s+(\w+)$)`)

func txTable(quote func(field string) string, name string, args ...any) (expr *clause.RawExpr, table string) {
	if strings.Contains(name, " ") || strings.Contains(name, "`") || len(args) > 0 {
		expr = lo.ToPtr(clause.Expr{SQL: name, Vars: args}.Compat())
		if results := tableRegexp.FindStringSubmatch(name); len(results) == 3 {
			if results[1] != "" {
				table = results[1]
			} else {
				table = results[2]
			}
		}
	} else if tables := strings.Split(name, "."); len(tables) == 2 {
		if name != "DUAL" {
			name = quote(name)
		}
		expr = lo.ToPtr(clause.Expr{SQL: name}.Compat())
		table = tables[1]
	} else if name != "" {
		if name != "DUAL" {
			name = quote(name)
		}
		expr = lo.ToPtr(clause.Expr{SQL: name}.Compat())
		table = name
	}
	return
}

func addSelects(stmt interface {
	AddClause(v clause.Interface)
}, distinct bool, selects []field.IField) {
	if len(selects) == 0 {
		return
	}
	var m = make(map[string]struct{}, len(selects))
	for _, s := range selects {
		name := s.Name()
		if len(name) == 0 {
			continue
		}
		_, ok := m[name]
		if ok {
			panic(fmt.Sprintf("conflict select field name: `%s`, check your select fields", name))
		}
		m[name] = struct{}{}
	}

	stmt.AddClause(clause.Select{
		Distinct: distinct,
		Expression: columnClause{
			commaExpr: clause.CommaExpression{
				Exprs: lo.Map(selects, func(item field.IField, index int) clause.Expression {
					return item.ToExpr()
				}),
			},
			distinct: distinct,
		},
	})
}

type columnClause struct {
	commaExpr clause.CommaExpression
	distinct  bool
}

func (c columnClause) Build(builder clause.Builder) {
	if c.distinct {
		_, _ = builder.WriteString("DISTINCT ")
	}
	c.commaExpr.Build(builder)
}

type Table struct {
	Name string
}

type Table2 struct {
	Name string
}

func (t Table2) TableName() string {
	return t.Name
}

func (t Table) Ptr() *Table {
	return &t
}

func (t *Table) TableName() string {
	return t.Name
}

type order struct {
	field clause.Expression
	asc   bool
}

func tableNameFn(name string) tableNameImpl {
	return tableNameImpl{name: name}
}

type tableNameImpl struct {
	name string
}

func (t tableNameImpl) TableName() string {
	return t.name
}

type templateTable[T any, Model any] struct {
	Fields    T
	tableName string
	expr      clause.Expression
}

func (t templateTable[T, Model]) ModelType() *Model {
	var def Model
	return &def
}

func (t templateTable[T, Model]) TableName() string {
	return t.tableName
}

func (t templateTable[T, Model]) ToExpr() clause.Expression {
	return t.expr
}
