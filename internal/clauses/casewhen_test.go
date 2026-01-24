package clauses

import (
	"fmt"
	"testing"

	"github.com/donutnomad/gsql/internal/fields"
	"github.com/donutnomad/gsql/internal/utils"
)

func TestT(t *testing.T) {
	var a = fields.IntColumn("age").From(TN("test"))
	var str = Case[string, fields.StringExpr[string]]().
		When(a.Gt(20), fields.StringV("C")).
		When(a.Gt(60), fields.StringV("B")).
		When(a.Gt(90), fields.StringV("A")).
		End().
		ToString()

	var score = Case[int, fields.IntExpr[int]]().
		When(str.Eq("C"), fields.IntV(30)).
		When(str.Eq("B"), fields.IntV(70)).
		When(str.Eq("A"), fields.IntV(100)).
		End()

	b := utils.NewMemoryBuilder()
	score.Build(b)
	fmt.Println(b.SQL.String())
}

func TN(tableName string) Table2 {
	return Table2{Name: tableName}
}

type Table2 struct {
	Name string
}

func (t Table2) TableName() string {
	return t.Name
}
