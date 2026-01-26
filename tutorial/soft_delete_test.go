package tutorial

import (
	"testing"
	"time"

	gsql "github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/fields"
	"gorm.io/gorm"
)

// SoftDeleteModel 带软删除的测试模型
type SoftDeleteModel struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string         `gorm:"column:name;size:100;not null"`
	Status    string         `gorm:"column:status;size:20;default:'active'"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"` // 软删除字段
}

func (SoftDeleteModel) TableName() string { return "soft_delete_models" }

// ==================== SoftDeleteModel Schema ====================

type SoftDeleteModelSchemaType struct {
	ID        fields.IntField[uint64]
	Name      fields.StringField[string]
	Status    fields.StringField[string]
	CreatedAt fields.IntField[time.Time]
	UpdatedAt fields.IntField[time.Time]
	DeletedAt fields.IntField[gorm.DeletedAt]
	fieldType SoftDeleteModel
	alias     string
	tableName string
}

func (t SoftDeleteModelSchemaType) TableName() string {
	return t.tableName
}

func (t SoftDeleteModelSchemaType) Alias() string {
	return t.alias
}

func (t *SoftDeleteModelSchemaType) WithTable(tableName string) {
	tn := gsql.TN(tableName)
	t.ID = t.ID.WithTable(&tn)
	t.Name = t.Name.WithTable(&tn)
	t.Status = t.Status.WithTable(&tn)
	t.CreatedAt = t.CreatedAt.WithTable(&tn)
	t.UpdatedAt = t.UpdatedAt.WithTable(&tn)
	t.DeletedAt = t.DeletedAt.WithTable(&tn)
}

func (t SoftDeleteModelSchemaType) As(alias string) SoftDeleteModelSchemaType {
	var ret = t
	ret.alias = alias
	ret.WithTable(alias)
	return ret
}

func (t SoftDeleteModelSchemaType) ModelType() *SoftDeleteModel {
	return &t.fieldType
}

func (t SoftDeleteModelSchemaType) ModelTypeAny() any {
	return &t.fieldType
}

func (t SoftDeleteModelSchemaType) AllFields() field.BaseFields {
	return field.BaseFields{
		t.ID,
		t.Name,
		t.Status,
		t.CreatedAt,
		t.UpdatedAt,
		t.DeletedAt,
	}
}

func (t SoftDeleteModelSchemaType) Star() field.IField {
	if t.alias != "" {
		return gsql.StarWith(t.alias)
	}
	return gsql.StarWith(t.tableName)
}

var SoftDeleteModelSchema = SoftDeleteModelSchemaType{
	tableName: "soft_delete_models",
	ID:        fields.IntFieldOf[uint64]("soft_delete_models", "id", field.FlagPrimaryKey),
	Name:      fields.StringFieldOf[string]("soft_delete_models", "name"),
	Status:    fields.StringFieldOf[string]("soft_delete_models", "status"),
	CreatedAt: field.NewComparable[time.Time]("soft_delete_models", "created_at"),
	UpdatedAt: field.NewComparable[time.Time]("soft_delete_models", "updated_at"),
	DeletedAt: field.NewComparable[gorm.DeletedAt]("soft_delete_models", "deleted_at"),
	fieldType: SoftDeleteModel{},
}

