package field

import (
	"github.com/donutnomad/gsql/clause"
)

type pointerImpl struct {
	IField
}

func (f pointerImpl) IsNil() Expression {
	if f.IsExpr() {
		panic("[pointerImpl] cannot operate on expr")
	}
	return clause.Eq{
		Column: f.ToColumn(),
		Value:  nil,
	}
}

func (f pointerImpl) NotNil() Expression {
	if f.IsExpr() {
		panic("[pointerImpl] cannot operate on expr")
	}
	return clause.Neq{
		Column: f.ToColumn(),
		Value:  nil,
	}
}
