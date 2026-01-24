package gsql

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/types"
	"github.com/donutnomad/gsql/internal/utils"
	"github.com/samber/lo"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// INSERT INTO ... VALUES
// INSERT INTO ... ON DUPLICATE KEY UPDATE

// INSERT IGNORE ... VALUES
// * 如果插入的行没有导致唯一性约束（主键或唯一索引）冲突，则正常插入。
// * 如果插入的行导致唯一性约束冲突，则 忽略该行，不会插入，也不会返回任何错误。

// INSERT INTO ... SELECT

// REPLACE INTO ... SELECT
// * DELETE + INSERT, 先删除后插入，id会自增
// * 如果新数据某些列没有提供值，它们将使用默认值，旧数据中未提供的列的值会丢失
// * 会触发 DELETE 和 INSERT 相关的触发器
// * 通常性能较低，特别是索引多或有复杂触发器时

// Assignment 表示 ON DUPLICATE KEY UPDATE 中的赋值表达式
// 用于支持自定义更新逻辑，如 column = IF(RowValue(version) > version, RowValue(column), column)
type Assignment struct {
	Column field.IField
	Value  field.Expression
}

// Set 创建一个赋值表达式，用于 ON DUPLICATE KEY UPDATE
// 示例:
//
//	// 简单更新：使用插入行的值更新
//	gsql.Set(t.Count, gsql.RowValue(t.Count))
//
//	// 条件更新：只有当新版本号更大时才更新
//	gsql.Set(t.Value, gsql.IF(
//	    gsql.Expr("? > ?", gsql.RowValue(t.Version), t.Version),
//	    gsql.RowValue(t.Value),
//	    t.Value,
//	))
func Set(column field.IField, value field.Expression) Assignment {
	return Assignment{
		Column: column,
		Value:  value,
	}
}

type InsertBuilder[T any] struct {
	ignore        bool
	selectColumns []field.IField
}

type insertBuilderWithValues[T any] struct {
	selectColumns     []field.IField
	duplicateUpdates  []Assignment
	values            *[]T
	onDuplicateUpdate bool
	ignore            bool
}

func InsertInto[T any, TableTypes interface {
	ModelType() T
}](_ TableTypes, columns ...field.IField) *InsertBuilder[T] {
	return &InsertBuilder[T]{
		selectColumns: columns,
	}
}

func InsertIgnore[T any, TableTypes interface {
	ModelType() T
}](_ TableTypes, columns ...field.IField) *InsertBuilder[T] {
	return &InsertBuilder[T]{
		ignore:        true,
		selectColumns: columns,
	}
}

func (b *InsertBuilder[T]) Value(value T) *insertBuilderWithValues[T] {
	return &insertBuilderWithValues[T]{
		selectColumns: b.selectColumns,
		values:        &[]T{value},
		ignore:        b.ignore,
	}
}

func (b *InsertBuilder[T]) Values(values *[]T) *insertBuilderWithValues[T] {
	return &insertBuilderWithValues[T]{
		selectColumns: b.selectColumns,
		values:        values,
		ignore:        b.ignore,
	}
}

func (b *insertBuilderWithValues[T]) DuplicateUpdate(columns ...field.IField) *insertBuilderWithValues[T] {
	b.onDuplicateUpdate = true
	// 将列转换为 Assignment，使用 RowValue 引用插入行的值
	for _, col := range columns {
		b.duplicateUpdates = append(b.duplicateUpdates, Assignment{
			Column: col,
			Value:  RowValue(col),
		})
	}
	return b
}

// DuplicateUpdateExpr 设置 ON DUPLICATE KEY UPDATE 使用自定义表达式
// 支持带条件的更新，如:
//
//	InsertInto(table).
//	    Value(row).
//	    DuplicateUpdateExpr(
//	        gsql.Set(t.LastConsumedMessageId,
//	            gsql.IF(
//	                gsql.Expr("? >= ?", gsql.RowValue(t.GenerationId), t.GenerationId),
//	                gsql.RowValue(t.LastConsumedMessageId),
//	                t.LastConsumedMessageId,
//	            ),
//	        ),
//	        gsql.Set(t.GenerationId,
//	            gsql.IF(
//	                gsql.Expr("? >= ?", gsql.RowValue(t.GenerationId), t.GenerationId),
//	                gsql.RowValue(t.GenerationId),
//	                t.GenerationId,
//	            ),
//	        ),
//	    )
func (b *insertBuilderWithValues[T]) DuplicateUpdateExpr(assignments ...Assignment) *insertBuilderWithValues[T] {
	b.onDuplicateUpdate = true
	b.duplicateUpdates = append(b.duplicateUpdates, assignments...)
	return b
}

