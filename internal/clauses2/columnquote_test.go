package clauses2

import (
	"fmt"
	"testing"

	"github.com/donutnomad/gsql/internal/utils"
)

func TestG(t *testing.T) {
	q := ColumnQuote{
		TableName:  "test",
		ColumnName: "apple",
		Alias:      "",
	}
	b := utils.NewMemoryBuilder()
	q.Build(b)
	// `test`.`apple`
	fmt.Println(b.SQL.String())
}

func TestG2(t *testing.T) {
	q := ColumnQuote{
		TableName:  "",
		ColumnName: "apple",
		Alias:      "",
	}
	b := utils.NewMemoryBuilder()
	q.Build(b)
	// `apple`
	fmt.Println(b.SQL.String())
}

func TestG3(t *testing.T) {
	q := ColumnQuote{
		TableName:  "",
		ColumnName: "apple",
		Alias:      "a",
	}
	b := utils.NewMemoryBuilder()
	q.Build(b)
	// `a`
	fmt.Println(b.SQL.String())
}
