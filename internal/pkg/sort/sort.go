package sort

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
)

const (
	Asc  = "ASC"
	Desc = "DESC"
)

type Sort struct {
	column string
	order  string
}

func New(column, order string) *Sort {
	order = strings.ToUpper(order)
	if order != Asc && order != Desc {
		order = Asc
	}

	return &Sort{
		column: column,
		order:  order,
	}
}

func (opt *Sort) UseSelectBuilder(builder sq.SelectBuilder) sq.SelectBuilder {
	return builder.OrderBy(opt.column + " " + opt.order)
}
