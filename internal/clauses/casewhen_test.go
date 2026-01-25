package clauses

import (
	"fmt"
	"testing"

	"github.com/donutnomad/gsql/internal/fields"
	"github.com/donutnomad/gsql/internal/utils"
)

func TestT(t *testing.T) {
	var a = fields.IntColumn[int64]("age").From(TN("test"))
	var str = Cases.String().
		When(a.Gt(20), fields.StringVal("C")).
		When(a.Gt(60), fields.StringVal("B")).
		When(a.Gt(90), fields.StringVal("A")).
		End()

	var score = Cases.Int().
		When(str.Eq("C"), fields.IntVal(30)).
		When(str.Eq("B"), fields.IntVal(70)).
		When(str.Eq("A"), fields.IntVal(100)).
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
