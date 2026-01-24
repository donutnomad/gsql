package fieldii

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
)

type PointerImpl = pointerImpl

type pointerImpl struct {
	fieldi.IField
}

func (f pointerImpl) IsNil() fieldi.Expression {
	if f.IsExpr() {
		panic("[pointerImpl] cannot operate on expr")
	}
	return clause.Eq{
		Column: f.ToColumn(),
		Value:  nil,
	}
}

func (f pointerImpl) NotNil() fieldi.Expression {
	if f.IsExpr() {
		panic("[pointerImpl] cannot operate on expr")
	}
	return clause.Neq{
		Column: f.ToColumn(),
		Value:  nil,
	}
}
