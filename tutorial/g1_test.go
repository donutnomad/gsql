package tutorial

import (
	"fmt"
	"testing"

	"github.com/donutnomad/gsql"
	"github.com/donutnomad/gsql/internal/utils"
)

func TestG1(t *testing.T) {
	var table1 = UserProfileSchema
	var table1Alias = table1.As("t")

	var table2 = SalesRecordSchema
	var table2Alias = table2.As("s")

	q1 := gsql.
		Select(
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("uid"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("company_name"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("company_id"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("table_company_id"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("banks"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("wallet_name"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("wallet_address"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("wallet_public_key"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("attribute"),
			gsql.IFF(table1Alias.ID.Eq(1), table1Alias.ID, table2Alias.ID).As("gldb_account"),
			table1Alias.ID,
		).
		From(table1Alias).
		Join(
			gsql.LeftJoin(table2Alias).On(table1Alias.ID.Eq(1)),
		).
		Where(
			gsql.NotExists(
				gsql.SelectOne().From(gsql.TN("body")).Where(),
			),
		)

	expr := gsql.UnionAll(q1, q1)

	type Result struct {
		Name string
	}

	gsql.DefineTable[Result]("t", Result{}, expr)

	b := utils.NewMemoryBuilder()
	expr.ToExpr().Build(b)

	fmt.Println(b.SQL.String())
}
