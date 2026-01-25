package gsql

//go:generate go run ./cmd/export-types -src ./internal/fields -dst ./typed_field.go -pkg gsql -exclude "AsJson"
//go:generate go run ./cmd/export-types -src ./internal/clauses/casewhen.go -dst ./clause_case.go -pkg gsql

// DbType 表示数据库类型
type DbType int

const (
	MySQL DbType = iota
	SQLite
	PostgresSQL
)

const (
	FUNC_VALUES FunctionName = "VALUES"
)