// TestSoftDelete_CountIssue 测试 Count 方法没有正确处理软删除的问题
func TestSoftDelete_CountIssue(t *testing.T) {
	var model SoftDeleteModel
	setupTable(t, &model)
	db := getDB()

	// 插入测试数据
	records := []SoftDeleteModel{
		{Name: "Active1", Status: "active"},
		{Name: "Active2", Status: "active"},
		{Name: "ToBeDeleted", Status: "active"},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// 软删除一条记录
	if err := db.Delete(&records[2]).Error; err != nil {
		t.Fatalf("Failed to soft delete record: %v", err)
	}

	// 验证数据库中实际有 3 条记录（包括软删除的）
	var totalCount int64
	if err := db.Unscoped().Model(&SoftDeleteModel{}).Count(&totalCount).Error; err != nil {
		t.Fatalf("Failed to count unscoped: %v", err)
	}
	if totalCount != 3 {
		t.Errorf("Expected 3 total records (including deleted), got %d", totalCount)
	}

	// 使用 GORM 原生的 Count（应该返回 2，因为有软删除过滤）
	var gormCount int64
	if err := db.Model(&SoftDeleteModel{}).Count(&gormCount).Error; err != nil {
		t.Fatalf("Failed to count with GORM: %v", err)
	}
	t.Logf("GORM native Count: %d (expected: 2)", gormCount)
	if gormCount != 2 {
		t.Errorf("GORM native Count should return 2 (excluding soft deleted), got %d", gormCount)
	}

	// 使用 gsql 的 Count（这是我们要测试的问题）
	s := SoftDeleteModelSchema

	gsqlCount, err := gsql.Select(s.AllFields()...).
		From(&s).
		Count(db)

	if err != nil {
		t.Fatalf("Failed to count with gsql: %v", err)
	}
	t.Logf("gsql Count: %d (expected: 2)", gsqlCount)

	// 这是问题所在：gsql 的 Count 返回 3（没有过滤软删除的记录）
	// 修复后应该返回 2
	if gsqlCount != 2 {
		t.Errorf("gsql Count should return 2 (excluding soft deleted), got %d. This confirms the soft delete issue in Count method.", gsqlCount)
	}
}

// TestSoftDelete_FindWorks 验证 Find 方法正确处理了软删除
func TestSoftDelete_FindWorks(t *testing.T) {
	var model SoftDeleteModel
	setupTable(t, &model)
	db := getDB()

	// 插入测试数据
	records := []SoftDeleteModel{
		{Name: "Active1", Status: "active"},
		{Name: "Active2", Status: "active"},
		{Name: "ToBeDeleted", Status: "active"},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// 软删除一条记录
	if err := db.Delete(&records[2]).Error; err != nil {
		t.Fatalf("Failed to soft delete record: %v", err)
	}

	// 使用 gsql 的 Find（应该正确过滤软删除的记录）
	s := SoftDeleteModelSchema

	var results []SoftDeleteModel
	err := gsql.Select(s.AllFields()...).
		From(&s).
		Find(db, &results)

	if err != nil {
		t.Fatalf("Failed to find with gsql: %v", err)
	}
	t.Logf("gsql Find returned %d records (expected: 2)", len(results))

	if len(results) != 2 {
		t.Errorf("gsql Find should return 2 records (excluding soft deleted), got %d", len(results))
	}
}

// TestSoftDelete_UnscopedCount 测试 Unscoped 可以绕过软删除过滤
func TestSoftDelete_UnscopedCount(t *testing.T) {
	var model SoftDeleteModel
	setupTable(t, &model)
	db := getDB()

	// 插入测试数据
	records := []SoftDeleteModel{
		{Name: "Active1", Status: "active"},
		{Name: "ToBeDeleted", Status: "active"},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// 软删除一条记录
	if err := db.Delete(&records[1]).Error; err != nil {
		t.Fatalf("Failed to soft delete record: %v", err)
	}

	s := SoftDeleteModelSchema

	// 使用 Unscoped 应该返回所有记录（包括软删除的）
	unscopedCount, err := gsql.Select(s.AllFields()...).
		From(&s).
		Unscoped().
		Count(db)

	if err != nil {
		t.Fatalf("Failed to count unscoped with gsql: %v", err)
	}
	t.Logf("gsql Unscoped Count: %d (expected: 2)", unscopedCount)

	if unscopedCount != 2 {
		t.Errorf("gsql Unscoped Count should return 2 (including soft deleted), got %d", unscopedCount)
	}
}

// TestSoftDelete_ExistIssue 测试 Exist 方法（基于 Count）没有正确处理软删除的问题
func TestSoftDelete_ExistIssue(t *testing.T) {
	var model SoftDeleteModel
	setupTable(t, &model)
	db := getDB()

	// 插入测试数据
	records := []SoftDeleteModel{
		{Name: "OnlyRecord", Status: "active"},
	}
	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create records: %v", err)
	}

	// 软删除唯一的记录
	if err := db.Delete(&records[0]).Error; err != nil {
		t.Fatalf("Failed to soft delete record: %v", err)
	}

	s := SoftDeleteModelSchema

	// 使用 gsql 的 Exist（基于 Count，应该返回 false）
	exists, err := gsql.Select(s.ID).
		From(&s).
		Exist(db)

	if err != nil {
		t.Fatalf("Failed to check exist with gsql: %v", err)
	}
	t.Logf("gsql Exist: %v (expected: false)", exists)

	// 修复后应该返回 false（因为唯一的记录已经被软删除了）
	if exists {
		t.Errorf("gsql Exist should return false (all records soft deleted), got true. This confirms the soft delete issue in Exist method (via Count).")
	}
}
