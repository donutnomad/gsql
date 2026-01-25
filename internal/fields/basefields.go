package fields

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/internal/fieldi"
	"github.com/samber/lo"
)

var _ fieldi.IField = (BaseFields)(nil)

type BaseFields []fieldi.IField

// SetAliasWith 设置别名
func (b BaseFields) SetAliasWith(fn func(fieldName string) string) BaseFields {
	var ret BaseFields
	for _, base := range b {
		val := reflect.ValueOf(base)
		withAlias, ok := val.Type().MethodByName("WithAlias")
		if !ok {
			panic(fmt.Sprintf("[BaseFields.WithPrefix] no WithAlias method found for %T", base))
		}
		oldFieldName := base.Name()
		newFieldName := fn(oldFieldName)
		rets := withAlias.Func.Call([]reflect.Value{val, reflect.ValueOf(newFieldName)})
		ret = append(ret, rets[0].Interface().(fieldi.IField))
	}
	return ret
}

func (b BaseFields) As(prefix string) fieldi.IField {
	return b.AsWith(prefix)
}

func (b BaseFields) AsWith(prefix string, suffix ...string) BaseFields {
	return b.SetAliasWith(func(fieldName string) string {
		if len(suffix) > 0 {
			return prefix + fieldName + suffix[0]
		}
		return prefix + fieldName
	})
}

func (b BaseFields) Exclude(f ...fieldi.IField) BaseFields {
	var names = lo.Map(f, func(item fieldi.IField, index int) string {
		return item.Name()
	})
	var ret BaseFields
	for _, base := range b {
		if !slices.Contains(names, base.Name()) {
			ret = append(ret, base)
		}
	}
	return ret
}

func (b BaseFields) Include(f ...fieldi.IField) BaseFields {
	var names = lo.Map(f, func(item fieldi.IField, index int) string {
		return item.Name()
	})
	var ret BaseFields
	for _, base := range b {
		if slices.Contains(names, base.Name()) {
			ret = append(ret, base)
		}
	}
	return ret
}

func (b BaseFields) Build(builder clause.Builder) {
	for _, item := range b {
		item.Build(builder)
	}
}

func (b BaseFields) ToExpr() clause.Expression {
	panic("BaseFields cannot ToExpr")
}

func (b BaseFields) Name() string {
	panic("BaseFields cannot get Name")
}

func (b BaseFields) FullName() string {
	panic("BaseFields cannot get FullName")
}

func (b BaseFields) Alias() string {
	panic("BaseFields cannot get Alias")
}
