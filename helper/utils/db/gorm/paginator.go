package gorm

import (
	"strings"

	pgn "pulse/helper/utils/db/gorm/cursor-paginator/paginator"
)

func NewCursorPaginator(limit int, prev, next, order string, keys []string) *pgn.Paginator {
	opts := []pgn.Option{
		&pgn.Config{
			Limit: limit,
		},
	}
	if len(keys) > 0 {
		opts = append(opts, pgn.WithKeys(keys...))
	}
	if len(order) > 0 {
		var orderBy = pgn.DESC
		if strings.ToUpper(order) == string(pgn.ASC) {
			orderBy = pgn.ASC
		}
		opts = append(opts, pgn.WithOrder(orderBy))
	}
	if len(next) > 0 {
		opts = append(opts, pgn.WithAfter(next))
	}
	if len(prev) > 0 {
		opts = append(opts, pgn.WithBefore(prev))
	}
	return pgn.New(opts...)
}
