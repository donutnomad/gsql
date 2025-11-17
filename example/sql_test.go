package example

import (
	"testing"

	"github.com/donutnomad/gsql"
)

func TestName2(t *testing.T) {
	s := gsql.Select().
		From(gsql.TN("aaa")).
		Where(
			gsql.Eq(
				gsql.JSON_EXTRACT(gsql.Expr("content"), "$.Address"),
				gsql.JSON_ARRAY(gsql.Slice("0x1111", "0x1112")),
			),
		).
		ToSQL()
	t.Log(s)
}
