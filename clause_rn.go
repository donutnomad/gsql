package gsql

import (
	"github.com/donutnomad/gsql/clause"
	"github.com/donutnomad/gsql/field"
	"github.com/donutnomad/gsql/internal/utils"
)

var _ clause.Expression = (*WindowFunctionBuilder)(nil)

// RowNumber 创建 ROW_NUMBER() OVER() 窗口函数
// SELECT ROW_NUMBER() OVER(PARTITION BY category ORDER BY price DESC) as row_num FROM products;
// SELECT *, ROW_NUMBER() OVER(ORDER BY created_at DESC) as rn FROM orders;
// SELECT ROW_NUMBER() OVER(PARTITION BY user_id, status ORDER BY amount) FROM transactions;
func RowNumber() *WindowFunctionBuilder {
	return &WindowFunctionBuilder{
		function: "ROW_NUMBER()",
	}
}

// Rank 创建 RANK() OVER() 窗口函数，相同值会得到相同排名，下一个排名会跳过
// SELECT RANK() OVER(ORDER BY score DESC) as rank FROM students;
// SELECT RANK() OVER(PARTITION BY department ORDER BY salary DESC) FROM employees;
func Rank() *WindowFunctionBuilder {
	return &WindowFunctionBuilder{
		function: "RANK()",
	}
}

// DenseRank 创建 DENSE_RANK() OVER() 窗口函数，相同值得到相同排名，下一个排名连续
// SELECT DENSE_RANK() OVER(ORDER BY score DESC) as dense_rank FROM students;
// SELECT DENSE_RANK() OVER(PARTITION BY category ORDER BY price) FROM products;
func DenseRank() *WindowFunctionBuilder {
	return &WindowFunctionBuilder{
		function: "DENSE_RANK()",
	}
}

// WindowFunctionBuilder 窗口函数构建器
type WindowFunctionBuilder struct {
	function    string              // ROW_NUMBER(), RANK(), DENSE_RANK() 等
	partitionBy []clause.Expression // PARTITION BY 子句
	orderBy     []FieldOrder        // ORDER BY 子句
}

// PartitionBy 添加 PARTITION BY 子句，支持多个字段
// RowNumber().PartitionBy(category).OrderBy(price, true)
// RowNumber().PartitionBy(user_id, status).OrderBy(created_at, false)
func (w *WindowFunctionBuilder) PartitionBy(exprs ...clause.Expression) *WindowFunctionBuilder {
	w.partitionBy = append(w.partitionBy, exprs...)
	return w
}

// OrderBy 添加 ORDER BY 子句
// desc 为 true 表示降序(DESC)，false 表示升序(ASC)
// RowNumber().OrderBy(price, true) // ORDER BY price DESC
// RowNumber().OrderBy(created_at, false) // ORDER BY created_at ASC
func (w *WindowFunctionBuilder) OrderBy(order FieldOrder) *WindowFunctionBuilder {
	w.orderBy = append(w.orderBy, order)
	return w
}

func (w *WindowFunctionBuilder) Build(builder clause.Builder) {
	// 写入函数名
	builder.WriteString(w.function)
	builder.WriteString(" OVER(")

	// PARTITION BY 子句
	if len(w.partitionBy) > 0 {
		builder.WriteString("PARTITION BY ")
		for idx, expr := range w.partitionBy {
			if idx > 0 {
				builder.WriteString(", ")
			}
			expr.Build(builder)
		}
	}

	// ORDER BY 子句
	if len(w.orderBy) > 0 {
		if len(w.partitionBy) > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString("ORDER BY ")
		for idx, item := range w.orderBy {
			if idx > 0 {
				builder.WriteString(", ")
			}
			item.Expr.Build(builder)
			if !item.Asc {
				builder.WriteString(" DESC")
			} else {
				builder.WriteString(" ASC")
			}
		}
	}

	builder.WriteString(")")
}

func (w *WindowFunctionBuilder) ToExpr() clause.Expression {
	return w
}

// AsF 创建带别名的字段
func (w *WindowFunctionBuilder) AsF(name ...string) field.IField {
	return FieldExpr(w.ToExpr(), utils.Optional(name, ""))
}
