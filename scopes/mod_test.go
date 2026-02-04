package scopes

import (
	"testing"
	"time"

	"github.com/donutnomad/gsql"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type UserWalletLogSchemaType struct {
	tableName  string
	alias      string
	UserID     gsql.IntField[uint64]
	BusinessID gsql.IntField[uint64]
	Address    gsql.IntField[string]
	CreatedAt  gsql.IntField[time.Time]
	UpdatedAt  gsql.IntField[time.Time]
	Bind       gsql.IntField[bool]
	UnbindAt   gsql.IntField[time.Time]
}

func (t UserWalletLogSchemaType) TableName() string {
	return t.tableName
}

func (t UserWalletLogSchemaType) Alias() string {
	return t.alias
}

var UserWalletLogSchema = UserWalletLogSchemaType{
	tableName:  "client_wallet_log",
	UserID:     gsql.IntFieldOf[uint64]("client_wallet_log", "client_id"), // << client_id
	BusinessID: gsql.IntFieldOf[uint64]("client_wallet_log", "business_id"),
	Address:    gsql.IntFieldOf[string]("client_wallet_log", "address"),
	CreatedAt:  gsql.IntFieldOf[time.Time]("client_wallet_log", "created_at"),
	UpdatedAt:  gsql.IntFieldOf[time.Time]("client_wallet_log", "updated_at"),
	Bind:       gsql.IntFieldOf[bool]("client_wallet_log", "bind"),
	UnbindAt:   gsql.IntFieldOf[time.Time]("client_wallet_log", "unbind_at"),
}

func TestMod(t *testing.T) {
	M := UserWalletLogSchema
	ordersMapping := SortNameMapping{
		"create": M.CreatedAt,
	}
	orders := []SortOrder{
		OrderBy("create", true),
		OrderBy(M.CreatedAt.Name(), true),
	}
	sql := gsql.SelectG[any]().
		From(M).
		Where(M.UserID.Between(lo.ToPtr[uint64](123), nil)).
		OrderBy(ordersMapping.Map(orders)...).
		Scope(
			TimeBetween(M.CreatedAt, TimestampRange{
				//From: mo.Some(int64(123)),
				To: mo.Some(int64(222)),
			}, ">", "<="),
		).ToSQL()
	t.Log(sql)
}

func TestMod2(t *testing.T) {
	sql := gsql.SelectG[any]().
		From(UserWalletLogSchema).
		Scope(
			TimeBetween(UserWalletLogSchema.CreatedAt, TimeRange{
				From: mo.Some(time.Now()),
				//To:    mo.Some(int64(222)),
			}),
		).ToSQL()
	t.Log(sql)
}