// ToSQL 生成 ON DUPLICATE KEY UPDATE 部分的 SQL 字符串（用于调试和测试）
func (b *insertBuilderWithValues[T]) ToSQL() string {
	if len(b.duplicateUpdates) == 0 {
		return ""
	}
	builder := utils.NewMemoryBuilder()
	onConflict := onConflictWithExprs{assignments: b.duplicateUpdates}
	onConflict.Build(builder)
	return builder.SQL.String()
}

func (b *insertBuilderWithValues[T]) Exec(db IGormDB) error {
	_, err := b.ExecWithResult(db)
	return err
}

func (b *insertBuilderWithValues[T]) ExecWithResult(db IGormDB) (int64, error) {
	var tx = db.Model(lo.Empty[T]())
	addSelects(tx.Statement, tx.Statement.Distinct, b.selectColumns)
	if b.ignore {
		tx = tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		})
	} else if b.onDuplicateUpdate {
		tx = tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		})
	}

	// 处理 ON DUPLICATE KEY UPDATE
	if len(b.duplicateUpdates) > 0 {
		// 使用 Recorder 捕获开始时间
		config := *tx.Config
		currentLogger, newLogger := config.Logger, logger.Recorder.New()
		config.Logger = newLogger
		tx.Config = &config

		tx.Statement.Dest = b.values
		tx.Statement.SQL.Reset()
		tx.Statement.Vars = nil

		tx = processerExec(createClauses, tx)
		if tx.Error != nil {
			return 0, tx.Error
		}
		stmt := tx.Statement

		stmt.SQL.Grow(180)
		stmt.AddClauseIfNotExists(clause.Insert{})

		// 添加 VALUES 子句，根据 MySQL 版本可能需要添加行别名
		valuesClause := callbacks.ConvertToCreateValues(stmt)
		if GetMySQLVersion() >= MySQLVersion8020 {
			stmt.AddClause(valuesWithAlias{valuesClause})
		} else {
			stmt.AddClause(valuesClause)
		}

		// 替换 OnConflict 子句，使用自定义表达式
		if v, ok := stmt.Clauses[clause.OnConflict{}.Name()]; ok {
			if _, ok := v.Expression.(clause.OnConflict); ok {
				customOnConflict := onConflictWithExprs{
					assignments: b.duplicateUpdates,
				}
				v.Name = "" // 清空名称，避免输出 "ON CONFLICT" 前缀
				v.Expression = customOnConflict
				stmt.Clauses[clause.OnConflict{}.Name()] = v
			}
		}

		stmt.Build(createClauses...)

		// 设置 newLogger.SQL，用于日志输出
		newLogger.SQL = tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...)

		// 直接执行已构建的 SQL
		result, err := tx.Statement.ConnPool.ExecContext(
			tx.Statement.Context,
			tx.Statement.SQL.String(),
			tx.Statement.Vars...,
		)
		rowsAffected := int64(0)
		if result != nil {
			rowsAffected, _ = result.RowsAffected()
		}

		// 输出 SQL 日志
		currentLogger.Trace(tx.Statement.Context, newLogger.BeginAt, func() (string, int64) {
			return newLogger.SQL, rowsAffected
		}, err)
		tx.Logger = currentLogger

		if err != nil {
			return 0, err
		}
		return rowsAffected, nil
	}

	ret := tx.Create(b.values)
	return ret.RowsAffected, ret.Error
}

// onConflictWithExprs 自定义的 OnConflict 表达式，支持复杂的更新逻辑
type onConflictWithExprs struct {
	assignments []Assignment
}

func (o onConflictWithExprs) Name() string {
	return "ON CONFLICT"
}

func (o onConflictWithExprs) Build(builder clause.Builder) {
	builder.WriteString("ON DUPLICATE KEY UPDATE ")
	for idx, assignment := range o.assignments {
		if idx > 0 {
			builder.WriteByte(',')
		}
		builder.WriteQuoted(assignment.Column.Name())
		builder.WriteByte('=')
		assignment.Value.Build(builder)
	}
}

