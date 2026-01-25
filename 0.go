package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/types"
)

//go:generate go run ./cmd/export-types -src ./internal/fields -dst ./typed_field.go -pkg gsql -exclude "AsJson"
//go:generate go run ./cmd/export-types -src ./internal/clauses/casewhen.go -dst ./clause_case.go -pkg gsql

type IFieldType[T any] = types.IFieldType[T]
type Expression = clause.Expression

// DbType 表示数据库类型
type DbType int

const (
	MySQL DbType = iota
	SQLite
	PostgresSQL
)

const (
	VALUES FunctionName = "VALUES"
)
