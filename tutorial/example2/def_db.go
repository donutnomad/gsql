package example2

import (
	"database/sql"

	"gorm.io/gorm"
)

type DefaultGormDB struct {
	getDB func() *gorm.DB
}

func (d *DefaultGormDB) DB() (*sql.DB, error) {
	return d.getDB().DB()
}

func NewDefaultGormDB(db *gorm.DB) *DefaultGormDB {
	return &DefaultGormDB{getDB: func() *gorm.DB {
		return db
	}}
}

func (d *DefaultGormDB) Model(value any) (tx *gorm.DB) {
	return d.getDB().Model(value)
}
func (d *DefaultGormDB) Session(config *gorm.Session) *gorm.DB {
	return d.getDB().Session(config)
}