func (o onConflictWithExprs) MergeClause(c *clause.Clause) {
	c.Name = "" // 清空名称，避免输出 "ON CONFLICT" 前缀
	c.Expression = o
}

// valuesWithAlias 包装 VALUES 子句，在 MySQL 8.0.20+ 模式下添加行别名
type valuesWithAlias struct {
	clause.Expression
}

func (v valuesWithAlias) Name() string {
	return "VALUES"
}

func (v valuesWithAlias) Build(builder clause.Builder) {
	v.Expression.Build(builder)
	// MySQL 8.0.20+ 需要添加行别名 AS _new
	builder.WriteString(" AS ")
	builder.WriteQuoted(insertRowAlias)
}

func (v valuesWithAlias) MergeClause(c *clause.Clause) {
	c.Name = "" // 清空名称，避免输出 "VALUES" 前缀
	c.Expression = v
}

func processerExec(pClauses []string, db *GormDB) *GormDB {
	// call scopes
	//for len(db.Statement.scopes) > 0 {
	//	db = db.executeScopes()
	//}

	var (
		curTime           = time.Now()
		stmt              = db.Statement
		resetBuildClauses bool
	)

	if len(stmt.BuildClauses) == 0 {
		stmt.BuildClauses = pClauses
		resetBuildClauses = true
	}

	if optimizer, ok := db.Statement.Dest.(StatementModifier); ok {
		optimizer.ModifyStatement(stmt)
	}

	// assign model values
	if stmt.Model == nil {
		stmt.Model = stmt.Dest
	} else if stmt.Dest == nil {
		stmt.Dest = stmt.Model
	}

	// parse model values
	if stmt.Model != nil {
		if err := stmt.Parse(stmt.Model); err != nil && (!errors.Is(err, schema.ErrUnsupportedDataType) || (stmt.Table == "" && stmt.TableExpr == nil && stmt.SQL.Len() == 0)) {
			if errors.Is(err, schema.ErrUnsupportedDataType) && stmt.Table == "" && stmt.TableExpr == nil {
				db.AddError(fmt.Errorf("%w: Table not set, please set it like: db.Model(&user) or db.Table(\"users\")", err))
			} else {
				db.AddError(err)
			}
		}
	}

	// assign stmt.ReflectValue
	if stmt.Dest != nil {
		stmt.ReflectValue = reflect.ValueOf(stmt.Dest)
		for stmt.ReflectValue.Kind() == reflect.Ptr {
			if stmt.ReflectValue.IsNil() && stmt.ReflectValue.CanAddr() {
				stmt.ReflectValue.Set(reflect.New(stmt.ReflectValue.Type().Elem()))
			}

			stmt.ReflectValue = stmt.ReflectValue.Elem()
		}
		if !stmt.ReflectValue.IsValid() {
			db.AddError(ErrInvalidValue)
		}
	}

	//for _, f := range p.fns {
	//	f(db)
	//}
	//
	//if stmt.SQL.Len() > 0 {
	//	db.Logger.Trace(stmt.Context, curTime, func() (string, int64) {
	//		sql, vars := stmt.SQL.String(), stmt.Vars
	//		if filter, ok := db.Logger.(ParamsFilter); ok {
	//			sql, vars = filter.ParamsFilter(stmt.Context, stmt.SQL.String(), stmt.Vars...)
	//		}
	//		return db.Dialector.Explain(sql, vars...), db.RowsAffected
	//	}, db.Error)
	//}
	//
	//if !stmt.DB.DryRun {
	//	stmt.SQL.Reset()
	//	stmt.Vars = nil
	//}
	//
	//if resetBuildClauses {
	//	stmt.BuildClauses = nil
	//}

	_ = curTime
	_ = resetBuildClauses
	return db
}

//////////////////////// select /////////////////////////////

func (b *InsertBuilder[T]) Select(q interface{ ToExpr() clause.Expr }) *insertBuilderWithSelect[T] {
	return &insertBuilderWithSelect[T]{
		selectColumns: b.selectColumns,
		ignore:        b.ignore,
		query:         q.ToExpr(),
	}
}

type insertBuilderWithSelect[T any] struct {
	ignore            bool
	onDuplicateUpdate bool
	duplicateUpdates  []Assignment
	query             clause.Expr
	selectColumns     []field.IField
}

