package types

import (
	"github.com/donutnomad/gsql/clause"
)

// OrderItem
// ORDER BY YEAR(order_date) ASC
// ORDER BY MONTH(order_date) DESC
// ORDER BY LOWER(product_name) ASC
// ORDER BY SUBSTRING_INDEX(full_name, ' ', 1) ASC
// ORDER BY ABS(difference) DESC
// ORDER BY
//
//	CASE status
//	    WHEN 'active' THEN 1
//	    WHEN 'pending' THEN 2
//	    WHEN 'closed' THEN 3
//	    ELSE 4
//	END ASC
type OrderItem struct {
	Expr clause.Expression
	Asc  bool
}

func NewOrder(expr clause.Expression, asc bool) OrderItem {
	return OrderItem{Expr: expr, Asc: asc}
}
