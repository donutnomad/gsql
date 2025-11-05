package gsql

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/donutnomad/gsql/clause"
	"gorm.io/driver/mysql"
)

var _ clause.Expression = (*JSONSetExpression)(nil)

// JSONSetExpression json set expression, implements clause.Expression interface to use as updater
type JSONSetExpression struct {
	column     string
	path2value map[string]any
	mutex      sync.RWMutex
}

// JSONSet update fields of json column
func JSONSet(column string) *JSONSetExpression {
	return &JSONSetExpression{
		column:     column,
		path2value: make(map[string]any),
	}
}

const jsonPrefix = "$."

// Set return clause.Expression.
//
//	{
//		"age": 20,
//		"name": "json-1",
//		"orgs": {"orga": "orgv"},
//		"tags": ["tag1", "tag2"]
//	}
//
//	// In MySQL/SQLite, path is `age`, `name`, `orgs.orga`, `tags[0]`, `tags[1]`.
//	DB.UpdateColumn("attr", JSONSet("attr").Set("orgs.orga", 42))
//
//	// In PostgreSQL, path is `{age}`, `{name}`, `{orgs,orga}`, `{tags, 0}`, `{tags, 1}`.
//	DB.UpdateColumn("attr", JSONSet("attr").Set("{orgs, orga}", "bar"))
func (jsonSet *JSONSetExpression) Set(path string, value interface{}) *JSONSetExpression {
	jsonSet.mutex.Lock()
	jsonSet.path2value[path] = value
	jsonSet.mutex.Unlock()
	return jsonSet
}

func (jsonSet *JSONSetExpression) Column() string {
	return jsonSet.column
}

func (jsonSet *JSONSetExpression) Len() int {
	return len(jsonSet.path2value)
}

// Build implements clause.Expression
// support mysql, sqlite and postgres
func (jsonSet *JSONSetExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql":

			var isMariaDB bool
			if v, ok := stmt.Dialector.(*mysql.Dialector); ok {
				isMariaDB = strings.Contains(v.ServerVersion, "MariaDB")
			}

			builder.WriteString("JSON_SET(")
			builder.WriteQuoted(jsonSet.column)
			for path, value := range jsonSet.path2value {
				builder.WriteByte(',')
				builder.AddVar(stmt, jsonPrefix+path)
				builder.WriteByte(',')

				if _, ok := value.(clause.Expression); ok {
					stmt.AddVar(builder, value)
					continue
				}

				rv := reflect.ValueOf(value)
				if rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}
				switch rv.Kind() {
				case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
					b, _ := json.Marshal(value)
					if isMariaDB {
						stmt.AddVar(builder, string(b))
						break
					}
					stmt.AddVar(builder, Expr("CAST(? AS JSON)", string(b)))
				case reflect.Bool:
					builder.WriteString(strconv.FormatBool(rv.Bool()))
				default:
					stmt.AddVar(builder, value)
				}
			}
			builder.WriteString(")")

		case "sqlite":
			builder.WriteString("JSON_SET(")
			builder.WriteQuoted(jsonSet.column)
			for path, value := range jsonSet.path2value {
				builder.WriteByte(',')
				builder.AddVar(stmt, jsonPrefix+path)
				builder.WriteByte(',')

				if _, ok := value.(clause.Expression); ok {
					stmt.AddVar(builder, value)
					continue
				}

				rv := reflect.ValueOf(value)
				if rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}
				switch rv.Kind() {
				case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
					b, _ := json.Marshal(value)
					stmt.AddVar(builder, Expr("JSON(?)", string(b)))
				default:
					stmt.AddVar(builder, value)
				}
			}
			builder.WriteString(")")

		case "postgres":
			var expr clause.Expression = columnExpression(jsonSet.column)
			for path, value := range jsonSet.path2value {
				if _, ok = value.(clause.Expression); ok {
					expr = Expr("JSONB_SET(?,?,?)", expr, path, value)
					continue
				} else {
					b, _ := json.Marshal(value)
					expr = Expr("JSONB_SET(?,?,?)", expr, path, string(b))
				}
			}
			stmt.AddVar(builder, expr)
		}
	}
}

type columnExpression string

func (col columnExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql", "sqlite", "postgres":
			builder.WriteString(stmt.Quote(string(col)))
		}
	}
}