func (b *insertBuilderWithSelect[T]) DuplicateUpdate(columns ...field.IField) *insertBuilderWithSelect[T] {
	b.onDuplicateUpdate = true
	// 将列转换为 Assignment，使用 RowValue 引用插入行的值
	for _, col := range columns {
		b.duplicateUpdates = append(b.duplicateUpdates, Assignment{
			Column: col,
			Value:  RowValue(col),
		})
	}
	return b
}

// DuplicateUpdateExpr 设置 ON DUPLICATE KEY UPDATE 使用自定义表达式
func (b *insertBuilderWithSelect[T]) DuplicateUpdateExpr(assignments ...Assignment) *insertBuilderWithSelect[T] {
	b.onDuplicateUpdate = true
	b.duplicateUpdates = append(b.duplicateUpdates, assignments...)
	return b
}

func (b *insertBuilderWithSelect[T]) Exec(db IGormDB) error {
	_, err := b.ExecWithResult(db)
	return err
}

func (b *insertBuilderWithSelect[T]) ExecWithResult(db IGormDB) (int64, error) {
	var tx = db.Model(lo.Empty[T]())
	if len(b.selectColumns) > 0 {
		tx = tx.Select(lo.Map(b.selectColumns, func(item field.IField, index int) string {
			return item.Name()
		}))
	}
	if b.ignore {
		tx = tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		})
	} else if b.onDuplicateUpdate {
		tx = tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		})
	}

	var def = []T{lo.Empty[T]()}

	tx.Statement.Dest = &def
	tx.Statement.SQL.Reset()
	tx.Statement.Vars = nil

	tx = processerExec(createClauses, tx)
	if tx.Error != nil {
		return 0, tx.Error
	}
	stmt := tx.Statement

	stmt.SQL.Grow(180)
	stmt.AddClauseIfNotExists(clause.Insert{})

	values := valuesWhere{Columns: make([]clause.Column, 0, len(stmt.Schema.DBNames))}
	selectColumns, restricted := stmt.SelectAndOmitColumns(true, false)
	for _, field := range stmt.Schema.FieldsWithDefaultDBValue {
		if v, ok := selectColumns[field.DBName]; (ok && v) || (!ok && !restricted) {
			values.Columns = append(values.Columns, clause.Column{Name: field.DBName})
		}
	}
	for _, dbName := range stmt.Schema.DBNames {
		if field := stmt.Schema.FieldsByDBName[dbName]; !field.HasDefaultValue || field.DefaultValueInterface != nil {
			if v, ok := selectColumns[dbName]; (ok && v) || (!ok && (!restricted || field.AutoCreateTime > 0 || field.AutoUpdateTime > 0)) {
				values.Columns = append(values.Columns, clause.Column{Name: dbName})
			}
		}
	}
	values.query = b.query
	stmt.AddClause(values)

	// 处理 ON DUPLICATE KEY UPDATE
	if len(b.duplicateUpdates) > 0 {
		if v, ok := stmt.Clauses[clause.OnConflict{}.Name()]; ok {
			if _, ok := v.Expression.(clause.OnConflict); ok {
				customOnConflict := onConflictWithExprs{
					assignments: b.duplicateUpdates,
				}
				v.Expression = customOnConflict
				stmt.Clauses[clause.OnConflict{}.Name()] = v
			}
		}
	}

	stmt.Build(createClauses...)

	ret := tx.Create(&def)
	return ret.RowsAffected, ret.Error
}

type valuesWhere struct {
	Columns []clause.Column
	query   clause.Expression
}

// Name from clause name
func (valuesWhere) Name() string {
	return "VALUES"
}

// Build from clause
func (values valuesWhere) Build(builder clause.Builder) {
	writer := &types.SafeWriter{builder}
	if len(values.Columns) > 0 {
		writer.WriteByte('(')
		for idx, column := range values.Columns {
			if idx > 0 {
				writer.WriteByte(',')
			}
			writer.WriteQuoted(column)
		}
		writer.WriteByte(')')
		writer.WriteByte(' ')
		values.query.Build(builder)
	} else {
		writer.WriteString("DEFAULT VALUES")
	}
}

// MergeClause merge values clauses
func (values valuesWhere) MergeClause(clause *clause.Clause) {
	clause.Name = ""
	clause.Expression = values
}
