package gsql

import (
	"time"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/utils"
	"gorm.io/gorm"
)

type ExprTo struct {
	clause.Expression
}

func (e ExprTo) AsF(name ...string) field.IField {
	return FieldExpr(e.Expression, utils.Optional(name, ""))
}

func (e ExprTo) ToExpr() field.Expression {
	return e.Expression
}

type primitive interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~string | time.Time | *time.Time
}

type DBResult struct {
	Error        error
	RowsAffected int64
}

type ICompactFrom interface {
	field.IToExpr
	TableName() string
}

type ITableName interface {
	TableName() string
}

type IDB interface {
	Session(config *gorm.Session) *gorm.DB
}

type IGormDB interface {
	Model(value any) (tx *gorm.DB)
	Session(config *gorm.Session) *gorm.DB
}

type GormDB = gorm.DB

var ErrRecordNotFound = gorm.ErrRecordNotFound
var ErrInvalidValue = gorm.ErrInvalidValue

type Statement = gorm.Statement
type Session = gorm.Session
type Config = gorm.Config
type StatementModifier = gorm.StatementModifier
